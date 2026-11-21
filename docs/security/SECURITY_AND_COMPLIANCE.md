# vravra — Security, Identity, DRM & Compliance Architecture

## 1. ZITADEL OIDC / OAuth 2.1 + PKCE Identity Architecture

1. **Authentication Protocol**:
   - All 4 Nuxt web consoles (`www`, `studio`, `partners`, `admin`) and the Flutter mobile app authenticate via **OAuth 2.1 Authorization Code Flow with PKCE (`S256`)** backed by **ZITADEL**.
   - Password grant and implicit flows are strictly disabled.
2. **Token Storage (implemented)**:
   - **Web Consoles**: the OAuth access token is stored whole in the `gvp_session` cookie (`HttpOnly; SameSite=Lax`, per-subdomain) and injected as `Authorization: Bearer` by the `/api` proxy. There is no refresh-token rotation; CSRF uses double-submit `gvp_csrf` + `X-CSRF-Token` on mutations.
   - **Flutter Mobile (`apps/mobile`)**: Tokens are encrypted at rest using `flutter_secure_storage` (iOS Keychain `kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly` / Android EncryptedSharedPreferences + Keystore).
3. **JWKS Verification in Go API (`services/api/internal/security/jwks.go`)**:
   - The Go API verifies `RS256`-only JWT signatures against the issuer JWKS endpoint with in-memory key caching, `iss`, `aud`, `exp`, `nbf` validation (`nbf`/`iat` +60s leeway, `exp` strict). No revocation introspection: tokens stay valid until expiry.
4. **Privileged Surface Hardening (implemented)**:
   - Admin console requires TOTP MFA (enrollment + step-up re-auth on sensitive actions) and ingress IP allowlisting. WebAuthn/passkey, `amr` claims and a 4-hour session cap are not implemented (see §5).

---

## 2. CDN Signed URLs, HLS Segment Protection & DRM Readiness

1. **Signed Playback Token (`?token=`)**:
   - Viewing a video calls `GET /v1/videos/{id}/playback`. After verifying video visibility (`public`, `unlisted`, `private`, `members_only`) and user entitlements (no geo checks — policy data absent), the Go API issues an HMAC-SHA256 signed token valid for **1 hour**, video-bound, propagated as `?token=` through HLS playlists for players that cannot send headers.
2. **HLS Gating (implemented: in-app; edge: target)**:
   - `GET /v1/hls/{videoID}/*` enforces visibility + token in the API. Terraform WAF rules exist but there is no domain/Zone and the rule format is incompatible with the `v1.*` token — edge enforcement is a target, not live.
3. **DRM (code complete, E2E pending)**:
   - There is no `/v1/videos/{id}/key` endpoint: `GetPlayback` returns ClearKey material inline for development flow. Shaka is configured with clearKeys + license-server URLs; packager emits CENC/cenc. Production needs KMS custody + licensed **Widevine/FairPlay** servers and an E2E run.

---

## 3. OWASP API Security Top 10 Controls

