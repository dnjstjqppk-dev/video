package main

import (
	"net/http"
	"strings"
	"testing"
)

// Shorts product rules (MASTER_DIRECTIVE 3.5 lists features, not numbers):
// at most 180 seconds, vertical or square. Mirrored in the API publish
// gate (services/api/internal/uploads/dispatch.go).
func TestValidateShortsProfile(t *testing.T) {
	cases := []struct {
		name      string
		videoType string
		meta      ProbeMetadata
		wantErr   string
	}{
		{"vertical shorts ok", "shorts", ProbeMetadata{Width: 1080, Height: 1920, DurationSeconds: 42}, ""},
		{"square shorts at cap ok", "shorts", ProbeMetadata{Width: 1080, Height: 1080, DurationSeconds: 180}, ""},
		{"overlong shorts rejected", "shorts", ProbeMetadata{Width: 1080, Height: 1920, DurationSeconds: 181}, "180"},
		{"landscape shorts rejected", "shorts", ProbeMetadata{Width: 1920, Height: 1080, DurationSeconds: 42}, "vertical"},
		{"standard ignores shape", "standard", ProbeMetadata{Width: 1920, Height: 1080, DurationSeconds: 3600}, ""},
		{"standard ignores length", "standard", ProbeMetadata{Width: 1080, Height: 1920, DurationSeconds: 3600}, ""},
		{"empty type ignores rules", "", ProbeMetadata{Width: 1920, Height: 1080, DurationSeconds: 3600}, ""},
	}
	for _, tc := range cases {
		err := ValidateShortsProfile(tc.videoType, tc.meta)
		if tc.wantErr == "" && err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if tc.wantErr != "" {
			if err == nil {
				t.Fatalf("%s: expected error containing %q, got nil", tc.name, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("%s: error %q should mention %q", tc.name, err.Error(), tc.wantErr)
			}
		}
	}
}

func TestSelectABRLadderForSource(t *testing.T) {
	// Landscape keeps the standard height-capped ladder.
	land := SelectABRLadderForSource(1920, 1080)
	if len(land) != 5 || land[0].Width != 426 || land[0].Height != 240 || land[4].Width != 1920 || land[4].Height != 1080 {
		t.Fatalf("landscape ladder wrong: %+v", land)
	}
	// Vertical sources get width-capped transposed rungs (no upscaling).
	vert := SelectABRLadderForSource(1080, 1920)
	wantDims := [][2]int{{240, 426}, {360, 640}, {480, 848}}
	if len(vert) != len(wantDims) {
		t.Fatalf("vertical ladder wants %d rungs, got %d: %+v", len(wantDims), len(vert), vert)
	}
	for i, want := range wantDims {
		if vert[i].Width != want[0] || vert[i].Height != want[1] {
			t.Fatalf("rung %d wants %dx%d, got %dx%d", i, want[0], want[1], vert[i].Width, vert[i].Height)
		}
	}
	if vert[1].Name != "360p" || vert[1].VideoBitrate != "800k" {
		t.Fatalf("transposed rung must keep name/bitrate: %+v", vert[1])
	}
	// Square follows the landscape branch (fits landscape boxes exactly).
	sq := SelectABRLadderForSource(1080, 1080)
	if len(sq) != 5 || sq[0].Width != 426 {
		t.Fatalf("square ladder wrong: %+v", sq)
	}
	// Tiny vertical sources fall back to one transposed rung, like landscape.
	tiny := SelectABRLadderForSource(100, 200)
	if len(tiny) != 1 || tiny[0].Width != 240 || tiny[0].Height != 426 {
		t.Fatalf("tiny vertical fallback wrong: %+v", tiny)
	}
	// Zero dimensions keep the legacy landscape fallback.
	zero := SelectABRLadderForSource(0, 0)
	if len(zero) != 1 || zero[0].Width != 426 || zero[0].Height != 240 {
		t.Fatalf("zero fallback wrong: %+v", zero)
	}
}

func TestTransposeLadder(t *testing.T) {
	in := []RenditionProfile{{Name: "360p", Width: 640, Height: 360, VideoBitrate: "800k"}}
	out := TransposeLadder(in)
	if out[0].Width != 360 || out[0].Height != 640 || out[0].Name != "360p" || out[0].VideoBitrate != "800k" {
		t.Fatalf("transpose wrong: %+v", out[0])
	}
	if in[0].Width != 640 || in[0].Height != 360 {
		t.Fatal("transpose must not mutate its input")
	}
}

func TestTranscodeIntakeRejectsInvalidShorts(t *testing.T) {
	t.Setenv("REQUIRE_PROBED_SOURCE", "false")
	srv := newHealthServer("media-worker", NewMediaWorker(t.TempDir()))

	// Landscape source filed as shorts is rejected, not transcoded.
	rec := postIntake(t, srv, `{"jobId":"job-shorts-1","videoId":"vid-shorts-1","sourceR2Key":"raw/u/v.mp4",`+
		`"videoType":"shorts","width":1920,"height":1080,"durationSeconds":42,"videoCodec":"h264","hasVideoStream":true}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for landscape shorts, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "vertical") {
		t.Fatalf("rejection should explain vertical rule: %s", rec.Body.String())
	}

	// Overlong vertical source is rejected.
	rec = postIntake(t, srv, `{"jobId":"job-shorts-2","videoId":"vid-shorts-2","sourceR2Key":"raw/u/v.mp4",`+
		`"videoType":"shorts","width":1080,"height":1920,"durationSeconds":200,"videoCodec":"h264","hasVideoStream":true}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for overlong shorts, got %d: %s", rec.Code, rec.Body.String())
	}

	// Valid vertical shorts still transcodes (legacy assumed path).
	rec = postIntake(t, srv, `{"jobId":"job-shorts-3","videoId":"vid-shorts-3","sourceR2Key":"raw/u/v.mp4",`+
		`"videoType":"shorts","width":1080,"height":1920,"durationSeconds":42,"videoCodec":"h264","hasVideoStream":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid shorts, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBuildStoryboardArgsAspectSafe(t *testing.T) {
	args := BuildStoryboardArgs("in.mp4", "tile-%03d.jpg")
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "force_original_aspect_ratio=decrease") || !strings.Contains(joined, "pad=160:90") {
		t.Fatalf("storyboard must letterbox instead of stretching: %s", joined)
	}
}
