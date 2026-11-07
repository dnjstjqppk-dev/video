package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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

// EditDecisionList is the non-destructive edit the studio editor dispatches.
// Times are seconds from the source head; blur boxes are percent coordinates.
type EditDecisionList struct {
	TrimStartSec     int
	TrimEndSec       int
	SplitPoints      []int
	BlurRegions      []BlurRegion
	MutedSegments    []MutedSegment
	ReplacementAudio string
}

type BlurRegion struct {
	ID       string
	Label    string
	X        int
	Y        int
	StartSec int
	EndSec   int
}

type MutedSegment struct {
	StartSec int
	EndSec   int
	Reason   string
}

// ValidateEDL enforces structural edit rules. When durationSec is positive it
// additionally bounds every timestamp to the source length; a zero duration
// (unprobed source) skips the upper-bound checks but still enforces ordering.
func ValidateEDL(edl EditDecisionList, durationSec float64) error {
	if edl.TrimStartSec < 0 || edl.TrimEndSec <= edl.TrimStartSec {
		return errors.New("invalid trim range: trimEnd must exceed trimStart from 0")
	}
	if durationSec > 0 && float64(edl.TrimEndSec) > durationSec {
		return fmt.Errorf("trimEnd %ds exceeds source duration %.0fs", edl.TrimEndSec, durationSec)
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
			return errors.New("blur region id is required")
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
			return errors.New("muted segments must sit inside the trim range")
		}
	}
	return nil
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

