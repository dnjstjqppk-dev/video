# vravra Design System — Vibe · Voice · Vivid

Implementation companion to [BRAND_GUIDELINES.md](BRAND_GUIDELINES.md).
Single source of truth: `packages/design-tokens/src/tokens.css` (spec
variables), `brand.ts` (names, slogans, hex), `brand-motion.css` (logo
wipe), and `packages/ui` (Vue primitives). All four consoles import
`@gvp/design-tokens/src/gvp.css` and extend the shared Tailwind preset.

**Brand concept:** *Vibe* (raw creator energy) · *Voice* (realtime talk
and reaction) · *Vivid* (crisp video immersion).
**Philosophy (Content-First Canvas):** UI chrome stays extremely
restrained so thumbnails and the player own the screen; brand color is
spent only on key actions (subscribe, live, upload, hero metrics).

## 1. Color tokens (dual theme)

Default theme is **Light (daytime white)** on every console, with a
persisted light/dark/system toggle and a True OLED Dark cinema theme.
Exception: the AI hall (`/ai`, `/series/[id]`) is always-dark cinema mode in
both themes (explicit `neutral-950` surfaces, not theme tokens).

| Token | Dark | Light |
| :--- | :--- | :--- |
| `bg-canvas` | `#0A0A0C` | `#FFFFFF` |
| `bg-surface` | `#141418` | `#F7F7F8` |
| `bg-elevated` | `#1F1F26` | `#F2F2F2` |
| `border-hairline` | `#26262E` | `#E8E9EB` |
| `text-primary` | `#F4F4F6` | `#0F0F0F` |
| `text-secondary` | `#9CA3AF` | `#606060` |
| `text-muted` | `#6B7280` | `#909090` |
| `brand-crimson` | `#FF2E4C` (hover `#E0223E`) | same |
| `studio-cobalt` | `#2563EB` (hover `#1D4ED8`) | same |
| `live-signal` | `#FF3B30` (streaming badge + avatar rings) | same |
| `premiere-amber` | `#FF9F1C` | same |

TypeScript mirrors live in `palette.ts` + `themes.ts`; legacy
`--gvp-*` variables alias the spec tokens. Never hard-code hex values
in components.

## 2. Typography & media

- Sans: `Plus Jakarta Sans`, `Pretendard`, system fallback.
- Mono: `JetBrains Mono` + `tabular-nums` (counts, durations,
  timestamps, CTR, revenue).
- Type scale (`.vravra-*` classes): Display 32/700/1.2/-0.02em ·
  Watch title 20/700/1.35/-0.015em · Section 18/600/1.4 · Card
  15/600/1.4 (2-line clamp) · Body 14/400/1.5 · Meta 12/500/1.3.
- Standard thumbnails `aspect-video (16:9)`; Shorts `aspect-[9/16]`
  (CLS = 0). Card radius `12px` → `4px` on hover preview.
- Spacing in `4px` multiples; pills `9999px`; studio inputs `8px`;
  time/LIVE badges `4px`.

## 3. Brand motion & layers

- Logo wipe (850ms, left → right) + `BrandSplash` first-session
  overlay + hover wave rhythm; `prefers-reduced-motion` disables all.
  See [BRAND_GUIDELINES.md](BRAND_GUIDELINES.md) §5.
- Z-order: chips `z-10`, LNB `z-30`, GNB `z-40`, dropdowns `z-50`,
  modals + splash `z-60`.
- Dropdown, drawer and hover motion: `180ms cubic-bezier(0.2, 0, 0, 1)`.
- Hover preview starts after `350ms`. All motion honors
  `prefers-reduced-motion`.

## 4. Dropdown and popover standard

Every dropdown across viewer, studio and admin follows the same contract:

1. Closes on outside click, `Escape`, and route change.
2. Arrow-key navigation with visible focus; `Enter` selects.
3. Placement flips (`bottom-end` to `top-end`) when space runs out.
4. Panel max height scrolls; width is fixed per use (notifications `420px`).

Reference implementation: `GvpDropdown.vue` + `dropdown.ts` helpers, and the
`v-click-outside` directive in the web console.

## 5. Lists, grids and cards

- Home grid: `repeat(auto-fill, minmax(310px, 1fr))`, `gap: 24px 16px`.
- Video cards: `16:9` rounded (`12px`) thumbnail, bottom-right duration
  or `LIVE` badge, crimson hover scrub bar, hover quick actions, avatar
  + 2-line title + channel + verified mark + views and relative time.
- Loading uses shimmer skeletons (`.gvp-shimmer`); errors show a retry
  banner; empty results show an empty state, never fake data.

## 5.1 AI badge (`AiBadge.vue`)

The AI label is a badge of honor, never hidden: brand-crimson pill, uppercase
`AI` + bolt glyph, `title` tooltip with tool and approval state.
Props: `tool?` (e.g. `Vravra Studio AI`), `approved?` (Originals pick),
`size` (`xs` default on cards, `sm` on watch/hero). Placement: top-left on
thumbnails and rails, beside the watch title, on series episode rows.

## 6. Usage rules

1. New or touched leaf components stay within 50-80 lines; split larger ones.
   Page-level compositions (e.g. `watch/[id].vue`, ~900 lines) are exempt but
   must move toward section components when touched (roadmap, not yet split).
2. Consoles default to light with a light/dark/system toggle.
3. Status colors: success `#10B981`, warning `#F59E0B`, danger
   `#EF4444`, info/studio `#2563EB`, live `#FF3B30`, brand `#FF2E4C`.
4. Icons are inline SVG; emoji is never used as iconography.
5. No hard-coded brand strings, colors, or slogans — import from
   `brand.ts` and `tokens.css`.
