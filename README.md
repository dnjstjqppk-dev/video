# Global Video Platform (GVP) — Enterprise Multi-Surface Video Ecosystem

An enterprise-grade reference architecture video platform monorepo built with **4 independently deployed Vue 3 / Nuxt 4 web consoles**, a **Flutter (iOS & Android) mobile application**, and a **Go Modular Monolith API + Separated Background Workers** backed by **PostgreSQL 18, Valkey 8, NATS JetStream, OpenSearch, ClickHouse, and Cloudflare R2 / Stream**.

---

## 1. Monorepo Directory Structure

```text
.
├── apps/
│   ├── web/                  # User Portal (www.brand.com) — Nuxt 4, Vue 3, Pinia, Shaka/HLS Player
│   ├── studio/               # Creator Studio (studio.brand.com) — Multipart R2 Upload, Analytics, Live
│   ├── partners/             # Partner Portal (partners.brand.com) — Ads, MCN, Content ID, Settlements
│   ├── admin/                # Admin Console (admin.brand.com) — Trust & Safety, Users, Audit Logs
│   └── mobile/               # Mobile App (iOS & Android) — Flutter, Riverpod, GoRouter, media_kit
├── packages/
│   ├── design-tokens/        # Shared Colors, Typography, Spacing, Z-Index & Dark/Light Theme Tokens
│   ├── ui/                   # Shared Vue 3 Headless/Accessible UI Primitives
│   ├── shared-types/         # TypeScript Interfaces aligned with OpenAPI 3.1 & Protobuf v1
│   ├── ts-api-client/        # Typed HTTP Client with Auth Interceptor & Standard Error Envelope
│   ├── eslint-config/        # Strict TypeScript & Vue 3 Lint Rules
│   └── tsconfig/             # Base TypeScript Configurations
├── services/
│   └── api/                  # Go Modular Monolith API (/v1 REST) & Worker Entrypoints
│       ├── cmd/server/       # HTTP Server Entrypoint (Graceful Shutdown, CORS, RequestID, RBAC)
│       └── internal/         # 18 Domain Modules (identity, users, channels, videos, uploads, etc.)
├── db/
│   ├── migrations/           # Ordered Forward/Rollback SQL Migrations (00001..00008, 32 Core Tables)
│   ├── queries/              # sqlc Type-Safe SQL Queries
│   ├── schema/               # Canonical PostgreSQL 18 Schema & Table Inventory
│   └── seeds/                # Deterministic Local & CI Seed Dataset
├── proto/                    # Protobuf v1 Contracts (Identity, Catalog, Media, Live, Analytics, etc.)
├── openapi/
│   └── gvp-v1.yaml           # OpenAPI 3.1.0 Specification for all /v1 REST Endpoints
├── infra/
│   ├── docker/               # Full Stack docker-compose.yml + Hardened Dockerfile.go & Dockerfile.nuxt
│   ├── kubernetes/base/      # K8s Deployments, Services, HPAs, Probes & Zero-Trust NetworkPolicies
│   ├── helm/gvp-chart/       # Production Helm Chart with Ingress Domains & Argo CD Annotations
│   ├── terraform/            # IaC for Cloudflare R2/Stream/WAF, PostgreSQL 18 HA, Valkey, OpenSearch, EKS
│   └── monitoring/           # Prometheus Scrape Config & SLO/FinOps Alert Rules (p95, 5xx, Queue, Cost)
├── tests/
│   ├── contract/             # OpenAPI 3.1 & Standard Error Envelope Contract Verification
│   ├── integration/          # Go End-to-End API Integration Suite
│   ├── e2e/                  # Playwright Multi-Surface E2E Suite (User, Studio, Partner, Admin, Security)
│   ├── performance/          # k6 Load Test Script (1,000 CCU, 100 Uploads, 10 Live Streams)
│   └── security/             # OWASP API Security Top 10 Verification (BOLA, RBAC, SSRF, Webhook Replay)
└── docs/
    ├── architecture/         # System Architecture (Phase 0~7) & Database ERDs
    ├── api/                  # Complete /v1 API Inventory Table
    ├── product/              # Feature Benchmark, 15-Role RBAC/ABAC Matrix & Monthly Unit Economics
    ├── security/             # ZITADEL OIDC+PKCE, Signed HLS CDN, DRM & OWASP Controls
    ├── operations/           # PostgreSQL 18 PITR Backup/Restore Runbook, RTO/RPO & DLQ Recovery
    └── audits/               # Phase 7 Independent Release Audit & Quality Gate Report
```

---

## 2. Quickstart (Local Development & Full Stack Docker)

