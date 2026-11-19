# vravra API Spec — Endpoints, Schemas and RBAC

Canonical contract sources: `openapi/gvp-v1.yaml` (119 paths),
`docs/api/ROUTE_CENSUS.md` + `docs/api/routes.json` (router census with auth
guards), and `packages/ts-api-client/src/client.ts` (typed caller used by all
consoles). This file summarizes request/response shape and access rules.

## 1. Conventions

- Base path `/v1`, JSON everywhere, standard error envelope
  `{ code, message, requestId, field? }` mapped to `ApiError` by the client.
- Auth: `Authorization: Bearer <token>` from the `gvp_access_token` cookie;
  `401` clears the session. Mutations through the Nuxt proxy also send the
  double-submit `X-CSRF-Token` header.
- Pagination: `page` (1-based) + `limit`, responses carry `{ items, total }`
  or `{ items, nextCursor }`.
- Money: integer minor units with explicit `currency` (KRW uses won units;
  verify per-endpoint docs before display math).

## 2. Domain endpoints (representative)

| Domain | Endpoints |
| :--- | :--- |
| Identity | `GET /v1/me`, `PATCH /v1/me`, sessions revoke, TOTP `POST /v1/me/mfa/*` |
| Channels | `GET /v1/channels/{id}`, create/update, videos list, subscribe/unsubscribe |
| Videos | `GET /v1/videos/{id}`, playback info, update/delete, like/unlike, captions CRUD |
| Uploads | multipart session create, part URLs, complete, studio edit jobs |
| Feed/search | home feed, shorts feed, subscriptions feed, `GET /v1/search`, suggestions, feedback |
| Social | comments list/add/delete/like/heart/pin, playlists CRUD + items, notifications + read |
| Live | streams CRUD, start/stop, key rotation, active poll, WebSocket chat |
| Studio | dashboard, videos, comments, analytics `?range=7d/28d/90d/365d`, earnings, settlements |
| Ads | decision, events, preview, diagnostics, creative upload, campaign status |
| Partners | overview, campaigns, contracts + revoke, claims, settlements, earnings, payout request |
| Copyright | studio/partner claims list, dispute, create claim |
| Admin | dashboard, users + status sanction, moderation queue + resolve, policies + flags, audit logs, finance summary + settlements + batch approve |
| Ops | service health, analytics event ingest |

## 3. RBAC and ownership rules

- 15 roles from `visitor` to `super_admin` (`packages/shared-types/src/enums.ts`).
- Channel, video, upload and comment mutations verify ownership (BOLA checks).
- `/v1/admin/*` requires platform roles plus an admin IP allowlist; sanctions,
  moderation resolves, policy/flag changes, creative approvals and settlement
  approvals additionally require a 6-digit step-up TOTP code.
- Advertiser routes enforce advertiser role + organization scope; new campaigns
  start as `draft` and creatives as `pending` until reviewed.
- Playback and search enforce visibility: `private` is owner-only, `unlisted`
  is link-only, `members_only` requires entitlement; playback tokens are HMAC
  signed, video-bound and expire after 1 hour.

## 4. Console wiring

Each console proxies `/api/*` to the Go API through `apps/api-proxy`, keeping
tunnel URLs out of console config. `useApi`, `useStudioApi`, `usePartnerApi`
and `useAdminApi` build the shared `ApiClient` with the console cookie, so all
screens listed in `docs/USER_GUIDE.md` read and write real backend state.
