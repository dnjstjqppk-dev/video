package store

import (
	"context"
	"log/slog"
	"time"
)

// Write-through persistence: MemoryStore stays the primary serving layer for
// speed and resilience, but every mutation below is also persisted to
// PostgreSQL when a backing pool is configured. Persistence failures are
// logged, never fatal, so the API keeps serving on memory during a DB outage.

const persistTimeout = 3 * time.Second

func (s *MemoryStore) persistEnabled() bool {
	return s.DB != nil
}

// PersistVideo writes a video row through to PostgreSQL (best effort).
func (s *MemoryStore) PersistVideo(v *Video) {
	if !s.persistEnabled() || v == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), persistTimeout)
	defer cancel()
	if err := s.DB.SaveVideo(ctx, v); err != nil {
		slog.Warn("Video write-through failed", "video_id", v.ID, "error", err)
	}
}

// PersistUser writes a user row through to PostgreSQL (best effort).
func (s *MemoryStore) PersistUser(u *User) {
	if !s.persistEnabled() || u == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), persistTimeout)
	defer cancel()
	if err := s.DB.SaveUser(ctx, u); err != nil {
		slog.Warn("User write-through failed", "user_id", u.ID, "error", err)
	}
}

// PersistTransaction writes a payment transaction and its ledger entries
// through to PostgreSQL (best effort). The in-memory copy remains the
// source of truth for the request lifecycle.
func (s *MemoryStore) PersistTransaction(tx *PaymentTransaction) {
	if !s.persistEnabled() || tx == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), persistTimeout)
	defer cancel()
	if err := s.DB.SaveTransaction(ctx, tx); err != nil {
		slog.Warn("Transaction write-through failed", "tx_id", tx.ID, "error", err)
	}
}
