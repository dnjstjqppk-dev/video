-- +goose Up
-- ==========================================
-- 00010: Forward fill for P0-1 persistence (roles + payment columns)
-- Idempotent: safe to re-apply. Never edits 00001..00009.
-- ==========================================

-- Roles referenced by the API role matrix / seed users but missing from 00001.
INSERT INTO roles (name, description) VALUES
    ('viewer', 'Authenticated viewer without a channel'),
    ('user', 'Standard registered consumer (alias)'),
    ('premium_member', 'Paid premium subscriber'),
    ('live_moderator', 'Live chat moderator'),
    ('partner_manager', 'MCN partner manager'),
    ('rights_holder', 'Copyright holder (portal alias)'),
    ('partner_finance', 'Partner finance operator'),
    ('finance_ops', 'Financial operations (alias)'),
    ('security_auditor', 'Read-only security auditor')
ON CONFLICT (name) DO NOTHING;

-- Columns carried by the API payment transaction model.
ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS tx_type VARCHAR(64);
ALTER TABLE payment_transactions ADD COLUMN IF NOT EXISTS checkout_url TEXT;

-- +goose Down
ALTER TABLE payment_transactions DROP COLUMN IF EXISTS checkout_url;
ALTER TABLE payment_transactions DROP COLUMN IF EXISTS tx_type;
DELETE FROM roles WHERE name IN (
    'viewer', 'user', 'premium_member', 'live_moderator', 'partner_manager',
    'rights_holder', 'partner_finance', 'finance_ops', 'security_auditor'
);
