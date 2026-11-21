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
// bitrate, aligned keyframes, fMP4 segments. A second scale pass forces
// even dimensions because x264 rejects odd sizes (any source aspect).
func rungArgs(inputPath, rungDir, rung string, profile RenditionProfile) []string {
	maxRate := profile.VideoBitrate
	bufSize := strings.TrimSuffix(strings.ToLower(maxRate), "k") + "k"
	// Forward slashes: ffmpeg derives the fMP4 init path from the playlist
	// path with slash semantics, so backslashes strand init.mp4 in the CWD.
	dir := filepath.ToSlash(rungDir)
	return []string{
		"-hide_banner", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2", profile.Width, profile.Height),
		"-c:v", "libx264", "-preset", "veryfast",
		"-b:v", profile.VideoBitrate, "-maxrate", maxRate, "-bufsize", bufSize,
		"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
		"-c:a", "aac", "-b:a", profile.AudioBitrate,
		"-hls_time", "4", "-hls_playlist_type", "vod", "-hls_segment_type", "fmp4",
		"-hls_segment_filename", dir + "/segment_%05d.m4s",
		dir + "/index.m3u8",
	}
}

func statSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

// DrmParams carries per-video CENC keying. A nil pointer disables
// protection and keeps the plain ffmpeg HLS path.
type DrmParams struct {
	KeyID string
	Key   string
}

// PackagerBin resolves the shaka-packager executable.
func PackagerBin() string {
	if v := os.Getenv("PACKAGER_BIN"); v != "" {
		return v
	}
	return "packager"
}

// RunTranscode executes the real ABR ladder plus poster, storyboard, and
// three auto-extracted thumbnail frames. It returns the emitted artifacts
// with local paths for the API to publish.
func RunTranscode(ctx context.Context, inputPath, outDir string, ladder []RenditionProfile, durationSec float64, drm *DrmParams) ([]TranscodeArtifact, error) {
	var artifacts []TranscodeArtifact
	addFile := func(key, path, contentType string) error {
		size := statSize(path)
		if size <= 0 {
			return fmt.Errorf("expected output missing or empty: %s", key)
		}
		artifacts = append(artifacts, TranscodeArtifact{Key: key, Path: path, ContentType: contentType, Bytes: size})
		return nil
	}

	if drm != nil {
		if err := runDrmLadder(ctx, inputPath, outDir, ladder, drm, addFile); err != nil {
			return nil, err
		}
	} else {
		if err := runClearLadder(ctx, inputPath, outDir, ladder, addFile); err != nil {
			return nil, err
		}
	}

	if err := collectStills(ctx, inputPath, outDir, durationSec, addFile); err != nil {
		return nil, err
	}
	return artifacts, nil
}

// runClearLadder encodes every rendition straight to HLS with ffmpeg.
func runClearLadder(ctx context.Context, inputPath, outDir string, ladder []RenditionProfile, addFile func(key, path, contentType string) error) error {
	for _, r := range ladder {
		rungDir := filepath.Join(outDir, r.Name)
		if err := os.MkdirAll(rungDir, 0o755); err != nil {
			return err
		}
		if err := runFFmpeg(ctx, rungArgs(inputPath, rungDir, r.Name, r)); err != nil {
			return fmt.Errorf("rung %s: %w", r.Name, err)
		}
		entries, err := os.ReadDir(rungDir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			name := e.Name()
			ct := "video/mp4"
			if strings.HasSuffix(name, ".m3u8") {
				ct = "application/vnd.apple.mpegurl"
			}
			if err := addFile(r.Name+"/"+name, filepath.Join(rungDir, name), ct); err != nil {
				return err
			}
		}
	}
	return nil
}

// runDrmLadder encodes renditions to mp4, then packages them into CENC cenc
// HLS with shaka-packager using the per-video raw key.
func runDrmLadder(ctx context.Context, inputPath, outDir string, ladder []RenditionProfile, drm *DrmParams, addFile func(key, path, contentType string) error) error {
	mp4dir := filepath.Join(outDir, "mp4tmp")
	if err := os.MkdirAll(mp4dir, 0o755); err != nil {
		return err
	}
	defer os.RemoveAll(mp4dir)
	mp4s := make(map[string]string, len(ladder))
	for i, r := range ladder {
		mp4Path := filepath.Join(mp4dir, r.Name+".mp4")
		if err := runFFmpeg(ctx, rungMP4Args(inputPath, mp4Path, r, i == 0)); err != nil {
			return fmt.Errorf("rung %s: %w", r.Name, err)
		}
		mp4s[r.Name] = mp4Path
	}
	if err := runPackager(ctx, outDir, ladder, mp4s, drm); err != nil {
		return fmt.Errorf("packager: %w", err)
	}
	collect := func(dir, prefix string) error {
		entries, err := os.ReadDir(filepath.Join(outDir, dir))
		if err != nil {
			return err
		}
		for _, e := range entries {
			name := e.Name()
			ct := "video/mp4"
			if strings.HasSuffix(name, ".m3u8") {
				ct = "application/vnd.apple.mpegurl"
			}
			if err := addFile(prefix+"/"+name, filepath.Join(outDir, dir, name), ct); err != nil {
				return err
			}
		}
		return nil
	}
	if err := addFile("master.m3u8", filepath.Join(outDir, "master.m3u8"), "application/vnd.apple.mpegurl"); err != nil {
		return err
	}
	if err := collect("audio", "audio"); err != nil {
		return err
	}
	for _, r := range ladder {
		if err := collect(r.Name, r.Name); err != nil {
			return err
		}
	}
	return nil
}

