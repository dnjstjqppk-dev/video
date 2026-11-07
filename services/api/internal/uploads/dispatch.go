package uploads

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"globalvideoplatform/services/api/internal/store"
)

// mediaWorkerBaseURL resolves the transcode worker intake base URL.
// Compose sets MEDIA_WORKER_URL to the worker service; local dev defaults
// to localhost. An empty value disables dispatch (jobs stay QUEUED).
func mediaWorkerBaseURL() string {
	if v := os.Getenv("MEDIA_WORKER_URL"); v != "" {
		return v
	}
	return "http://localhost:8081"
}

type transcodeDispatch struct {
	JobID       string `json:"jobId"`
	VideoID     string `json:"videoId"`
	SourceR2Key string `json:"sourceR2Key"`
	Attempt     int    `json:"attempt"`
	MaxAttempts int    `json:"maxAttempts"`
}

type transcodeDispatchResult struct {
	State  string `json:"state"`
	Probed bool   `json:"probed"`
	Error  string `json:"error"`
}

// dispatchTranscodeJob POSTs a completed-upload job to the media worker
// intake and folds the resulting state back into the in-memory job and
// video records. It is best-effort: any transport or worker failure only
// logs, leaving the job QUEUED for a later retry instead of failing the
// already-acknowledged upload completion.
func (h *Handler) dispatchTranscodeJob(job *store.MediaJob, video *store.Video, sourceR2Key string) {
	base := mediaWorkerBaseURL()
	if base == "" || base == "disabled" {
		return
	}
	body, err := json.Marshal(transcodeDispatch{
		JobID:       job.ID,
		VideoID:     video.ID,
		SourceR2Key: sourceR2Key,
		MaxAttempts: 3,
	})
	if err != nil {
		slog.Error("Transcode dispatch encode failed", "jobId", job.ID, "error", err)
		return
	}
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Post(base+"/jobs/transcode", "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Error("Transcode dispatch unreachable", "jobId", job.ID, "worker", base, "error", err)
		return
	}
	defer resp.Body.Close()
	var result transcodeDispatchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Transcode dispatch bad response", "jobId", job.ID, "status", resp.StatusCode, "error", err)
		return
	}
	now := time.Now()
	switch {
	case resp.StatusCode == http.StatusOK && result.State == "READY":
		job.Status = "READY"
		job.ProgressPct = 100
		job.CompletedAt = &now
		video.Status = "READY"
		video.Visibility = "public"
		slog.Info("Transcode job completed", "jobId", job.ID, "videoId", video.ID, "probed", result.Probed)
	case resp.StatusCode == http.StatusUnprocessableEntity || result.State == "FAILED":
		job.Status = "FAILED"
		job.Error = result.Error
		if job.Error == "" {
			job.Error = "media worker rejected transcode job"
		}
		video.Status = "FAILED"
		slog.Error("Transcode job failed", "jobId", job.ID, "videoId", video.ID, "error", job.Error)
	default:
		slog.Error("Transcode dispatch unexpected state", "jobId", job.ID, "status", resp.StatusCode, "state", result.State)
	}
}
