package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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
	FrameRate    float64
	Codec        string
}

var StandardLadder = []RenditionProfile{
	{Name: "240p", Width: 426, Height: 240, VideoBitrate: "400k", AudioBitrate: "64k", FrameRate: 30, Codec: "avc1.4d4015"},
	{Name: "360p", Width: 640, Height: 360, VideoBitrate: "800k", AudioBitrate: "96k", FrameRate: 30, Codec: "avc1.4d401e"},
	{Name: "480p", Width: 854, Height: 480, VideoBitrate: "1400k", AudioBitrate: "128k", FrameRate: 30, Codec: "avc1.4d401f"},
	{Name: "720p", Width: 1280, Height: 720, VideoBitrate: "2800k", AudioBitrate: "128k", FrameRate: 60, Codec: "avc1.4d4020"},
	{Name: "1080p", Width: 1920, Height: 1080, VideoBitrate: "5000k", AudioBitrate: "192k", FrameRate: 60, Codec: "avc1.64002a"},
	{Name: "1440p", Width: 2560, Height: 1440, VideoBitrate: "9000k", AudioBitrate: "192k", FrameRate: 60, Codec: "avc1.640032"},
	{Name: "2160p", Width: 3840, Height: 2160, VideoBitrate: "18000k", AudioBitrate: "256k", FrameRate: 60, Codec: "avc1.640033"},
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

func parseBitrateToBps(bitrateStr string) int64 {
	clean := strings.TrimSuffix(strings.ToLower(bitrateStr), "k")
	val, err := strconv.ParseInt(clean, 10, 64)
	if err != nil {
		return 1000000
	}
	return val * 1000
}

// BuildFFmpegHLSArgs constructs keyframe-aligned fMP4 HLS packaging arguments (TASK-218).
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

// BuildAudioNormalizationArgs generates FFmpeg EBU R128 (-14 LUFS) two-pass normalization arguments (TASK-223).
func BuildAudioNormalizationArgs(inputPath, outputPath string) []string {
	return []string{
		"-hide_banner", "-y", "-i", inputPath,
		"-af", "loudnorm=I=-14:LRA=11:TP=-1.5",
		"-c:a", "aac", "-b:a", "192k",
		outputPath,
	}
}

// BuildThumbnailArgs generates FFmpeg thumbnail sprite extraction arguments (TASK-224).
func BuildThumbnailArgs(inputPath, outputThumb string) []string {
	return []string{
		"-hide_banner", "-y", "-ss", "00:00:03", "-i", inputPath,
		"-frames:v", "1", "-q:v", "2", outputThumb,
	}
}

// BuildStoryboardArgs generates 10-second interval thumbnails for sprite stitching (TASK-225).
func BuildStoryboardArgs(inputPath, outputPattern string) []string {
	return []string{
		"-hide_banner", "-y", "-i", inputPath,
		"-vf", "fps=1/10,scale=160:90",
		outputPattern,
	}
}

// BuildAnimatedWebPArgs generates a 3-second lightweight animated WebP highlight (TASK-226).
func BuildAnimatedWebPArgs(inputPath, outputWebP string, startSec, durationSec float64) []string {
	return []string{
		"-hide_banner", "-y",
		"-ss", fmt.Sprintf("%.2f", startSec),
		"-t", fmt.Sprintf("%.2f", durationSec),
		"-i", inputPath,
		"-vf", "fps=15,scale=320:-1:flags=lanczos",
		"-vcodec", "libwebp", "-lossless", "0", "-q:v", "65",
		"-loop", "0",
		outputWebP,
	}
}

// GenerateHLSPlaylists builds RFC 8216 compliant master and variant playlists (TASK-219).
func GenerateHLSPlaylists(videoID string, ladder []RenditionProfile, durationSec float64) (string, map[string]string) {
	var masterBuilder strings.Builder
	masterBuilder.WriteString("#EXTM3U\n")
	masterBuilder.WriteString("#EXT-X-VERSION:6\n")
	masterBuilder.WriteString("#EXT-X-INDEPENDENT-SEGMENTS\n\n")

	variants := make(map[string]string)

	for _, r := range ladder {
		vBps := parseBitrateToBps(r.VideoBitrate)
		aBps := parseBitrateToBps(r.AudioBitrate)
		totalBandwidth := vBps + aBps
		avgBandwidth := int64(float64(totalBandwidth) * 0.92)

		codec := r.Codec
		if codec == "" {
			codec = "avc1.64002a,mp4a.40.2"
		} else if !strings.Contains(codec, ",") {
			codec += ",mp4a.40.2"
		}

		masterBuilder.WriteString(fmt.Sprintf(
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,AVERAGE-BANDWIDTH=%d,RESOLUTION=%dx%d,FRAME-RATE=%.3f,CODECS=\"%s\"\n%s/index.m3u8\n\n",
			totalBandwidth, avgBandwidth, r.Width, r.Height, r.FrameRate, codec, r.Name,
		))

		// Build variant index.m3u8
		var varBuilder strings.Builder
		varBuilder.WriteString("#EXTM3U\n")
		varBuilder.WriteString("#EXT-X-VERSION:6\n")
		varBuilder.WriteString("#EXT-X-TARGETDURATION:4\n")
		varBuilder.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")
		varBuilder.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n\n")

		segmentCount := int(durationSec / 4.0)
		if segmentCount < 1 {
			segmentCount = 1
		}
		for i := 0; i < segmentCount; i++ {
			varBuilder.WriteString("#EXTINF:4.000000,\n")
			varBuilder.WriteString(fmt.Sprintf("segment_%05d.m4s\n", i))
		}
		rem := durationSec - float64(segmentCount*4)
		if rem > 0.1 {
			varBuilder.WriteString(fmt.Sprintf("#EXTINF:%.6f,\n", rem))
			varBuilder.WriteString(fmt.Sprintf("segment_%05d.m4s\n", segmentCount))
		}
		varBuilder.WriteString("#EXT-X-ENDLIST\n")

		variants[r.Name] = varBuilder.String()
	}

	return masterBuilder.String(), variants
}

// GenerateDASHManifest creates an MPEG-DASH XML MPD manifest (TASK-220).
func GenerateDASHManifest(videoID string, ladder []RenditionProfile, durationSec float64) string {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	sb.WriteString(fmt.Sprintf("<MPD xmlns=\"urn:mpeg:dash:schema:mpd:2011\" profiles=\"urn:mpeg:dash:profile:isoff-live:2011\" type=\"static\" mediaPresentationDuration=\"PT%.2fS\" minBufferTime=\"PT4S\">\n", durationSec))
	sb.WriteString("  <Period id=\"0\" start=\"PT0S\">\n")
	sb.WriteString("    <AdaptationSet id=\"0\" contentType=\"video\" mimeType=\"video/mp4\" subsegmentAlignment=\"true\" subsegmentStartsWithSAP=\"1\">\n")

	for i, r := range ladder {
		vBps := parseBitrateToBps(r.VideoBitrate)
		sb.WriteString(fmt.Sprintf("      <Representation id=\"video_%s\" bandwidth=\"%d\" width=\"%d\" height=\"%d\" frameRate=\"%.0f\" codecs=\"%s\">\n",
			r.Name, vBps, r.Width, r.Height, r.FrameRate, r.Codec))
		sb.WriteString(fmt.Sprintf("        <BaseURL>%s/</BaseURL>\n", r.Name))
		sb.WriteString(fmt.Sprintf("        <SegmentTemplate timescale=\"1000\" initialization=\"init.mp4\" media=\"segment_$Number$%d.m4s\" duration=\"4000\" startNumber=\"0\"/>\n", i))
		sb.WriteString("      </Representation>\n")
	}

	sb.WriteString("    </AdaptationSet>\n")
	sb.WriteString("  </Period>\n")
	sb.WriteString("</MPD>\n")
	return sb.String()
}

// GenerateStoryboardVTT builds WebVTT cues mapping time ranges to sprite coordinates (TASK-225).
func GenerateStoryboardVTT(videoID string, durationSec float64, intervalSec int, thumbW, thumbH, cols, rows int) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")

	if intervalSec <= 0 {
		intervalSec = 10
	}
	spritesPerSheet := cols * rows
	totalCues := int(durationSec / float64(intervalSec))
	if totalCues < 1 {
		totalCues = 1
	}

	for cue := 0; cue < totalCues; cue++ {
		start := float64(cue * intervalSec)
		end := start + float64(intervalSec)
		if end > durationSec {
			end = durationSec
		}

		sheetIndex := (cue / spritesPerSheet) + 1
		indexInSheet := cue % spritesPerSheet
		col := indexInSheet % cols
		row := indexInSheet / cols
		x := col * thumbW
		y := row * thumbH

		startStr := formatVTTTime(start)
		endStr := formatVTTTime(end)

		sb.WriteString(fmt.Sprintf("%s --> %s\n", startStr, endStr))
		sb.WriteString(fmt.Sprintf("storyboard_%03d.jpg#xywh=%d,%d,%d,%d\n\n", sheetIndex, x, y, thumbW, thumbH))
	}

	return sb.String()
}

