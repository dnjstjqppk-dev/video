-- Generated from ordered migrations; use cmd/migrate for real deployments.
-- Phase 1: identity, organizations and access control. Credentials live in the OIDC provider.
CREATE TABLE users (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 oidc_issuer text NOT NULL, oidc_subject text NOT NULL,
 verified_email text, display_name varchar(120) NOT NULL DEFAULT '',
 locale varchar(12) NOT NULL DEFAULT 'ko-KR',
 status varchar(20) NOT NULL DEFAULT 'active' CHECK(status IN ('active','suspended','closed')),
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 deleted_at timestamptz,
 UNIQUE(oidc_issuer,oidc_subject)
);
CREATE INDEX users_email_idx ON users(lower(verified_email)) WHERE verified_email IS NOT NULL;
CREATE TABLE user_preferences (
 user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 autoplay boolean NOT NULL DEFAULT true, history_enabled boolean NOT NULL DEFAULT true,
 personalized_ads boolean NOT NULL DEFAULT false,
 theme varchar(12) NOT NULL DEFAULT 'system',
 content_language varchar(12) NOT NULL DEFAULT 'ko',
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE user_devices (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL REFERENCES users(id),
 device_fingerprint_hash text NOT NULL, platform varchar(30) NOT NULL,
 push_token_ciphertext bytea, last_seen_at timestamptz, revoked_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(user_id,device_fingerprint_hash)
);
CREATE TABLE organizations (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name varchar(160) NOT NULL,
 kind varchar(30) NOT NULL CHECK(kind IN ('advertiser','mcn','rights_owner','merchant','agency','internal')),
 verification_status varchar(24) NOT NULL DEFAULT 'pending' CHECK(verification_status IN ('pending','verified','rejected','suspended')),
 created_by uuid NOT NULL REFERENCES users(id), created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz
);
CREATE TABLE organization_members (
 organization_id uuid NOT NULL REFERENCES organizations(id), user_id uuid NOT NULL REFERENCES users(id),
 role varchar(40) NOT NULL CHECK(role IN ('owner','admin','finance','manager','viewer')),
 joined_at timestamptz NOT NULL DEFAULT now(), revoked_at timestamptz,
 PRIMARY KEY(organization_id,user_id)
);
CREATE TABLE roles (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), code varchar(60) NOT NULL UNIQUE, description text NOT NULL DEFAULT '');
CREATE TABLE permissions (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), code varchar(100) NOT NULL UNIQUE);
CREATE TABLE role_permissions (role_id uuid NOT NULL REFERENCES roles(id),permission_id uuid NOT NULL REFERENCES permissions(id),PRIMARY KEY(role_id,permission_id));
CREATE TABLE user_global_roles(user_id uuid NOT NULL REFERENCES users(id),role_id uuid NOT NULL REFERENCES roles(id),granted_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(user_id,role_id));
CREATE TABLE user_security_events(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),user_id uuid REFERENCES users(id),event_code varchar(60) NOT NULL,ip_hash text,metadata jsonb NOT NULL DEFAULT '{}'::jsonb,created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX user_security_events_user_created_idx ON user_security_events(user_id,created_at DESC);