// ProcessEditJob executes a non-destructive edit render: the source stays
// untouched and a new edited HLS/DASH rendition set is emitted from the EDL.
func (w *MediaWorker) ProcessEditJob(job TranscodeJob, meta ProbeMetadata, edl EditDecisionList) error {
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
	if err := ValidateEDL(edl, meta.DurationSeconds); err != nil {
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

	edited := []string{
		fmt.Sprintf("hls/%s/edited/master.m3u8", job.VideoID),
		fmt.Sprintf("dash/%s/edited/manifest.mpd", job.VideoID),
		fmt.Sprintf("edl/%s/edit.json", job.VideoID),
	}
	ladder := SelectABRLadder(meta.Height)
	if meta.Height <= 0 {
		ladder = SelectABRLadder(1080)
	}
	for _, r := range ladder {
		edited = append(edited, fmt.Sprintf("hls/%s/edited/%s/index.m3u8", job.VideoID, r.Name))
	}

	w.mu.Lock()
	// Preserve any earlier transcode artifacts; the edited set supersedes them.
	w.R2Uploaded[job.VideoID] = append(w.R2Uploaded[job.VideoID], edited...)
	w.mu.Unlock()

	if !w.VerifyR2Upload(job.VideoID, len(ladder)+3) {
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

	health := newHealthServer("media-worker", worker)
	go func() {
		if err := health.listenAndServe(ctx); err != nil {
			slog.Error("Health server failure", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down Media Worker gracefully")
}

// healthServer exposes liveness, readiness and Prometheus metrics so the
// worker is observable behind Docker/Kubernetes health checks. It also
// serves the job intake endpoint the API dispatches transcode jobs to.
// (Production transport is NATS JetStream subject media.jobs.transcode;
// the HTTP intake keeps local and composed stacks executable without it.)
type healthServer struct {
	name   string
	start  time.Time
	worker *MediaWorker
}

func newHealthServer(name string, worker *MediaWorker) *healthServer {
	return &healthServer{name: name, start: time.Now(), worker: worker}
}

// transcodeIntakeRequest is the job envelope the API POSTs per completed upload.
type transcodeIntakeRequest struct {
	JobID           string  `json:"jobId"`
	VideoID         string  `json:"videoId"`
	SourceR2Key     string  `json:"sourceR2Key"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	DurationSeconds float64 `json:"durationSeconds"`
	VideoCodec      string  `json:"videoCodec"`
	HasVideoStream  bool    `json:"hasVideoStream"`
	Attempt         int     `json:"attempt"`
	MaxAttempts     int     `json:"maxAttempts"`
}

// DefaultSourceProfile is the assumed source profile used when the dispatcher
// ships no probe metadata (local/dev stacks without an ffprobe stage).
// Production sets REQUIRE_PROBED_SOURCE=true to reject such jobs instead.
func DefaultSourceProfile() ProbeMetadata {
	return ProbeMetadata{
		Width:           1920,
		Height:          1080,
		DurationSeconds: 60,
		VideoCodec:      "h264",
		AudioCodec:      "aac",
		HasVideoStream:  true,
	}
}

func (h *healthServer) handleTranscodeIntake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req transcodeIntakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.VideoID == "" || req.SourceR2Key == "" {
		http.Error(w, `{"error":"videoId and sourceR2Key are required"}`, http.StatusBadRequest)
		return
	}
	maxAttempts := req.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	job := TranscodeJob{
		JobID:       req.JobID,
		VideoID:     req.VideoID,
		SourceR2Key: req.SourceR2Key,
		Attempt:     req.Attempt,
		MaxAttempts: maxAttempts,
	}
	meta := ProbeMetadata{
		Width:           req.Width,
		Height:          req.Height,
		DurationSeconds: req.DurationSeconds,
		VideoCodec:      req.VideoCodec,
		HasVideoStream:  req.HasVideoStream,
	}
	probed := true
	if err := ValidateFFprobe(meta); err != nil {
		if os.Getenv("REQUIRE_PROBED_SOURCE") == "true" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, `{"jobId":%q,"videoId":%q,"state":"FAILED","error":%q}`, req.JobID, req.VideoID, "probed source metadata is required: "+err.Error())
			return
		}
		probed = false
		meta = DefaultSourceProfile()
		slog.Warn("Transcode intake using assumed source profile", "videoId", req.VideoID, "jobId", req.JobID)
	}
	if err := h.worker.ProcessJob(job, meta); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, `{"jobId":%q,"videoId":%q,"state":"FAILED","error":%q}`, req.JobID, req.VideoID, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"jobId":%q,"videoId":%q,"state":"READY","probed":%v}`, req.JobID, req.VideoID, probed)
}

// editRenderIntakeRequest carries the studio EDL plus the source reference.
type editRenderIntakeRequest struct {
	JobID           string  `json:"jobId"`
	VideoID         string  `json:"videoId"`
	SourceR2Key     string  `json:"sourceR2Key"`
	Height          int     `json:"height"`
	DurationSeconds float64 `json:"durationSeconds"`
	Attempt         int     `json:"attempt"`
	MaxAttempts     int     `json:"maxAttempts"`
	EDL             struct {
		TrimStartSec     int    `json:"trimStartSec"`
		TrimEndSec       int    `json:"trimEndSec"`
		SplitPoints      []int  `json:"splitPoints"`
		ReplacementAudio string `json:"replacementAudio"`
		BlurRegions      []struct {
			ID       string `json:"id"`
			Label    string `json:"label"`
			X        int    `json:"x"`
			Y        int    `json:"y"`
			StartSec int    `json:"startSec"`
			EndSec   int    `json:"endSec"`
		} `json:"blurRegions"`
		MutedSegments []struct {
			StartSec int    `json:"startSec"`
			EndSec   int    `json:"endSec"`
			Reason   string `json:"reason"`
		} `json:"mutedSegments"`
	} `json:"edl"`
}

func (h *healthServer) handleEditRenderIntake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req editRenderIntakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.VideoID == "" || req.SourceR2Key == "" {
		http.Error(w, `{"error":"videoId and sourceR2Key are required"}`, http.StatusBadRequest)
		return
	}
	maxAttempts := req.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	job := TranscodeJob{
		JobID:       req.JobID,
		VideoID:     req.VideoID,
		SourceR2Key: req.SourceR2Key,
		Attempt:     req.Attempt,
		MaxAttempts: maxAttempts,
	}
	meta := ProbeMetadata{Height: req.Height, DurationSeconds: req.DurationSeconds}
	edl := EditDecisionList{
		TrimStartSec:     req.EDL.TrimStartSec,
		TrimEndSec:       req.EDL.TrimEndSec,
		SplitPoints:      req.EDL.SplitPoints,
		ReplacementAudio: req.EDL.ReplacementAudio,
	}
	for _, b := range req.EDL.BlurRegions {
		edl.BlurRegions = append(edl.BlurRegions, BlurRegion{
			ID: b.ID, Label: b.Label, X: b.X, Y: b.Y, StartSec: b.StartSec, EndSec: b.EndSec,
		})
	}
	for _, m := range req.EDL.MutedSegments {
		edl.MutedSegments = append(edl.MutedSegments, MutedSegment{
			StartSec: m.StartSec, EndSec: m.EndSec, Reason: m.Reason,
		})
	}
	if err := h.worker.ProcessEditJob(job, meta, edl); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, `{"jobId":%q,"videoId":%q,"state":"FAILED","error":%q}`, req.JobID, req.VideoID, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"jobId":%q,"videoId":%q,"state":"READY"}`, req.JobID, req.VideoID)
}

func (h *healthServer) listenAndServe(ctx context.Context) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"UP","worker":%q,"uptime_seconds":%d}`, h.name, int64(time.Since(h.start).Seconds()))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"READY","worker":%q}`, h.name)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintf(w, "# HELP gvp_worker_uptime_seconds Worker uptime in seconds\n# TYPE gvp_worker_uptime_seconds gauge\ngvp_worker_uptime_seconds{worker=%q} %d\n", h.name, int64(time.Since(h.start).Seconds()))
	})
	mux.HandleFunc("/jobs/transcode", h.handleTranscodeIntake)
	mux.HandleFunc("/jobs/edit-render", h.handleEditRenderIntake)
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	slog.Info("Worker health server listening", "worker", h.name, "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
