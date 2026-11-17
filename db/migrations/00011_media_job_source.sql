-- +goose Up
-- ==========================================
-- 00011: Media job source reference (P1-13 durable dispatch)
-- Idempotent: safe to re-apply. Never edits earlier migrations.
-- ==========================================

-- Canonical transcode source carried by each job: a file:// URI for
-- local-relay uploads or a raw R2 object key (presigned fresh at every
-- dispatch attempt so restarts never reuse an expired URL).
ALTER TABLE media_jobs ADD COLUMN IF NOT EXISTS source_ref TEXT;

-- +goose Down
ALTER TABLE media_jobs DROP COLUMN IF EXISTS source_ref;
