package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMediaWorkerABRLadderAndLifecycle(t *testing.T) {
	w := NewMediaWorker("/var/gvp/work")

	// 1080p source should produce 240p, 360p, 480p, 720p, 1080p (5 renditions, no 1440p/2160p upscaling)
	ladder := SelectABRLadder(1080)
	if len(ladder) != 5 || ladder[len(ladder)-1].Name != "1080p" {
		t.Fatalf("expected 5 renditions capped at 1080p, got %d", len(ladder))
	}

	// Verify keyframe alignment & fMP4 flags (TASK-218)
	args := BuildFFmpegHLSArgs("input.mp4", "/out", ladder)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-g 48 -keyint_min 48 -sc_threshold 0") || !strings.Contains(joined, "fmp4") {
		t.Fatalf("missing keyframe alignment or fmp4 flags: %s", joined)
	}

	// Execute full job lifecycle: UPLOADED -> VALIDATING -> QUEUED -> PROCESSING -> READY
	job := TranscodeJob{
		JobID:       "job-1",
		VideoID:     "vid-100",
		SourceR2Key: "raw/user1/vid-100.mp4",
		MaxAttempts: 3,
	}
	meta := ProbeMetadata{
		Width:           1920,
		Height:          1080,
		DurationSeconds: 185.4,
		VideoCodec:      "h264",
		AudioCodec:      "aac",
		HasVideoStream:  true,
	}
	if err := w.ProcessJob(job, meta); err != nil {
		t.Fatalf("expected successful job execution, got %v", err)
	}
	if w.States["vid-100"] != StateReady {
		t.Fatalf("expected state READY, got %s", w.States["vid-100"])
	}
}

func TestMediaWorkerPathTraversalAndDLQ(t *testing.T) {
	w := NewMediaWorker("/var/gvp/work")

	// Path traversal rejection
	if _, err := w.SanitizeWorkspacePath("../../etc/passwd"); err == nil {
		t.Fatal("expected path traversal error")
	}

	// Corrupted video routes to DLQ and transitions to FAILED
	badJob := TranscodeJob{
		JobID:       "job-bad",
		VideoID:     "vid-bad",
		SourceR2Key: "raw/user1/vid-bad.mp4",
	}
	badMeta := ProbeMetadata{HasVideoStream: false}
	if err := w.ProcessJob(badJob, badMeta); err == nil {
		t.Fatal("expected error for invalid probe metadata")
	}
	if w.States["vid-bad"] != StateFailed {
		t.Fatalf("expected FAILED state, got %s", w.States["vid-bad"])
	}
	if len(w.DLQ) != 1 || w.DLQ[0].JobID != "job-bad" {
		t.Fatalf("expected job routed to DLQ, got %+v", w.DLQ)
	}
}

func TestGenerateHLSPlaylists(t *testing.T) {
	ladder := SelectABRLadder(1080)
	master, variants := GenerateHLSPlaylists("vid-abc", ladder, 120.0)

	if !strings.HasPrefix(master, "#EXTM3U") {
		t.Fatalf("master playlist missing EXTM3U header: %s", master)
	}
	if !strings.Contains(master, "#EXT-X-STREAM-INF:BANDWIDTH=5192000,AVERAGE-BANDWIDTH=4776640,RESOLUTION=1920x1080") {
		t.Fatalf("missing 1080p stream inf in master: %s", master)
	}
	if len(variants) != len(ladder) {
		t.Fatalf("expected %d variant playlists, got %d", len(ladder), len(variants))
	}
	v1080 := variants["1080p"]
	if !strings.Contains(v1080, "#EXT-X-TARGETDURATION:4") || !strings.Contains(v1080, "segment_00000.m4s") {
		t.Fatalf("invalid variant playlist format: %s", v1080)
	}
}

func TestGenerateDASHManifest(t *testing.T) {
	ladder := SelectABRLadder(1080)
	mpd := GenerateDASHManifest("vid-abc", ladder, 120.0)

	if !strings.HasPrefix(mpd, "<?xml") || !strings.Contains(mpd, "<MPD") {
		t.Fatalf("invalid DASH MPD header: %s", mpd)
	}
	if !strings.Contains(mpd, `id="video_1080p"`) {
		t.Fatalf("missing 1080p representation in MPD: %s", mpd)
	}
}

