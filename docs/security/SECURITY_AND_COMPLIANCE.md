# Global Video Platform — Security, Identity, DRM & Compliance Architecture

## 1. ZITADEL OIDC / OAuth 2.1 + PKCE Identity Architecture

1. **Authentication Protocol**:
   - All 4 Nuxt web consoles (`www`, `studio`, `partners`, `admin`) and the Flutter mobile app authenticate via **OAuth 2.1 Authorization Code Flow with PKCE (`S256`)** backed by **ZITADEL**.
   - Password grant and implicit flows are strictly disabled.
2. **Token Storage & Rotation**:
   - **Web Consoles**: Access tokens (15-minute TTL) are held in memory; Refresh tokens (7-day sliding TTL with reuse detection) are stored in `HttpOnly; Secure; SameSite=Lax` cookies scoped per subdomain.
   - **Flutter Mobile (`apps/mobile`)**: Tokens are encrypted at rest using `flutter_secure_storage` (iOS Keychain `kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly` / Android EncryptedSharedPreferences + Keystore).
3. **JWKS Verification in Go API (`services/api/internal/security/jwks.go`)**:
   - The Go API verifies `RS256` / `ES256` JWT signatures against ZITADEL's JWKS endpoint with in-memory key caching, clock-skew tolerance (`<= 30s`), `iss`, `aud`, `exp`, and `nbf` validation.
4. **Privileged Surface Hardening (`admin.brand.com`)**:
   - Requires hardware WebAuthn/Passkey or TOTP MFA claim (`amr: ["mfa"]`), maximum session lifetime of 4 hours, and ingress IP allowlisting.

---

## 2. CDN Signed URLs, HLS Segment Protection & DRM Readiness

1. **Signed Playback Token (`pb_token`)**:
   - Viewing a video calls `GET /v1/videos/{id}/playback`. After verifying video visibility (`public`, `unlisted`, `private`, `members_only`), geo-restrictions, and user entitlements, the Go API issues an HMAC-SHA256 signed URL token valid for **1 hour** bound to the video path prefix `/hls/{video_id}/*`.
2. **Cloudflare Edge WAF Enforcement (`infra/terraform/main.tf`)**:
   - `media.brand.com` rejects any `.m3u8` manifest, `.ts`/`.m4s` segment, or `.vtt` caption request with a missing, tampered, or expired HMAC signature (`403 Forbidden`) at the Cloudflare edge before hitting R2 storage.
3. **HLS AES-128 & Multi-DRM Envelope**:
   - Premium and `members_only` streams support `#EXT-X-KEY:METHOD=AES-128,URI="https://api.brand.com/v1/videos/{id}/key"` where the Go API only releases the 16-byte decryption key to authenticated subscribers with an active entitlement.
   - Packaging metadata is structured for seamless upgrade to **Widevine (Android/Chrome)** and **FairPlay Streaming (iOS/Safari)** CENC/cbcs encryption for Studio partner content.

---

## 3. OWASP API Security Top 10 Controls

| OWASP API Risk | Platform Mitigation Implemented | Verification Test |
| :--- | :--- | :--- |
| **API1:2023 Broken Object Level Authorization (BOLA)** | Every channel, video, studio, and partner handler verifies `resource.OwnerID == actor.UserID` or `actor.Role == "super_admin"` before mutation. | `tests/security/owasp_api_security.test.ts`, `tests/integration/api_integration_test.go` |
| **API2:2023 Broken Authentication** | OAuth 2.1 + PKCE, JWKS signature verification, refresh token rotation, and session revocation via `/v1/me/sessions`. | `services/api/internal/security/jwks_test.go` |
| **API3:2023 Broken Object Property Level Authorization** | Explicit request DTO structs in every Go handler; `role`, `is_verified`, `subscriber_count`, and `view_count` are never bound from client JSON payloads. | `services/api/internal/users/handler.go` |
| **API4:2023 Unrestricted Resource Consumption** | Nginx Ingress rate limit (`150 req/s`), Valkey token-bucket rate limiter per route, 10MB multipart chunk enforcement, and `http.Server` timeouts (`ReadTimeout: 15s`, `WriteTimeout: 15s`). | `infra/helm/gvp-chart/values.yaml` |
| **API5:2023 Broken Function Level Authorization** | `/v1/admin/*` routes enforce `RequireRole("super_admin", "platform_admin")` and `/v1/admin/moderation` enforces `RequireRole("super_admin", "moderator")`. | `tests/e2e/platform.spec.ts` |
| **API7:2023 Server-Side Request Forgery (SSRF)** | Media workers only read R2 keys matching `^raw/[a-zA-Z0-9-]+/[a-zA-Z0-9-]+\.(mp4\|mov\|mkv\|webm)$`; external URLs and `..` traversal are blocked. | `tests/security/owasp_api_security.test.ts` |
| **API8:2023 Security Misconfiguration** | Strict CORS origin allowlist, non-root containers (`UID 10001`), read-only root filesystem, and Kubernetes `NetworkPolicy` default-deny. | `infra/kubernetes/base/deployments.yaml` |
| **Webhook Replay & Forgery** | Payment (Stripe/Toss) and Cloudflare Stream webhooks verify `HMAC-SHA256(timestamp.nonce.payload)`, reject timestamps older than 300s, and deduplicate `event_id` in PostgreSQL/Valkey. | `tests/security/owasp_api_security.test.ts` |

---

## 4. PII Masking, GDPR / PIPA Privacy & Audit Logging

- **PII Minimization & Encryption**: Payout bank account numbers and tax identifiers in `payout_accounts` are encrypted at the application layer using AES-256-GCM before storage in PostgreSQL 18.
- **Structured Log Redaction**: Go `slog` JSON loggers never log raw `Authorization` headers, cookies, refresh tokens, or full email addresses in production traces.
- **Right to Erasure & Account Deletion**: Account deletion anonymizes PII in `users` while retaining legally required financial ledger rows (`ledger_entries`, `settlements`) under statutory tax retention periods.
- **Immutable Administrative Audit Trail**: Every user sanction, moderation decision, and payout approval writes an append-only entry to `audit_logs` with `actor_id`, `action`, `target_id`, `reason`, and RFC3339 timestamp.
