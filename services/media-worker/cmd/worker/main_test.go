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

	// Verify keyframe alignment & fMP4 flags
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