func TestGenerateStoryboardVTT(t *testing.T) {
	vtt := GenerateStoryboardVTT("vid-abc", 65.0, 10, 160, 90, 5, 5)

	if !strings.HasPrefix(vtt, "WEBVTT") {
		t.Fatalf("missing WEBVTT header: %s", vtt)
	}
	if !strings.Contains(vtt, "00:00:00.000 --> 00:00:10.000") || !strings.Contains(vtt, "storyboard_001.jpg#xywh=0,0,160,90") {
		t.Fatalf("missing expected first cue in storyboard: %s", vtt)
	}
	if !strings.Contains(vtt, "storyboard_001.jpg#xywh=160,0,160,90") {
		t.Fatalf("missing expected second cue coordinates: %s", vtt)
	}
}

func TestAudioNormalizationAndQualityGates(t *testing.T) {
	// Audio normalization arguments (TASK-223)
	audioArgs := BuildAudioNormalizationArgs("input.wav", "norm.aac")
	joined := strings.Join(audioArgs, " ")
	if !strings.Contains(joined, "loudnorm=I=-14:LRA=11:TP=-1.5") {
		t.Fatalf("missing loudnorm filter in audio args: %s", joined)
	}

	// Quality gate evaluation (TASK-227)
	approved, grade := EvaluateVideoQuality(94.2, 42.1, 0.98)
	if !approved || grade != "EXCELLENT" {
		t.Fatalf("expected EXCELLENT approved, got %v / %s", approved, grade)
	}

	approved, grade = EvaluateVideoQuality(75.0, 32.0, 0.85)
	if approved || grade != "POOR" {
		t.Fatalf("expected POOR rejected, got %v / %s", approved, grade)
	}
}

func postIntake(t *testing.T, srv *healthServer, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/jobs/transcode", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.handleTranscodeIntake(rec, req)
	return rec
}

