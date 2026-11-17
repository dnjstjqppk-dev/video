-- Canonical PostgreSQL schema snapshot (generated, do not hand-edit).
-- Source of truth: db/migrations/*.sql (goose Up sections).
-- Regenerate: python docs/generate_schema.py

-- ========== 00001_create_accounts.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00001: Accounts, Users, Sessions, Organizations, RBAC
-- ==========================================

CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(64) PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
    ('visitor', 'Unauthenticated visitor'),
    ('registered_user', 'Standard registered consumer'),
    ('creator', 'Creator with an active channel'),
    ('channel_manager', 'Delegated channel manager'),
    ('channel_editor', 'Delegated channel editor'),
    ('channel_analyst', 'Delegated channel analytics viewer'),
    ('advertiser', 'Ad campaign manager'),
    ('mcn_manager', 'Multi-channel network representative'),
    ('rights_owner', 'Copyright owner / content manager'),
    ('partner_admin', 'Partner organization administrator'),
    ('moderator', 'Trust & Safety content reviewer'),
    ('support_agent', 'Customer support specialist'),
    ('finance_admin', 'Settlement & ledger auditor'),
    ('platform_admin', 'System operations administrator'),
    ('super_admin', 'Unrestricted root administrator')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(64) PRIMARY KEY,
    description TEXT,
    category VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name VARCHAR(64) REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(64) REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    role VARCHAR(64) NOT NULL REFERENCES roles(name) DEFAULT 'registered_user',
    subscription_tier VARCHAR(32) NOT NULL DEFAULT 'free',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    zitadel_user_id VARCHAR(128) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(128) NOT NULL,
    bio TEXT,
    avatar_url TEXT,
    banner_url TEXT,
    country_code VARCHAR(4),
    language VARCHAR(8) DEFAULT 'ko',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    dark_mode BOOLEAN NOT NULL DEFAULT TRUE,
    autoplay BOOLEAN NOT NULL DEFAULT TRUE,
    default_quality VARCHAR(16) NOT NULL DEFAULT 'auto',
    email_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    push_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    restricted_mode BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(128) NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user ON sessions(user_id, is_revoked);

CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(128) NOT NULL,
    device_type VARCHAR(32) NOT NULL, -- web, ios, android
    fcm_token TEXT,
    apns_token TEXT,
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL, -- advertiser, mcn, rights_owner
    contact_email VARCHAR(255) NOT NULL,
    tax_id VARCHAR(64),
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS organization_members (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_role VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);

-- ========== 00002_create_content.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00002: Channels, Videos, Media Pipeline, Playlists
-- ==========================================

CREATE TABLE IF NOT EXISTS channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    handle VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    avatar_url TEXT,
    banner_url TEXT,
    subscriber_count BIGINT NOT NULL DEFAULT 0,
    video_count INT NOT NULL DEFAULT 0,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    monetization_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channels_owner ON channels(owner_id);
CREATE INDEX idx_channels_handle ON channels(handle);

CREATE TABLE IF NOT EXISTS channel_members (
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL, -- manager, editor, viewer
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, user_id)
);

CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    video_type VARCHAR(32) NOT NULL DEFAULT 'standard', -- standard, shorts, live_vod
    visibility VARCHAR(32) NOT NULL DEFAULT 'private', -- public, unlisted, private, members_only, scheduled
    status VARCHAR(32) NOT NULL DEFAULT 'CREATED', -- CREATED, UPLOADING, UPLOADED, VALIDATING, QUEUED, PROCESSING, READY, FAILED, BLOCKED, DELETED
    duration_seconds INT NOT NULL DEFAULT 0,
    thumbnail_url TEXT,
    preview_url TEXT,
    hls_master_url TEXT,
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count BIGINT NOT NULL DEFAULT 0,
    dislike_count BIGINT NOT NULL DEFAULT 0,
    comment_count BIGINT NOT NULL DEFAULT 0,
    tags TEXT[] NOT NULL DEFAULT '{}',
    category VARCHAR(64) NOT NULL DEFAULT 'General',
    is_age_restricted BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ,
    scheduled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_videos_channel ON videos(channel_id);
CREATE INDEX idx_videos_visibility_status ON videos(visibility, status);
CREATE INDEX idx_videos_published_at ON videos(published_at DESC) WHERE visibility = 'public' AND status = 'READY';
CREATE INDEX idx_videos_type ON videos(video_type) WHERE visibility = 'public' AND status = 'READY';

CREATE TABLE IF NOT EXISTS video_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    r2_bucket VARCHAR(128) NOT NULL,
    r2_key TEXT NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    mime_type VARCHAR(64) NOT NULL,
    checksum_sha256 VARCHAR(64),
    video_codec VARCHAR(32),
    audio_codec VARCHAR(32),
    width INT,
    height INT,
    fps DOUBLE PRECISION,
    bitrate BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS video_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    quality VARCHAR(16) NOT NULL, -- 240p, 360p, 480p, 720p, 1080p, 1440p, 2160p
    manifest_r2_key TEXT NOT NULL,
    width INT NOT NULL,
    height INT NOT NULL,
    bitrate BIGINT NOT NULL,
    fps INT NOT NULL DEFAULT 30,
    codec VARCHAR(32) NOT NULL DEFAULT 'avc1.64001f,mp4a.40.2',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS upload_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    r2_upload_id TEXT NOT NULL,
    r2_key TEXT NOT NULL,
    part_size BIGINT NOT NULL,
    total_parts INT NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, COMPLETED, ABORTED, EXPIRED
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS upload_parts (
    upload_session_id UUID NOT NULL REFERENCES upload_sessions(id) ON DELETE CASCADE,
    part_number INT NOT NULL,
    etag VARCHAR(128) NOT NULL,
    part_size_bytes BIGINT NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (upload_session_id, part_number)
);

CREATE TABLE IF NOT EXISTS media_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    job_type VARCHAR(32) NOT NULL, -- encode, thumbnail, storyboard, blur, audio_norm
    status VARCHAR(32) NOT NULL DEFAULT 'QUEUED', -- QUEUED, RUNNING, SUCCEEDED, FAILED
    progress_percent INT NOT NULL DEFAULT 0,
    worker_id VARCHAR(64),
    error_message TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS thumbnails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    r2_key TEXT NOT NULL,
    url TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    timestamp_seconds DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS captions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    language VARCHAR(8) NOT NULL,
    label VARCHAR(64) NOT NULL,
    vtt_r2_key TEXT NOT NULL,
    vtt_url TEXT NOT NULL,
    is_auto_generated BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    title VARCHAR(128) NOT NULL,
    start_time_seconds INT NOT NULL,
    end_time_seconds INT NOT NULL,
    thumbnail_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS playlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    title VARCHAR(128) NOT NULL,
    description TEXT,
    visibility VARCHAR(32) NOT NULL DEFAULT 'public',
    video_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS playlist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    playlist_id UUID NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    position INT NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (playlist_id, video_id)
);

CREATE TABLE IF NOT EXISTS visibility_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    allowed_countries TEXT[] NOT NULL DEFAULT '{}',
    blocked_countries TEXT[] NOT NULL DEFAULT '{}',
    require_membership_tier VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ========== 00003_create_social.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00003: Subscriptions, Reactions, Comments, History, Notifications, Community
-- ==========================================

CREATE TABLE IF NOT EXISTS subscriptions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    notification_level VARCHAR(16) NOT NULL DEFAULT 'all', -- all, personalized, none
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, channel_id)
);

CREATE INDEX idx_subscriptions_channel ON subscriptions(channel_id);

CREATE TABLE IF NOT EXISTS video_reactions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    reaction_type VARCHAR(16) NOT NULL, -- like, dislike
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, video_id)
);

CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    like_count INT NOT NULL DEFAULT 0,
    dislike_count INT NOT NULL DEFAULT 0,
    reply_count INT NOT NULL DEFAULT 0,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    is_hearted BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_video ON comments(video_id, created_at DESC) WHERE parent_id IS NULL AND is_deleted = FALSE;
CREATE INDEX idx_comments_parent ON comments(parent_id) WHERE parent_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS comment_reactions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    reaction_type VARCHAR(16) NOT NULL, -- like, dislike
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, comment_id)
);

CREATE TABLE IF NOT EXISTS watch_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    watched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, video_id)
);

CREATE TABLE IF NOT EXISTS watch_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    last_position_seconds INT NOT NULL DEFAULT 0,
    duration_seconds INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, video_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(32) NOT NULL, -- upload, reply, mention, live, billing, moderation
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    target_url TEXT NOT NULL,
    thumbnail_url TEXT,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);

CREATE TABLE IF NOT EXISTS community_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    image_urls TEXT[] NOT NULL DEFAULT '{}',
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS community_votes (
    post_id UUID NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    poll_option_index INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

-- ========== 00004_create_live.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00004: Live Streaming Infrastructure & Chat
-- ==========================================

CREATE TABLE IF NOT EXISTS live_streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'idle', -- idle, preparing, live, disconnected, ended
    stream_key VARCHAR(128) NOT NULL UNIQUE,
    cloudflare_stream_uid VARCHAR(128),
    rtmps_ingest_url TEXT NOT NULL,
    srt_ingest_url TEXT NOT NULL,
    playback_hls_url TEXT NOT NULL,
    viewer_count INT NOT NULL DEFAULT 0,
    peak_viewer_count INT NOT NULL DEFAULT 0,
    scheduled_start_time TIMESTAMPTZ,
    actual_start_time TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_live_streams_channel ON live_streams(channel_id);
CREATE INDEX idx_live_streams_status ON live_streams(status);

CREATE TABLE IF NOT EXISTS live_inputs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    live_stream_id UUID NOT NULL REFERENCES live_streams(id) ON DELETE CASCADE,
    client_ip INET,
    encoder_name VARCHAR(128),
    resolution VARCHAR(32),
    bitrate_kbps INT,
    fps DOUBLE PRECISION,
    connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disconnected_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS live_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    live_stream_id UUID NOT NULL REFERENCES live_streams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    donation_amount_cents BIGINT,
    currency VARCHAR(4) DEFAULT 'KRW',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_live_chat_stream ON live_chat_messages(live_stream_id, created_at DESC) WHERE is_deleted = FALSE;

CREATE TABLE IF NOT EXISTS live_recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    live_stream_id UUID NOT NULL REFERENCES live_streams(id) ON DELETE CASCADE,
    video_id UUID REFERENCES videos(id) ON DELETE SET NULL,
    r2_key TEXT NOT NULL,
    duration_seconds INT NOT NULL DEFAULT 0,
    file_size_bytes BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS live_moderators (
    live_stream_id UUID NOT NULL REFERENCES live_streams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (live_stream_id, user_id)
);

-- ========== 00005_create_ads.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00005: Advertising Engine & Inventory
-- ==========================================

CREATE TABLE IF NOT EXISTS advertisers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    billing_address TEXT NOT NULL,
    tax_number VARCHAR(64),
    balance_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ad_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES advertisers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    format VARCHAR(32) NOT NULL, -- preroll, midroll, feed, shorts, display
    status VARCHAR(32) NOT NULL DEFAULT 'draft', -- draft, pending_review, active, paused, completed, rejected
    total_budget_cents BIGINT NOT NULL,
    daily_budget_cents BIGINT NOT NULL,
    spent_cents BIGINT NOT NULL DEFAULT 0,
    cpm_bid_micros BIGINT NOT NULL DEFAULT 0,
    cpc_bid_micros BIGINT NOT NULL DEFAULT 0,
    target_categories TEXT[] NOT NULL DEFAULT '{}',
    target_countries TEXT[] NOT NULL DEFAULT '{}',
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    impressions BIGINT NOT NULL DEFAULT 0,
    clicks BIGINT NOT NULL DEFAULT 0,
    views BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaigns_status ON ad_campaigns(status, start_date, end_date);

CREATE TABLE IF NOT EXISTS ad_creatives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    title VARCHAR(128) NOT NULL,
    video_asset_id UUID REFERENCES video_assets(id) ON DELETE SET NULL,
    media_url TEXT NOT NULL,
    destination_url TEXT NOT NULL,
    cta_text VARCHAR(32) NOT NULL DEFAULT 'Learn More',
    duration_seconds INT NOT NULL DEFAULT 15,
    skip_offset_seconds INT NOT NULL DEFAULT 5,
    is_approved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ad_placements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    time_offset_seconds INT NOT NULL DEFAULT 0,
    placement_type VARCHAR(32) NOT NULL, -- preroll, midroll, postroll
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ad_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    creative_id UUID NOT NULL REFERENCES ad_creatives(id) ON DELETE CASCADE,
    video_id UUID REFERENCES videos(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(32) NOT NULL, -- impression, click, view_25, view_50, view_75, complete, skip
    billed_amount_cents BIGINT NOT NULL DEFAULT 0,
    ip_hash VARCHAR(64) NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ad_events_campaign ON ad_events(campaign_id, event_type, created_at);

CREATE TABLE IF NOT EXISTS ad_budget_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    allocated_date DATE NOT NULL,
    budget_cents BIGINT NOT NULL,
    spent_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (campaign_id, allocated_date)
);

-- ========== 00006_create_payments_and_ledger.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00006: Monetization, Billing, Double-Entry Ledger, Payouts
-- ==========================================

CREATE TABLE IF NOT EXISTS monetization_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE UNIQUE,
    bank_name VARCHAR(64) NOT NULL,
    account_number_encrypted TEXT NOT NULL,
    account_holder_name VARCHAR(128) NOT NULL,
    tax_identification_number VARCHAR(64),
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending, active, suspended, rejected
    payout_currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    tier_name VARCHAR(64) NOT NULL,
    monthly_price_cents BIGINT NOT NULL,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    perks TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscriptions_billing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier VARCHAR(32) NOT NULL, -- premium, premium_family, channel_membership
    channel_id UUID REFERENCES channels(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active', -- active, past_due, canceled, paused
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    external_subscription_id VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    provider VARCHAR(32) NOT NULL, -- toss, stripe
    external_transaction_id VARCHAR(128) NOT NULL UNIQUE,
    amount_cents BIGINT NOT NULL,
    fee_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    status VARCHAR(32) NOT NULL, -- pending, succeeded, failed, refunded, partially_refunded
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    payment_method VARCHAR(32) NOT NULL, -- card, virtual_account, easy_pay
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_transactions_user ON payment_transactions(user_id);

CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES payment_transactions(id) ON DELETE RESTRICT,
    refund_amount_cents BIGINT NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Strict Double-Entry Ledger System (Section 14.3)
CREATE TABLE IF NOT EXISTS ledger_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_number VARCHAR(64) NOT NULL UNIQUE,
    account_name VARCHAR(128) NOT NULL,
    account_type VARCHAR(32) NOT NULL, -- asset, liability, equity, revenue, expense
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    balance_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Essential System Ledger Accounts
INSERT INTO ledger_accounts (account_number, account_name, account_type, currency) VALUES
    ('1000-CASH', 'Platform Operating Cash', 'asset', 'KRW'),
    ('2000-CREATOR-PAYABLE', 'Creator Earnings Payable', 'liability', 'KRW'),
    ('2100-MCN-PAYABLE', 'MCN Share Payable', 'liability', 'KRW'),
    ('2200-RIGHTS-PAYABLE', 'Copyright Royalties Payable', 'liability', 'KRW'),
    ('4000-AD-REVENUE', 'Gross Ad Revenue', 'revenue', 'KRW'),
    ('4100-SUB-REVENUE', 'Premium Subscription Revenue', 'revenue', 'KRW'),
    ('5000-PG-FEE', 'Payment Gateway Processing Fees', 'expense', 'KRW')
ON CONFLICT (account_number) DO NOTHING;

CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    account_id UUID NOT NULL REFERENCES ledger_accounts(id) ON DELETE RESTRICT,
    debit_cents BIGINT NOT NULL DEFAULT 0,
    credit_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_transaction ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_entries_account ON ledger_entries(account_id, created_at);

CREATE TABLE IF NOT EXISTS creator_earnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    period_date DATE NOT NULL,
    ad_revenue_cents BIGINT NOT NULL DEFAULT 0,
    membership_revenue_cents BIGINT NOT NULL DEFAULT 0,
    donation_revenue_cents BIGINT NOT NULL DEFAULT 0,
    platform_fee_cents BIGINT NOT NULL DEFAULT 0,
    net_earnings_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (channel_id, period_date)
);

CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    organization_id UUID REFERENCES organizations(id) ON DELETE RESTRICT,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    gross_amount_cents BIGINT NOT NULL,
    tax_withheld_cents BIGINT NOT NULL DEFAULT 0,
    net_amount_cents BIGINT NOT NULL,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending, approved, processing, paid, failed, on_hold
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    payout_batch_id VARCHAR(128),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES advertisers(id) ON DELETE RESTRICT,
    invoice_number VARCHAR(64) NOT NULL UNIQUE,
    amount_cents BIGINT NOT NULL,
    tax_cents BIGINT NOT NULL,
    total_cents BIGINT NOT NULL,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    due_date DATE NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'issued', -- issued, paid, overdue, canceled
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ========== 00007_create_copyright_and_moderation.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00007: Rights Management, Content Claims, Moderation & Safety
-- ==========================================

CREATE TABLE IF NOT EXISTS rights_owners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    legal_name VARCHAR(255) NOT NULL,
    country VARCHAR(4) NOT NULL,
    verification_document_url TEXT,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rights_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES rights_owners(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    isrc_code VARCHAR(32),
    iswc_code VARCHAR(32),
    custom_id VARCHAR(64),
    fingerprint_hash VARCHAR(128) NOT NULL,
    fingerprint_data_url TEXT,
    asset_type VARCHAR(32) NOT NULL, -- sound_recording, composition, visual, audiovisual
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rights_territories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES rights_assets(id) ON DELETE CASCADE,
    territory_code VARCHAR(4) NOT NULL, -- ISO-3166-1 alpha-2 or 'WORLD'
    ownership_percentage DOUBLE PRECISION NOT NULL DEFAULT 100.0,
    policy VARCHAR(32) NOT NULL DEFAULT 'monetize', -- monetize, block, track
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (asset_id, territory_code)
);

CREATE TABLE IF NOT EXISTS copyright_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES rights_assets(id) ON DELETE RESTRICT,
    claimed_video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    rights_owner_id UUID NOT NULL REFERENCES rights_owners(id) ON DELETE RESTRICT,
    claim_action VARCHAR(32) NOT NULL DEFAULT 'monetize', -- block, monetize, track
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending, active, disputed, resolved, withdrawn
    match_start_seconds INT NOT NULL,
    match_end_seconds INT NOT NULL,
    match_confidence DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_copyright_claims_video ON copyright_claims(claimed_video_id);
CREATE INDEX idx_copyright_claims_owner ON copyright_claims(rights_owner_id);

CREATE TABLE IF NOT EXISTS copyright_disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id UUID NOT NULL REFERENCES copyright_claims(id) ON DELETE CASCADE,
    disputant_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reason VARCHAR(64) NOT NULL, -- original_content, licensed, public_domain, fair_use
    explanation TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'under_review', -- under_review, upheld, rejected, expired
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS dispute_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES copyright_disputes(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    target_type VARCHAR(32) NOT NULL, -- video, comment, channel, post
    target_id UUID NOT NULL,
    category VARCHAR(64) NOT NULL, -- hate_speech, violence, harassment, spam, copyright, adult
    description TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending, reviewing, resolved, dismissed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reports_target ON reports(target_type, target_id);

CREATE TABLE IF NOT EXISTS moderation_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES reports(id) ON DELETE SET NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id UUID NOT NULL,
    assigned_moderator_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, appealed, escalated
    violation_code VARCHAR(64),
    action_taken VARCHAR(64), -- none, age_restrict, geo_block, remove, ban_user
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- ========== 00008_create_operations_and_outbox.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00008: Operations, Audit Trail, Feature Flags, Transactional Outbox
-- ==========================================

CREATE TABLE IF NOT EXISTS admin_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action_type VARCHAR(64) NOT NULL, -- user_ban, video_takedown, policy_change, payout_approval
    target_entity_type VARCHAR(64) NOT NULL,
    target_entity_id UUID NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_role VARCHAR(64) NOT NULL,
    action VARCHAR(128) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id VARCHAR(128) NOT NULL,
    request_id VARCHAR(128),
    ip_address INET,
    user_agent TEXT,
    old_state JSONB,
    new_state JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

CREATE TABLE IF NOT EXISTS feature_flags (
    name VARCHAR(64) PRIMARY KEY,
    description TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    percentage_rollout INT NOT NULL DEFAULT 0,
    allowed_roles TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO feature_flags (name, description, is_enabled, percentage_rollout) VALUES
    ('enable_4k_encoding', 'Support 4K 2160p HLS rendering', TRUE, 100),
    ('enable_live_chat_donations', 'SuperChat donation feature', TRUE, 100),
    ('enable_shorts_feed', 'Enable infinite swipe vertical shorts', TRUE, 100),
    ('enable_ai_subtitle_generation', 'Automatic speech to text captions', TRUE, 50)
ON CONFLICT (name) DO NOTHING;

-- Transactional Outbox Pattern (Section 4.4)
CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(64) NOT NULL, -- video, user, payment, moderation
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(64) NOT NULL, -- VideoCreated, EncodingRequested, PaymentSucceeded
    payload JSONB NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING', -- PENDING, PUBLISHED, FAILED
    retry_count INT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_events_pending ON outbox_events(created_at) WHERE status = 'PENDING';

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    request_hash VARCHAR(64) NOT NULL,
    status_code INT NOT NULL,
    response_body JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- ========== 00009_create_production_ad_engine.sql (Up) ==========
-- +goose Up
-- ============================================================================
-- 00009: Production Ad Engine, Decision Logging, Telemetry & Diagnostics (ARTICLE 7)
-- ============================================================================

-- 1. AD INVENTORY SLOTS DEFINITION
CREATE TABLE IF NOT EXISTS ad_slots (
    id VARCHAR(64) PRIMARY KEY, -- e.g. watch_preroll, home_feed_banner, shorts_sponsored
    name VARCHAR(128) NOT NULL,
    format VARCHAR(32) NOT NULL, -- preroll, feed_banner, shorts_sponsored, midroll, display
    width INT NOT NULL DEFAULT 1920,
    height INT NOT NULL DEFAULT 1080,
    min_duration_seconds INT NOT NULL DEFAULT 0,
    max_duration_seconds INT NOT NULL DEFAULT 30,
    is_skippable BOOLEAN NOT NULL DEFAULT TRUE,
    skip_offset_seconds INT NOT NULL DEFAULT 5,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed production ad slots
INSERT INTO ad_slots (id, name, format, width, height, min_duration_seconds, max_duration_seconds, is_skippable, skip_offset_seconds, is_active)
VALUES 
    ('watch_preroll', 'Main Player Pre-Roll Video', 'preroll', 1920, 1080, 5, 30, TRUE, 5, TRUE),
    ('home_feed_banner', 'Home Feed Native Sponsored Card', 'feed_banner', 1280, 720, 0, 0, FALSE, 0, TRUE),
    ('shorts_sponsored', 'Shorts Vertical Sponsored Feed Item', 'shorts_sponsored', 1080, 1920, 5, 60, TRUE, 5, TRUE)
ON CONFLICT (id) DO UPDATE SET 
    name = EXCLUDED.name,
    format = EXCLUDED.format,
    is_active = EXCLUDED.is_active;

-- 2. AD FLIGHTS (Campaign pacing and flight schedule)
CREATE TABLE IF NOT EXISTS ad_flights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    pacing_strategy VARCHAR(32) NOT NULL DEFAULT 'even', -- even, asap, front_loaded
    daily_budget_cents BIGINT NOT NULL DEFAULT 0,
    spent_cents BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'active', -- active, paused, completed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ad_flights_campaign ON ad_flights(campaign_id, status);

-- 3. AD TARGETING RULES
CREATE TABLE IF NOT EXISTS ad_targeting_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    target_categories TEXT[] NOT NULL DEFAULT '{}',
    target_platforms TEXT[] NOT NULL DEFAULT '{WEB,ANDROID,IOS}',
    target_countries TEXT[] NOT NULL DEFAULT '{}',
    target_languages TEXT[] NOT NULL DEFAULT '{}',
    min_age INT DEFAULT NULL,
    max_age INT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ad_targeting_campaign ON ad_targeting_rules(campaign_id);

-- 4. AD CREATIVE REVIEWS & GOVERNANCE
CREATE TABLE IF NOT EXISTS ad_creative_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creative_id UUID NOT NULL REFERENCES ad_creatives(id) ON DELETE CASCADE,
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    review_status VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    policy_reason TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_creative_reviews_status ON ad_creative_reviews(review_status);

-- 5. AD DELIVERY DECISIONS LOG (Decision Engine Telemetry)
CREATE TABLE IF NOT EXISTS ad_delivery_decisions (
    decision_id VARCHAR(64) PRIMARY KEY,
    slot_id VARCHAR(64) NOT NULL REFERENCES ad_slots(id) ON DELETE CASCADE,
    campaign_id UUID REFERENCES ad_campaigns(id) ON DELETE SET NULL,
    creative_id UUID REFERENCES ad_creatives(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    decided_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(32) NOT NULL, -- served, unfilled
    reason VARCHAR(64) NOT NULL, -- WINNER_SELECTED, UNMATCHED_SLOT, NO_ELIGIBLE_CAMPAIGN, BUDGET_EXHAUSTED
    bid_cpm_cents BIGINT NOT NULL DEFAULT 0,
    decision_latency_ms INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_decisions_slot ON ad_delivery_decisions(slot_id, decided_at DESC);
CREATE INDEX IF NOT EXISTS idx_decisions_campaign ON ad_delivery_decisions(campaign_id, decided_at DESC);

-- 6. AD DELIVERY VERIFIED EVENTS (Audit-Proof Real Telemetry)
CREATE TABLE IF NOT EXISTS ad_delivery_events (
    event_id VARCHAR(64) PRIMARY KEY,
    decision_id VARCHAR(64) NOT NULL REFERENCES ad_delivery_decisions(decision_id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    creative_id UUID REFERENCES ad_creatives(id) ON DELETE SET NULL,
    event_type VARCHAR(32) NOT NULL, -- impression, start, first_quartile, midpoint, third_quartile, complete, click, skip
    billed_cost_cents BIGINT NOT NULL DEFAULT 0,
    idempotency_key VARCHAR(128) UNIQUE,
    client_timestamp TIMESTAMPTZ NOT NULL,
    server_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_hash VARCHAR(64) NOT NULL,
    user_agent TEXT
);

CREATE INDEX IF NOT EXISTS idx_delivery_events_type ON ad_delivery_events(campaign_id, event_type, server_timestamp DESC);

-- 7. AD FREQUENCY CAPS
CREATE TABLE IF NOT EXISTS ad_frequency_caps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(128),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    window_seconds INT NOT NULL DEFAULT 86400, -- 24 hours
    max_impressions INT NOT NULL DEFAULT 3,
    current_count INT NOT NULL DEFAULT 1,
    last_impression_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_campaign_cap UNIQUE (user_id, campaign_id)
);

CREATE INDEX IF NOT EXISTS idx_frequency_caps_user ON ad_frequency_caps(user_id, campaign_id);

-- 8. AD REPORTING DAILY (Aggregated Performance & P&L)
CREATE TABLE IF NOT EXISTS ad_reporting_daily (
    report_date DATE NOT NULL,
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    slot_id VARCHAR(64) NOT NULL REFERENCES ad_slots(id) ON DELETE CASCADE,
    format VARCHAR(32) NOT NULL,
    requests BIGINT NOT NULL DEFAULT 0,
    impressions BIGINT NOT NULL DEFAULT 0,
    clicks BIGINT NOT NULL DEFAULT 0,
    completes BIGINT NOT NULL DEFAULT 0,
    skips BIGINT NOT NULL DEFAULT 0,
    gross_spend_cents BIGINT NOT NULL DEFAULT 0,
    net_revenue_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (report_date, campaign_id, slot_id)
);

CREATE INDEX IF NOT EXISTS idx_ad_reporting_date ON ad_reporting_daily(report_date DESC);

-- ========== 00010_forward_roles_and_payments.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00010: Forward fill for P0-1 persistence (roles + payment columns)
-- Idempotent: safe to re-apply. Never edits 00001..00009.
-- ==========================================

-- Roles referenced by the API role matrix / seed users but missing from 00001.
INSERT INTO roles (name, description) VALUES
    ('viewer', 'Authenticated viewer without a channel'),
    ('user', 'Standard registered consumer (alias)'),
    ('premium_member', 'Paid premium subscriber'),
    ('live_moderator', 'Live chat moderator'),
    ('partner_manager', 'MCN partner manager'),
    ('rights_holder', 'Copyright holder (portal alias)'),
    ('partner_finance', 'Partner finance operator'),
    ('finance_ops', 'Financial operations (alias)'),
    ('security_auditor', 'Read-only security auditor')
ON CONFLICT (name) DO NOTHING;

-- Columns carried by the API payment transaction model.
ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS tx_type VARCHAR(64);
ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS checkout_url TEXT;

-- ========== 00011_media_job_source.sql (Up) ==========
-- +goose Up
-- ==========================================
-- 00011: Media job source reference (P1-13 durable dispatch)
-- Idempotent: safe to re-apply. Never edits earlier migrations.
-- ==========================================

-- Canonical transcode source carried by each job: a file:// URI for
-- local-relay uploads or a raw R2 object key (presigned fresh at every
-- dispatch attempt so restarts never reuse an expired URL).
ALTER TABLE media_jobs ADD COLUMN IF NOT EXISTS source_ref TEXT;
