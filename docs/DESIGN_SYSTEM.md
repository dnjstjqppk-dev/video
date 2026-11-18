# StreamVerse Design System — Next-Gen Media Platform

Single source of truth: `packages/design-tokens` (tokens + CSS) and `packages/ui`
(Vue primitives). All four consoles import `@gvp/design-tokens/src/gvp.css`.

## 1. Color tokens (dual theme)

Default theme is **Deep Obsidian Dark**. Creator Studio, Partner and Admin consoles
default to **Pure Crisp Light** with a persisted toggle. Brand colors are identical
in both themes.

| Token | Dark | Light |
| :--- | :--- | :--- |
| `bg-canvas` | `#0B0D12` | `#F8FAFC` |
| `bg-surface` | `#131722` | `#FFFFFF` |
| `bg-elevated` | `#1C2232` | `#F1F5F9` |
| `border-subtle` | white at 8% opacity | `#E2E8F0` |
| `text-primary` | `#F8FAFC` | `#0F172A` |
| `text-secondary` | `#94A3B8` | `#475569` |
| `brand-primary` | `#6366F1` (hover `#4F46E5`) | same |
| `brand-accent` | `#06B6D4` (live pulse, progress, unread) | same |
| `semantic-live` | `#F43F5E` (streaming badge only) | same |

CSS variables (`--gvp-*`) are defined in `gvp-themes.css`. TypeScript mirrors live
in `palette.ts` + `themes.ts`. Never hard-code hex values in components; use the
variables so both themes keep working.

## 2. Typography

- Sans: `Pretendard Variable`, `Inter`, system fallback.
- Mono: `JetBrains Mono` (time codes, metrics, IDs).
- View counts, durations, revenue and time codes always use `tabular-nums`.
- Titles clamp to 2 lines (`.clamp-2`); meta lines are single-line truncated.

## 3. Elevation and motion

- Sticky headers and dropdowns use glass treatment: `blur(16px)` + subtle border
  (`.gvp-glass`, `.gvp-elevated`).
- Dropdown, drawer and hover motion: `180ms cubic-bezier(0.2, 0, 0, 1)`.
- Hover preview starts after `400ms`. All motion honors `prefers-reduced-motion`.

## 4. Dropdown and popover standard

Every dropdown across viewer, studio and admin follows the same contract:

1. Closes on outside click, `Escape`, and route change.
2. Arrow-key navigation with visible focus; `Enter` selects.
3. Placement flips (`bottom-end` to `top-end`) when space runs out.
4. Panel max height scrolls; width is fixed per use (notifications `420px`).

Reference implementation: `GvpDropdown.vue` + `dropdown.ts` helpers, and the
`v-click-outside` directive in the web console.

## 5. Lists, grids and cards

- Grid: `>=1600px` 4 cols, `1200-1599px` 3 cols, `768-1199px` 2 cols, `<768px` 1 col.
- Video cards: `16:9` rounded (`12px`) thumbnail, bottom-right duration or `LIVE`
  badge, `3px` brand progress bar, hover quick actions, avatar + 2-line title +
  channel + verified mark + views and relative time.
- Loading uses shimmer skeletons (`.gvp-shimmer`); errors show a retry banner;
  empty results show an empty state, never fake data.

## 6. Usage rules

1. New or touched components stay within 50-80 lines; split larger ones.
2. Consoles default: web dark, studio/partners/admin light with toggle.
3. Status colors: success `#22C55E`, warning `#F59E0B`, danger `#EF4444`,
   info/accent `#06B6D4`, live `#F43F5E`.
4. Icons are inline SVG; emoji is never used as iconography.