-- Phase 2: channels, video catalog, assets, ingest and publication.
CREATE TABLE channels (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),owner_id uuid NOT NULL REFERENCES users(id),
 handle varchar(30) NOT NULL CHECK(handle ~ '^[a-z0-9_-]{3,30}$'),
 name varchar(200) NOT NULL, description varchar(5000) NOT NULL DEFAULT '',
 avatar_asset_key text,banner_asset_key text,country char(2),
 status varchar(20) NOT NULL DEFAULT 'active' CHECK(status IN ('active','suspended','closed')),
 created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now(),deleted_at timestamptz
);
CREATE UNIQUE INDEX channels_handle_active_unique ON channels(lower(handle)) WHERE deleted_at IS NULL;
CREATE INDEX channels_owner_idx ON channels(owner_id);
CREATE TABLE channel_members(channel_id uuid NOT NULL REFERENCES channels(id),user_id uuid NOT NULL REFERENCES users(id),role varchar(24) NOT NULL CHECK(role IN('owner','manager','editor','analyst','moderator')),created_at timestamptz NOT NULL DEFAULT now(),revoked_at timestamptz,PRIMARY KEY(channel_id,user_id));
CREATE TABLE videos (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), channel_id uuid NOT NULL REFERENCES channels(id),
 title varchar(200) NOT NULL,description varchar(5000) NOT NULL DEFAULT '',
 visibility varchar(20) NOT NULL DEFAULT 'private' CHECK(visibility IN('private','unlisted','public','members','paid')),
 status varchar(24) NOT NULL DEFAULT 'draft' CHECK(status IN('draft','uploading','uploaded','validating','queued','processing','ready','failed','blocked','deleted')),
 kind varchar(12) NOT NULL DEFAULT 'video' CHECK(kind IN ('video','short','live_replay')),
 duration_ms bigint CHECK(duration_ms>=0),original_filename text,
 published_at timestamptz, scheduled_for timestamptz,
 age_restricted boolean NOT NULL DEFAULT false,allow_comments boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now(),deleted_at timestamptz,
 CONSTRAINT public_requires_ready CHECK(visibility NOT IN ('public','unlisted','members','paid') OR status='ready')
);
CREATE INDEX videos_public_idx ON videos(published_at DESC,id DESC) WHERE visibility='public' AND status='ready' AND deleted_at IS NULL;
CREATE INDEX videos_channel_idx ON videos(channel_id,created_at DESC);
CREATE TABLE video_assets (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),video_id uuid NOT NULL REFERENCES videos(id),
 asset_kind varchar(30) NOT NULL CHECK(asset_kind IN ('original','hls_manifest','dash_manifest','thumbnail','caption','audio','preview')),
 storage_provider varchar(25) NOT NULL DEFAULT 'r2',object_key text NOT NULL UNIQUE,content_type text NOT NULL,
 size_bytes bigint NOT NULL CHECK(size_bytes>=0),checksum_sha256 char(64),created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX video_assets_video_idx ON video_assets(video_id);
CREATE TABLE video_variants (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),video_id uuid NOT NULL REFERENCES videos(id),
 manifest_asset_id uuid REFERENCES video_assets(id),height integer CHECK(height>0),width integer CHECK(width>0),
 bitrate_bps integer CHECK(bitrate_bps>0),codec varchar(32) NOT NULL,
 state varchar(20) NOT NULL DEFAULT 'pending',created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(video_id,height,codec)
);
CREATE TABLE upload_sessions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),video_id uuid NOT NULL REFERENCES videos(id),
 user_id uuid NOT NULL REFERENCES users(id),object_key text NOT NULL UNIQUE,
 storage_upload_id text,upload_state varchar(24) NOT NULL DEFAULT 'created',
 expected_size_bytes bigint NOT NULL CHECK(expected_size_bytes>0),expected_sha256 char(64),
 expires_at timestamptz NOT NULL,completed_at timestamptz,created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE upload_parts (upload_id uuid NOT NULL REFERENCES upload_sessions(id),part_number integer NOT NULL CHECK(part_number BETWEEN 1 AND 10000),etag text NOT NULL,size_bytes bigint NOT NULL CHECK(size_bytes>0),PRIMARY KEY(upload_id,part_number));
CREATE TABLE media_jobs (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),video_id uuid NOT NULL REFERENCES videos(id),
 job_type varchar(32) NOT NULL, state varchar(20) NOT NULL DEFAULT 'queued',attempts integer NOT NULL DEFAULT 0,
 max_attempts integer NOT NULL DEFAULT 4,locked_until timestamptz,error_code text,
 created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX media_jobs_state_idx ON media_jobs(state,created_at);
