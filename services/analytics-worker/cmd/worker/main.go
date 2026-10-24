package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type TelemetryEvent struct {
	EventID         string
	EventType       string // valid_view, watch_progress, ad_impression, ad_click
	VideoID         string
	UserID          string
	UserAgent       string
	WatchSeconds    float64
	PlaybackRate    float64
	OccurredAt      time.Time
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
	_ = proc

	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
	}
}
