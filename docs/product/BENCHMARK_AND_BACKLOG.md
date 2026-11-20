# vravra — Benchmark Matrix, RBAC/ABAC & Unit Economics

> Implementation status key (audited 2026-11-20, commit d4964ea): **구현** = code + gate/E2E verified ·
> **목표** = documented target, not wired. Strategy: [GLOBAL_STRATEGY.md](GLOBAL_STRATEGY.md).

## 1. Competitive Feature Benchmark Matrix

| Domain / Capability | YouTube / YouTube Studio | Twitch / IVS | Vimeo Enterprise | vravra (GVP v2.0) |
| :--- | :--- | :--- | :--- | :--- |
| **AI Originals Vertical** | Unfiltered AI flood | — | Staff picks (generic) | **Review-gated shelf + creator-run series + cinema hall, E2E green** (구현 — 핵심 승부처) |
| **Playback Engine** | Custom HTML5 / DASH / HLS | Low-Latency HLS | HLS + CDN | **Shaka Player (Web) & `video_player` (Flutter) with Signed ABR HLS** (구현; HLS.js fallback 미확인) |
| **Upload Architecture** | Resumable HTTP Chunked | RTMP Live Only | tus.io Resumable | **S3-Compatible Presigned Multipart (10MB chunks, Resume & Checksum)** (구현; R2/MinIO/Seaweed 호환) |
| **Transcoding Pipeline** | Custom ASICs (Argos) | Managed Transcode | Managed AWS | **FFmpeg Workers, 7-rung ladder (240p–2160p)** (구현; Spot K8s + Stream Live는 목표) |
| **vravra studio** | Dedicated SPA | Creator Dashboard | Team Library | **Independent Nuxt 4 `apps/studio` with series + AI submission** (구현; analytics 인메모리, ClickHouse는 목표) |
| **Partner / MCN / Content ID** | CMS + Content ID | Affiliate Portal | Enterprise Admin | **Independent Nuxt 4 `apps/partners` with Ads, MCN & Claims** (구현, 메모리 기반; 지문 매칭 MISSING) |
| **Admin & Trust/Safety** | Internal Tools | Internal Mod | Internal Admin | **Independent Nuxt 4 `apps/admin` with audit logs + AI review queue** (구현, 메모리 기반; 불변성 미확인) |

---

## 2. 15-Role RBAC + ABAC Authorization Matrix

GVP enforces dual-layer **RBAC (Role-Based Access Control)** via Casbin and **ABAC (Attribute-Based Access Control)** verifying `owner_id == actor.user_id` or `org_id == actor.org_id` on every mutation.

| # | Role Identifier | Surface Access | Key Permitted Scopes & ABAC Constraints |
| :--- | :--- | :--- | :--- |
| 1 | `visitor` | `www.brand.com`, Mobile | Public VOD/Shorts playback, public comments read, search |
| 2 | `user` | `www.brand.com`, Mobile | Subscribe, like, comment, playlists, watch history, premium subscription |
| 3 | `premium_member` | `www.brand.com`, Mobile | Ad-free playback, background/offline download tokens, `members_only` VOD access |
| 4 | `creator` | `studio.brand.com` | Full CRUD on own channel (`channel.owner_id == actor.id`), multipart upload, live stream, revenue view |
| 5 | `channel_manager` | `studio.brand.com` | Delegated upload, metadata edit, and comment moderation on assigned `channel_id` (no payout bank edit) |
| 6 | `channel_editor` | `studio.brand.com` | Draft video upload, chapters, captions, and thumbnail management (cannot publish or delete) |
| 7 | `live_moderator` | `www.brand.com`, `studio` | Chat message deletion, slow-mode toggle, and 300s user timeout within assigned `channel_id` |
| 8 | `partner_manager` | `partners.brand.com` | MCN multi-channel linking, contract management, and aggregated partner analytics (`org_id` scoped) |
| 9 | `advertiser` | `partners.brand.com` | VAST/VMAP ad campaign creation, budget pacing, and creative upload (`advertiser_id` scoped) |
| 10 | `rights_holder` | `partners.brand.com` | Reference fingerprint asset upload, Content ID copyright claim submission & dispute resolution |
| 11 | `partner_finance` | `partners.brand.com` | Monthly settlement statements, tax invoice download, and payout bank verification (`org_id` scoped) |
| 12 | `moderator` | `admin.brand.com` | Trust & Safety report queue triage, video/comment takedown, strike issuance (logged to `audit_logs`) |
| 13 | `finance_ops` | `admin.brand.com` | Double-entry ledger reconciliation, payout batch approval, refund/chargeback handling |
| 14 | `security_auditor` | `admin.brand.com` | Read-only inspection of immutable `audit_logs`, active sessions, and compliance exports |
| 15 | `super_admin` | `admin.brand.com` (All) | Platform-wide configuration, role assignment, kill-switch & FinOps budget controls (MFA + IP allowlist mandatory) |