CREATE TABLE video_chapters(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),video_id uuid NOT NULL REFERENCES videos(id),start_ms bigint NOT NULL CHECK(start_ms>=0),title varchar(200) NOT NULL,UNIQUE(video_id,start_ms));
CREATE TABLE captions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),video_id uuid NOT NULL REFERENCES videos(id),language varchar(16) NOT NULL,kind varchar(16) NOT NULL,asset_id uuid REFERENCES video_assets(id),reviewed boolean NOT NULL DEFAULT false,UNIQUE(video_id,language,kind));
CREATE TABLE video_territories(video_id uuid NOT NULL REFERENCES videos(id),country char(2) NOT NULL,allowed boolean NOT NULL,PRIMARY KEY(video_id,country));


-- Phase 2/3: interactions, history, community, playlists.
CREATE TABLE channel_subscriptions(user_id uuid NOT NULL REFERENCES users(id),channel_id uuid NOT NULL REFERENCES channels(id),notification_level varchar(15) NOT NULL DEFAULT 'personalized',created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(user_id,channel_id));
CREATE INDEX channel_subscriptions_channel_idx ON channel_subscriptions(channel_id);
CREATE TABLE video_reactions(user_id uuid NOT NULL REFERENCES users(id),video_id uuid NOT NULL REFERENCES videos(id),reaction varchar(10) NOT NULL CHECK(reaction IN ('like','dislike')),created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(user_id,video_id));
CREATE TABLE comments(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),video_id uuid NOT NULL REFERENCES videos(id),author_id uuid NOT NULL REFERENCES users(id),parent_id uuid REFERENCES comments(id),body varchar(10000) NOT NULL,status varchar(20) NOT NULL DEFAULT 'visible',is_pinned boolean NOT NULL DEFAULT false,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now(),deleted_at timestamptz);
CREATE INDEX comments_video_idx ON comments(video_id,created_at DESC);
CREATE TABLE comment_reactions(user_id uuid NOT NULL REFERENCES users(id),comment_id uuid NOT NULL REFERENCES comments(id),reaction varchar(10) NOT NULL CHECK(reaction IN('like','dislike')),PRIMARY KEY(user_id,comment_id));
CREATE TABLE playlists(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),owner_id uuid NOT NULL REFERENCES users(id),channel_id uuid REFERENCES channels(id),title varchar(200) NOT NULL,visibility varchar(16) NOT NULL DEFAULT 'private' CHECK(visibility IN('private','unlisted','public')),created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE playlist_items(playlist_id uuid NOT NULL REFERENCES playlists(id),video_id uuid NOT NULL REFERENCES videos(id),position integer NOT NULL CHECK(position>=0),added_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(playlist_id,video_id),UNIQUE(playlist_id,position));
CREATE TABLE watch_progress(user_id uuid NOT NULL REFERENCES users(id),video_id uuid NOT NULL REFERENCES videos(id),position_ms bigint NOT NULL DEFAULT 0 CHECK(position_ms>=0),updated_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(user_id,video_id));
CREATE TABLE watch_history(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),user_id uuid NOT NULL REFERENCES users(id),video_id uuid NOT NULL REFERENCES videos(id),watched_at timestamptz NOT NULL DEFAULT now(),duration_ms bigint NOT NULL DEFAULT 0 CHECK(duration_ms>=0));
CREATE INDEX watch_history_user_idx ON watch_history(user_id,watched_at DESC);
CREATE TABLE notifications(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),user_id uuid NOT NULL REFERENCES users(id),type varchar(40) NOT NULL,payload jsonb NOT NULL DEFAULT '{}'::jsonb,read_at timestamptz,created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX notifications_unread_idx ON notifications(user_id,created_at DESC) WHERE read_at IS NULL;
CREATE TABLE community_posts(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),channel_id uuid NOT NULL REFERENCES channels(id),author_id uuid NOT NULL REFERENCES users(id),body varchar(10000) NOT NULL,media_asset_key text,status varchar(16) NOT NULL DEFAULT 'visible',created_at timestamptz NOT NULL DEFAULT now(),deleted_at timestamptz);
CREATE TABLE community_votes(post_id uuid NOT NULL REFERENCES community_posts(id),user_id uuid NOT NULL REFERENCES users(id),option_index smallint NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(post_id,user_id));


