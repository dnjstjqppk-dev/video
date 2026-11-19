# vravra Design System — Vibe · Voice · Vivid

Single source of truth: `packages/design-tokens` (tokens + CSS) and `packages/ui`
(Vue primitives). All four consoles import `@gvp/design-tokens/src/gvp.css`.

**Brand concept:** *Vibe* (raw creator energy) · *Voice* (realtime talk and
reaction) · *Vivid* (crisp video immersion).
**Philosophy (Content-First Canvas):** UI chrome stays extremely restrained so
thumbnails and the player own the screen; brand color is spent only on key
actions (subscribe, live, upload, hero metrics).

## 1. Color tokens (dual theme)

Default theme is **Light (daytime)** on every console, with a persisted
light/dark/system toggle. **Dark** is the True OLED cinema/night mode.

| Token | Light | Dark |
| :--- | :--- | :--- |
| `bg-canvas` | `#FFFFFF` | `#0A0A0C` |
| `bg-surface` | `#F8F9FA` | `#141418` |
| `bg-elevated` / hover & pill | `#F2F4F7` | `#1F1F26` |
| `border-subtle` | `#E5E7EB` | `#26262E` |
| `text-primary` | `#0F0F0F` | `#F4F4F6` |
| `text-secondary` | `#606060` | `#9CA3AF` |
| `brand-primary` | `#FF2E4C` Vivid Electric Coral-Crimson (hover `#E02643`) | same |
| `semantic-live` | `#FF3B30` Signal Amber-Red (premiere gradient to `#FF9F1C`) | same |
| `studio-accent` | `#2563EB` Studio Cobalt (vravra studio charts, analytics, upload progress) | same |

CSS variables (`--gvp-*`) are defined in `gvp-themes.css`. TypeScript mirrors live
in `palette.ts` + `themes.ts`. Never hard-code hex values in components; use the
variables so both themes keep working.

## 2. Typography

- Sans: `Pretendard` (Korean) + `Plus Jakarta Sans` / `Inter` (Latin), system fallback.
- Mono: `JetBrains Mono` (view counts, timestamps, CTR, revenue, IDs).
- View counts, durations, CTR, revenue and time codes always use `tabular-nums`.
- Titles clamp to 2 lines (`.clamp-2`); meta lines are single-line truncated.

## 3. Media ratios and layout

- Standard video thumbnails are always `aspect-video (16:9)`; Shorts are always
  `aspect-[9/16]`, fixed before images load so **CLS = 0**.
- Viewer header is `56px` sticky with `backdrop-blur-md`; sidebar is `240px`
  expanded / `72px` mini rail / `56px` mobile bottom tabs.

## 4. Elevation and motion

- Sticky headers and dropdowns use glass treatment: `blur(16px)` + subtle border
  (`.gvp-glass`, `.gvp-elevated`).
- Dropdown, drawer and hover motion: `180ms cubic-bezier(0.2, 0, 0, 1)`.
- Hover preview starts after `400ms`. All motion honors `prefers-reduced-motion`.

## 5. Dropdown and popover standard

Every dropdown across viewer, studio and admin follows the same contract:

1. Closes on outside click, `Escape`, and route change.
2. Arrow-key navigation with visible focus; `Enter` selects.
3. Placement flips (`bottom-end` to `top-end`) when space runs out.
4. Panel max height scrolls; width is fixed per use (notifications `420px`).

Reference implementation: `GvpDropdown.vue` + `dropdown.ts` helpers, and the
`v-click-outside` directive in the web console.

## 6. Lists, grids and cards

- Home grid: `repeat(auto-fill, minmax(310px, 1fr))`, `gap: 24px 16px`.
- Video cards: `16:9` rounded (`12px`) thumbnail, bottom-right duration or `LIVE`
  badge, `3px` brand progress bar, hover quick actions, avatar + 2-line title +
  channel + verified mark + views and relative time.
- Loading uses shimmer skeletons (`.gvp-shimmer`); errors show a retry banner;
  empty results show an empty state, never fake data.

## 7. Usage rules

1. New or touched components stay within 50-80 lines; split larger ones.
2. Consoles default to light with a light/dark/system toggle.
3. Status colors: success `#22C55E`, warning `#F59E0B`, danger `#EF4444`,
   info/studio `#2563EB`, live `#FF3B30`, brand `#FF2E4C`.
4. Icons are inline SVG; emoji is never used as iconography.
