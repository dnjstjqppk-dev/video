package main

import (
	"strings"
	"testing"
)

const sampleFFprobeJSON = `{
  "streams": [
    {"index": 0, "codec_name": "h264", "codec_type": "video", "width": 1280, "height": 720,
     "avg_frame_rate": "60/1", "duration": "63.400000"},
    {"index": 1, "codec_name": "aac", "codec_type": "audio", "duration": "63.381000"}
  ],
  "format": {"filename": "source.mp4", "duration": "63.400000", "size": "12345678"}
}`

// P0-2: ffprobe JSON output parses into validated probe metadata.
func TestParseFFprobeJSON(t *testing.T) {
	meta, err := ParseFFprobeJSON([]byte(sampleFFprobeJSON))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if meta.Width != 1280 || meta.Height != 720 {
		t.Fatalf("bad dimensions: %+v", meta)
	}
	if meta.DurationSeconds < 63 || meta.DurationSeconds > 64 {
		t.Fatalf("bad duration: %+v", meta)
	}
	if meta.VideoCodec != "h264" || meta.AudioCodec != "aac" || !meta.HasVideoStream {
		t.Fatalf("bad codecs: %+v", meta)
	}
	if err := ValidateFFprobe(meta); err != nil {
		t.Fatalf("valid probe rejected: %v", err)
	}
}

// P0-2: audio-only or empty probe output is rejected, never assumed.
func TestParseFFprobeJSONRejectsNonVideo(t *testing.T) {
	audioOnly := `{"streams": [{"codec_name": "aac", "codec_type": "audio"}], "format": {"duration": "10"}}`
	if _, err := ParseFFprobeJSON([]byte(audioOnly)); err == nil {
		t.Fatal("audio-only source must fail probe validation")
	}
	if _, err := ParseFFprobeJSON([]byte(`{}`)); err == nil {
		t.Fatal("empty probe output must fail")
	}
	if _, err := ParseFFprobeJSON([]byte(`not json`)); err == nil {
		t.Fatal("malformed probe output must fail")
	}
}

// P0-2: source references resolve safely: allowlisted local files and https
// downloads only; bare keys and escapes are rejected.
func TestResolveSourceAllowlist(t *testing.T) {
	t.Setenv("MEDIA_SOURCE_ROOTS", t.TempDir())
	if _, _, err := ResolveSourceRef("raw/vid123/source.mp4"); err == nil {
		t.Fatal("bare R2 key must be rejected without a presigned URL")
	}
	if _, _, err := ResolveSourceRef("file:///etc/passwd"); err == nil {
		t.Fatal("file outside allowlist must be rejected")
	}
	kind, target, err := ResolveSourceRef("https://r2.test/presigned?sig=abc")
	if err != nil {
		t.Fatalf("presigned URL rejected: %v", err)
	}
	if kind != SourceRemote || !strings.HasPrefix(target, "https://") {
		t.Fatalf("bad remote resolution: %v %v", kind, target)
	}
}