-- Phase 4: live control & websocket messages. Managed ingest is controlled by Go.
CREATE TABLE live_streams(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),channel_id uuid NOT NULL REFERENCES channels(id),title varchar(200) NOT NULL,description text NOT NULL DEFAULT '',state varchar(25) NOT NULL DEFAULT 'scheduled' CHECK(state IN('scheduled','ready','live','ended','failed','blocked')),provider varchar(30) NOT NULL DEFAULT 'cloudflare',provider_stream_id text UNIQUE,stream_key_secret_ref text,planned_start_at timestamptz,started_at timestamptz,ended_at timestamptz,replay_video_id uuid REFERENCES videos(id),created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX live_streams_channel_idx ON live_streams(channel_id,created_at DESC);
CREATE TABLE live_sessions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),stream_id uuid NOT NULL REFERENCES live_streams(id),started_at timestamptz NOT NULL DEFAULT now(),ended_at timestamptz,provider_event_id text UNIQUE,disconnects integer NOT NULL DEFAULT 0);
CREATE TABLE live_chat_messages(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),stream_id uuid NOT NULL REFERENCES live_streams(id),author_id uuid NOT NULL REFERENCES users(id),body varchar(2000) NOT NULL,status varchar(16) NOT NULL DEFAULT 'visible',sent_at timestamptz NOT NULL DEFAULT now(),deleted_at timestamptz);
CREATE INDEX live_chat_messages_stream_idx ON live_chat_messages(stream_id,sent_at DESC);
CREATE TABLE live_chat_moderators(stream_id uuid NOT NULL REFERENCES live_streams(id),user_id uuid NOT NULL REFERENCES users(id),created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(stream_id,user_id));
CREATE TABLE live_chat_bans(stream_id uuid NOT NULL REFERENCES live_streams(id),user_id uuid NOT NULL REFERENCES users(id),reason text NOT NULL DEFAULT '',expires_at timestamptz,created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(stream_id,user_id));


-- Phase 3/6: metadata for recommendation and analytics; high-volume events live in ClickHouse.
CREATE TABLE analytics_rollups_daily(video_id uuid NOT NULL REFERENCES videos(id),metric_date date NOT NULL,views bigint NOT NULL DEFAULT 0 CHECK(views>=0),watch_ms bigint NOT NULL DEFAULT 0 CHECK(watch_ms>=0),impressions bigint NOT NULL DEFAULT 0 CHECK(impressions>=0),likes bigint NOT NULL DEFAULT 0 CHECK(likes>=0),updated_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(video_id,metric_date));
CREATE TABLE channel_analytics_daily(channel_id uuid NOT NULL REFERENCES channels(id),metric_date date NOT NULL,subscribers_total bigint NOT NULL DEFAULT 0,views bigint NOT NULL DEFAULT 0,watch_ms bigint NOT NULL DEFAULT 0,updated_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(channel_id,metric_date));
CREATE TABLE video_tags(video_id uuid NOT NULL REFERENCES videos(id),tag varchar(60) NOT NULL,PRIMARY KEY(video_id,tag));
CREATE TABLE recommendation_feedback(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),user_id uuid REFERENCES users(id),video_id uuid NOT NULL REFERENCES videos(id),feedback varchar(24) NOT NULL CHECK(feedback IN('not_interested','dont_recommend_channel','liked','saved')),created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE recommendation_experiments(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),name varchar(120) NOT NULL UNIQUE,config jsonb NOT NULL DEFAULT '{}'::jsonb,active boolean NOT NULL DEFAULT false,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE search_index_jobs(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),video_id uuid REFERENCES videos(id),channel_id uuid REFERENCES channels(id),action varchar(16) NOT NULL CHECK(action IN('upsert','delete')),state varchar(16) NOT NULL DEFAULT 'queued',created_at timestamptz NOT NULL DEFAULT now(),CHECK(video_id IS NOT NULL OR channel_id IS NOT NULL));


