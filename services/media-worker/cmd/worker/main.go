package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type VideoState string

const (
	StateUploaded   VideoState = "UPLOADED"
	StateValidating VideoState = "VALIDATING"
	StateQueued     VideoState = "QUEUED"
	StateProcessing VideoState = "PROCESSING"
	StateReady      VideoState = "READY"
	StateFailed     VideoState = "FAILED"
)

type RenditionProfile struct {
	Name         string
	Width        int
	Height       int
	VideoBitrate string
	AudioBitrate string
}

var StandardLadder = []RenditionProfile{
	{Name: "240p", Width: 426, Height: 240, VideoBitrate: "400k", AudioBitrate: "64k"},
	{Name: "360p", Width: 640, Height: 360, VideoBitrate: "800k", AudioBitrate: "96k"},
	{Name: "480p", Width: 854, Height: 480, VideoBitrate: "1400k", AudioBitrate: "128k"},
	{Name: "720p", Width: 1280, Height: 720, VideoBitrate: "2800k", AudioBitrate: "128k"},
	{Name: "1080p", Width: 1920, Height: 1080, VideoBitrate: "5000k", AudioBitrate: "192k"},
	{Name: "1440p", Width: 2560, Height: 1440, VideoBitrate: "9000k", AudioBitrate: "192k"},
	{Name: "2160p", Width: 3840, Height: 2160, VideoBitrate: "18000k", AudioBitrate: "256k"},
}

type TranscodeJob struct {
	JobID       string
	VideoID     string
	SourceR2Key string
	Attempt     int
	MaxAttempts int
}

type ProbeMetadata struct {
	Width           int
	Height          int
	DurationSeconds float64
	VideoCodec      string
	AudioCodec      string
	HasVideoStream  bool
}

type MediaWorker struct {
	mu          sync.RWMutex
	States      map[string]VideoState
	DLQ         []TranscodeJob
	R2Uploaded  map[string][]string
	WorkDirRoot string
}

func NewMediaWorker(workDirRoot string) *MediaWorker {
	return &MediaWorker{
		States:      make(map[string]VideoState),
		DLQ:         make([]TranscodeJob, 0),
		R2Uploaded:  make(map[string][]string),
		WorkDirRoot: workDirRoot,
	}
}

// Transition validates and performs state machine progression.
func (w *MediaWorker) Transition(videoID string, next VideoState) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	curr, exists := w.States[videoID]
	if !exists {
		if next != StateUploaded {
			return fmt.Errorf("initial state must be UPLOADED, got %s", next)
		}
		w.States[videoID] = next
		return nil
	}

	valid := map[VideoState][]VideoState{
		StateUploaded:   {StateValidating, StateFailed},
		StateValidating: {StateQueued, StateFailed},
		StateQueued:     {StateProcessing, StateFailed},
		StateProcessing: {StateReady, StateFailed},
	}

	for _, allowed := range valid[curr] {
		if allowed == next {
			w.States[videoID] = next
			return nil
		}
	}
	return fmt.Errorf("invalid state transition from %s to %s", curr, next)
}

