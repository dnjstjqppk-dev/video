package studio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

// editWorkerBaseURL resolves the media worker edit-render intake base URL.
func editWorkerBaseURL() string {
	if v := os.Getenv("MEDIA_WORKER_URL"); v != "" {
		return v
	}
	return "http://localhost:8081"
}

// validateEDL mirrors the media worker's structural edit rules so the editor
// gets immediate 400s instead of a FAILED job round-trip.
func validateEDL(edl store.EditDecisionList, durationSec int) error {
	if edl.TrimStartSec < 0 || edl.TrimEndSec <= edl.TrimStartSec {
		return fmt.Errorf("invalid trim range: trimEnd must exceed trimStart from 0")
	}
	if durationSec > 0 && edl.TrimEndSec > durationSec {
		return fmt.Errorf("trimEnd %ds exceeds video duration %ds", edl.TrimEndSec, durationSec)
	}
	prev := edl.TrimStartSec
	for _, sp := range edl.SplitPoints {
		if sp <= prev || sp >= edl.TrimEndSec {
			return fmt.Errorf("split points must be ascending inside the trim range, got %d", sp)
		}
		prev = sp
	}
	for _, b := range edl.BlurRegions {
		if b.ID == "" {
			return fmt.Errorf("blur region id is required")
		}
		if b.X < 0 || b.X > 100 || b.Y < 0 || b.Y > 100 {
			return fmt.Errorf("blur region %q coordinates must be 0-100 percent", b.ID)
		}
		if b.StartSec < edl.TrimStartSec || b.EndSec > edl.TrimEndSec || b.EndSec <= b.StartSec {
			return fmt.Errorf("blur region %q must sit inside the trim range", b.ID)
		}
	}
	for _, m := range edl.MutedSegments {
		if m.StartSec < edl.TrimStartSec || m.EndSec > edl.TrimEndSec || m.EndSec <= m.StartSec {
			return fmt.Errorf("muted segments must sit inside the trim range")
		}
	}
	return nil
}

func (h *Handler) ownedVideo(w http.ResponseWriter, r *http.Request) (*store.Video, *store.Channel, bool) {
	user := identity.GetUserFromContext(r.Context())
	videoID := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[videoID]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return nil, nil, false
	}
	channel := h.Store.ChannelForUser(user.ID)
	if channel == nil || video.ChannelID != channel.ID {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot edit a video from another channel")
		return nil, nil, false
	}
	return video, channel, true
}

// CreateEditJob validates a studio EDL, queues a render job and dispatches it
// to the media worker. The source video record is never mutated: the render
// produces a separate edited rendition set (non-destructive).
func (h *Handler) CreateEditJob(w http.ResponseWriter, r *http.Request) {
	video, channel, ok := h.ownedVideo(w, r)
	if !ok {
		return
	}
	var edl store.EditDecisionList
	if err := json.NewDecoder(r.Body).Decode(&edl); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid edit decision list")
		return
	}
	if err := validateEDL(edl, video.DurationSeconds); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_EDL", err.Error())
		return
	}

	now := time.Now()
	job := &store.EditJob{
		ID:          "edit-job-" + uuid.NewString(),
		VideoID:     video.ID,
		ChannelID:   channel.ID,
		EDL:         edl,
		Status:      "QUEUED",
		ProgressPct: 0,
		CreatedAt:   now,
	}
	h.Store.EditJobs[job.ID] = job

	go h.dispatchEditRender(job, video, channel)

	identity.WriteJSON(w, http.StatusAccepted, editJobResponse(job))
}

// GetEditJob returns one render job's live status for editor polling.
func (h *Handler) GetEditJob(w http.ResponseWriter, r *http.Request) {
	video, _, ok := h.ownedVideo(w, r)
	if !ok {
		return
	}
	jobID := chi.URLParam(r, "jobId")
	job, ok := h.Store.EditJobs[jobID]
	if !ok || job.VideoID != video.ID {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Edit job not found")
		return
	}
	identity.WriteJSON(w, http.StatusOK, editJobResponse(job))
}

func editJobResponse(job *store.EditJob) map[string]interface{} {
	return map[string]interface{}{
		"jobId":        job.ID,
		"videoId":      job.VideoID,
		"status":       job.Status,
		"progressPct":  job.ProgressPct,
		"outputPrefix": "r2://edited-hls/" + job.VideoID + "/",
		"error":        job.Error,
		"createdAt":    job.CreatedAt.Format(time.RFC3339),
	}
}

type editRenderDispatch struct {
	JobID           string                 `json:"jobId"`
	VideoID         string                 `json:"videoId"`
	SourceR2Key     string                 `json:"sourceR2Key"`
	Height          int                    `json:"height"`
	DurationSeconds float64                `json:"durationSeconds"`
	MaxAttempts     int                    `json:"maxAttempts"`
	EDL             store.EditDecisionList `json:"edl"`
}

type editRenderResult struct {
	State string `json:"state"`
	Error string `json:"error"`
}

// dispatchEditRender POSTs the EDL to the media worker and folds the result
// back into the job record. Best-effort: transport failures leave the job
// QUEUED for retry instead of failing the accepted dispatch.
func (h *Handler) dispatchEditRender(job *store.EditJob, video *store.Video, channel *store.Channel) {
	base := editWorkerBaseURL()
	if base == "" || base == "disabled" {
		return
	}
	body, err := json.Marshal(editRenderDispatch{
		JobID:           job.ID,
		VideoID:         video.ID,
		SourceR2Key:     fmt.Sprintf("raw/%s/%s.mp4", channel.OwnerID, video.ID),
		Height:          1080,
		DurationSeconds: float64(video.DurationSeconds),
		MaxAttempts:     3,
		EDL:             job.EDL,
	})
	if err != nil {
		slog.Error("Edit render dispatch encode failed", "jobId", job.ID, "error", err)
		return
	}
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Post(base+"/jobs/edit-render", "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Error("Edit render dispatch unreachable", "jobId", job.ID, "worker", base, "error", err)
		return
	}
	defer resp.Body.Close()
	var result editRenderResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Edit render dispatch bad response", "jobId", job.ID, "status", resp.StatusCode, "error", err)
		return
	}
	now := time.Now()
	switch {
	case resp.StatusCode == http.StatusOK && result.State == "READY":
		job.Status = "READY"
		job.ProgressPct = 100
		job.CompletedAt = &now
		slog.Info("Edit render job completed", "jobId", job.ID, "videoId", video.ID)
	case resp.StatusCode == http.StatusUnprocessableEntity || result.State == "FAILED":
		job.Status = "FAILED"
		job.Error = result.Error
		if job.Error == "" {
			job.Error = "media worker rejected edit render job"
		}
		slog.Error("Edit render job failed", "jobId", job.ID, "videoId", video.ID, "error", job.Error)
	default:
		slog.Error("Edit render dispatch unexpected state", "jobId", job.ID, "status", resp.StatusCode, "state", result.State)
	}
}
