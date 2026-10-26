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

-- +goose Down
DROP TABLE IF EXISTS ad_reporting_daily;
DROP TABLE IF EXISTS ad_frequency_caps;
DROP TABLE IF EXISTS ad_delivery_events;
DROP TABLE IF EXISTS ad_delivery_decisions;
DROP TABLE IF EXISTS ad_creative_reviews;
DROP TABLE IF EXISTS ad_targeting_rules;
DROP TABLE IF EXISTS ad_flights;
DROP TABLE IF EXISTS ad_slots;

