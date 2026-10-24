# Global Video Platform — Complete `/v1` API Inventory

All HTTP endpoints are versioned under `/v1` and return errors strictly conforming to the standard envelope:

```json
{
  "error": {
    "code": "FORBIDDEN_RESOURCE",
    "message": "Access denied: insufficient role permissions",
    "request_id": "req_01J9X8Z2M4K1",
    "field": "optional_field_name"
  }
}
```

## 1. System Health & Identity (`/healthz`, `/v1/me`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | Public | `visitor`+ | 600/min | Safe | None | Liveness/Readiness check |
| `GET` | `/v1/me` | `openid profile email` | `user`+ | 120/min | Safe | None | Reads current authenticated profile |
| `PATCH` | `/v1/me` | `profile:write` | `user`+ | 30/min | Idempotent | None | Updates user `display_name`, `bio`; emits `user.profile.updated` |
| `GET` | `/v1/me/sessions` | `security:read` | `user`+ | 60/min | Safe | None | Lists active device/IP sessions |
| `GET` | `/v1/me/history` | `history:read` | `user`+ | 60/min | Safe | Cursor/Page | Lists user watch history |
| `DELETE` | `/v1/me/history` | `history:write` | `user`+ | 10/min | Idempotent | None | Clears watch history; emits `user.history.cleared` |

## 2. Channels & Subscriptions (`/v1/channels`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/v1/channels/{id}` | Public | `visitor`+ | 300/min | Safe | None | Resolves channel by UUID or `@handle` |
| `GET` | `/v1/channels/{id}/videos` | Public | `visitor`+ | 300/min | Safe | Page/Limit | Lists channel public VODs & Shorts |
| `POST` | `/v1/channels` | `channel:write` | `user`+ | 10/min | `Idempotency-Key` | None | Creates channel; emits `channel.created` |
| `PATCH` | `/v1/channels/{id}` | `channel:write` | `creator` (Owner) / `super_admin` | 30/min | Idempotent | None | BOLA-checked channel update; emits `channel.updated` |
| `POST` | `/v1/channels/{id}/subscribe` | `social:write` | `user`+ | 60/min | Idempotent | None | Subscribes user & increments counter; emits `channel.subscribed` |
| `DELETE` | `/v1/channels/{id}/subscribe` | `social:write` | `user`+ | 60/min | Idempotent | None | Unsubscribes user; emits `channel.unsubscribed` |

## 3. Videos, Playback & Social Comments (`/v1/videos`, `/v1/comments`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/v1/videos/{id}` | Public | `visitor`+ | 600/min | Safe | None | Returns video metadata |
| `GET` | `/v1/videos/{id}/playback` | Public / Token | `visitor`+ | 600/min | Safe | None | Issues signed `playback_token`, ABR HLS manifest, chapters, VTT captions |
| `PATCH` | `/v1/videos/{id}` | `video:write` | `creator` (Owner) | 60/min | Idempotent | None | Updates title, description, visibility; emits `video.metadata.updated` |
| `DELETE` | `/v1/videos/{id}` | `video:write` | `creator` (Owner) / `super_admin` | 20/min | Idempotent | None | Soft-deletes video (`DELETED`); emits `video.deleted` |
| `POST` | `/v1/videos/{id}/like` | `social:write` | `user`+ | 120/min | Idempotent | None | Records video like; emits `video.liked` |
| `DELETE` | `/v1/videos/{id}/like` | `social:write` | `user`+ | 120/min | Idempotent | None | Removes video like; emits `video.unliked` |
| `GET` | `/v1/videos/{id}/comments` | Public | `visitor`+ | 300/min | Safe | Page/Limit | Lists threaded comments for video |
| `POST` | `/v1/videos/{id}/comments` | `comment:write` | `user`+ | 30/min | `Idempotency-Key` | None | Creates comment; emits `comment.created` |
| `DELETE` | `/v1/comments/{id}` | `comment:write` | `user` (Author) / `moderator` | 30/min | Idempotent | None | Deletes comment; emits `comment.deleted` |

## 4. Resumable Multipart Uploads (`/v1/uploads`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `POST` | `/v1/uploads` | `upload:write` | `creator`+ | 30/min | `Idempotency-Key` | None | Creates R2 multipart session & `CREATED` video record |
| `POST` | `/v1/uploads/{id}/parts` | `upload:write` | `creator` (Owner) | 600/min | Idempotent | None | Generates presigned R2 PUT URL for `partNumber` |
| `POST` | `/v1/uploads/{id}/complete` | `upload:write` | `creator` (Owner) | 30/min | Idempotent | None | Completes R2 multipart upload; emits `video.upload.completed` to NATS `MEDIA_JOBS` |

## 5. Recommendation Feeds & Search (`/v1/feed`, `/v1/search`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/v1/feed/home` | Public / Optional Auth | `visitor`+ | 300/min | Safe | `page`, `limit`, `category` | Personalized or trending VOD recommendations |
| `GET` | `/v1/feed/shorts` | Public / Optional Auth | `visitor`+ | 300/min | Safe | Cursor | Vertical infinite-scroll Shorts feed |
| `GET` | `/v1/feed/subscriptions` | `feed:read` | `user`+ | 180/min | Safe | `page`, `limit` | Chronological feed from subscribed channels |
| `POST` | `/v1/feed/feedback` | `feed:write` | `user`+ | 120/min | Idempotent | None | Records `not_interested` / `dont_recommend_channel` signal |
| `GET` | `/v1/search` | Public | `visitor`+ | 180/min | Safe | `q`, `page` | Queries OpenSearch across videos & channels |
| `GET` | `/v1/search/suggestions` | Public | `visitor`+ | 600/min | Safe | None | Low-latency prefix autocomplete suggestions |

