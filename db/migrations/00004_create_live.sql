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

-- +goose Down
DROP TABLE IF EXISTS live_moderators;
DROP TABLE IF EXISTS live_recordings;
DROP TABLE IF EXISTS live_chat_messages;
DROP TABLE IF EXISTS live_inputs;
DROP TABLE IF EXISTS live_streams;

