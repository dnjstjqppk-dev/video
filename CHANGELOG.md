# Changelog

All notable product changes. Dates are UTC.

## [2.0.0] — 2026-11-18 — Next-gen UI/UX commercialization

### Added

- Dual-theme design system: Deep Obsidian Dark default on web, Pure Crisp
  Light default on Studio/Partners/Admin, each with a persisted toggle.
- Electric Indigo brand (`#6366F1`), Cyan pulse (`#06B6D4`) and Rose live
  (`#F43F5E`) tokens across all consoles.
- Modular viewer shell: glass header, search autocomplete with recent terms
  and type filters, `+ Create` menu, tabbed notification center, workspace
  switcher profile menu, 240px/72px drawer with live subscriptions, and a
  mobile bottom tab bar.
- Feed video cards with 400ms hover preview, progress bar, relative time and
  a 5-action kebab menu (save, share, not interested, don't recommend, report).
- Spec category rail (10 chips incl. recently uploaded and watched) with
  scroll arrows and spec grid breakpoints (4/3/2/1 columns).
- Studio content status filter and per-row quick actions (edit, analytics,
  comments, watch-on-web); brand analytics chart palette.
- Admin split moderation review with status filter and source links;
  settlement CSV export.
- Docs: `DESIGN_SYSTEM.md`, `ARCHITECTURE.md`, `API_SPEC.md`, `USER_GUIDE.md`.

### Changed

- Migrated 56 console views from fixed light grays to theme variables with
  zero visual regressions; all four production builds pass.
- Sidebar subscriptions now derive from the real subscription feed and live
  list instead of a fixed list.

### Fixed

- Live-status check now matches the `live` stream status value.
- Analytics chart axis and grid colors render correctly on dark theme.
- Test-run artifacts are git-ignored instead of committed.

### Verification

- `pnpm --filter @gvp/{web,studio,admin,partners} build`: all pass.
- `pnpm -r typecheck`: all packages and consoles pass.
- `cspell` over 98 touched frontend sources: 0 issues.
- Homepage SSR smoke: HTTP 200 with feed, shorts and live markers.

## [1.0.0] — 2026-11-11 — Enterprise testbed baseline

- Four Nuxt consoles, Flutter app, Go modular monolith, 75-table Postgres
  schema, Docker/K8s/Helm/Terraform delivery, and the audit-documented
  prototype scope (see `docs/IMPLEMENTATION_STATUS.md`).
