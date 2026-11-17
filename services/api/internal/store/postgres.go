package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// PostgresStore manages persistent relational storage on PostgreSQL.
//
// P0-1: the API never creates schema. Migrations in db/migrations are an
// operator concern (CI/compose/psql); the API only verifies the schema it
// needs (VerifySchema) and persists entities the migrations define.
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
	db.SetConnMaxLifetime(5 * time.Second)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresStore{DB: db}, nil
}

// requiredTables are the migration tables the persistence layer writes.
var requiredTables = []string{
	"roles", "users", "user_profiles", "channels", "videos",
	"payment_transactions", "ledger_accounts", "ledger_entries",
}

// VerifySchema checks that the migrated schema (db/migrations, incl. 00010)
// is present. It creates nothing; operators apply migrations out of band.
func (p *PostgresStore) VerifySchema(ctx context.Context) error {
	var missing []string
	for _, table := range requiredTables {
		var exists bool
		err := p.DB.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`,
			table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("schema check failed for %s: %w", table, err)
		}
		if !exists {
			missing = append(missing, "table "+table)
		}
	}
	// 00010 marker column.
	var colExists bool
	err := p.DB.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'payment_transactions' AND column_name = 'tx_type')`,
	).Scan(&colExists)
	if err != nil {
		return fmt.Errorf("schema check failed for payment_transactions.tx_type: %w", err)
	}
	if !colExists {
		missing = append(missing, "column payment_transactions.tx_type (apply db/migrations/00010)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("postgres schema incomplete, run db/migrations against DATABASE_URL: %s", strings.Join(missing, ", "))
	}
	return nil
}

// validUUID reports whether s parses as a UUID. Migration PKs are UUID-typed;
// legacy prefixed memory IDs are skipped by persistence, never coerced.
func validUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// SaveUser inserts or updates a user (users + user_profiles split) in PostgreSQL.
func (p *PostgresStore) SaveUser(ctx context.Context, u *User) error {
	if u == nil {
		return fmt.Errorf("nil user")
	}
	if !validUUID(u.ID) {
		return fmt.Errorf("user id %q is not a UUID", u.ID)
	}
	dbTx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTx.Rollback()

	_, err = dbTx.ExecContext(ctx, `INSERT INTO users (
		id, email, username, role, subscription_tier, is_verified, is_active,
		two_factor_enabled, two_factor_secret, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	ON CONFLICT (id) DO UPDATE SET
		email = EXCLUDED.email,
		username = EXCLUDED.username,
		role = EXCLUDED.role,
		subscription_tier = EXCLUDED.subscription_tier,
		is_verified = EXCLUDED.is_verified,
		is_active = EXCLUDED.is_active,
		two_factor_enabled = EXCLUDED.two_factor_enabled,
		two_factor_secret = EXCLUDED.two_factor_secret,
		updated_at = NOW();`,
		u.ID, u.Email, u.Username, u.Role, u.SubscriptionTier, u.IsVerified, u.IsActive,
		u.MFAEnabled, nullIfEmpty(u.TOTPSecret), u.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert users: %w", err)
	}
	_, err = dbTx.ExecContext(ctx, `INSERT INTO user_profiles (
		user_id, display_name, bio, avatar_url, banner_url
	) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (user_id) DO UPDATE SET
		display_name = EXCLUDED.display_name,
		bio = EXCLUDED.bio,
		avatar_url = EXCLUDED.avatar_url,
		banner_url = EXCLUDED.banner_url,
		updated_at = NOW();`,
		u.ID, displayOr(u.DisplayName, u.Username), nullIfEmpty(u.Bio),
		nullIfEmpty(u.AvatarURL), nullIfEmpty(u.BannerURL),
	)
	if err != nil {
		return fmt.Errorf("upsert user_profiles: %w", err)
	}
	return dbTx.Commit()
}

// SaveChannel inserts or updates a channel in PostgreSQL.
func (p *PostgresStore) SaveChannel(ctx context.Context, c *Channel) error {
	if c == nil {
		return fmt.Errorf("nil channel")
	}
	if !validUUID(c.ID) {
		return fmt.Errorf("channel id %q is not a UUID", c.ID)
	}
	if !validUUID(c.OwnerID) {
		return fmt.Errorf("channel owner id %q is not a UUID", c.OwnerID)
	}
	_, err := p.DB.ExecContext(ctx, `INSERT INTO channels (
		id, owner_id, handle, name, description, avatar_url, banner_url,
		subscriber_count, video_count, is_verified, monetization_enabled,
		created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
	ON CONFLICT (id) DO UPDATE SET
		owner_id = EXCLUDED.owner_id,
		handle = EXCLUDED.handle,
		name = EXCLUDED.name,
		description = EXCLUDED.description,
		avatar_url = EXCLUDED.avatar_url,
		banner_url = EXCLUDED.banner_url,
		subscriber_count = EXCLUDED.subscriber_count,
		video_count = EXCLUDED.video_count,
		is_verified = EXCLUDED.is_verified,
		monetization_enabled = EXCLUDED.monetization_enabled,
		updated_at = NOW();`,
		c.ID, c.OwnerID, c.Handle, c.Name, nullIfEmpty(c.Description),
		nullIfEmpty(c.AvatarURL), nullIfEmpty(c.BannerURL),
		c.SubscriberCount, c.VideoCount, c.IsVerified, c.MonetizationEnabled, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert channels: %w", err)
	}
	return nil
}

// SaveVideo inserts or updates a video in PostgreSQL.
func (p *PostgresStore) SaveVideo(ctx context.Context, v *Video) error {
	if v == nil {
		return fmt.Errorf("nil video")
	}
	if !validUUID(v.ID) {
		return fmt.Errorf("video id %q is not a UUID", v.ID)
	}
	if !validUUID(v.ChannelID) {
		return fmt.Errorf("video channel id %q is not a UUID", v.ChannelID)
	}
	tags := v.Tags
	if tags == nil {
		tags = []string{}
	}
	_, err := p.DB.ExecContext(ctx, `INSERT INTO videos (
		id, channel_id, title, description, video_type, visibility, status,
		duration_seconds, thumbnail_url, hls_master_url, view_count, like_count,
		dislike_count, comment_count, tags, category, published_at, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NOW())
	ON CONFLICT (id) DO UPDATE SET
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		video_type = EXCLUDED.video_type,
		visibility = EXCLUDED.visibility,
		status = EXCLUDED.status,
		duration_seconds = EXCLUDED.duration_seconds,
		thumbnail_url = EXCLUDED.thumbnail_url,
		hls_master_url = EXCLUDED.hls_master_url,
		view_count = EXCLUDED.view_count,
		like_count = EXCLUDED.like_count,
		dislike_count = EXCLUDED.dislike_count,
		comment_count = EXCLUDED.comment_count,
		tags = EXCLUDED.tags,
		category = EXCLUDED.category,
		published_at = EXCLUDED.published_at,
		updated_at = NOW();`,
		v.ID, v.ChannelID, v.Title, nullIfEmpty(v.Description), v.VideoType, v.Visibility, v.Status,
		v.DurationSeconds, nullIfEmpty(v.ThumbnailURL), nullIfEmpty(v.HlsMasterURL),
		v.ViewCount, v.LikeCount, v.DislikeCount, v.CommentCount,
		pq.Array(tags), v.Category, nullTime(v.PublishedAt), v.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert videos: %w", err)
	}
	return nil
}

// ledgerAccountNumbers maps memory account types to seeded ledger_accounts rows.
var ledgerAccountNumbers = map[string]string{
	"platform_escrow":  "1000-CASH",
	"creator_payable":  "2000-CREATOR-PAYABLE",
	"platform_revenue": "4000-AD-REVENUE",
}

var ledgerAccountTypes = map[string]string{
	"1000-CASH":            "platform_escrow",
	"2000-CREATOR-PAYABLE": "creator_payable",
	"4000-AD-REVENUE":      "platform_revenue",
	"4100-SUB-REVENUE":     "platform_revenue",
}

func ledgerAccountFor(txType, accountType string) string {
	if accountType == "platform_revenue" && (txType == "membership" || txType == "premium") {
		return "4100-SUB-REVENUE"
	}
	return ledgerAccountNumbers[accountType]
}

// SaveTransaction persists a payment transaction and its ledger entries atomically.
func (p *PostgresStore) SaveTransaction(ctx context.Context, tx *PaymentTransaction) error {
	if tx == nil {
		return fmt.Errorf("nil transaction")
	}
	if !validUUID(tx.ID) {
		return fmt.Errorf("transaction id %q is not a UUID", tx.ID)
	}
	if !validUUID(tx.UserID) {
		return fmt.Errorf("transaction user id %q is not a UUID", tx.UserID)
	}
	dbTx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTx.Rollback()

	_, err = dbTx.ExecContext(ctx, `INSERT INTO payment_transactions (
		id, user_id, provider, external_transaction_id, amount_cents, currency,
		status, idempotency_key, payment_method, tx_type, checkout_url, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
	ON CONFLICT (id) DO UPDATE SET
		status = EXCLUDED.status,
		tx_type = EXCLUDED.tx_type,
		checkout_url = EXCLUDED.checkout_url,
		updated_at = NOW();`,
		tx.ID, tx.UserID, tx.Provider, tx.ID, tx.AmountCents, tx.Currency,
		tx.Status, tx.IdempotencyKey, "unknown", nullIfEmpty(tx.Type), nullIfEmpty(tx.CheckoutURL), tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert payment_transactions: %w", err)
	}

	for _, entry := range tx.LedgerEntries {
		if !validUUID(entry.ID) {
			return fmt.Errorf("ledger entry id %q is not a UUID", entry.ID)
		}
		accountNumber := ledgerAccountFor(tx.Type, entry.AccountType)
		if accountNumber == "" {
			return fmt.Errorf("unknown ledger account type %q", entry.AccountType)
		}
		var accountID string
		err = dbTx.QueryRowContext(ctx,
			`SELECT id FROM ledger_accounts WHERE account_number = $1`, accountNumber).Scan(&accountID)
		if err != nil {
			return fmt.Errorf("resolve ledger account %s: %w", accountNumber, err)
		}
		var debit, credit int64
		if entry.Direction == "DEBIT" {
			debit = entry.AmountCents
		} else {
			credit = entry.AmountCents
		}
		_, err = dbTx.ExecContext(ctx, `INSERT INTO ledger_entries (
			id, transaction_id, account_id, debit_cents, credit_cents, currency, description, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING;`,
			entry.ID, tx.ID, accountID, debit, credit, entry.Currency,
			entry.AccountType+" "+tx.ID, entry.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert ledger_entries: %w", err)
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

func nullIfEmpty(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func displayOr(display, fallback string) string {
	if display != "" {
		return display
	}
	return fallback
}

func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}
