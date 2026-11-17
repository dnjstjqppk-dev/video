package store

import (
	"context"
	"database/sql"
	"fmt"

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
		duration_seconds, COALESCE(thumbnail_url, ''), COALESCE(hls_master_url, ''),
		view_count, like_count, dislike_count, comment_count, tags, category,
		published_at, created_at FROM videos`)
	if err != nil {
		return nil, fmt.Errorf("load videos: %w", err)
	}
	defer rows.Close()
	var out []*Video
	for rows.Next() {
		var v Video
		var published sql.NullTime
		if err := rows.Scan(
			&v.ID, &v.ChannelID, &v.Title, &v.Description, &v.VideoType, &v.Visibility, &v.Status,
			&v.DurationSeconds, &v.ThumbnailURL, &v.HlsMasterURL,
			&v.ViewCount, &v.LikeCount, &v.DislikeCount, &v.CommentCount,
			pq.Array(&v.Tags), &v.Category, &published, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan video: %w", err)
		}
		if published.Valid {
			v.PublishedAt = published.Time
		}
		out = append(out, &v)
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
