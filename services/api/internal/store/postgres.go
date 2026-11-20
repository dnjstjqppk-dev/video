package store

import (
	"context"
	"database/sql"
	"encoding/json"
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
	"payment_transactions", "ledger_accounts", "ledger_entries", "media_jobs",
	"series",
}

// VerifySchema checks that the migrated schema (db/migrations, incl. 00013)
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
	var jobColExists bool
	err = p.DB.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'media_jobs' AND column_name = 'source_ref')`,
	).Scan(&jobColExists)
	if err != nil {
		return fmt.Errorf("schema check failed for media_jobs.source_ref: %w", err)
	}
	if !jobColExists {
		missing = append(missing, "column media_jobs.source_ref (apply db/migrations/00011)")
	}
	var studioColExists bool
	err = p.DB.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'elements')`,
	).Scan(&studioColExists)
	if err != nil {
		return fmt.Errorf("schema check failed for videos.elements: %w", err)
	}
	if !studioColExists {
		missing = append(missing, "column videos.elements (apply db/migrations/00012)")
	}
	var hourlyExists bool
	err = p.DB.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'channel_hourly_views')`,
	).Scan(&hourlyExists)
	if err != nil {
		return fmt.Errorf("schema check failed for channel_hourly_views: %w", err)
	}
	if !hourlyExists {
		missing = append(missing, "table channel_hourly_views (apply db/migrations/00012)")
	}
	var drmColExists bool
	err = p.DB.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'drm_key_id')`,
	).Scan(&drmColExists)
	if err != nil {
		return fmt.Errorf("schema check failed for videos.drm_key_id: %w", err)
	}
	if !drmColExists {
		missing = append(missing, "column videos.drm_key_id (apply db/migrations/00013)")
	}
	var aiColExists bool
	err = p.DB.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'ai_review_status')`,
	).Scan(&aiColExists)
	if err != nil {
		return fmt.Errorf("schema check failed for videos.ai_review_status: %w", err)
	}
	if !aiColExists {
		missing = append(missing, "column videos.ai_review_status (apply db/migrations/00014)")
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
	if v.SeriesID != "" && !validUUID(v.SeriesID) {
		return fmt.Errorf("video series id %q is not a UUID", v.SeriesID)
	}
	tags := v.Tags
	if tags == nil {
		tags = []string{}
	}
	frames := v.AutoFrameURLs
	if frames == nil {
		frames = []string{}
	}
	elements, err := json.Marshal(v.Elements)
	if err != nil {
		return fmt.Errorf("encode video elements: %w", err)
	}
	_, err = p.DB.ExecContext(ctx, `INSERT INTO videos (
		id, channel_id, title, description, video_type, visibility, status,
		duration_seconds, thumbnail_url, custom_thumbnail_url, auto_frame_urls,
		hls_master_url, view_count, like_count,
		dislike_count, comment_count, tags, category, elements, published_at,
		scheduled_at, drm_protected, drm_key_id, drm_key_hex, ai_generated, ai_tool,
		ai_review_status, ai_review_note, series_id, season_no, episode_no, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, NOW())
	ON CONFLICT (id) DO UPDATE SET
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		video_type = EXCLUDED.video_type,
		visibility = EXCLUDED.visibility,
		status = EXCLUDED.status,
		duration_seconds = EXCLUDED.duration_seconds,
		thumbnail_url = EXCLUDED.thumbnail_url,
		custom_thumbnail_url = EXCLUDED.custom_thumbnail_url,
		auto_frame_urls = EXCLUDED.auto_frame_urls,
		hls_master_url = EXCLUDED.hls_master_url,
		view_count = EXCLUDED.view_count,
		like_count = EXCLUDED.like_count,
		dislike_count = EXCLUDED.dislike_count,
		comment_count = EXCLUDED.comment_count,
		tags = EXCLUDED.tags,
		category = EXCLUDED.category,
		elements = EXCLUDED.elements,
		published_at = EXCLUDED.published_at,
		scheduled_at = EXCLUDED.scheduled_at,
		drm_protected = EXCLUDED.drm_protected,
		drm_key_id = EXCLUDED.drm_key_id,
		drm_key_hex = EXCLUDED.drm_key_hex,
		ai_generated = EXCLUDED.ai_generated,
		ai_tool = EXCLUDED.ai_tool,
		ai_review_status = EXCLUDED.ai_review_status,
		ai_review_note = EXCLUDED.ai_review_note,
		series_id = EXCLUDED.series_id,
		season_no = EXCLUDED.season_no,
		episode_no = EXCLUDED.episode_no,
		updated_at = NOW();`,
		v.ID, v.ChannelID, v.Title, nullIfEmpty(v.Description), v.VideoType, v.Visibility, v.Status,
		v.DurationSeconds, nullIfEmpty(v.ThumbnailURL), nullIfEmpty(v.CustomThumbnailURL),
		pq.Array(frames), nullIfEmpty(v.HlsMasterURL),
		v.ViewCount, v.LikeCount, v.DislikeCount, v.CommentCount,
		pq.Array(tags), v.Category, elements, nullTime(v.PublishedAt), nullTimePtr(v.ScheduledAt),
		v.DrmProtected, nullIfEmpty(v.DrmKeyID), nullIfEmpty(v.DrmKeyHex),
		v.AiGenerated, nullIfEmpty(v.AiTool), v.AiReviewStatus, nullIfEmpty(v.AiReviewNote),
		nullIfEmpty(v.SeriesID), v.SeasonNo, v.EpisodeNo, v.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert videos: %w", err)
	}
	return nil
}

// SaveSeries inserts or updates a series row in PostgreSQL.
func (p *PostgresStore) SaveSeries(ctx context.Context, s *Series) error {
	if s == nil {
		return fmt.Errorf("nil series")
	}
	if !validUUID(s.ID) {
		return fmt.Errorf("series id %q is not a UUID", s.ID)
	}
	if !validUUID(s.ChannelID) {
		return fmt.Errorf("series channel id %q is not a UUID", s.ChannelID)
	}
	_, err := p.DB.ExecContext(ctx, `INSERT INTO series (
		id, channel_id, title, description, cover_url, status, ai_original,
		created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	ON CONFLICT (id) DO UPDATE SET
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		cover_url = EXCLUDED.cover_url,
		status = EXCLUDED.status,
		ai_original = EXCLUDED.ai_original,
		updated_at = NOW();`,
		s.ID, s.ChannelID, s.Title, nullIfEmpty(s.Description),
		nullIfEmpty(s.CoverURL), s.Status, s.AiOriginal, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert series: %w", err)
	}
	return nil
}

// DeleteSeries removes a series row; episodes are detached first by the
// caller via SaveVideo so no FK-bearing rows dangle.
func (p *PostgresStore) DeleteSeries(ctx context.Context, id string) error {
	if !validUUID(id) {
		return fmt.Errorf("series id %q is not a UUID", id)
	}
	if _, err := p.DB.ExecContext(ctx, `DELETE FROM series WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete series: %w", err)
	}
	return nil
}

