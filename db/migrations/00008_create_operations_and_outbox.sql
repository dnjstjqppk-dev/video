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

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS admin_actions;

