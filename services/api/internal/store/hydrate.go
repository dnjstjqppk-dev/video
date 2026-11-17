package store

import (
	"context"
	"fmt"
	"log/slog"
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
	txs, err := s.DB.LoadTransactions(ctx)
	if err != nil {
		return fmt.Errorf("hydrate transactions: %w", err)
	}
	for _, tx := range txs {
		s.Transactions[tx.ID] = tx
	}
	slog.Info("Memory store hydrated from PostgreSQL",
		"users", len(users), "channels", len(channels),
		"videos", len(videos), "transactions", len(txs))
	return nil
}
