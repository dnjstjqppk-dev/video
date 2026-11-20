# vravra Architecture — Consoles, State, API and Media Pipeline

Companion to `docs/architecture/SYSTEM_ARCHITECTURE.md` (system view) and
`docs/IMPLEMENTATION_STATUS.md` (build status). This file describes the
frontend module layout, state management, and how consoles reach the backend.

## 1. Console overview

| Console | Path | Port | Theme default | Purpose |
| :--- | :--- | :--- | :--- | :--- |
| Viewer web | `apps/web` | 3000 | light | home, watch, shorts, live, channels, search, AI hall (`/ai`, `/series`) |
| vravra studio | `apps/studio` | 3001 | light | dashboard, content, upload, series, analytics, comments, earn |
| Partners | `apps/partners` | 3002 | light | advertiser, campaigns, MCN, rights, settlements |
| Admin | `apps/admin` | 3003 | light | overview, moderation, users, finance, policies, ads, AI reviews |

All consoles are light-first with a persisted light/dark/system toggle; the AI
hall pages are an always-dark cinema exception (see DESIGN_SYSTEM §1).

Each console is an independent Nuxt 4 + Vue 3 + Pinia + Tailwind app deployed to
Cloudflare Workers. Shared code lives in `packages/`: `design-tokens`, `ui`,
`shared-types`, `ts-api-client`, `tsconfig`, `eslint-config`.

## 2. Viewer shell module tree

Leaf components stay within 50-80 lines; page compositions are exempt with a
split roadmap (see DESIGN_SYSTEM §6):

```text
apps/web/app/
├── layouts/default.vue            # header + drawer + main + footer + tabs
├── components/ (flat)             # AppHeader, AppSearchBar, AppCreateMenu,
│                                  # AppNotificationCenter, AppProfileMenu,
│                                  # AppSideDrawer, AppBottomTabs, FeedVideoCard,
│                                  # VideoPlayer.client, WatchSeriesPanel,
│                                  # AiBadge, Watch* (share/save/subscribe)
├── composables/
│   ├── useApi.ts                  # typed client bound to runtime config
│   ├── useRecentSearch.ts         # localStorage recent terms
│   ├── useSubscriptionChannels.ts # channels from subscription feed
│   └── useThemeMode.ts            # light-first theme control
├── plugins/click-outside.ts       # dropdown dismiss directive
└── stores/ (app.ts, library.ts)   # session, options, watch history, library
```

Studio, partners and admin share the same pattern with `useStudioApi`,
`usePartnerApi`, `useAdminApi` and `useConsoleTheme` (light default,
light/dark/system toggle).

## 3. State management

- `app.ts` (web): session user, sidebar, locale, currency, watch progress map,
  hidden videos and blocked channels.
- `library.ts` (web): history, saved, liked, queue, subscriptions
  persisted to `localStorage` (`gvp_library_v1`, capped at 200 entries
  each). The queue feeds watch-page autoplay-next.
- Consoles: per-app session state plus `console-theme` preference.
- Server state uses `useAsyncData` with refresh on user action; new code must
  render empty or error states, never seed fallbacks (legacy seed fallbacks
  remain on some shelves — P1).

## 4. Backend API and media pipeline

```text
browser -> console worker (/api) -> api-proxy worker
        -> cloudflared tunnel -> Go API :18092 + media worker :8081
```

- Go modular monolith serves `/v1` REST (see `docs/API_SPEC.md` and
  `openapi/gvp-v1.yaml`). Auth is OIDC + PKCE with the `gvp_session` cookie
  (proxy-injected `Authorization`); admin mutations require step-up TOTP and
  write audit logs.
- Uploads: multipart session, presigned part URLs, completion check, then HTTP
  dispatch to the media worker, which probes with ffprobe, encodes the 7-rung
  HLS ladder with ffmpeg, and publishes to S3-compatible storage before the
  API flips READY.
- Playback: signed HMAC playback tokens (1h, video-bound) with visibility and
  membership checks served through the in-app HLS gateway; Shaka adaptive
  player with quality, speed, captions, chapters, PiP and theater mode.
- Live: CRUD, stream-key rotation, ll-HLS fields, WebSocket chat hub, slow mode,
  SuperChat and polls; real ingest is missing (controller only assembles
  Stream URLs), DVR missing.
- Ads: decision API with budget/schedule/category checks, creative upload with
  review states, frequency caps, and decision-linked billing events.
