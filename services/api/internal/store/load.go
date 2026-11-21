package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// LoadUser reads one user with its profile from PostgreSQL.
func (p *PostgresStore) LoadUser(ctx context.Context, id string) (*User, error) {
	var u User
	var profileDisplay, profileBio, profileAvatar, profileBanner sql.NullString
	err := p.DB.QueryRowContext(ctx, `SELECT
		u.id, u.email, u.username, u.role, u.subscription_tier, u.is_verified, u.is_active,
		u.two_factor_enabled, COALESCE(u.two_factor_secret, ''), u.created_at,
		p.display_name, p.bio, p.avatar_url, p.banner_url
		FROM users u LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE u.id = $1`, id).Scan(
		&u.ID, &u.Email, &u.Username, &u.Role, &u.SubscriptionTier, &u.IsVerified, &u.IsActive,
		&u.MFAEnabled, &u.TOTPSecret, &u.CreatedAt,
		&profileDisplay, &profileBio, &profileAvatar, &profileBanner,
	)
	if err != nil {
		return nil, fmt.Errorf("load user %s: %w", id, err)
	}
	u.DisplayName = profileDisplay.String
	u.Bio = profileBio.String
	u.AvatarURL = profileAvatar.String
	u.BannerURL = profileBanner.String
	return &u, nil
}