-- Phase 5: ad campaigns, review, budgets, delivery and fraud prevention.
CREATE TABLE advertiser_accounts(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),organization_id uuid NOT NULL UNIQUE REFERENCES organizations(id),billing_currency char(3) NOT NULL DEFAULT 'KRW',state varchar(20) NOT NULL DEFAULT 'pending',created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE ad_campaigns(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),advertiser_id uuid NOT NULL REFERENCES advertiser_accounts(id),name varchar(200) NOT NULL,objective varchar(40) NOT NULL,pricing_model varchar(12) NOT NULL CHECK(pricing_model IN('CPM','CPC','CPV')),daily_budget_minor bigint NOT NULL CHECK(daily_budget_minor>=0),total_budget_minor bigint NOT NULL CHECK(total_budget_minor>=0),currency char(3) NOT NULL,start_at timestamptz NOT NULL,end_at timestamptz NOT NULL,state varchar(20) NOT NULL DEFAULT 'draft',created_at timestamptz NOT NULL DEFAULT now(),CHECK(end_at>start_at));
CREATE INDEX ad_campaigns_advertiser_idx ON ad_campaigns(advertiser_id,state);
CREATE TABLE ad_creatives(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),campaign_id uuid NOT NULL REFERENCES ad_campaigns(id),video_id uuid REFERENCES videos(id),asset_key text,landing_url text,review_state varchar(20) NOT NULL DEFAULT 'pending',reviewed_by uuid REFERENCES users(id),reviewed_at timestamptz,created_at timestamptz NOT NULL DEFAULT now(),CHECK(video_id IS NOT NULL OR asset_key IS NOT NULL));
CREATE TABLE ad_targeting_rules(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),campaign_id uuid NOT NULL REFERENCES ad_campaigns(id),rule_type varchar(32) NOT NULL,rule_json jsonb NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE ad_placements(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),code varchar(60) NOT NULL UNIQUE,format varchar(40) NOT NULL,active boolean NOT NULL DEFAULT true);
CREATE TABLE ad_budget_reservations(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),campaign_id uuid NOT NULL REFERENCES ad_campaigns(id),request_id uuid NOT NULL UNIQUE,reserved_minor bigint NOT NULL CHECK(reserved_minor>=0),settled_minor bigint NOT NULL DEFAULT 0 CHECK(settled_minor>=0),status varchar(20) NOT NULL DEFAULT 'reserved',expires_at timestamptz NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE ad_delivery_events(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),campaign_id uuid NOT NULL REFERENCES ad_campaigns(id),creative_id uuid NOT NULL REFERENCES ad_creatives(id),placement_id uuid REFERENCES ad_placements(id),event_kind varchar(20) NOT NULL CHECK(event_kind IN('request','impression','view','click','complete','skip','invalid')),event_id uuid NOT NULL UNIQUE,invalid_reason text,occurred_at timestamptz NOT NULL,received_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX ad_delivery_events_campaign_idx ON ad_delivery_events(campaign_id,occurred_at DESC);
CREATE TABLE ad_invoices(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),advertiser_id uuid NOT NULL REFERENCES advertiser_accounts(id),period_start date NOT NULL,period_end date NOT NULL,currency char(3) NOT NULL,amount_minor bigint NOT NULL CHECK(amount_minor>=0),state varchar(20) NOT NULL DEFAULT 'draft',created_at timestamptz NOT NULL DEFAULT now(),CHECK(period_end>=period_start));


