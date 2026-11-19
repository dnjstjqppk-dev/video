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
//
// P0-1: writes target the real migration schema (db/migrations). Parent rows
// are ensured first so FK constraints hold; legacy non-UUID fixture IDs are
// skipped with a warning instead of failing silently row by row.

const persistTimeout = 3 * time.Second

func (s *MemoryStore) persistEnabled() bool {
	return s.DB != nil
}

func persistCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), persistTimeout)
}

// ensureUserRow upserts the in-memory user row so FK children can persist.
// It returns false when there is nothing persistable (caller skips).
func (s *MemoryStore) ensureUserRow(id string) bool {
	if !s.persistEnabled() || !validUUID(id) {
		return false
	}
	u, ok := s.Users[id]
	if !ok || u == nil {
		return false
	}
	ctx, cancel := persistCtx()
	defer cancel()
	var exists bool
	if err := s.DB.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, id).Scan(&exists); err != nil {
		slog.Warn("User existence check failed", "user_id", id, "error", err)
		return false
	}
	if exists {
		return true
	}
	if err := s.DB.SaveUser(ctx, u); err != nil {
		slog.Warn("User write-through failed", "user_id", id, "error", err)
		return false
	}
	return true
}

// ensureChannelRow upserts the in-memory channel row (and its owner first).
func (s *MemoryStore) ensureChannelRow(id string) bool {
	if !s.persistEnabled() || !validUUID(id) {
		return false
	}
	c, ok := s.Channels[id]
	if !ok || c == nil {
		return false
	}
	if !s.ensureUserRow(c.OwnerID) {
		return false
	}
	ctx, cancel := persistCtx()
	defer cancel()
	var exists bool
	if err := s.DB.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM channels WHERE id = $1)`, id).Scan(&exists); err != nil {
		slog.Warn("Channel existence check failed", "channel_id", id, "error", err)
		return false
	}
	if exists {
		return true
	}
	if err := s.DB.SaveChannel(ctx, c); err != nil {
		slog.Warn("Channel write-through failed", "channel_id", id, "error", err)
		return false
	}
	return true
}

// PersistChannel writes a channel row through to PostgreSQL (best effort).
func (s *MemoryStore) PersistChannel(c *Channel) {
	if !s.persistEnabled() || c == nil {
		return
	}
	if !validUUID(c.ID) {
		slog.Warn("Channel write-through skipped: non-UUID fixture id", "channel_id", c.ID)
		return
	}
	if !s.ensureUserRow(c.OwnerID) {
		slog.Warn("Channel write-through skipped: owner row unavailable", "channel_id", c.ID)
		return
	}
	ctx, cancel := persistCtx()
	defer cancel()
	if err := s.DB.SaveChannel(ctx, c); err != nil {
		slog.Warn("Channel write-through failed", "channel_id", c.ID, "error", err)
	}
}

// PersistVideo writes a video row through to PostgreSQL (best effort).
func (s *MemoryStore) PersistVideo(v *Video) {
	if !s.persistEnabled() || v == nil {
		return
	}
	if !validUUID(v.ID) {
		slog.Warn("Video write-through skipped: non-UUID fixture id", "video_id", v.ID)
		return
	}
	if !s.ensureChannelRow(v.ChannelID) {
		slog.Warn("Video write-through skipped: channel row unavailable", "video_id", v.ID)
		return
	}
	ctx, cancel := persistCtx()
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
	if !validUUID(u.ID) {
		slog.Warn("User write-through skipped: non-UUID fixture id", "user_id", u.ID)
		return
	}
	ctx, cancel := persistCtx()
	defer cancel()
	if err := s.DB.SaveUser(ctx, u); err != nil {
		slog.Warn("User write-through failed", "user_id", u.ID, "error", err)
	}
}

// PersistHourlyView increments a channel hour bucket in PostgreSQL.
func (s *MemoryStore) PersistHourlyView(channelID string, hour time.Time, delta int64) {
	if !s.persistEnabled() || channelID == "" || delta == 0 {
		return
	}
	if !validUUID(channelID) {
		return
	}
	ctx, cancel := persistCtx()
	defer cancel()
	if err := s.DB.SaveHourlyViews(ctx, channelID, hour, delta); err != nil {
		slog.Warn("Hourly views write-through failed", "channel_id", channelID, "error", err)
	}
}

// PersistMediaJob writes a transcode job through to PostgreSQL (best effort).
func (s *MemoryStore) PersistMediaJob(job *MediaJob) {
	if !s.persistEnabled() || job == nil {
		return
	}
	if !validUUID(job.ID) {
		slog.Warn("MediaJob write-through skipped: non-UUID id", "job_id", job.ID)
		return
	}
	ctx, cancel := persistCtx()
	defer cancel()
	if err := s.DB.SaveMediaJob(ctx, job); err != nil {
		slog.Warn("MediaJob write-through failed", "job_id", job.ID, "error", err)
	}
}

// PersistTransaction writes a payment transaction and its ledger entries
// through to PostgreSQL (best effort). The in-memory copy remains the
// source of truth for the request lifecycle.
func (s *MemoryStore) PersistTransaction(tx *PaymentTransaction) {
	if !s.persistEnabled() || tx == nil {
		return
	}
	if !validUUID(tx.ID) {
		slog.Warn("Transaction write-through skipped: non-UUID id", "tx_id", tx.ID)
		return
	}
	if !s.ensureUserRow(tx.UserID) {
		slog.Warn("Transaction write-through skipped: user row unavailable", "tx_id", tx.ID)
		return
	}
	ctx, cancel := persistCtx()
	defer cancel()
	if err := s.DB.SaveTransaction(ctx, tx); err != nil {
		slog.Warn("Transaction write-through failed", "tx_id", tx.ID, "error", err)
	}
}
