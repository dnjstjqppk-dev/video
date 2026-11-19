package main

import (
	"strings"
	"testing"
)

func TestBuildFrameArgs_SeeksToOffset(t *testing.T) {
	args := BuildFrameArgs("/src/in.mp4", "/out/auto_2.jpg", 188.5)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-ss 188.50") {
		t.Fatalf("frame args must seek to the offset: %v", args)
	}
	if !strings.Contains(joined, "-frames:v 1") || !strings.Contains(joined, "/out/auto_2.jpg") {
		t.Fatalf("frame args must emit one still: %v", args)
	}
	if neg := BuildFrameArgs("/src", "/out", -4); strings.Contains(strings.Join(neg, " "), "-ss -4") {
		t.Fatal("negative offsets must clamp to zero")
	}
}