### 2.1 Option A: Full Stack via Docker Compose
Starts PostgreSQL 18, Valkey 8, NATS JetStream, OpenSearch, ClickHouse, MinIO (R2-compatible S3), Go API + 5 Workers, all 4 Nuxt Web Consoles, Prometheus, and Grafana:

```bash
docker compose -f infra/docker/docker-compose.yml up --build -d
```

| Surface / Service | Local URL | Default Credentials / Auth Token |
| :--- | :--- | :--- |
| **Go API (`/v1` & `/healthz`)** | `http://localhost:8080/healthz` | Configured via OIDC/JWT in local `.env` |
| **User Web Portal (`apps/web`)** | `http://localhost:3000` | Public / Creator Session |
| **Creator Studio (`apps/studio`)** | `http://localhost:3001` | Creator Session (`creator1@brand.com`) |
| **Partner Portal (`apps/partners`)** | `http://localhost:3002` | Partner Manager Session |
| **Admin Console (`apps/admin`)** | `http://localhost:3003` | Super Admin (`admin@brand.com`) |
| **Prometheus / Grafana** | `http://localhost:9090` / `http://localhost:3005` | `admin` / `<configured_in_secrets>` |

### 2.2 Option B: Run Services Natively
```bash
# 1. Start Go API Server (Port 8080)
cd services/api
go run ./cmd/server

# 2. Install Frontend Workspace Dependencies & Start Nuxt 4 Consoles
pnpm install
pnpm dev:web       # http://localhost:3000
pnpm dev:studio    # http://localhost:3001
pnpm dev:partners  # http://localhost:3002
pnpm dev:admin     # http://localhost:3003

# 3. Run Flutter Mobile App (apps/mobile)
cd apps/mobile
flutter pub get
flutter run
```

---

## 3. Verification, Testing & Audit Commands

```bash
# 1. Run Go End-to-End Integration Test Suite
cd tests/integration && go test -v ./...

# 2. Run Go Security & JWKS Unit Tests
cd services/api && go test -v ./...

# 3. Run OpenAPI 3.1 Contract & OWASP API Top 10 Security Verification
node --experimental-strip-types --test tests/contract/openapi_contract.test.ts tests/security/owasp_api_security.test.ts

# 4. Run Playwright End-to-End Multi-Surface Tests
npx playwright test tests/e2e/platform.spec.ts

# 5. Run k6 Commercial Baseline Load Test (1,000 Concurrent Viewers)
k6 run tests/performance/k6-load-test.js
```

---

## 4. Phase 0~11 Engineering Deliverables & Remediation Index

- **Phase 0~2 System Architecture & Data Flows**: [docs/architecture/SYSTEM_ARCHITECTURE.md](docs/architecture/SYSTEM_ARCHITECTURE.md)
- **Phase 3 Database ERDs (Identity, Catalog, Social, Live, Analytics, Ads, Finance, Rights, Moderation, Ops)**: [docs/architecture/erd/README.md](docs/architecture/erd/README.md)
- **Phase 4 Complete `/v1` API Inventory & OpenAPI 3.1 Specification**: [docs/api/API_INVENTORY.md](docs/api/API_INVENTORY.md) · [openapi/gvp-v1.yaml](openapi/gvp-v1.yaml)
- **Phase 5 Competitive Benchmark, 15-Role RBAC/ABAC & Monthly Unit Economics**: [docs/product/BENCHMARK_AND_BACKLOG.md](docs/product/BENCHMARK_AND_BACKLOG.md)
- **Phase 6 Security, OIDC+PKCE, Signed HLS CDN & OWASP Top 10 Controls**: [docs/security/SECURITY_AND_COMPLIANCE.md](docs/security/SECURITY_AND_COMPLIANCE.md)
- **Phase 7 Operations Runbook & PostgreSQL 18 PITR Disaster Recovery**: [docs/operations/RUNBOOK_AND_DR.md](docs/operations/RUNBOOK_AND_DR.md)
- **Phase 8 Cross-Platform Mobile Application (Flutter iOS/Android)**: [apps/mobile/](apps/mobile/)
- **Phase 9 End-to-End Multi-Surface Integration & OWASP Security Verification**: [tests/](tests/)
- **Phase 10 Multi-Surface Cloudflare Edge Deployment & Canonical Routing**: [wrangler.toml](wrangler.toml) · [.github/workflows/cloudflare-deploy.yml](.github/workflows/cloudflare-deploy.yml)
- **Phase 11 Production Remediation, Fail-Closed Security & Governance Audit**: [docs/audits/COMMERCIAL_REMEDIATION_AUDIT.md](docs/audits/COMMERCIAL_REMEDIATION_AUDIT.md) · [docs/audits/INDEPENDENT_RELEASE_AUDIT.md](docs/audits/INDEPENDENT_RELEASE_AUDIT.md)
