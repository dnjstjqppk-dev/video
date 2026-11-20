-- +goose Up
-- ==========================================
-- 00014: AI Originals (AI home ground) — series + AI review columns
-- ==========================================
-- A video belongs to at most one series; membership is denormalized onto
-- videos(series_id, season_no, episode_no) so feeds and playback enrich
-- without joins. AI review gates the Originals shelf only: pending and
-- rejected AI videos stay visible under their normal visibility rule.

ALTER TABLE videos ADD COLUMN IF NOT EXISTS ai_generated BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS ai_tool TEXT;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS ai_review_status TEXT NOT NULL DEFAULT '';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS ai_review_note TEXT;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS series_id UUID;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS season_no INT NOT NULL DEFAULT 0;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS episode_no INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_url TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'ongoing', -- ongoing, completed
    ai_original BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_series_channel ON series(channel_id);
CREATE INDEX IF NOT EXISTS idx_videos_series ON videos(series_id, season_no, episode_no) WHERE series_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_videos_ai_review ON videos(ai_review_status) WHERE ai_generated = TRUE;
