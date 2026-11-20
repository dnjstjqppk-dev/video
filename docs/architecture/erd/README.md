# Database ERD — actual migration schema (2026-11-20, code-audited)

Authoritative schema: `db/migrations/00001..00014` with **77 tables** and **104 foreign keys**, listed below.
Per-phase Mermaid diagrams are generated deterministically by `python docs/generate_erd.py` — do not hand-edit outputs.

## Phase diagrams

- [00001_create_accounts](./00001_create_accounts.md) — 10 tables, identity/RBAC/orgs
- [00002_create_content](./00002_create_content.md) — 14 tables, catalog/media ingest
- [00003_create_social](./00003_create_social.md) — 9 tables, engagement
- [00004_create_live](./00004_create_live.md) — 5 tables, live
- [00005_create_ads](./00005_create_ads.md) — 6 tables, base advertising
- [00006_create_payments_and_ledger](./00006_create_payments_and_ledger.md) — 10 tables, monetization
- [00007_create_copyright_and_moderation](./00007_create_copyright_and_moderation.md) — 8 tables, rights/moderation
- [00008_create_operations_and_outbox](./00008_create_operations_and_outbox.md) — 5 tables, ops/outbox
- [00009_create_production_ad_engine](./00009_create_production_ad_engine.md) — 8 tables, ad engine
- [00010_forward_roles_and_payments](./00010_forward_roles_and_payments.md) — 0 tables, roles fill + 2 payment columns
- [00011_media_job_source](./00011_media_job_source.md) — 0 tables, `media_jobs.source_ref`
- [00012_vravra_studio_completion](./00012_vravra_studio_completion.md) — 1 table (`channel_hourly_views`) + 3 video columns
- [00013_video_drm](./00013_video_drm.md) — 0 tables, 3 DRM columns on `videos`
- [00014_ai_originals](./00014_ai_originals.md) — 1 table (`series`) + 7 AI/series columns on `videos`

## Actual tables by migration (77 total)

### 00001_create_accounts.sql — 10 tables (identity, RBAC, orgs)

`roles`, `permissions`, `role_permissions`, `users` (UUID PK, role FK, `zitadel_user_id`),
`user_profiles` (display_name/bio/avatar/banner split from users), `user_preferences`,
`sessions`, `devices`, `organizations`, `organization_members`

### 00002_create_content.sql — 14 tables (catalog + media ingest)

`channels`, `channel_members`, `videos`, `video_assets`, `video_variants`,
`upload_sessions`, `upload_parts`, `media_jobs`, `thumbnails`, `captions`,
`chapters`, `playlists`, `playlist_items`, `visibility_rules`

### 00003_create_social.sql — 9 tables (engagement)

`subscriptions`, `video_reactions`, `comments`, `comment_reactions`,
`watch_history`, `watch_progress`, `notifications`, `community_posts`, `community_votes`

### 00004_create_live.sql — 5 tables

`live_streams`, `live_inputs`, `live_chat_messages`, `live_recordings`, `live_moderators`

### 00005_create_ads.sql — 6 tables (base advertising)

`advertisers`, `ad_campaigns`, `ad_creatives`, `ad_placements`, `ad_events`, `ad_budget_allocations`

### 00006_create_payments_and_ledger.sql — 10 tables (monetization)

`monetization_accounts`, `memberships`, `subscriptions_billing`, `payment_transactions`,
`refunds`, `ledger_accounts`, `ledger_entries`, `creator_earnings`, `payouts`, `invoices`

### 00007_create_copyright_and_moderation.sql — 8 tables (rights + trust & safety)

`rights_owners`, `rights_assets`, `rights_territories`, `copyright_claims`,
`copyright_disputes`, `dispute_evidence`, `reports`, `moderation_cases`

### 00008_create_operations_and_outbox.sql — 5 tables (ops)

`admin_actions`, `audit_logs`, `feature_flags`, `outbox_events`, `idempotency_keys`

### 00009_create_production_ad_engine.sql — 8 tables (production ad engine)

`ad_slots`, `ad_flights`, `ad_targeting_rules`, `ad_creative_reviews`,
`ad_delivery_decisions`, `ad_delivery_events`, `ad_frequency_caps`, `ad_reporting_daily`

### 00010_forward_roles_and_payments.sql — 0 tables (P0-1 forward fill)

Roles fill (9 rows: `viewer`, `user`, `premium_member`, `live_moderator`,
`partner_manager`, `rights_holder`, `partner_finance`, `finance_ops`,
`security_auditor`) and `payment_transactions.tx_type` + `checkout_url` columns.

### 00011_media_job_source.sql — 0 tables

`media_jobs.source_ref` column.

### 00012_vravra_studio_completion.sql — 1 table

`channel_hourly_views` (channel FK) + `videos.custom_thumbnail_url`,
`auto_frame_urls`, `elements`.

### 00013_video_drm.sql — 0 tables

`videos.drm_protected`, `drm_key_id`, `drm_key_hex` (dev ClearKey custody;
production must move to KMS + licensed Widevine/FairPlay).

### 00014_ai_originals.sql — 1 table (AI home ground)

`series` (channel FK) + `videos.ai_generated`, `ai_tool`, `ai_review_status`,
`ai_review_note`, `series_id` (denormalized, no FK), `season_no`, `episode_no`.

## Known schema hazards

- ~~`InitSchema()` second-schema conflict~~ — **FIXED 2026-11-17:** the simplified DDL is deleted; the API verifies the
  migrated schema (incl. `00010`) and refuses to boot on mismatch. Users/channels/videos/transactions persist to the real schema.
- `db/schema/schema.sql` and `db/schema/inventory.csv` also describe the stale 88-table schema, not the 77-table one.