---

## 3. Monthly Unit Economics & Cost Model (Section 24)

### Baseline Workload Assumptions (Initial Commercial Launch)
- **Registered Users**: 10,000 MAU | **Concurrent Viewers (Peak)**: 1,000 CCU
- **Monthly VOD Catalog Growth**: 5,000 new VODs/month (avg 10 minutes = 50,000 minutes ingested/month, ~1.2 TB raw + ABR rungs)
- **Monthly VOD Playback Volume**: 1,500,000 minutes watched/month (~45 TB CDN egress at average 4 Mbps ABR)
- **Live Streaming Volume**: 10 concurrent channels, 15,000 live broadcast minutes/month, 300,000 live viewer minutes/month

### Option A (Adopted): Hybrid Self-Hosted VOD (Cloudflare R2 + Spot FFmpeg Workers) + Managed Cloudflare Stream Live

| Cost Component | Pricing Basis | Monthly Volume | Estimated Monthly Cost (USD) |
| :--- | :--- | :--- | :--- |
| **Cloudflare R2 Storage (Raw + HLS)** | \$0.015 / GB-month | 4,000 GB cumulative | **\$60.00** |
| **Cloudflare R2 CDN Egress (`media.brand.com`)** | **\$0.00 / GB (Zero Egress Fee)** | 45,000 GB delivered | **\$0.00** |
| **R2 Class A / Class B Operations** | \$4.50 / 1M PUT, \$0.36 / 1M GET | 3M PUT / 40M GET | **\$27.90** |
| **Spot FFmpeg Media Workers (`c7i.4xlarge`)** | ~\$0.24 / hr Spot (3 nodes avg) | 50,000 VOD mins encoded | **\$518.40** |
| **Cloudflare Stream Live (Ingest + Delivery)** | \$5/1k mins stored + \$1/1k mins viewed | 15k mins stored + 300k viewed | **\$375.00** |
| **Control Plane (EKS + PG18 HA + Valkey + CH + OS)** | Reserved/Savings Plan compute & storage | Core production cluster | **\$1,420.00** |
| **Total Monthly Infrastructure Cost (Hybrid Adopted)** | — | — | **\$2,401.30 / month** |

### Option B (Rejected for VOD): 100% Managed Video SaaS (e.g., AWS Elemental + CloudFront or Full Managed VOD)

| Cost Component | Pricing Basis | Monthly Volume | Estimated Monthly Cost (USD) |
| :--- | :--- | :--- | :--- |
| **Managed VOD Encoding (4 ABR rungs)** | \$0.024 / min | 50,000 mins | \$1,200.00 |
| **Standard CDN Egress (45 TB @ \$0.085/GB)** | \$0.085 / GB | 45,000 GB | \$3,825.00 |
| **Managed VOD Storage & Live Streaming** | \$5/1k mins + \$1/1k mins | Catalog + Live | \$1,875.00 |
| **Control Plane (DB, Cache, API)** | Standard HA | Core cluster | \$1,420.00 |
| **Total Monthly Cost (100% Managed VOD)** | — | — | **\$8,320.00 / month** |

> **Unit Economics Conclusion (estimates, unmeasured 2026-11-20)**: By combining **Cloudflare R2 (zero egress fee)** + **Spot Kubernetes FFmpeg workers** for VOD/Shorts with **Cloudflare Stream** for low-latency live broadcasting, GVP would reduce monthly media delivery and encoding costs by **71.1% (\$2,401/mo vs \$8,320/mo)**. Caveats: ladder is 7 rungs (not 4), Spot/Stream are targets (current: PC-hosted FFmpeg + URL-assembly live), and no load/cost run has been measured. Do not quote as measured figures.
