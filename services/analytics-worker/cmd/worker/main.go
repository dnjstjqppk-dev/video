package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type TelemetryEvent struct {
	EventID      string
	EventType    string // valid_view, watch_progress, ad_impression, ad_click
	VideoID      string
	UserID       string
	UserAgent    string
	WatchSeconds float64
	PlaybackRate float64
	OccurredAt   time.Time
}

type AnalyticsProcessor struct {
	mu           sync.Mutex
	SeenEventIDs map[string]bool
	BatchBuffer  []TelemetryEvent
	FlushedRows  []TelemetryEvent
	BatchSize    int
}

func NewAnalyticsProcessor(batchSize int) *AnalyticsProcessor {
	if batchSize <= 0 {
		batchSize = 100
	}
	return &AnalyticsProcessor{
		SeenEventIDs: make(map[string]bool),
		BatchBuffer:  make([]TelemetryEvent, 0, batchSize),
		FlushedRows:  make([]TelemetryEvent, 0),
		BatchSize:    batchSize,
	}
}

func IsInvalidBotTraffic(ev TelemetryEvent) bool {
	ua := strings.ToLower(ev.UserAgent)
	botSignatures := []string{"headlesschrome", "phantomjs", "selenium", "puppeteer", "curl/", "python-requests", "scrapy"}
	for _, sig := range botSignatures {
		if strings.Contains(ua, sig) {
			return true
		}
	}
	if ev.PlaybackRate > 4.0 || ev.WatchSeconds < 0 || ev.WatchSeconds > 86400 {
		return true
	}
	return false
}

func (p *AnalyticsProcessor) Ingest(ev TelemetryEvent) (accepted bool, err error) {
	if ev.EventID == "" {
		return false, errors.New("missing event_id")
	}

	validTypes := map[string]bool{
		"valid_view":     true,
		"watch_progress": true,
		"ad_impression":  true,
		"ad_click":       true,
	}
	if !validTypes[ev.EventType] {
		return false, errors.New("unsupported event_type")
	}

	if IsInvalidBotTraffic(ev) {
		return false, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Deduplication by event_id
	if p.SeenEventIDs[ev.EventID] {
		return false, nil
	}
	p.SeenEventIDs[ev.EventID] = true

	p.BatchBuffer = append(p.BatchBuffer, ev)
	if len(p.BatchBuffer) >= p.BatchSize {
		p.flushLocked()
	}
	return true, nil
}

func (p *AnalyticsProcessor) FlushToClickHouse() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.flushLocked()
}

func (p *AnalyticsProcessor) flushLocked() int {
	count := len(p.BatchBuffer)
	if count == 0 {
		return 0
	}
	p.FlushedRows = append(p.FlushedRows, p.BatchBuffer...)
	p.BatchBuffer = p.BatchBuffer[:0]
	return count
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.Info("Starting Analytics Worker & ClickHouse Batch Sink")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	proc := NewAnalyticsProcessor(500)

	health := newHealthServer("analytics-worker", proc)
	go func() {
		if err := health.listenAndServe(ctx); err != nil {
			slog.Error("Health server failure", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down Analytics Worker gracefully")
}

// healthServer exposes liveness, readiness and Prometheus metrics so the
// worker is observable behind Docker/Kubernetes health checks.
type healthServer struct {
	name  string
	start time.Time
	proc  *AnalyticsProcessor
}

func newHealthServer(name string, proc *AnalyticsProcessor) *healthServer {
	return &healthServer{name: name, start: time.Now(), proc: proc}
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