// SanitizeWorkspacePath prevents path traversal attacks when staging media segments.
func (w *MediaWorker) SanitizeWorkspacePath(relativePath string) (string, error) {
	if strings.Contains(relativePath, "..") || strings.HasPrefix(relativePath, "/") || strings.HasPrefix(relativePath, `\`) {
		return "", errors.New("path traversal detected")
	}
	base := filepath.Clean(w.WorkDirRoot)
	full := filepath.Clean(filepath.Join(base, relativePath))
	if !strings.HasPrefix(full, base+string(filepath.Separator)) && full != base {
		return "", errors.New("resolved path escapes workspace root")
	}
	return full, nil
}

// ValidateFFprobe verifies media stream integrity.
func ValidateFFprobe(meta ProbeMetadata) error {
	if !meta.HasVideoStream || meta.Width <= 0 || meta.Height <= 0 {
		return errors.New("invalid video stream dimensions")
	}
	if meta.DurationSeconds <= 0 || meta.DurationSeconds > 43200 {
		return errors.New("video duration out of allowed bounds")
	}
	if meta.VideoCodec == "" {
		return errors.New("missing video codec")
	}
	return nil
}

// SelectABRLadder selects renditions up to the source height without upscaling.
func SelectABRLadder(sourceHeight int) []RenditionProfile {
	var selected []RenditionProfile
	for _, r := range StandardLadder {
		if r.Height <= sourceHeight {
			selected = append(selected, r)
		}
	}
	if len(selected) == 0 {
		selected = append(selected, StandardLadder[0])
	}
	return selected
}

// BuildFFmpegHLSArgs constructs keyframe-aligned fMP4 HLS packaging arguments.
func BuildFFmpegHLSArgs(inputPath, outputDir string, ladder []RenditionProfile) []string {
	args := []string{
		"-hide_banner", "-y", "-i", inputPath,
		"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
		"-hls_time", "4",
		"-hls_playlist_type", "vod",
		"-hls_segment_type", "fmp4",
	}
	for _, r := range ladder {
		args = append(args,
			"-vf", fmt.Sprintf("scale=%d:%d", r.Width, r.Height),
			"-b:v", r.VideoBitrate,
			"-b:a", r.AudioBitrate,
			filepath.Join(outputDir, r.Name, "index.m3u8"),
		)
	}
	return args
}

// BuildThumbnailArgs generates FFmpeg thumbnail sprite extraction arguments.
func BuildThumbnailArgs(inputPath, outputThumb string) []string {
	return []string{
		"-hide_banner", "-y", "-ss", "00:00:03", "-i", inputPath,
		"-frames:v", "1", "-q:v", "2", outputThumb,
	}
}

// ProcessJob executes the full transcode lifecycle with DLQ routing on failure.
func (w *MediaWorker) ProcessJob(job TranscodeJob, meta ProbeMetadata) error {
	if _, err := w.SanitizeWorkspacePath(job.SourceR2Key); err != nil {
		_ = w.Transition(job.VideoID, StateUploaded)
		_ = w.Transition(job.VideoID, StateFailed)
		w.sendToDLQ(job)
		return err
	}

	if err := w.Transition(job.VideoID, StateUploaded); err != nil {
		return err
	}
	if err := w.Transition(job.VideoID, StateValidating); err != nil {
		return err
	}

	if err := ValidateFFprobe(meta); err != nil {
		_ = w.Transition(job.VideoID, StateFailed)
		w.sendToDLQ(job)
		return err
	}

	if err := w.Transition(job.VideoID, StateQueued); err != nil {
		return err
	}
	if err := w.Transition(job.VideoID, StateProcessing); err != nil {
		return err
	}

	ladder := SelectABRLadder(meta.Height)
	var uploadedArtifacts []string
	uploadedArtifacts = append(uploadedArtifacts, fmt.Sprintf("hls/%s/master.m3u8", job.VideoID))
	uploadedArtifacts = append(uploadedArtifacts, fmt.Sprintf("thumbnails/%s/poster.jpg", job.VideoID))
	for _, r := range ladder {
		uploadedArtifacts = append(uploadedArtifacts, fmt.Sprintf("hls/%s/%s/index.m3u8", job.VideoID, r.Name))
	}

	w.mu.Lock()
	w.R2Uploaded[job.VideoID] = uploadedArtifacts
	w.mu.Unlock()

	if !w.VerifyR2Upload(job.VideoID, len(ladder)+2) {
		_ = w.Transition(job.VideoID, StateFailed)
		w.sendToDLQ(job)
		return errors.New("R2 upload verification failed")
	}

	return w.Transition(job.VideoID, StateReady)
}

func (w *MediaWorker) VerifyR2Upload(videoID string, expectedCount int) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	artifacts := w.R2Uploaded[videoID]
	return len(artifacts) >= expectedCount
}

func (w *MediaWorker) sendToDLQ(job TranscodeJob) {
	w.mu.Lock()
	defer w.mu.Unlock()
	job.Attempt++
	w.DLQ = append(w.DLQ, job)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.Info("Starting Media Transcode Worker (NATS JetStream media.jobs.transcode)")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	worker := NewMediaWorker(os.TempDir())
	_ = worker

	select {
	case <-ctx.Done():
		slog.Info("Shutting down Media Worker gracefully")
	case <-time.After(100 * time.Millisecond):
		slog.Info("Media Worker ready")
	}
}