// SaveHourlyViews increments a channel hour bucket in PostgreSQL.
func (p *PostgresStore) SaveHourlyViews(ctx context.Context, channelID string, hour time.Time, delta int64) error {
	if !validUUID(channelID) {
		return fmt.Errorf("channel id %q is not a UUID", channelID)
	}
	_, err := p.DB.ExecContext(ctx, `INSERT INTO channel_hourly_views (
		channel_id, hour, views, updated_at
	) VALUES ($1, $2, $3, NOW())
	ON CONFLICT (channel_id, hour) DO UPDATE SET
		views = channel_hourly_views.views + EXCLUDED.views,
		updated_at = NOW();`,
		channelID, hour.UTC().Truncate(time.Hour), delta,
	)
	if err != nil {
		return fmt.Errorf("upsert hourly views: %w", err)
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

// memoryToPGJobStatus maps memory lifecycle states to migration states.
func memoryToPGJobStatus(status string) string {
	switch status {
	case "PROCESSING":
		return "RUNNING"
	case "READY":
		return "SUCCEEDED"
	case "QUEUED", "FAILED":
		return status
	default:
		return "QUEUED"
	}
}

// pgToMemoryJobStatus maps migration states back. Stale RUNNING rows come
// back as QUEUED so the boot sweep re-dispatches them (P1-13).
func pgToMemoryJobStatus(status string) string {
	switch status {
	case "RUNNING":
		return "QUEUED"
	case "SUCCEEDED":
		return "READY"
	case "QUEUED", "FAILED":
		return status
	default:
		return "QUEUED"
	}
}

// SaveMediaJob inserts or updates a transcode job in PostgreSQL.
func (p *PostgresStore) SaveMediaJob(ctx context.Context, job *MediaJob) error {
	if job == nil {
		return fmt.Errorf("nil job")
	}
	if !validUUID(job.ID) {
		return fmt.Errorf("job id %q is not a UUID", job.ID)
	}
	if !validUUID(job.VideoID) {
		return fmt.Errorf("job video id %q is not a UUID", job.VideoID)
	}
	_, err := p.DB.ExecContext(ctx, `INSERT INTO media_jobs (
		id, video_id, job_type, status, progress_percent, error_message,
		source_ref, started_at, completed_at
	) VALUES ($1, $2, 'encode', $3, $4, $5, $6, $7, $8)
	ON CONFLICT (id) DO UPDATE SET
		status = EXCLUDED.status,
		progress_percent = EXCLUDED.progress_percent,
		error_message = EXCLUDED.error_message,
		source_ref = EXCLUDED.source_ref,
		started_at = EXCLUDED.started_at,
		completed_at = EXCLUDED.completed_at;`,
		job.ID, job.VideoID, memoryToPGJobStatus(job.Status), job.ProgressPct,
		nullIfEmpty(job.Error), nullIfEmpty(job.InputURI),
		nullTime(job.CreatedAt), nullTimePtr(job.CompletedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert media_jobs: %w", err)
	}
	return nil
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

func nullTimePtr(t *time.Time) sql.NullTime {
	if t == nil || t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
