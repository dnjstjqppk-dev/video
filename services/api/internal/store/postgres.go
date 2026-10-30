package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgresStore manages persistent relational storage on PostgreSQL 18.
type PostgresStore struct {
	DB *sql.DB
}

// NewPostgresStore creates a connection pool to PostgreSQL.
func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Production connection pool tuning
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresStore{DB: db}, nil
}

// InitSchema ensures the essential production tables exist.
func (p *PostgresStore) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(64) PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			username VARCHAR(64) NOT NULL UNIQUE,
			role VARCHAR(64) NOT NULL DEFAULT 'registered_user',
			subscription_tier VARCHAR(32) NOT NULL DEFAULT 'free',
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			display_name VARCHAR(128) NOT NULL DEFAULT '',
			bio TEXT NOT NULL DEFAULT '',
			avatar_url TEXT NOT NULL DEFAULT '',
			banner_url TEXT NOT NULL DEFAULT '',
			two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
			two_factor_secret VARCHAR(255) NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS channels (
			id VARCHAR(64) PRIMARY KEY,
			owner_id VARCHAR(64) NOT NULL,
			handle VARCHAR(64) NOT NULL UNIQUE,
			name VARCHAR(128) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			avatar_url TEXT NOT NULL DEFAULT '',
			banner_url TEXT NOT NULL DEFAULT '',
			subscriber_count BIGINT NOT NULL DEFAULT 0,
			video_count INT NOT NULL DEFAULT 0,
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			monetization_enabled BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS videos (
			id VARCHAR(64) PRIMARY KEY,
			channel_id VARCHAR(64) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			video_type VARCHAR(32) NOT NULL DEFAULT 'standard',
			visibility VARCHAR(32) NOT NULL DEFAULT 'public',
			status VARCHAR(32) NOT NULL DEFAULT 'READY',
			duration_seconds INT NOT NULL DEFAULT 0,
			thumbnail_url TEXT NOT NULL DEFAULT '',
			hls_master_url TEXT NOT NULL DEFAULT '',
			view_count BIGINT NOT NULL DEFAULT 0,
			like_count BIGINT NOT NULL DEFAULT 0,
			dislike_count BIGINT NOT NULL DEFAULT 0,
			comment_count BIGINT NOT NULL DEFAULT 0,
			category VARCHAR(64) NOT NULL DEFAULT 'General',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS comments (
			id VARCHAR(64) PRIMARY KEY,
			video_id VARCHAR(64) NOT NULL,
			user_id VARCHAR(64) NOT NULL,
			display_name VARCHAR(128) NOT NULL,
			avatar_url TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL,
			parent_id VARCHAR(64),
			like_count INT NOT NULL DEFAULT 0,
			is_hearted BOOLEAN NOT NULL DEFAULT FALSE,
			is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS payment_transactions (
			id VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(64) NOT NULL,
			provider VARCHAR(32) NOT NULL,
			idempotency_key VARCHAR(128) NOT NULL,
			tx_type VARCHAR(64) NOT NULL,
			amount_cents BIGINT NOT NULL,
			currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
			status VARCHAR(32) NOT NULL,
			checkout_url TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS ledger_entries (
			id VARCHAR(64) PRIMARY KEY,
			transaction_id VARCHAR(64) NOT NULL,
			account_type VARCHAR(64) NOT NULL,
			direction VARCHAR(16) NOT NULL,
			amount_cents BIGINT NOT NULL,
			currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS live_streams (
			id VARCHAR(64) PRIMARY KEY,
			channel_id VARCHAR(64) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL DEFAULT 'idle',
			stream_key VARCHAR(128) NOT NULL,
			rtmp_ingest_url TEXT NOT NULL DEFAULT '',
			hls_playback_url TEXT NOT NULL DEFAULT '',
			current_viewers INT NOT NULL DEFAULT 0,
			peak_viewers INT NOT NULL DEFAULT 0,
			slow_mode_sec INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS copyright_claims (
			id VARCHAR(64) PRIMARY KEY,
			claimed_video_id VARCHAR(64) NOT NULL,
			owner_name VARCHAR(128) NOT NULL,
			asset_title VARCHAR(255) NOT NULL,
			claim_type VARCHAR(32) NOT NULL DEFAULT 'AUDIO',
			match_confidence_pct INT NOT NULL DEFAULT 0,
			policy VARCHAR(32) NOT NULL DEFAULT 'MONETIZE',
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}

	for _, q := range queries {
		if _, err := p.DB.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("failed schema initialization query: %w", err)
		}
	}
	return nil
}

// SaveUser inserts or updates a user in PostgreSQL.
func (p *PostgresStore) SaveUser(ctx context.Context, u *User) error {
	query := `INSERT INTO users (
		id, email, username, role, subscription_tier, is_verified, is_active,
		display_name, bio, avatar_url, banner_url, two_factor_enabled, two_factor_secret, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT (id) DO UPDATE SET
		email = EXCLUDED.email,
		username = EXCLUDED.username,
		role = EXCLUDED.role,
		subscription_tier = EXCLUDED.subscription_tier,
		is_verified = EXCLUDED.is_verified,
		is_active = EXCLUDED.is_active,
		display_name = EXCLUDED.display_name,
		bio = EXCLUDED.bio,
		avatar_url = EXCLUDED.avatar_url,
		banner_url = EXCLUDED.banner_url,
		two_factor_enabled = EXCLUDED.two_factor_enabled,
		two_factor_secret = EXCLUDED.two_factor_secret;`

	_, err := p.DB.ExecContext(ctx, query,
		u.ID, u.Email, u.Username, u.Role, u.SubscriptionTier, u.IsVerified, u.IsActive,
		u.DisplayName, u.Bio, u.AvatarURL, u.BannerURL, u.MFAEnabled, u.TOTPSecret, u.CreatedAt,
	)
	return err
}

// SaveVideo inserts or updates a video in PostgreSQL.
func (p *PostgresStore) SaveVideo(ctx context.Context, v *Video) error {
	query := `INSERT INTO videos (
		id, channel_id, title, description, video_type, visibility, status,
		duration_seconds, thumbnail_url, hls_master_url, view_count, like_count,
		dislike_count, comment_count, category, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (id) DO UPDATE SET
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		visibility = EXCLUDED.visibility,
		status = EXCLUDED.status,
		duration_seconds = EXCLUDED.duration_seconds,
		thumbnail_url = EXCLUDED.thumbnail_url,
		hls_master_url = EXCLUDED.hls_master_url,
		view_count = EXCLUDED.view_count,
		like_count = EXCLUDED.like_count,
		dislike_count = EXCLUDED.dislike_count,
		comment_count = EXCLUDED.comment_count;`

	_, err := p.DB.ExecContext(ctx, query,
		v.ID, v.ChannelID, v.Title, v.Description, v.VideoType, v.Visibility, v.Status,
		v.DurationSeconds, v.ThumbnailURL, v.HlsMasterURL, v.ViewCount, v.LikeCount,
		v.DislikeCount, v.CommentCount, v.Category, v.CreatedAt,
	)
	return err
}

// SaveTransaction persists a payment transaction and its ledger entries atomically.
func (p *PostgresStore) SaveTransaction(ctx context.Context, tx *PaymentTransaction) error {
	dbTx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTx.Rollback()

	query := `INSERT INTO payment_transactions (
		id, user_id, provider, idempotency_key, tx_type, amount_cents, currency, status, checkout_url, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT (id) DO UPDATE SET
		status = EXCLUDED.status,
		checkout_url = EXCLUDED.checkout_url;`

	_, err = dbTx.ExecContext(ctx, query,
		tx.ID, tx.UserID, tx.Provider, tx.IdempotencyKey, tx.Type, tx.AmountCents,
		tx.Currency, tx.Status, tx.CheckoutURL, tx.CreatedAt,
	)
	if err != nil {
		return err
	}

	for _, entry := range tx.LedgerEntries {
		entryQuery := `INSERT INTO ledger_entries (
			id, transaction_id, account_type, direction, amount_cents, currency, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING;`

		_, err = dbTx.ExecContext(ctx, entryQuery,
			entry.ID, entry.TransactionID, entry.AccountType, entry.Direction,
			entry.AmountCents, entry.Currency, entry.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}

// Close closes the database connection pool.
func (p *PostgresStore) Close() error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}

// SanitizeDatabaseURL removes passwords for safe logging.
func SanitizeDatabaseURL(raw string) string {
	if !strings.Contains(raw, "@") {
		return raw
	}
	parts := strings.SplitN(raw, "@", 2)
	prefix := parts[0]
	if idx := strings.Index(prefix, ":"); idx != -1 {
		if subIdx := strings.LastIndex(prefix, ":"); subIdx > idx {
			prefix = prefix[:subIdx] + ":****"
		}
	}
	return prefix + "@" + parts[1]
}