// LoadUsers reads all users with profiles (bounded by operator table size).
func (p *PostgresStore) LoadUsers(ctx context.Context) ([]*User, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT
		u.id, u.email, u.username, u.role, u.subscription_tier, u.is_verified, u.is_active,
		u.two_factor_enabled, COALESCE(u.two_factor_secret, ''), u.created_at,
		p.display_name, p.bio, p.avatar_url, p.banner_url
		FROM users u LEFT JOIN user_profiles p ON p.user_id = u.id`)
	if err != nil {
		return nil, fmt.Errorf("load users: %w", err)
	}
	defer rows.Close()
	var out []*User
	for rows.Next() {
		var u User
		var profileDisplay, profileBio, profileAvatar, profileBanner sql.NullString
		if err := rows.Scan(
			&u.ID, &u.Email, &u.Username, &u.Role, &u.SubscriptionTier, &u.IsVerified, &u.IsActive,
			&u.MFAEnabled, &u.TOTPSecret, &u.CreatedAt,
			&profileDisplay, &profileBio, &profileAvatar, &profileBanner,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		u.DisplayName = profileDisplay.String
		u.Bio = profileBio.String
		u.AvatarURL = profileAvatar.String
		u.BannerURL = profileBanner.String
		out = append(out, &u)
	}
	return out, rows.Err()
}

// LoadChannels reads all channels from PostgreSQL.
func (p *PostgresStore) LoadChannels(ctx context.Context) ([]*Channel, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT
		id, owner_id, handle, name, COALESCE(description, ''), COALESCE(avatar_url, ''),
		COALESCE(banner_url, ''), subscriber_count, video_count, is_verified,
		monetization_enabled, created_at FROM channels`)
	if err != nil {
		return nil, fmt.Errorf("load channels: %w", err)
	}
	defer rows.Close()
	var out []*Channel
	for rows.Next() {
		var c Channel
		if err := rows.Scan(
			&c.ID, &c.OwnerID, &c.Handle, &c.Name, &c.Description, &c.AvatarURL,
			&c.BannerURL, &c.SubscriberCount, &c.VideoCount, &c.IsVerified,
			&c.MonetizationEnabled, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}

// LoadVideos reads all videos from PostgreSQL.
func (p *PostgresStore) LoadVideos(ctx context.Context) ([]*Video, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT
		id, channel_id, title, COALESCE(description, ''), video_type, visibility, status,
		duration_seconds, COALESCE(thumbnail_url, ''), COALESCE(custom_thumbnail_url, ''),
		COALESCE(auto_frame_urls, '{}'), COALESCE(hls_master_url, ''),
		view_count, like_count, dislike_count, comment_count, tags, category,
		COALESCE(elements, '[]'), published_at, scheduled_at,
		COALESCE(drm_protected, false), COALESCE(drm_key_id, ''), COALESCE(drm_key_hex, ''),
		COALESCE(ai_generated, false), COALESCE(ai_tool, ''), COALESCE(ai_review_status, ''),
		COALESCE(ai_review_note, ''), series_id, COALESCE(season_no, 0), COALESCE(episode_no, 0),
		created_at FROM videos`)
	if err != nil {
		return nil, fmt.Errorf("load videos: %w", err)
	}
	defer rows.Close()
	var out []*Video
	for rows.Next() {
		var v Video
		var published, scheduled sql.NullTime
		var seriesID sql.NullString
		var elementsJSON []byte
		if err := rows.Scan(
			&v.ID, &v.ChannelID, &v.Title, &v.Description, &v.VideoType, &v.Visibility, &v.Status,
			&v.DurationSeconds, &v.ThumbnailURL, &v.CustomThumbnailURL,
			pq.Array(&v.AutoFrameURLs), &v.HlsMasterURL,
			&v.ViewCount, &v.LikeCount, &v.DislikeCount, &v.CommentCount,
			pq.Array(&v.Tags), &v.Category, &elementsJSON, &published, &scheduled,
			&v.DrmProtected, &v.DrmKeyID, &v.DrmKeyHex,
			&v.AiGenerated, &v.AiTool, &v.AiReviewStatus, &v.AiReviewNote,
			&seriesID, &v.SeasonNo, &v.EpisodeNo, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan video: %w", err)
		}
		v.SeriesID = seriesID.String
		if published.Valid {
			v.PublishedAt = published.Time
		}
		if scheduled.Valid {
			v.ScheduledAt = &scheduled.Time
		}
		if len(elementsJSON) > 0 {
			if err := json.Unmarshal(elementsJSON, &v.Elements); err != nil {
				return nil, fmt.Errorf("decode video elements: %w", err)
			}
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

// LoadSeries reads all series rows from PostgreSQL.
func (p *PostgresStore) LoadSeries(ctx context.Context) ([]*Series, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT
		id, channel_id, title, COALESCE(description, ''), COALESCE(cover_url, ''),
		status, COALESCE(ai_original, false), created_at, updated_at
		FROM series`)
	if err != nil {
		return nil, fmt.Errorf("load series: %w", err)
	}
	defer rows.Close()
	var out []*Series
	for rows.Next() {
		var s Series
		if err := rows.Scan(
			&s.ID, &s.ChannelID, &s.Title, &s.Description, &s.CoverURL,
			&s.Status, &s.AiOriginal, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan series: %w", err)
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

// HourlyBucket is one persisted per-channel hour of counted views.
type HourlyBucket struct {
	ChannelID string
	Hour      time.Time
	Views     int64
}

// LoadHourlyViews reads per-channel hourly views since the given cutoff.
func (p *PostgresStore) LoadHourlyViews(ctx context.Context, since time.Time) ([]HourlyBucket, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT channel_id, hour, views
		FROM channel_hourly_views WHERE hour >= $1 ORDER BY hour ASC`, since)
	if err != nil {
		return nil, fmt.Errorf("load hourly views: %w", err)
	}
	defer rows.Close()
	var out []HourlyBucket
	for rows.Next() {
		var b HourlyBucket
		if err := rows.Scan(&b.ChannelID, &b.Hour, &b.Views); err != nil {
			return nil, fmt.Errorf("scan hourly views: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// LoadTransactions reads all payment transactions with ledger entries.
func (p *PostgresStore) LoadTransactions(ctx context.Context) ([]*PaymentTransaction, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT
		id, user_id, provider, amount_cents, currency, status, idempotency_key,
		COALESCE(tx_type, ''), COALESCE(checkout_url, ''), created_at
		FROM payment_transactions`)
	if err != nil {
		return nil, fmt.Errorf("load transactions: %w", err)
	}
	defer rows.Close()
	var out []*PaymentTransaction
	for rows.Next() {
		var tx PaymentTransaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.Provider, &tx.AmountCents, &tx.Currency,
			&tx.Status, &tx.IdempotencyKey, &tx.Type, &tx.CheckoutURL, &tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		entries, err := p.loadLedgerEntries(ctx, tx.ID)
		if err != nil {
			return nil, err
		}
		tx.LedgerEntries = entries
		out = append(out, &tx)
	}
	return out, rows.Err()
}

// LoadMediaJobs reads all transcode jobs, mapping stale RUNNING back to QUEUED.
func (p *PostgresStore) LoadMediaJobs(ctx context.Context) ([]*MediaJob, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT
		id, video_id, status, progress_percent, COALESCE(error_message, ''),
		COALESCE(source_ref, ''), completed_at, created_at
		FROM media_jobs`)
	if err != nil {
		return nil, fmt.Errorf("load media jobs: %w", err)
	}
	defer rows.Close()
	var out []*MediaJob
	for rows.Next() {
		var job MediaJob
		var pgStatus string
		var completed sql.NullTime
		if err := rows.Scan(
			&job.ID, &job.VideoID, &pgStatus, &job.ProgressPct, &job.Error,
			&job.InputURI, &completed, &job.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan media job: %w", err)
		}
		job.Status = pgToMemoryJobStatus(pgStatus)
		if completed.Valid {
			at := completed.Time
			job.CompletedAt = &at
		}
		out = append(out, &job)
	}
	return out, rows.Err()
}

func (p *PostgresStore) loadLedgerEntries(ctx context.Context, txID string) ([]LedgerEntry, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT
		e.id, e.transaction_id, a.account_number, e.debit_cents, e.credit_cents,
		e.currency, e.created_at
		FROM ledger_entries e JOIN ledger_accounts a ON a.id = e.account_id
		WHERE e.transaction_id = $1`, txID)
	if err != nil {
		return nil, fmt.Errorf("load ledger entries: %w", err)
	}
	defer rows.Close()
	var out []LedgerEntry
	for rows.Next() {
		var e LedgerEntry
		var accountNumber string
		var debit, credit int64
		if err := rows.Scan(&e.ID, &e.TransactionID, &accountNumber, &debit, &credit, &e.Currency, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ledger entry: %w", err)
		}
		accountType, ok := ledgerAccountTypes[accountNumber]
		if !ok {
			return nil, fmt.Errorf("unknown ledger account number %q", accountNumber)
		}
		e.AccountType = accountType
		if debit > 0 {
			e.Direction = "DEBIT"
			e.AmountCents = debit
		} else {
			e.Direction = "CREDIT"
			e.AmountCents = credit
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// scanDecision maps one ad_delivery_decisions row, restoring non-UUID memory
// IDs from metadata refs when the UUID columns are NULL. The Creative blob
// is memory-only and stays nil; event validation needs Status/CampaignID.
func scanDecision(row interface{ Scan(...interface{}) error }) (*AdDeliveryDecision, error) {
	var dec AdDeliveryDecision
	var campaignID, creativeID, userID sql.NullString
	var meta string
	if err := row.Scan(&dec.DecisionID, &dec.SlotID, &campaignID, &creativeID, &userID,
		&dec.DecidedAt, &dec.Status, &dec.Reason, &dec.BidCpmCents, &meta); err != nil {
		return nil, err
	}
	dec.CampaignID, dec.CreativeID, dec.UserID = campaignID.String, creativeID.String, userID.String
	if meta != "" {
		var refs map[string]string
		if err := json.Unmarshal([]byte(meta), &refs); err == nil {
			if dec.CampaignID == "" {
				dec.CampaignID = refs[decisionCampaignRefKey]
			}
			if dec.CreativeID == "" {
				dec.CreativeID = refs[decisionCreativeRefKey]
			}
			if dec.UserID == "" {
				dec.UserID = refs[decisionUserRefKey]
			}
		}
	}
	return &dec, nil
}

const decisionColumns = `decision_id, slot_id, campaign_id, creative_id, user_id,
	decided_at, status, reason, bid_cpm_cents, metadata`

// LoadDecision reads one ad delivery decision by id. It returns (nil, nil)
// when no row exists so callers can fail closed on unknown decisions.
func (p *PostgresStore) LoadDecision(ctx context.Context, id string) (*AdDeliveryDecision, error) {
	dec, err := scanDecision(p.DB.QueryRowContext(ctx,
		`SELECT `+decisionColumns+` FROM ad_delivery_decisions WHERE decision_id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load decision %s: %w", id, err)
	}
	return dec, nil
}

// LoadRecentDecisions reads the newest decisions since the cutoff for boot
// hydration. A non-positive limit selects the default cap.
func (p *PostgresStore) LoadRecentDecisions(ctx context.Context, since time.Time, limit int) ([]*AdDeliveryDecision, error) {
	if limit <= 0 {
		limit = 5000
	}
	rows, err := p.DB.QueryContext(ctx,
		`SELECT `+decisionColumns+` FROM ad_delivery_decisions WHERE decided_at >= $1 ORDER BY decided_at DESC LIMIT $2`,
		since, limit)
	if err != nil {
		return nil, fmt.Errorf("load recent decisions: %w", err)
	}
	defer rows.Close()
	var out []*AdDeliveryDecision
	for rows.Next() {
		dec, err := scanDecision(rows)
		if err != nil {
			return nil, fmt.Errorf("scan decision: %w", err)
		}
		out = append(out, dec)
	}
	return out, rows.Err()
}