-- Phase 5: earnings, double-entry finance, premium, refunds & payout controls.
CREATE TABLE monetization_accounts(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),channel_id uuid NOT NULL UNIQUE REFERENCES channels(id),state varchar(20) NOT NULL DEFAULT 'pending',verified_at timestamptz,paused_at timestamptz,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE premium_plans(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),code varchar(50) NOT NULL UNIQUE,currency char(3) NOT NULL,price_minor bigint NOT NULL CHECK(price_minor>=0),billing_interval varchar(12) NOT NULL CHECK(billing_interval IN('month','year')),active boolean NOT NULL DEFAULT false);
CREATE TABLE user_premium_subscriptions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),user_id uuid NOT NULL REFERENCES users(id),plan_id uuid NOT NULL REFERENCES premium_plans(id),provider varchar(30) NOT NULL,provider_subscription_id text UNIQUE,status varchar(20) NOT NULL,period_start timestamptz NOT NULL,period_end timestamptz NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE channel_membership_plans(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),channel_id uuid NOT NULL REFERENCES channels(id),name varchar(120) NOT NULL,price_minor bigint NOT NULL CHECK(price_minor>0),currency char(3) NOT NULL,active boolean NOT NULL DEFAULT true);
CREATE TABLE channel_memberships(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),user_id uuid NOT NULL REFERENCES users(id),plan_id uuid NOT NULL REFERENCES channel_membership_plans(id),status varchar(20) NOT NULL,started_at timestamptz NOT NULL DEFAULT now(),expires_at timestamptz);
CREATE TABLE payment_transactions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),user_id uuid REFERENCES users(id),provider varchar(30) NOT NULL,provider_reference text UNIQUE,kind varchar(30) NOT NULL,amount_minor bigint NOT NULL CHECK(amount_minor>=0),currency char(3) NOT NULL,status varchar(20) NOT NULL DEFAULT 'pending',idempotency_key varchar(160) UNIQUE,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE payment_webhook_events(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),provider varchar(30) NOT NULL,provider_event_id text NOT NULL,payload_hash char(64) NOT NULL,processed_at timestamptz,received_at timestamptz NOT NULL DEFAULT now(),UNIQUE(provider,provider_event_id));
CREATE TABLE refunds(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),payment_id uuid NOT NULL REFERENCES payment_transactions(id),provider_ref text UNIQUE,amount_minor bigint NOT NULL CHECK(amount_minor>0),reason text NOT NULL,status varchar(20) NOT NULL DEFAULT 'requested',created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE ledger_accounts(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),owner_user_id uuid REFERENCES users(id),owner_org_id uuid REFERENCES organizations(id),owner_channel_id uuid REFERENCES channels(id),code varchar(120) NOT NULL UNIQUE,currency char(3) NOT NULL,account_type varchar(20) NOT NULL CHECK(account_type IN('asset','liability','revenue','expense','equity')),created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE ledger_transactions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),external_ref text UNIQUE,source_type varchar(40) NOT NULL,description text NOT NULL DEFAULT '',posted_at timestamptz NOT NULL DEFAULT now(),created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE ledger_entries(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),transaction_id uuid NOT NULL REFERENCES ledger_transactions(id),account_id uuid NOT NULL REFERENCES ledger_accounts(id),direction varchar(6) NOT NULL CHECK(direction IN('debit','credit')),amount_minor bigint NOT NULL CHECK(amount_minor>0),created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX ledger_entries_tx_idx ON ledger_entries(transaction_id);
-- Ledger balancing MUST be enforced by the Go settlement transaction with row locking,
-- plus deferred database validation/immutable-write protection before finance GO.
CREATE TABLE creator_earnings(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),channel_id uuid NOT NULL REFERENCES channels(id),source varchar(35) NOT NULL,source_ref text NOT NULL UNIQUE,amount_minor bigint NOT NULL,currency char(3) NOT NULL,status varchar(20) NOT NULL DEFAULT 'pending',ledger_transaction_id uuid REFERENCES ledger_transactions(id),earned_at timestamptz NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE payout_accounts(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),organization_id uuid REFERENCES organizations(id),user_id uuid REFERENCES users(id),provider varchar(30) NOT NULL,provider_account_token_ref text NOT NULL,verified boolean NOT NULL DEFAULT false,created_at timestamptz NOT NULL DEFAULT now(),CHECK((organization_id IS NULL) <> (user_id IS NULL)));
CREATE TABLE payouts(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),payee_account_id uuid NOT NULL REFERENCES payout_accounts(id),amount_minor bigint NOT NULL CHECK(amount_minor>0),currency char(3) NOT NULL,state varchar(20) NOT NULL DEFAULT 'pending',provider_payout_ref text UNIQUE,approved_by uuid REFERENCES users(id),approved_at timestamptz,created_at timestamptz NOT NULL DEFAULT now());


