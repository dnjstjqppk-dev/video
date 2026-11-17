package uploads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
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

// transcoderTimeout bounds one transcode intake call (P0-2 real encodes run
// minutes; the default covers typical VODs and stays overridable).
func transcoderTimeout() time.Duration {
	if v := os.Getenv("TRANSCODER_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 30 * time.Minute
}

type transcodeDispatch struct {
	JobID       string `json:"jobId"`
	VideoID     string `json:"videoId"`
	SourceR2Key string `json:"sourceR2Key"`
	SourceURL   string `json:"sourceUrl,omitempty"`
	Attempt     int    `json:"attempt"`
	MaxAttempts int    `json:"maxAttempts"`
}

type transcodeArtifactRef struct {
	Key         string `json:"key"`
	ContentType string `json:"contentType"`
	Bytes       int64  `json:"bytes"`
}

type transcodeProbe struct {
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	DurationSeconds float64 `json:"durationSeconds"`
	VideoCodec      string  `json:"videoCodec"`
}

type transcodeDispatchResult struct {
	State     string                 `json:"state"`
	Probed    bool                   `json:"probed"`
	Error     string                 `json:"error"`
	Probe     transcodeProbe         `json:"probe"`
	Artifacts []transcodeArtifactRef `json:"artifacts"`
}

// publicHLSBase resolves the player-facing base URL for published HLS.
func publicHLSBase(s3cfg S3Config) string {
	if v := strings.TrimSpace(os.Getenv("R2_PUBLIC_HLS_BASE")); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	public := strings.TrimSuffix(s3cfg.PublicEndpoint, "/")
	if public == "" {
		public = strings.TrimSuffix(s3cfg.Endpoint, "/")
	}
	return public + "/" + s3cfg.HLSBucket
}

// dispatchTranscodeJob POSTs a completed-upload job to the media worker
// intake and folds the resulting state back into the in-memory job and
// video records.
//
// P0-2: the worker really encodes; on READY the dispatcher fetches staged
// artifacts from the worker and publishes them to object storage before the
// video may turn READY. Transport failures stay QUEUED for a later retry;
// worker-reported failures and unpublishable results fail the job with an
// explicit error instead of a fake READY.
func (h *Handler) dispatchTranscodeJob(job *store.MediaJob, video *store.Video, session *store.UploadSession) {
	base := mediaWorkerBaseURL()
	if base == "" || base == "disabled" {
		return
	}
	sourceURL := ""
	if session.Mode == "local" && session.StoredPath != "" {
		sourceURL = "file://" + session.StoredPath
	} else if session.Mode == "s3" {
		s3cfg := S3ConfigFromEnv()
		presigned, err := s3cfg.PresignGet(session.R2Key, 2*time.Hour)
		if err != nil {
			h.failJob(job, video, fmt.Sprintf("could not presign worker source fetch: %v", err))
			return
		}
		sourceURL = presigned
	}
	body, err := json.Marshal(transcodeDispatch{
		JobID:       job.ID,
		VideoID:     video.ID,
		SourceR2Key: session.R2Key,
		SourceURL:   sourceURL,
		MaxAttempts: 3,
	})
	if err != nil {
		slog.Error("Transcode dispatch encode failed", "jobId", job.ID, "error", err)
		return
	}
	client := &http.Client{Timeout: transcoderTimeout()}
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
	switch {
	case resp.StatusCode == http.StatusOK && result.State == "READY":
		h.publishReady(job, video, base, &result)
	case resp.StatusCode == http.StatusUnprocessableEntity || result.State == "FAILED":
		msg := result.Error
		if msg == "" {
			msg = "media worker rejected transcode job"
		}
		h.failJob(job, video, msg)
	default:
		slog.Error("Transcode dispatch unexpected state", "jobId", job.ID, "status", resp.StatusCode, "state", result.State)
	}
}

func (h *Handler) failJob(job *store.MediaJob, video *store.Video, msg string) {
	job.Status = "FAILED"
	job.Error = msg
	video.Status = "FAILED"
	h.Store.PersistVideo(video)
	slog.Error("Transcode job failed", "jobId", job.ID, "videoId", video.ID, "error", msg)
}

// publishReady fetches worker-staged artifacts, uploads them to the public
// HLS bucket, and completes the video with real probe metadata and URLs.
// Visibility is left untouched: READY never flips a private draft public.
func (h *Handler) publishReady(job *store.MediaJob, video *store.Video, workerBase string, result *transcodeDispatchResult) {
	s3cfg := S3ConfigFromEnv()
	if !s3cfg.Configured() {
		h.failJob(job, video, "artifact storage unavailable: set R2_ACCESS_KEY_ID/R2_SECRET_ACCESS_KEY (or run MinIO) so encodes can publish")
		return
	}
	if len(result.Artifacts) == 0 {
		h.failJob(job, video, "worker reported READY without artifacts")
		return
	}
	ensureCtx, ensureCancel := context.WithTimeout(context.Background(), 30*time.Second)
	ensureErr := s3cfg.EnsureBucketNamed(ensureCtx, s3cfg.HLSBucket)
	ensureCancel()
	if ensureErr != nil {
		h.failJob(job, video, fmt.Sprintf("HLS bucket unavailable: %v", ensureErr))
		return
	}
	client := &http.Client{Timeout: 10 * time.Minute}
	published := 0
	for _, a := range result.Artifacts {
		if a.Key == "" || strings.Contains(a.Key, "..") {
			h.failJob(job, video, fmt.Sprintf("worker returned unsafe artifact key %q", a.Key))
			return
		}
		content, err := fetchWorkerArtifact(client, workerBase, video.ID, a.Key)
		if err != nil {
			h.failJob(job, video, fmt.Sprintf("artifact fetch failed for %s: %v", a.Key, err))
			return
		}
		key := "hls/" + video.ID + "/" + a.Key
		contentType := a.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		err = s3cfg.PutObject(ctx, s3cfg.HLSBucket, key, contentType, content)
		cancel()
		if err != nil {
			h.failJob(job, video, fmt.Sprintf("artifact publish failed for %s: %v", key, err))
			return
		}
		published++
	}
	now := time.Now()
	base := publicHLSBase(s3cfg)
	video.HlsMasterURL = base + "/hls/" + video.ID + "/master.m3u8"
	video.ThumbnailURL = base + "/hls/" + video.ID + "/poster.jpg"
	if result.Probe.DurationSeconds > 0 {
		video.DurationSeconds = int(result.Probe.DurationSeconds)
	}
	video.Status = "READY"
	job.Status = "READY"
	job.ProgressPct = 100
	job.CompletedAt = &now
	h.Store.PersistVideo(video)
	slog.Info("Transcode job completed", "jobId", job.ID, "videoId", video.ID,
		"probed", result.Probed, "artifacts", published,
		"duration", video.DurationSeconds, "hls", video.HlsMasterURL)
}

func fetchWorkerArtifact(client *http.Client, workerBase, videoID, key string) ([]byte, error) {
	target := strings.TrimSuffix(workerBase, "/") + "/artifacts/" + videoID + "/" + key
	resp, err := client.Get(target)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("worker artifact HTTP %d", resp.StatusCode)
	}
	const maxArtifactBytes = 512 << 20
	content, err := io.ReadAll(io.LimitReader(resp.Body, maxArtifactBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > maxArtifactBytes {
		return nil, fmt.Errorf("artifact exceeds the 512 MiB fetch limit")
	}
	return content, nil
}