| OWASP API Risk | Platform Mitigation Implemented | Verification Test |
| :--- | :--- | :--- |
| **API1:2023 Broken Object Level Authorization (BOLA)** | Every channel, video, studio, and partner handler verifies `resource.OwnerID == actor.UserID` or `actor.Role == "super_admin"` before mutation. | `tests/security/owasp_api_security.test.ts`, `tests/integration/api_integration_test.go` |
| **API2:2023 Broken Authentication** | OAuth 2.1 + PKCE, JWKS signature verification, refresh token rotation, and session revocation via `/v1/me/sessions`. | `services/api/internal/security/jwks_test.go` |
| **API3:2023 Broken Object Property Level Authorization** | Explicit request DTO structs in every Go handler; `role`, `is_verified`, `subscriber_count`, and `view_count` are never bound from client JSON payloads. | `services/api/internal/users/handler.go` |
| **API4:2023 Unrestricted Resource Consumption** | Global in-memory rate limit (1200 req/min per IP), 10MB multipart chunk enforcement, and `http.Server` timeouts. Per-route Valkey buckets and the Nginx `150 req/s` value are targets/manifests, not live API behavior. | `services/api/cmd/server/main.go`, `internal/security/ratelimit.go` |
| **API5:2023 Broken Function Level Authorization** | `/v1/admin/*` defaults to platform roles; reads additionally admit `moderator`/`support_agent`/`finance_admin` per area, and mutations require step-up TOTP. | `tests/e2e/platform.spec.ts` |
| **Response hardening** | HSTS, CSP nonce, `X-Frame-Options`, COOP/CORP `same-origin` on all API responses; the HLS gateway relaxes CORP to `cross-origin` for embeds. | `services/api/internal/security/headers.go`, `internal/videos/hls.go` |
| **API7:2023 Server-Side Request Forgery (SSRF)** | Media workers only read R2 keys matching `^raw/[a-zA-Z0-9-]+/[a-zA-Z0-9-]+\.(mp4\|mov\|mkv\|webm)$`; external URLs and `..` traversal are blocked. | `tests/security/owasp_api_security.test.ts` |
| **API8:2023 Security Misconfiguration** | Strict CORS origin allowlist, non-root containers (`UID 10001`), read-only root filesystem, and Kubernetes `NetworkPolicy` default-deny. | `infra/kubernetes/base/deployments.yaml` |
| **Webhook Replay & Forgery** | Payment (Stripe/Toss) and Cloudflare Stream webhooks verify `HMAC-SHA256(timestamp.nonce.payload)`, reject timestamps older than 300s, and deduplicate `event_id` in PostgreSQL/Valkey. | `tests/security/owasp_api_security.test.ts` |

---

## 4. PII Masking, GDPR / PIPA Privacy & Audit Logging

- **PII Minimization & Encryption**: Payout bank account numbers and tax identifiers in `payout_accounts` are encrypted at the application layer using AES-256-GCM before storage in PostgreSQL 18.
- **Structured Log Redaction**: Go `slog` JSON loggers never log raw `Authorization` headers, cookies, refresh tokens, or full email addresses in production traces.
- **Right to Erasure & Account Deletion**: Account deletion anonymizes PII in `users` while retaining legally required financial ledger rows (`ledger_entries`, `settlements`) under statutory tax retention periods.
- **Immutable Administrative Audit Trail**: Every user sanction, moderation decision, and payout approval writes an append-only entry to `audit_logs` with `actor_id`, `action`, `target_id`, `reason`, and RFC3339 timestamp.

---

## 5. Implementation Gaps (2026-11-16, code-audited — target vs. code)

The controls in §1–§4 are the **target posture**. Code audit found these gaps (P0 unless noted):

1. **Playback tokens are unsigned** — **FIXED 2026-11-17**: HMAC-SHA256 signed `v1.*` tokens (1h, video-bound) with
   private/members-only entitlement checks (`videos/playback_token.go`, `videos/visibility.go`). Geo/age checks remain open.
2. **Payment confirmation fail-open** — **FIXED 2026-11-17**: `ConfirmPayment` refuses unconfigured/unknown providers
   and verifies Toss DONE + Stripe paid sessions server-side (`payments/handler.go`, `payments/stripe.go`).
3. **Current deployment runs dev tokens** (`GVP_ENV=test`, `GVP_ALLOW_DEV_TOKENS=true`); production OIDC tenant not provisioned.
4. **Admin MFA / step-up re-auth** — **FIXED 2026-11-17**: all admin mutations (sanctions, moderation, policy,
   creative review, settlement approval) require TOTP step-up via `identity.RequireStepUp` with full audit logging.
   OIDC `amr` claim verification still lands with the production tenant (P0-6).
5. **AES-128 key release / Widevine / FairPlay**: no key-release endpoint or DRM packaging in code.
6. **PII at-rest encryption** (AES-256-GCM payout accounts) and log-redaction coverage: UNVERIFIED in code — needs proof test.
7. Rate limiting is single-instance in-memory (no Valkey); multi-replica enforcement pending.
8. **Ad surface unauthenticated** — **FIXED 2026-11-17:** role gates + advertiser scoping on campaign/report/preview routes,
   admin-only review with audit logs, DRAFT→review state machine, duplicate mount removed (137 routes, 0 holes).
   See `docs/api/ROUTE_CENSUS.md` ads/advertiser groups.
9. **Search leaks non-public videos** — **FIXED 2026-11-17:** `/v1/search` and suggestions exclude private/unlisted and gate
   members-only by entitlement (`videos.SearchVisible`).

Full list: [docs/IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md) §4–§5.
