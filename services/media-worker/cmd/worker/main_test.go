package main

import (
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
