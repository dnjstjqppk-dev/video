package main

import (
	"context"
	"log/slog"
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
	ID          string
	Visibility  string // scheduled, public
	PublishAt   time.Time
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
	_ = eng

	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
	}
}
