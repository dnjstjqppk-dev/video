-- 00012: vravra studio completion — custom thumbnails, auto frames,
-- end-screen/card elements, hourly channel views for realtime analytics.
-- videos.scheduled_at already exists (00002); the API now reads/writes it.

ALTER TABLE videos ADD COLUMN IF NOT EXISTS custom_thumbnail_url TEXT;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS auto_frame_urls TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS elements JSONB NOT NULL DEFAULT '[]';

CREATE TABLE IF NOT EXISTS channel_hourly_views (
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    hour TIMESTAMPTZ NOT NULL,
    views BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, hour)
);

CREATE INDEX IF NOT EXISTS idx_channel_hourly_views_hour
    ON channel_hourly_views(hour DESC);
CREATE INDEX IF NOT EXISTS idx_videos_scheduled
    ON videos(scheduled_at) WHERE visibility = 'scheduled';
