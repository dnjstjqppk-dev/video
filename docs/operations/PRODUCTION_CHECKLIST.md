# Production Cutover Checklist — StreamVerse Commercial Launch

Target reader: the operator performing the testbed → commercial cutover.
Authority: Master Directive `BN-GVP-YOUTUBE-CLASS-PRODUCTION-001` ARTICLES 17–19.
Current status baseline: [docs/IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md).

> **Do not cut over while any P0 item is open.** Each gate below names its proof artifact.
> A gate without its artifact is NOT passed.

## Gate 0 — Code gates (all P0 closed)

| # | Gate | Proof |
| :--- | :--- | :--- |
| 0-1 | P0-1 persistent store: PG repository layer, restart durability | restart test log + `SELECT count(*)` before/after |
| 0-2 | P0-2 real transcoding behind `REQUIRE_PROBED_SOURCE=true` | upload→READY E2E with ffprobe/ffmpeg logs + R2 objects |
| 0-3 | P0-3 payment fail-closed without secrets | refused-confirm test (no secret) + live PG webhook test |
| 0-4 | P0-4 signed playback + search visibility filter | forged-token 403 test + private-video search exclusion test |
| 0-5 | P0-5 ad E2E (register→review→preroll→billed) | full-funnel E2E report with campaign/decision/event IDs |
| 0-6 | P0-6 production OIDC, dev tokens dead | login flow per console + dev-token 401 proof |
| 0-7 | P0-7 admin MFA + step-up + audit integrity | MFA-on test + sensitive-action re-auth test |
| 0-8 | P0-8 ad surface auth + tenant isolation | anonymous 401/403 matrix on `/v1/ads/*`, `/v1/advertiser/*` |

## Gate 1 — Secrets & credentials matrix

Set every value in the deployment secret store (K8s `gvp-backend-secrets` or equivalent).
Never commit values; this table lists NAMES only.

| Secret | Consumer | Notes |
| :--- | :--- | :--- |
| `DATABASE_URL` | api, workers | PG18 HA primary, pooled, PITR enabled |
| `OIDC_ISSUER`, `OIDC_AUDIENCE`, `OIDC_JWKS_URL` | api | ZITADEL tenant; no `auth.example.com` |
| `NUXT_OIDC_*` (×4 consoles) | web/studio/partners/admin | per-console client id/secret + redirect URI |
| `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_ENDPOINT` | api, media-worker | raw + HLS buckets, WAF HMAC secret |
| `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET` | api | live merchant, KYB done |
| `TOSS_SECRET_KEY` (+ webhook secret) | api | live merchant |
| `JWT_SECRET` | api | 32+ chars, unique per env |
| `GVP_ADMIN_WHITELIST_CIDR` | api | operator egress ranges only |
| `SENTRY_DSN`, `NUXT_PUBLIC_SENTRY_DSN` | api, consoles | error tracking live |
| `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID` | CI | least-privilege Workers deploy token |
| `GVP_ENV=production`, `GVP_ALLOW_DEV_TOKENS` unset | api | verified at boot log |

## Gate 2 — Infrastructure cutover order

1. Provision production data plane (PG18 HA + PITR, Valkey, NATS JetStream, OpenSearch, ClickHouse, R2/Stream).
2. Run migrations `00001..00009` against production + verify `SELECT count(*)` per table family.
3. Deploy API + 6 workers to hosted compute (leave the operator-PC tunnel behind; keep it as break-glass only).
4. Switch consoles' `NUXT_INTERNAL_API_URL` from api-proxy/tunnel to the hosted API; redeploy 4 consoles + api-proxy (or retire api-proxy).
5. Move DNS from `*.workers.dev` to canonical domains (`www/studio/partners/admin/media/api.brand.com`) with WAF signed-URL rules.
6. Run Gate 3 verification, then open traffic (canary 1% → 25% → 100%).

## Gate 3 — Release verification (ARTICLE 17)

- Static: lint/typecheck/build, migration verify, contract suite (incl. census gate), dep scan, secret scan.
- Functional: signup→login→logout, real upload→encode→play, history/comment/subscribe persistence, live ingest+play,
  ad register→review→serve→click→bill, pay→refund→ledger, admin isolation matrix.
- E2E: Chromium/Firefox/WebKit × 360–1920px, Android+iOS smoke, player/ad/shorts/auth transitions, a11y pass.
- Security: unauthenticated 401 matrix, cross-tenant 403 matrix, unapproved-ad suppression, forged-webhook rejection, replay idempotency.
- Ops: PITR restore drill, DLQ replay drill, k6 1k-CCU baseline within SLO, cost meter vs. unit-economics model, rollback drill.

## Rollback

- Consoles: `wrangler rollback` (or re-deploy previous commit) per app; verify `/api/v1/version` commit hash.
- API/workers: re-deploy previous image; DB changes are forward-only — never roll back migrations, roll forward a fix.
- Keep the pre-cutover tunnel stack stopped but ready until canary completes.
