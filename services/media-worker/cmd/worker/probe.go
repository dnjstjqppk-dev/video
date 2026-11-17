package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// FFprobeBin resolves the ffprobe executable (P0-2 real media probing).
func FFprobeBin() string {
	if v := os.Getenv("FFPROBE_BIN"); v != "" {
		return v
	}
	return "ffprobe"
}

// FFmpegBin resolves the ffmpeg executable.
func FFmpegBin() string {
	if v := os.Getenv("FFMPEG_BIN"); v != "" {
		return v
	}
	return "ffmpeg"
}

type ffprobeStream struct {
	CodecName    string `json:"codec_name"`
	CodecType    string `json:"codec_type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	AvgFrameRate string `json:"avg_frame_rate"`
	Duration     string `json:"duration"`
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func parseSeconds(raw string) float64 {
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return 0
	}
	return v
}

// ParseFFprobeJSON converts ffprobe -print_format json output into probe
// metadata. Anything that is not a decodable video stream fails: the worker
// never assumes dimensions, duration, or codecs.
func ParseFFprobeJSON(data []byte) (ProbeMetadata, error) {
	var out ffprobeOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return ProbeMetadata{}, fmt.Errorf("decode ffprobe output: %w", err)
	}
	var meta ProbeMetadata
	for _, s := range out.Streams {
		switch s.CodecType {
		case "video":
			if !meta.HasVideoStream {
				meta.HasVideoStream = true
				meta.Width = s.Width
				meta.Height = s.Height
				meta.VideoCodec = s.CodecName
				meta.DurationSeconds = parseSeconds(s.Duration)
			}
		case "audio":
			if meta.AudioCodec == "" {
				meta.AudioCodec = s.CodecName
			}
		}
	}
	if meta.DurationSeconds <= 0 {
		meta.DurationSeconds = parseSeconds(out.Format.Duration)
	}
	if err := ValidateFFprobe(meta); err != nil {
		return ProbeMetadata{}, err
	}
	return meta, nil
}

// RunFFprobe executes ffprobe against a local media file.
func RunFFprobe(ctx context.Context, inputPath string) (ProbeMetadata, error) {
	bin := FFprobeBin()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin,
		"-v", "error", "-print_format", "json", "-show_format", "-show_streams", inputPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ProbeMetadata{}, fmt.Errorf("ffprobe timed out on %s", inputPath)
		}
		return ProbeMetadata{}, fmt.Errorf("ffprobe failed: %v: %s", err, truncate(stderr.String(), 500))
	}
	return ParseFFprobeJSON(stdout.Bytes())
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
