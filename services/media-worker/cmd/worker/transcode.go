package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TranscodeArtifact is one emitted file: its storage key suffix, local path,
// and content type for the API to publish.
type TranscodeArtifact struct {
	Key         string `json:"key"`
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
	Bytes       int64  `json:"bytes"`
}

func ffmpegTimeout() time.Duration {
	if v := os.Getenv("FFMPEG_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 30 * time.Minute
}

func runFFmpeg(ctx context.Context, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, ffmpegTimeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, FFmpegBin(), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %v: %s", err, truncate(stderr.String(), 800))
	}
	return nil
}

// rungArgs builds one HLS variant encode: aspect-preserving scale, capped
// bitrate, aligned keyframes, fMP4 segments.
func rungArgs(inputPath, rungDir, rung string, profile RenditionProfile) []string {
	maxRate := profile.VideoBitrate
	bufSize := strings.TrimSuffix(strings.ToLower(maxRate), "k") + "k"
	return []string{
		"-hide_banner", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", profile.Width, profile.Height),
		"-c:v", "libx264", "-preset", "veryfast",
		"-b:v", profile.VideoBitrate, "-maxrate", maxRate, "-bufsize", bufSize,
		"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
		"-c:a", "aac", "-b:a", profile.AudioBitrate,
		"-hls_time", "4", "-hls_playlist_type", "vod", "-hls_segment_type", "fmp4",
		"-hls_segment_filename", filepath.Join(rungDir, "segment_%05d.m4s"),
		filepath.Join(rungDir, "index.m3u8"),
	}
}

func statSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

// RunTranscode executes the real ABR ladder plus poster and storyboard stills.
// It returns the emitted artifacts with local paths for the API to publish.
func RunTranscode(ctx context.Context, inputPath, outDir string, ladder []RenditionProfile) ([]TranscodeArtifact, error) {
	var artifacts []TranscodeArtifact
	addFile := func(key, path, contentType string) error {
		size := statSize(path)
		if size <= 0 {
			return fmt.Errorf("expected output missing or empty: %s", key)
		}
		artifacts = append(artifacts, TranscodeArtifact{Key: key, Path: path, ContentType: contentType, Bytes: size})
		return nil
	}

	for _, r := range ladder {
		rungDir := filepath.Join(outDir, r.Name)
		if err := os.MkdirAll(rungDir, 0o755); err != nil {
			return nil, err
		}
		if err := runFFmpeg(ctx, rungArgs(inputPath, rungDir, r.Name, r)); err != nil {
			return nil, fmt.Errorf("rung %s: %w", r.Name, err)
		}
		entries, err := os.ReadDir(rungDir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			name := e.Name()
			ct := "video/mp4"
			if strings.HasSuffix(name, ".m3u8") {
				ct = "application/vnd.apple.mpegurl"
			}
			if err := addFile(r.Name+"/"+name, filepath.Join(rungDir, name), ct); err != nil {
				return nil, err
			}
		}
	}

	posterPath := filepath.Join(outDir, "poster.jpg")
	if err := runFFmpeg(ctx, BuildThumbnailArgs(inputPath, posterPath)); err != nil {
		return nil, fmt.Errorf("poster: %w", err)
	}
	if err := addFile("poster.jpg", posterPath, "image/jpeg"); err != nil {
		return nil, err
	}

	storyDir := filepath.Join(outDir, "storyboard")
	if err := os.MkdirAll(storyDir, 0o755); err != nil {
		return nil, err
	}
	pattern := filepath.Join(storyDir, "thumb_%03d.jpg")
	if err := runFFmpeg(ctx, BuildStoryboardArgs(inputPath, pattern)); err != nil {
		return nil, fmt.Errorf("storyboard: %w", err)
	}
	entries, err := os.ReadDir(storyDir)
	if err != nil {
		return nil, err
	}
	var thumbs []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jpg") {
			thumbs = append(thumbs, e.Name())
			if err := addFile("storyboard/"+e.Name(), filepath.Join(storyDir, e.Name()), "image/jpeg"); err != nil {
				return nil, err
			}
		}
	}
	vttPath := filepath.Join(storyDir, "storyboard.vtt")
	if err := os.WriteFile(vttPath, []byte(GenerateFileCueVTT(thumbs, 10)), 0o644); err != nil {
		return nil, err
	}
	if err := addFile("storyboard/storyboard.vtt", vttPath, "text/vtt"); err != nil {
		return nil, err
	}

	return artifacts, nil
}

// GenerateFileCueVTT builds a WebVTT storyboard with one image file per cue.
func GenerateFileCueVTT(thumbs []string, intervalSec int) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")
	if intervalSec <= 0 {
		intervalSec = 10
	}
	for i, name := range thumbs {
		start := float64(i * intervalSec)
		end := start + float64(intervalSec)
		sb.WriteString(fmt.Sprintf("%s --> %s\n", formatVTTTime(start), formatVTTTime(end)))
		sb.WriteString(fmt.Sprintf("%s\n\n", name))
	}
	return sb.String()
}