func TestTranscodeIntakeExecutesJob(t *testing.T) {
	srv := newHealthServer("media-worker", NewMediaWorker(t.TempDir()))
	payload := `{"jobId":"job-intake-1","videoId":"vid-intake-1","sourceR2Key":"raw/user1/vid-intake-1.mp4",` +
		`"width":1920,"height":1080,"durationSeconds":120,"videoCodec":"h264","hasVideoStream":true}`
	rec := postIntake(t, srv, payload)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result struct {
		State  string `json:"state"`
		Probed bool   `json:"probed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid intake response: %v", err)
	}
	if result.State != "READY" || !result.Probed {
		t.Fatalf("expected READY/probed response, got %+v", result)
	}
	if srv.worker.States["vid-intake-1"] != StateReady {
		t.Fatalf("expected worker state READY, got %s", srv.worker.States["vid-intake-1"])
	}
}

func TestTranscodeIntakeAssumedProfileAndValidation(t *testing.T) {
	srv := newHealthServer("media-worker", NewMediaWorker(t.TempDir()))

	// No probe metadata: dev default assumes the 1080p source profile.
	rec := postIntake(t, srv, `{"jobId":"job-intake-2","videoId":"vid-intake-2","sourceR2Key":"raw/user1/v.mp4"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with assumed profile, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"probed":false`) {
		t.Fatalf("expected probed=false flag, got %s", rec.Body.String())
	}

	// Production gate rejects unprobed jobs.
	t.Setenv("REQUIRE_PROBED_SOURCE", "true")
	rec = postIntake(t, srv, `{"jobId":"job-intake-3","videoId":"vid-intake-3","sourceR2Key":"raw/user1/v.mp4"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 under REQUIRE_PROBED_SOURCE, got %d", rec.Code)
	}
	if os.Getenv("REQUIRE_PROBED_SOURCE") != "true" {
		t.Fatal("expected env gate to hold during test")
	}

	// Missing identifiers are a bad request, not a job failure.
	rec = postIntake(t, srv, `{"videoId":"vid-no-key"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing sourceR2Key, got %d", rec.Code)
	}

	// Non-POST methods are rejected.
	req := httptest.NewRequest(http.MethodGet, "/jobs/transcode", nil)
	getRec := httptest.NewRecorder()
	srv.handleTranscodeIntake(getRec, req)
	if getRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET, got %d", getRec.Code)
	}
}

func TestEditRenderIntakeExecutesEDL(t *testing.T) {
	srv := newHealthServer("media-worker", NewMediaWorker(t.TempDir()))
	payload := `{"jobId":"job-edit-1","videoId":"vid-edit-1","sourceR2Key":"raw/user1/vid-edit-1.mp4",` +
		`"height":1080,"durationSeconds":1900,"edl":{"trimStartSec":12,"trimEndSec":1820,` +
		`"splitPoints":[310,890],"replacementAudio":"Royalty-Free Synthwave","blurRegions":` +
		`[{"id":"blur_1","label":"plate","x":42,"y":68,"startSec":120,"endSec":165}],` +
		`"mutedSegments":[{"startSec":410,"endSec":445,"reason":"copyrighted music"}]}}`
	req := httptest.NewRequest(http.MethodPost, "/jobs/edit-render", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	srv.handleEditRenderIntake(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"state":"READY"`) {
		t.Fatalf("expected READY state, got %s", rec.Body.String())
	}
	if srv.worker.States["vid-edit-1"] != StateReady {
		t.Fatalf("expected worker state READY, got %s", srv.worker.States["vid-edit-1"])
	}
	found := false
	for _, key := range srv.worker.R2Uploaded["vid-edit-1"] {
		if strings.Contains(key, "/edited/master.m3u8") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected edited HLS master artifact, got %v", srv.worker.R2Uploaded["vid-edit-1"])
	}
}

func TestEditRenderIntakeRejectsBadEDL(t *testing.T) {
	srv := newHealthServer("media-worker", NewMediaWorker(t.TempDir()))

	// Inverted trim range.
	payload := `{"jobId":"job-edit-2","videoId":"vid-edit-2","sourceR2Key":"raw/u/v.mp4",` +
		`"edl":{"trimStartSec":500,"trimEndSec":100}}`
	req := httptest.NewRequest(http.MethodPost, "/jobs/edit-render", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	srv.handleEditRenderIntake(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for inverted trim, got %d: %s", rec.Code, rec.Body.String())
	}
	if srv.worker.States["vid-edit-2"] != StateFailed {
		t.Fatalf("expected worker state FAILED, got %s", srv.worker.States["vid-edit-2"])
	}

	// Pure validator: descending splits and out-of-range blur.
	if err := ValidateEDL(EditDecisionList{TrimStartSec: 0, TrimEndSec: 100, SplitPoints: []int{80, 40}}, 200); err == nil {
		t.Fatal("expected error for descending split points")
	}
	badBlur := EditDecisionList{TrimStartSec: 0, TrimEndSec: 100,
		BlurRegions: []BlurRegion{{ID: "b1", X: 150, Y: 10, StartSec: 5, EndSec: 10}}}
	if err := ValidateEDL(badBlur, 200); err == nil {
		t.Fatal("expected error for out-of-range blur coordinates")
	}
	if err := ValidateEDL(EditDecisionList{TrimStartSec: 0, TrimEndSec: 500}, 200); err == nil {
		t.Fatal("expected error when trimEnd exceeds known duration")
	}
}

func TestTranscodeIntakeRejectsTraversalKey(t *testing.T) {
	srv := newHealthServer("media-worker", NewMediaWorker(t.TempDir()))
	var buf bytes.Buffer
	buf.WriteString(`{"jobId":"job-intake-4","videoId":"vid-intake-4","sourceR2Key":"../../etc/passwd",`)
	buf.WriteString(`"width":1920,"height":1080,"durationSeconds":60,"videoCodec":"h264","hasVideoStream":true}`)
	rec := postIntake(t, srv, buf.String())
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for traversal key, got %d: %s", rec.Code, rec.Body.String())
	}
	if srv.worker.States["vid-intake-4"] != StateFailed {
		t.Fatalf("expected worker state FAILED, got %s", srv.worker.States["vid-intake-4"])
	}
}