// rungMP4Args encodes one rendition to a plain mp4 (audio only when asked).
func rungMP4Args(inputPath, outPath string, profile RenditionProfile, withAudio bool) []string {
	args := []string{
		"-hide_banner", "-y", "-i", inputPath,
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2", profile.Width, profile.Height),
		"-c:v", "libx264", "-preset", "veryfast",
		"-b:v", profile.VideoBitrate, "-maxrate", profile.VideoBitrate,
		"-bufsize", strings.TrimSuffix(strings.ToLower(profile.VideoBitrate), "k") + "k",
		"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
	}
	if withAudio {
		args = append(args, "-c:a", "aac", "-b:a", profile.AudioBitrate)
	} else {
		args = append(args, "-an")
	}
	return append(args, "-movflags", "+faststart", outPath)
}

// packagerArgs builds one shaka-packager run: per-rung video, one shared
// audio group, CENC cenc with raw keys, HLS master + variants. cenc (not
// cbcs) because Chrome ClearKey/Widevine cannot play cbcs; FairPlay dual
// output is a separate follow-up.
func packagerArgs(mp4s map[string]string, audioMP4 string, ladder []RenditionProfile, drm *DrmParams) []string {
	args := []string{
		fmt.Sprintf("in=%s,stream=audio,init_segment=audio/init.mp4,segment_template=audio/segment_$Number%%05d$.m4s,playlist_name=audio/index.m3u8,hls_group_id=audio,hls_name=ENGLISH", audioMP4),
	}
	for _, r := range ladder {
		args = append(args, fmt.Sprintf("in=%s,stream=video,init_segment=%s/init.mp4,segment_template=%s/segment_$Number%%05d$.m4s,playlist_name=%s/index.m3u8", mp4s[r.Name], r.Name, r.Name, r.Name))
	}
	return append(args,
		"--enable_raw_key_encryption",
		"--keys", "label=:key_id="+drm.KeyID+":key="+drm.Key,
		"--protection_systems", "Widevine",
		"--protection_scheme", "cenc",
		"--hls_master_playlist_output", "master.m3u8",
	)
}

func runPackager(ctx context.Context, outDir string, ladder []RenditionProfile, mp4s map[string]string, drm *DrmParams) error {
	if len(ladder) == 0 {
		return fmt.Errorf("empty ladder")
	}
	ctx, cancel := context.WithTimeout(ctx, ffmpegTimeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, PackagerBin(), packagerArgs(mp4s, mp4s[ladder[0].Name], ladder, drm)...)
	cmd.Dir = outDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("packager failed: %v: %s", err, truncate(stderr.String(), 800))
	}
	return nil
}

// collectStills emits poster, auto frames, and storyboard stills.
func collectStills(ctx context.Context, inputPath, outDir string, durationSec float64, addFile func(key, path, contentType string) error) error {
	posterPath := filepath.Join(outDir, "poster.jpg")
	if err := runFFmpeg(ctx, BuildThumbnailArgs(inputPath, posterPath)); err != nil {
		return fmt.Errorf("poster: %w", err)
	}
	if err := addFile("poster.jpg", posterPath, "image/jpeg"); err != nil {
		return err
	}

	base := durationSec
	if base <= 0 {
		base = 60
	}
	for i, frac := range []float64{0.25, 0.5, 0.75} {
		offset := base * frac
		if offset < 1 {
			offset = 1
		}
		name := fmt.Sprintf("auto_%d.jpg", i+1)
		framePath := filepath.Join(outDir, name)
		if err := runFFmpeg(ctx, BuildFrameArgs(inputPath, framePath, offset)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if err := addFile(name, framePath, "image/jpeg"); err != nil {
			return err
		}
	}

	storyDir := filepath.Join(outDir, "storyboard")
	if err := os.MkdirAll(storyDir, 0o755); err != nil {
		return err
	}
	pattern := filepath.Join(storyDir, "thumb_%03d.jpg")
	if err := runFFmpeg(ctx, BuildStoryboardArgs(inputPath, pattern)); err != nil {
		return fmt.Errorf("storyboard: %w", err)
	}
	entries, err := os.ReadDir(storyDir)
	if err != nil {
		return err
	}
	var thumbs []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jpg") {
			thumbs = append(thumbs, e.Name())
			if err := addFile("storyboard/"+e.Name(), filepath.Join(storyDir, e.Name()), "image/jpeg"); err != nil {
				return err
			}
		}
	}
	vttPath := filepath.Join(storyDir, "storyboard.vtt")
	if err := os.WriteFile(vttPath, []byte(GenerateFileCueVTT(thumbs, 10)), 0o644); err != nil {
		return err
	}
	if err := addFile("storyboard/storyboard.vtt", vttPath, "text/vtt"); err != nil {
		return err
	}

	return nil
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