## 6. Live Streaming (`/v1/live-streams`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/v1/live-streams/{id}` | Public | `visitor`+ | 600/min | Safe | None | Returns live status, viewer count, and LL-HLS playback URL |
| `POST` | `/v1/live-streams` | `live:write` | `creator`+ | 15/min | `Idempotency-Key` | None | Provisions Cloudflare Stream RTMPS/SRT ingest & stream key |
| `POST` | `/v1/live-streams/{id}/start` | `live:write` | `creator` (Owner) | 15/min | Idempotent | None | Transitions live stream to `live`; emits `livestream.started` |
| `POST` | `/v1/live-streams/{id}/stop` | `live:write` | `creator` (Owner) | 15/min | Idempotent | None | Ends live broadcast & archives DVR VOD; emits `livestream.ended` |
| `POST` | `/v1/live-streams/{id}/rotate-key` | `live:write` | `creator` (Owner) | 10/min | Idempotent | None | Invalidates old stream key and issues new RTMPS key |

## 7. Creator Studio Console (`/v1/studio`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/v1/studio/dashboard` | `studio:read` | `creator`+ | 120/min | Safe | None | Aggregated channel KPIs & recent uploads |
| `GET` | `/v1/studio/videos` | `studio:read` | `creator`+ | 120/min | Safe | Page/Limit | Lists creator's private/unlisted/public videos |
| `GET` | `/v1/studio/analytics` | `studio:read` | `creator`+ | 120/min | Safe | Date Range | ClickHouse time-series for views, watch time, subscribers, revenue |
| `GET` | `/v1/studio/comments` | `studio:read` | `creator`+ | 120/min | Safe | Page/Limit | Unified comment moderation inbox for creator |
| `GET` | `/v1/studio/earnings` | `studio:finance` | `creator`+ | 60/min | Safe | None | Breakdown of Ad, Membership, and Super Chat revenue |

## 8. Partner Portal (`/v1/partners`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/v1/partners/me` | `partner:read` | `partner_manager`+ | 120/min | Safe | None | Partner organization summary & active contract counts |
| `GET` | `/v1/partners/channels` | `partner:read` | `partner_manager`+ | 120/min | Safe | Page/Limit | Lists MCN-managed channels |
| `GET` | `/v1/partners/contracts` | `partner:read` | `partner_manager`+ | 60/min | Safe | None | Lists revenue share & SLA agreements |
| `GET` | `/v1/partners/campaigns` | `ads:read` | `advertiser` / `partner_manager` | 120/min | Safe | None | Lists VAST/VMAP ad campaigns & pacing metrics |
| `POST` | `/v1/partners/campaigns` | `ads:write` | `advertiser` / `partner_manager` | 30/min | `Idempotency-Key` | None | Creates ad campaign; emits `ad.campaign.created` |
| `GET` | `/v1/partners/settlements` | `finance:read` | `partner_finance`+ | 60/min | Safe | Period | Monthly double-entry payout ledger & tax withholding |
| `GET` | `/v1/partners/rights/claims` | `rights:read` | `rights_holder`+ | 120/min | Safe | Page/Limit | Lists active Content ID copyright claims |
| `POST` | `/v1/partners/rights/claims` | `rights:write` | `rights_holder`+ | 30/min | `Idempotency-Key` | None | Files copyright claim (`monetize`, `track`, `block`); emits `copyright.claim.created` |

## 9. Admin Console & Analytics Ingestion (`/v1/admin`, `/v1/events`)

| Method | URL | Auth Scope | Required Role | Rate Limit | Idempotency | Pagination | Side Effects & Events |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/v1/admin/dashboard` | `admin:read` | `super_admin`, `platform_admin` | 120/min | Safe | None | Platform-wide health, active streams, and revenue KPIs |
| `GET` | `/v1/admin/users` | `admin:users` | `super_admin`, `platform_admin` | 120/min | Safe | Page/Limit | Searches user directory |
| `PATCH` | `/v1/admin/users/{id}/status` | `admin:users:write` | `super_admin`, `platform_admin` | 30/min | Idempotent | None | Suspends/bans/activates user; appends immutable `audit_logs` record |
| `GET` | `/v1/admin/moderation` | `admin:moderation` | `super_admin`, `moderator` | 120/min | Safe | Page/Limit | Lists pending Trust & Safety moderation queue |
| `POST` | `/v1/admin/moderation/{id}/resolve` | `admin:moderation` | `super_admin`, `moderator` | 60/min | Idempotent | None | Resolves moderation case (`approve`, `reject`, `remove`); logs audit event |
| `GET` | `/v1/admin/audit-logs` | `admin:audit` | `super_admin`, `security_auditor` | 60/min | Safe | Page/Limit | Reads append-only administrative audit trail |
| `POST` | `/v1/events/analytics` | Public / Telemetry | `visitor`+ | 1200/min | `event_id` Dedup | None | Asynchronous playback beacon ingestion to NATS & ClickHouse (`202 Accepted`) |
