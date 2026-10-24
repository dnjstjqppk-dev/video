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

-- +goose Down
DROP TABLE IF EXISTS visibility_rules;
DROP TABLE IF EXISTS playlist_items;
DROP TABLE IF EXISTS playlists;
DROP TABLE IF EXISTS chapters;
DROP TABLE IF EXISTS captions;
DROP TABLE IF EXISTS thumbnails;
DROP TABLE IF EXISTS media_jobs;
DROP TABLE IF EXISTS upload_parts;
DROP TABLE IF EXISTS upload_sessions;
DROP TABLE IF EXISTS video_variants;
DROP TABLE IF EXISTS video_assets;
DROP TABLE IF EXISTS videos;
DROP TABLE IF EXISTS channel_members;
DROP TABLE IF EXISTS channels;

