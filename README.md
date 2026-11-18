# StreamVerse — Global Video Platform (Enterprise Multi-Surface Ecosystem)

StreamVerse is a next-generation video content platform: **4 independently
deployed Vue 3 / Nuxt 4 consoles** (viewer web, Creator Studio, partners,
admin), a **Flutter mobile app**, and a **Go modular monolith API + workers**
backed by **PostgreSQL 18, Valkey 8, NATS JetStream, OpenSearch, ClickHouse,
and Cloudflare R2 / Stream**.

> **Release 2.0.0 (2026-11-18):** premium UI/UX redesign is complete — dual
> theme (Deep Obsidian Dark + Pure Crisp Light), modular app shell, real
> API-backed viewer/studio/admin flows, and full docs. See [CHANGELOG.md](CHANGELOG.md)
> and [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for scope
> and remaining backend backlog.

---

## 1. Design system (summary)

- **Brand:** Electric Indigo `#6366F1` primary, Cyan Pulse `#06B6D4` accent,
  Rose `#F43F5E` reserved for live. Full spec: [docs/DESIGN_SYSTEM.md](docs/DESIGN_SYSTEM.md).
- **Themes:** web defaults to dark; studio/partners/admin default to light.
  Every surface keeps a working toggle; colors are CSS variables, never
  hard-coded in components.
- **Shell:** glass sticky header, autocomplete search, `+ Create` menu, tabbed
  notification center, workspace switcher, 240px/72px drawer, mobile bottom tabs.
- **Modules:** every new or touched component stays within 50-80 lines and
  talks to real API state — no mock, dummy or placeholder data.

## 2. Three-domain architecture

| Domain | Console | Routes (highlights) |
| :--- | :--- | :--- |
| Viewer | `apps/web` `:3000` | `/` feed, `/watch/:id`, `/shorts`, `/live/:id`, `/channels/:id`, `/search` |
| Creator Studio | `apps/studio` `:3001` | dashboard, `/content`, `/upload`, `/analytics`, `/comments`, `/monetization` |
| Partners | `apps/partners` `:3002` | advertiser, campaigns, MCN, rights, settlements |
| Admin | `apps/admin` `:3003` | overview, `/moderation`, `/users`, `/finance`, `/policies`, `/ads` |

Details: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) ·
API + RBAC: [docs/API_SPEC.md](docs/API_SPEC.md) ·
Manuals: [docs/USER_GUIDE.md](docs/USER_GUIDE.md).

```text
browser -> console workers (web/studio/partners/admin, /api)
       -> api-proxy worker (API_UPSTREAM secret)
       -> cloudflared tunnel -> Go API :18092 + media worker :8081
```

## 3. Monorepo layout

```text
apps/web|studio|partners|admin   # Nuxt 4 consoles (see table above)
apps/api-proxy                   # Cloudflare Worker, stable /api upstream
apps/mobile                      # Flutter iOS/Android app
packages/design-tokens           # Dual-theme tokens + gvp.css system
packages/ui                      # Accessible Vue primitives + formatters
packages/shared-types            # TS models aligned with OpenAPI/Protobuf
packages/ts-api-client           # Typed client, auth + error envelope
services/api                     # Go API (/v1) + worker entrypoints
db/ proto/ openapi/ infra/ tests/ docs/  # schema, contracts, delivery, QA
```

## 4. Local run and build

```bash
# Full stack (Postgres, Valkey, NATS, OpenSearch, ClickHouse, MinIO, API, consoles)
docker compose -f infra/docker/docker-compose.yml up --build -d

# Or run natively
cd services/api && go run ./cmd/server   # :8080 (local) / :18092 (tunnel host)
pnpm install
pnpm dev:web        # http://localhost:3000
pnpm dev:studio     # http://localhost:3001
pnpm dev:partners   # http://localhost:3002
pnpm dev:admin      # http://localhost:3003
```

| Check | Command |
| :--- | :--- |
| Production builds | `pnpm -r --filter "./apps/*" build` |
| Type checks | `pnpm typecheck` |
| Spelling + types | `pnpm lint` |
| Contract tests | `node --experimental-strip-types --test tests/contract/*.test.ts` |
| E2E | `npx playwright test tests/e2e/platform.spec.ts` |

Default demo logins are documented in `docs/operations/PRODUCTION_CHECKLIST.md`;
production requires a real ZITADEL tenant (`docs/operations/OIDC_SETUP.md`).
