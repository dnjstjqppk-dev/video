# vravra Architecture — Consoles, State, API and Media Pipeline

Companion to `docs/architecture/SYSTEM_ARCHITECTURE.md` (system view) and
`docs/IMPLEMENTATION_STATUS.md` (build status). This file describes the
frontend module layout, state management, and how consoles reach the backend.

## 1. Console overview

| Console | Path | Port | Theme default | Purpose |
| :--- | :--- | :--- | :--- | :--- |
| Viewer web | `apps/web` | 3000 | dark | home, watch, shorts, live, channels, search |
| vravra studio | `apps/studio` | 3001 | light | dashboard, content, upload, analytics, comments, earn |
| Partners | `apps/partners` | 3002 | light | advertiser, campaigns, MCN, rights, settlements |
| Admin | `apps/admin` | 3003 | light | overview, moderation, users, finance, policies, ads |

Each console is an independent Nuxt 4 + Vue 3 + Pinia + Tailwind app deployed to
Cloudflare Workers. Shared code lives in `packages/`: `design-tokens`, `ui`,
`shared-types`, `ts-api-client`, `tsconfig`, `eslint-config`.

## 2. Viewer shell module tree

New or rewritten modules stay within 50-80 lines each:

```text
apps/web/app/
├── layouts/default.vue            # header + drawer + main + footer + tabs
├── components/layout/
│   ├── AppHeader.vue              # glass header shell
│   ├── AppSearchBar.vue           # search + autocomplete dropdown
│   ├── AppCreateMenu.vue          # + Create menu (studio deep links)
│   ├── AppNotificationCenter.vue  # 420px tabbed notifications
│   ├── AppProfileMenu.vue         # workspace switcher + theme + locale
│   ├── AppSideDrawer.vue          # 240px / 72px drawer + subscriptions
│   └── AppBottomTabs.vue          # mobile bottom navigation
├── components/feed/FeedVideoCard.vue  # hover preview + kebab actions
├── composables/
│   ├── useApi.ts                  # typed client bound to runtime config
│   ├── useRecentSearch.ts         # localStorage recent terms
│   ├── useSubscriptionChannels.ts # channels from subscription feed
│   └── useThemeMode.ts            # dark-first theme control
├── plugins/click-outside.ts       # dropdown dismiss directive
└── stores/ (app.ts, library.ts)   # session, options, watch history, library
```

Studio, partners and admin share the same pattern with `useStudioApi`,
`usePartnerApi`, `useAdminApi` and `useConsoleTheme` (dark default,
light/dark/system toggle).

## 3. State management

- `app.ts` (web): session user, sidebar, locale, currency, watch progress map,
  hidden videos and blocked channels.
- `library.ts` (web): history, saved, liked, queue, subscriptions
  persisted to `localStorage` (`gvp_library_v1`, capped at 200 entries
  each). The queue feeds watch-page autoplay-next.
- Consoles: per-app session state plus `console-theme` preference.
- Server state uses `useAsyncData` with refresh on user action; failures render
  empty or error states, never seed fallbacks.

## 4. Backend API and media pipeline

```text
browser -> console worker (/api) -> api-proxy worker
        -> cloudflared tunnel -> Go API :18092 + media worker :8081
```

- Go modular monolith serves `/v1` REST (see `docs/API_SPEC.md` and
  `openapi/gvp-v1.yaml`). Auth is OIDC + PKCE with `gvp_access_token` cookie;
  admin mutations require step-up TOTP and write audit logs.
- Uploads: multipart session, presigned part URLs, completion check, then HTTP
  dispatch to the media worker, which probes with ffprobe, encodes the HLS
  ladder with ffmpeg, and publishes to R2/MinIO before the API flips READY.
- Playback: signed HMAC playback tokens (1h, video-bound) with visibility and
  membership checks; Shaka/HLS adaptive player with quality, speed, captions,
  chapters, PiP and theater mode.
- Live: CRUD, stream-key rotation, ll-HLS fields, WebSocket chat hub, slow mode,
  SuperChat and polls; ingest via Cloudflare Stream provisioning.
- Ads: decision API with budget/schedule/category checks, creative upload with
  review states, frequency caps, and decision-linked billing events.