-- Phase 5/6: multi-tenant contracts, IP claims, disputes and ownership evidence.
CREATE TABLE partner_contracts(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),organization_id uuid NOT NULL REFERENCES organizations(id),channel_id uuid REFERENCES channels(id),contract_type varchar(30) NOT NULL,version integer NOT NULL DEFAULT 1,terms_hash char(64) NOT NULL,document_key text NOT NULL,starts_at timestamptz NOT NULL,ends_at timestamptz,signature_state varchar(24) NOT NULL DEFAULT 'pending',created_at timestamptz NOT NULL DEFAULT now(),CHECK(ends_at IS NULL OR ends_at>starts_at));
CREATE TABLE channel_partner_links(channel_id uuid NOT NULL REFERENCES channels(id),organization_id uuid NOT NULL REFERENCES organizations(id),contract_id uuid NOT NULL REFERENCES partner_contracts(id),share_basis_points integer NOT NULL DEFAULT 0 CHECK(share_basis_points BETWEEN 0 AND 10000),starts_at timestamptz NOT NULL,ends_at timestamptz,PRIMARY KEY(channel_id,organization_id,contract_id));
CREATE TABLE rights_assets(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),organization_id uuid NOT NULL REFERENCES organizations(id),title varchar(300) NOT NULL,original_asset_key text NOT NULL,sha256 char(64),fingerprint_ref text,verified_status varchar(24) NOT NULL DEFAULT 'pending',created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX rights_assets_sha_idx ON rights_assets(sha256) WHERE sha256 IS NOT NULL;
CREATE TABLE rights_territories(asset_id uuid NOT NULL REFERENCES rights_assets(id),country char(2) NOT NULL,rights_kind varchar(32) NOT NULL,valid_from date NOT NULL,valid_until date,PRIMARY KEY(asset_id,country,rights_kind,valid_from));
CREATE TABLE rights_evidence(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),asset_id uuid NOT NULL REFERENCES rights_assets(id),storage_key text NOT NULL,submitted_by uuid NOT NULL REFERENCES users(id),created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE copyright_claims(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),asset_id uuid NOT NULL REFERENCES rights_assets(id),video_id uuid NOT NULL REFERENCES videos(id),claimant_org_id uuid NOT NULL REFERENCES organizations(id),claim_kind varchar(24) NOT NULL CHECK(claim_kind IN('block','monetize','track')),status varchar(24) NOT NULL DEFAULT 'pending',reviewed_by uuid REFERENCES users(id),created_at timestamptz NOT NULL DEFAULT now(),resolved_at timestamptz);
CREATE INDEX copyright_claims_video_idx ON copyright_claims(video_id,status);
CREATE TABLE copyright_disputes(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),claim_id uuid NOT NULL REFERENCES copyright_claims(id),opened_by uuid NOT NULL REFERENCES users(id),reason text NOT NULL,status varchar(24) NOT NULL DEFAULT 'open',decided_by uuid REFERENCES users(id),decision_reason text,created_at timestamptz NOT NULL DEFAULT now(),resolved_at timestamptz);
CREATE TABLE dispute_evidence(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),dispute_id uuid NOT NULL REFERENCES copyright_disputes(id),uploader_id uuid NOT NULL REFERENCES users(id),storage_key text NOT NULL,created_at timestamptz NOT NULL DEFAULT now());


