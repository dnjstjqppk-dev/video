package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Hydrate overlays persisted rows onto the memory store after boot. Database
// rows win over same-ID seeds, so runtime-created entities survive restarts
// while fixtures still fill IDs the database never saw.
func (s *MemoryStore) Hydrate(ctx context.Context) error {
	if s.DB == nil {
		return nil
	}
	users, err := s.DB.LoadUsers(ctx)
	if err != nil {
		return fmt.Errorf("hydrate users: %w", err)
	}
	for _, u := range users {
		s.Users[u.ID] = u
	}
	channels, err := s.DB.LoadChannels(ctx)
	if err != nil {
		return fmt.Errorf("hydrate channels: %w", err)
	}
	for _, c := range channels {
		s.Channels[c.ID] = c
	}
	videos, err := s.DB.LoadVideos(ctx)
	if err != nil {
		return fmt.Errorf("hydrate videos: %w", err)
	}
	for _, v := range videos {
		s.Videos[v.ID] = v
	}
	series, err := s.DB.LoadSeries(ctx)
	if err != nil {
		return fmt.Errorf("hydrate series: %w", err)
	}
	for _, sr := range series {
		s.Series[sr.ID] = sr
	}
	txs, err := s.DB.LoadTransactions(ctx)
	if err != nil {
		return fmt.Errorf("hydrate transactions: %w", err)
	}
	for _, tx := range txs {
		s.Transactions[tx.ID] = tx
	}
	jobs, err := s.DB.LoadMediaJobs(ctx)
	if err != nil {
		return fmt.Errorf("hydrate media jobs: %w", err)
	}
	for _, job := range jobs {
		s.MediaJobs[job.ID] = job
	}
	buckets, err := s.DB.LoadHourlyViews(ctx, time.Now().Add(-48*time.Hour))
	if err != nil {
		return fmt.Errorf("hydrate hourly views: %w", err)
	}
	for _, b := range buckets {
		s.RecordHourlyView(b.ChannelID, b.Hour, b.Views)
	}
	slog.Info("Memory store hydrated from PostgreSQL",
		"users", len(users), "channels", len(channels),
		"videos", len(videos), "series", len(series),
		"transactions", len(txs), "jobs", len(jobs),
		"hourly_buckets", len(buckets))
	return nil
}
