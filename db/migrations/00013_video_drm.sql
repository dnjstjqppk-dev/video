-- 00013: video DRM protection metadata. Keys are stored as hex for the
-- development ClearKey flow; production must move key custody to a KMS
-- and serve Widevine/FairPlay licenses from a licensed provider.
ALTER TABLE videos ADD COLUMN IF NOT EXISTS drm_protected BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS drm_key_id TEXT;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS drm_key_hex TEXT;
