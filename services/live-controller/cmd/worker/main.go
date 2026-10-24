package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type LiveState string

const (
	LiveStateCreated LiveState = "CREATED"
	LiveStateLive    LiveState = "LIVE"
	LiveStateEnded   LiveState = "ENDED"
)

type StreamHealth struct {
	BitrateKbps     int
	FPS             int
	DroppedFramePct float64
	Status          string // HEALTHY, DEGRADED, UNHEALTHY
}

type LiveSession struct {
	ID             string
	ChannelID      string
	StreamKey      string
	RtmpsURL       string
	SrtURL         string
	PlaybackHLSURL string
	State          LiveState
	DvrEnabled     bool
	VodVideoID     string
	CreatedAt      time.Time
	EndedAt        *time.Time
}

type LiveController struct {
	mu       sync.RWMutex
	Sessions map[string]*LiveSession
}

func NewLiveController() *LiveController {
	return &LiveController{
		Sessions: make(map[string]*LiveSession),
	}
}

func generateStreamKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "live_sk_" + hex.EncodeToString(b)
}

func (c *LiveController) ProvisionInput(sessionID, channelID string, dvrEnabled bool) *LiveSession {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := generateStreamKey()
	sess := &LiveSession{
		ID:             sessionID,
		ChannelID:      channelID,
		StreamKey:      key,
		RtmpsURL:       "rtmps://live.cloudflare.com:443/live/",
		SrtURL:         fmt.Sprintf("srt://live.cloudflare.com:778?streamid=%s", key),
		PlaybackHLSURL: fmt.Sprintf("https://customer-gvp.cloudflarestream.com/%s/manifest/video.m3u8", sessionID),
		State:          LiveStateCreated,
		DvrEnabled:     dvrEnabled,
		CreatedAt:      time.Now(),
	}
	c.Sessions[sessionID] = sess
	return sess
}

func (c *LiveController) RotateStreamKey(sessionID string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sess, ok := c.Sessions[sessionID]
	if !ok {
		return "", errors.New("live session not found")
	}
	if sess.State == LiveStateEnded {
		return "", errors.New("cannot rotate key for ended session")
	}
	sess.StreamKey = generateStreamKey()
	sess.SrtURL = fmt.Sprintf("srt://live.cloudflare.com:778?streamid=%s", sess.StreamKey)
	return sess.StreamKey, nil
}

func (c *LiveController) TransitionState(sessionID string, target LiveState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sess, ok := c.Sessions[sessionID]
	if !ok {
		return errors.New("live session not found")
	}

	switch sess.State {
	case LiveStateCreated:
		if target != LiveStateLive && target != LiveStateEnded {
			return fmt.Errorf("invalid transition %s -> %s", sess.State, target)
		}
	case LiveStateLive:
		if target != LiveStateEnded {
			return fmt.Errorf("invalid transition %s -> %s", sess.State, target)
		}
	default:
		return fmt.Errorf("session already in terminal state %s", sess.State)
	}

	sess.State = target
	if target == LiveStateEnded {
		now := time.Now()
		sess.EndedAt = &now
		if sess.DvrEnabled {
			sess.VodVideoID = "vod_" + sess.ID
		}
	}
	return nil
}

func EvaluateHealth(bitrateKbps, fps int, droppedPct float64) StreamHealth {
	status := "HEALTHY"
	if bitrateKbps < 1500 || fps < 24 || droppedPct > 5.0 {
		status = "DEGRADED"
	}
	if bitrateKbps < 500 || fps < 10 || droppedPct > 20.0 {
		status = "UNHEALTHY"
	}
	return StreamHealth{
		BitrateKbps:     bitrateKbps,
		FPS:             fps,
		DroppedFramePct: droppedPct,
		Status:          status,
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.Info("Starting Live Stream Controller (Cloudflare Stream RTMPS/SRT)")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctrl := NewLiveController()
	_ = ctrl

	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
	}
}