-- Phase 6: community safety, policy versions, immutable decision chronology.
CREATE TABLE policies(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),code varchar(80) NOT NULL,version integer NOT NULL,body text NOT NULL,effective_at timestamptz NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),UNIQUE(code,version));
CREATE TABLE content_reports(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),reporter_id uuid REFERENCES users(id),video_id uuid REFERENCES videos(id),comment_id uuid REFERENCES comments(id),live_id uuid REFERENCES live_streams(id),channel_id uuid REFERENCES channels(id),reason_code varchar(60) NOT NULL,details text NOT NULL DEFAULT '',status varchar(24) NOT NULL DEFAULT 'open',created_at timestamptz NOT NULL DEFAULT now(),CHECK(num_nonnulls(video_id,comment_id,live_id,channel_id)=1));
CREATE TABLE moderation_cases(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),report_id uuid REFERENCES content_reports(id),assigned_to uuid REFERENCES users(id),priority smallint NOT NULL DEFAULT 3,status varchar(24) NOT NULL DEFAULT 'open',created_at timestamptz NOT NULL DEFAULT now(),closed_at timestamptz);
CREATE TABLE moderation_decisions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),case_id uuid NOT NULL REFERENCES moderation_cases(id),policy_id uuid REFERENCES policies(id),actor_id uuid NOT NULL REFERENCES users(id),action varchar(60) NOT NULL,reason text NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE moderation_appeals(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),decision_id uuid NOT NULL REFERENCES moderation_decisions(id),appellant_id uuid NOT NULL REFERENCES users(id),reason text NOT NULL,status varchar(24) NOT NULL DEFAULT 'open',reviewed_by uuid REFERENCES users(id),created_at timestamptz NOT NULL DEFAULT now(),resolved_at timestamptz);
CREATE TABLE user_sanctions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),user_id uuid NOT NULL REFERENCES users(id),decision_id uuid NOT NULL REFERENCES moderation_decisions(id),sanction_type varchar(40) NOT NULL,starts_at timestamptz NOT NULL,ends_at timestamptz,revoked_at timestamptz);
CREATE TABLE blocked_terms(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),scope varchar(30) NOT NULL,channel_id uuid REFERENCES channels(id),term varchar(160) NOT NULL,created_at timestamptz NOT NULL DEFAULT now());


-- Phase 0-7: auditable operations, outbox, flags and rollout tracking.
CREATE TABLE audit_logs(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),actor_user_id uuid REFERENCES users(id),actor_service varchar(80),organization_id uuid REFERENCES organizations(id),action varchar(120) NOT NULL,object_type varchar(60) NOT NULL,object_id uuid,request_id text,details jsonb NOT NULL DEFAULT '{}'::jsonb,created_at timestamptz NOT NULL DEFAULT now(),CHECK(actor_user_id IS NOT NULL OR actor_service IS NOT NULL));
CREATE INDEX audit_logs_created_idx ON audit_logs(created_at DESC);
CREATE TABLE admin_approvals(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),action varchar(80) NOT NULL,object_type varchar(60) NOT NULL,object_id uuid NOT NULL,requested_by uuid NOT NULL REFERENCES users(id),approved_by uuid REFERENCES users(id),decision varchar(20) NOT NULL DEFAULT 'pending',created_at timestamptz NOT NULL DEFAULT now(),decided_at timestamptz,CHECK(approved_by IS NULL OR approved_by<>requested_by));
CREATE TABLE feature_flags(key varchar(120) PRIMARY KEY,enabled boolean NOT NULL DEFAULT false,rollout_percentage integer NOT NULL DEFAULT 0 CHECK(rollout_percentage BETWEEN 0 AND 100),rules jsonb NOT NULL DEFAULT '{}'::jsonb,updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE outbox_events(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),aggregate_type varchar(60) NOT NULL,aggregate_id uuid NOT NULL,event_type varchar(100) NOT NULL,payload jsonb NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),published_at timestamptz,retry_count integer NOT NULL DEFAULT 0);
CREATE INDEX outbox_unpublished_idx ON outbox_events(created_at) WHERE published_at IS NULL;
CREATE TABLE inbox_events(consumer_name varchar(100) NOT NULL,event_id uuid NOT NULL,processed_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(consumer_name,event_id));
CREATE TABLE idempotency_keys(key varchar(160) NOT NULL,scope varchar(100) NOT NULL,request_hash char(64) NOT NULL,response_code integer,result_ref text,expires_at timestamptz NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(scope,key));
CREATE TABLE system_incidents(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),severity varchar(12) NOT NULL CHECK(severity IN('P0','P1','P2','P3')),title varchar(200) NOT NULL,details text,status varchar(24) NOT NULL DEFAULT 'open',opened_at timestamptz NOT NULL DEFAULT now(),closed_at timestamptz);
CREATE TABLE deployment_records(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),service varchar(80) NOT NULL,environment varchar(20) NOT NULL,commit_sha char(40) NOT NULL,image_digest text,deployed_at timestamptz NOT NULL DEFAULT now(),result varchar(20) NOT NULL);

