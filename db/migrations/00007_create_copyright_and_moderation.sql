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

-- +goose Down
DROP TABLE IF EXISTS moderation_cases;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS dispute_evidence;
DROP TABLE IF EXISTS copyright_disputes;
DROP TABLE IF EXISTS copyright_claims;
DROP TABLE IF EXISTS rights_territories;
DROP TABLE IF EXISTS rights_assets;
DROP TABLE IF EXISTS rights_owners;

