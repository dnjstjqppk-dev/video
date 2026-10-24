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

-- +goose Down
DROP TABLE IF EXISTS ad_budget_allocations;
DROP TABLE IF EXISTS ad_events;
DROP TABLE IF EXISTS ad_placements;
DROP TABLE IF EXISTS ad_creatives;
DROP TABLE IF EXISTS ad_campaigns;
DROP TABLE IF EXISTS advertisers;

