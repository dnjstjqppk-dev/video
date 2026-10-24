# Global Video Platform — Benchmark Matrix, RBAC/ABAC & Unit Economics

## 1. Competitive Feature Benchmark Matrix

| Domain / Capability | YouTube / YouTube Studio | Twitch / IVS | Vimeo Enterprise | Global Video Platform (GVP v2.0) |
| :--- | :--- | :--- | :--- | :--- |
| **Playback Engine** | Custom HTML5 / DASH / HLS | Low-Latency HLS | HLS + CDN | **Shaka Player + HLS.js (Web) & `media_kit` (Flutter) with Signed ABR HLS** |
| **Upload Architecture** | Resumable HTTP Chunked | RTMP Live Only | tus.io Resumable | **Direct-to-R2 Presigned Multipart (10MB chunks, Resume & Checksum)** |
| **Transcoding Pipeline** | Custom ASICs (Argos) | Managed Transcode | Managed AWS | **Self-Hosted Spot FFmpeg Workers on R2 (Zero Egress) + Cloudflare Stream Live** |
| **Creator Studio** | Dedicated SPA | Creator Dashboard | Team Library | **Independent Nuxt 4 `apps/studio` (`studio.brand.com`) with ClickHouse Analytics** |
| **Partner / MCN / Content ID** | CMS + Content ID | Affiliate Portal | Enterprise Admin | **Independent Nuxt 4 `apps/partners` (`partners.brand.com`) with Ads, MCN & Claims** |
| **Admin & Trust/Safety** | Internal Tools | Internal Mod | Internal Admin | **Independent Nuxt 4 `apps/admin` (`admin.brand.com`) with Immutable Audit Logs** |

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

> **Unit Economics Conclusion**: By combining **Cloudflare R2 (zero egress fee)** + **Spot Kubernetes FFmpeg workers** for VOD/Shorts with **Cloudflare Stream** for low-latency live broadcasting, GVP reduces monthly media delivery and encoding costs by **71.1% (\$2,401/mo vs \$8,320/mo)** while preserving sub-second CDN TTFB globally.