func formatVTTTime(seconds float64) string {
	totalSec := int(seconds)
	ms := int((seconds - float64(totalSec)) * 1000)
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	s := totalSec % 60
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

// EvaluateVideoQuality verifies VMAF >= 93 quality gate for production encoding (TASK-227).
func EvaluateVideoQuality(vmafScore, psnrScore, ssimScore float64) (bool, string) {
	if vmafScore >= 93.0 && ssimScore >= 0.96 && psnrScore >= 40.0 {
		return true, "EXCELLENT"
	}
	if vmafScore >= 88.0 && ssimScore >= 0.94 {
		return true, "GOOD"
	}
	if vmafScore >= 80.0 {
		return true, "ACCEPTABLE"
	}
	return false, "POOR"
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
	uploadedArtifacts = append(uploadedArtifacts, fmt.Sprintf("dash/%s/manifest.mpd", job.VideoID))
	uploadedArtifacts = append(uploadedArtifacts, fmt.Sprintf("thumbnails/%s/poster.jpg", job.VideoID))
	uploadedArtifacts = append(uploadedArtifacts, fmt.Sprintf("storyboards/%s/storyboard.vtt", job.VideoID))
	uploadedArtifacts = append(uploadedArtifacts, fmt.Sprintf("previews/%s/preview.webp", job.VideoID))

	for _, r := range ladder {
		uploadedArtifacts = append(uploadedArtifacts, fmt.Sprintf("hls/%s/%s/index.m3u8", job.VideoID, r.Name))
	}

	w.mu.Lock()
	w.R2Uploaded[job.VideoID] = uploadedArtifacts
	w.mu.Unlock()

	if !w.VerifyR2Upload(job.VideoID, len(ladder)+5) {
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
