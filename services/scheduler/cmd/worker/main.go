package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type OutboxEvent struct {
	ID          string
	Topic       string
	PayloadJSON string
	Published   bool
	CreatedAt   time.Time
}

type ScheduledVideo struct {
	ID         string
	Visibility string // scheduled, public
	PublishAt  time.Time
}

type UploadSession struct {
	UploadID  string
	Status    string // pending, completed, expired
	ExpiresAt time.Time
}

type SettlementPeriod struct {
	PeriodID string
	Status   string // open, closed
}

type SchedulerEngine struct {
	mu          sync.Mutex
	Outbox      []*OutboxEvent
	Videos      []*ScheduledVideo
	Uploads     []*UploadSession
	Settlements map[string]*SettlementPeriod
	RelayedIDs  []string
}

func NewSchedulerEngine() *SchedulerEngine {
	return &SchedulerEngine{
		Outbox:      make([]*OutboxEvent, 0),
		Videos:      make([]*ScheduledVideo, 0),
		Uploads:     make([]*UploadSession, 0),
		Settlements: make(map[string]*SettlementPeriod),
		RelayedIDs:  make([]string, 0),
	}
}

// RelayOutboxEvents polls unpublished transactional outbox events and publishes them.
func (s *SchedulerEngine) RelayOutboxEvents() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	relayed := 0
	for _, ev := range s.Outbox {
		if !ev.Published {
			ev.Published = true
			s.RelayedIDs = append(s.RelayedIDs, ev.ID)
			relayed++
		}
	}
	return relayed
}

// PublishDueVideos transitions scheduled videos to public when PublishAt <= now.
func (s *SchedulerEngine) PublishDueVideos(now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	published := 0
	for _, v := range s.Videos {
		if v.Visibility == "scheduled" && !now.Before(v.PublishAt) {
			v.Visibility = "public"
			published++
		}
	}
	return published
}

// CleanupExpiredUploads marks stale multipart upload sessions as expired.
func (s *SchedulerEngine) CleanupExpiredUploads(now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleaned := 0
	for _, u := range s.Uploads {
		if u.Status == "pending" && now.After(u.ExpiresAt) {
			u.Status = "expired"
			cleaned++
		}
	}
	return cleaned
}

// CloseSettlementPeriod locks an open monthly financial period for payout calculation.
func (s *SchedulerEngine) CloseSettlementPeriod(periodID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, exists := s.Settlements[periodID]
	if !exists || p.Status == "closed" {
		return false
	}
	p.Status = "closed"
	return true
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.Info("Starting Distributed Scheduler & Transactional Outbox Relay")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	eng := NewSchedulerEngine()

	health := newHealthServer("scheduler", eng)
	go func() {
		if err := health.listenAndServe(ctx); err != nil {
			slog.Error("Health server failure", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down Scheduler gracefully")
}

// healthServer exposes liveness, readiness and Prometheus metrics so the
// worker is observable behind Docker/Kubernetes health checks.
type healthServer struct {
	name  string
	start time.Time
	eng   *SchedulerEngine
}

func newHealthServer(name string, eng *SchedulerEngine) *healthServer {
	return &healthServer{name: name, start: time.Now(), eng: eng}
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
