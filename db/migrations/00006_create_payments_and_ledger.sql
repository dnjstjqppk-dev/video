-- +goose Up
-- ==========================================
-- 00006: Monetization, Billing, Double-Entry Ledger, Payouts
-- ==========================================

CREATE TABLE IF NOT EXISTS monetization_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE UNIQUE,
    bank_name VARCHAR(64) NOT NULL,
    account_number_encrypted TEXT NOT NULL,
    account_holder_name VARCHAR(128) NOT NULL,
    tax_identification_number VARCHAR(64),
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending, active, suspended, rejected
    payout_currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    tier_name VARCHAR(64) NOT NULL,
    monthly_price_cents BIGINT NOT NULL,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    perks TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscriptions_billing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier VARCHAR(32) NOT NULL, -- premium, premium_family, channel_membership
    channel_id UUID REFERENCES channels(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active', -- active, past_due, canceled, paused
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    external_subscription_id VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    provider VARCHAR(32) NOT NULL, -- toss, stripe
    external_transaction_id VARCHAR(128) NOT NULL UNIQUE,
    amount_cents BIGINT NOT NULL,
    fee_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    status VARCHAR(32) NOT NULL, -- pending, succeeded, failed, refunded, partially_refunded
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    payment_method VARCHAR(32) NOT NULL, -- card, virtual_account, easy_pay
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_transactions_user ON payment_transactions(user_id);

CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES payment_transactions(id) ON DELETE RESTRICT,
    refund_amount_cents BIGINT NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Strict Double-Entry Ledger System (Section 14.3)
CREATE TABLE IF NOT EXISTS ledger_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_number VARCHAR(64) NOT NULL UNIQUE,
    account_name VARCHAR(128) NOT NULL,
    account_type VARCHAR(32) NOT NULL, -- asset, liability, equity, revenue, expense
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    balance_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Essential System Ledger Accounts
INSERT INTO ledger_accounts (account_number, account_name, account_type, currency) VALUES
    ('1000-CASH', 'Platform Operating Cash', 'asset', 'KRW'),
    ('2000-CREATOR-PAYABLE', 'Creator Earnings Payable', 'liability', 'KRW'),
    ('2100-MCN-PAYABLE', 'MCN Share Payable', 'liability', 'KRW'),
    ('2200-RIGHTS-PAYABLE', 'Copyright Royalties Payable', 'liability', 'KRW'),
    ('4000-AD-REVENUE', 'Gross Ad Revenue', 'revenue', 'KRW'),
    ('4100-SUB-REVENUE', 'Premium Subscription Revenue', 'revenue', 'KRW'),
    ('5000-PG-FEE', 'Payment Gateway Processing Fees', 'expense', 'KRW')
ON CONFLICT (account_number) DO NOTHING;

CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    account_id UUID NOT NULL REFERENCES ledger_accounts(id) ON DELETE RESTRICT,
    debit_cents BIGINT NOT NULL DEFAULT 0,
    credit_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_transaction ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_entries_account ON ledger_entries(account_id, created_at);

CREATE TABLE IF NOT EXISTS creator_earnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    period_date DATE NOT NULL,
    ad_revenue_cents BIGINT NOT NULL DEFAULT 0,
    membership_revenue_cents BIGINT NOT NULL DEFAULT 0,
    donation_revenue_cents BIGINT NOT NULL DEFAULT 0,
    platform_fee_cents BIGINT NOT NULL DEFAULT 0,
    net_earnings_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (channel_id, period_date)
);

CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    organization_id UUID REFERENCES organizations(id) ON DELETE RESTRICT,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    gross_amount_cents BIGINT NOT NULL,
    tax_withheld_cents BIGINT NOT NULL DEFAULT 0,
    net_amount_cents BIGINT NOT NULL,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending, approved, processing, paid, failed, on_hold
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    payout_batch_id VARCHAR(128),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES advertisers(id) ON DELETE RESTRICT,
    invoice_number VARCHAR(64) NOT NULL UNIQUE,
    amount_cents BIGINT NOT NULL,
    tax_cents BIGINT NOT NULL,
    total_cents BIGINT NOT NULL,
    currency VARCHAR(4) NOT NULL DEFAULT 'KRW',
    due_date DATE NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'issued', -- issued, paid, overdue, canceled
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS payouts;
DROP TABLE IF EXISTS creator_earnings;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS ledger_accounts;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_transactions;
DROP TABLE IF EXISTS subscriptions_billing;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS monetization_accounts;

