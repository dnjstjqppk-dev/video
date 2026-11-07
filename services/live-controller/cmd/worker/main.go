package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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

	health := newHealthServer("live-controller", ctrl)
	go func() {
		if err := health.listenAndServe(ctx); err != nil {
			slog.Error("Health server failure", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down Live Controller gracefully")
}

// healthServer exposes liveness, readiness and Prometheus metrics so the
// worker is observable behind Docker/Kubernetes health checks.
type healthServer struct {
	name  string
	start time.Time
	ctrl  *LiveController
}

func newHealthServer(name string, ctrl *LiveController) *healthServer {
	return &healthServer{name: name, start: time.Now(), ctrl: ctrl}
}

func (h *healthServer) listenAndServe(ctx context.Context) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"UP","worker":%q,"uptime_seconds":%d}`, h.name, int64(time.Since(h.start).Seconds()))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"READY","worker":%q}`, h.name)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintf(w, "# HELP gvp_worker_uptime_seconds Worker uptime in seconds\n# TYPE gvp_worker_uptime_seconds gauge\ngvp_worker_uptime_seconds{worker=%q} %d\n", h.name, int64(time.Since(h.start).Seconds()))
	})
	h.registerIntakeRoutes(mux)
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	slog.Info("Worker health server listening", "worker", h.name, "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
