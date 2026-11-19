# vravra Brand Guidelines (SSOT)

Single source of truth for the **vravra** viewer platform and
**vravra studio** creator console. Code sources:
`packages/design-tokens/src/tokens.css` (color/type tokens),
`packages/design-tokens/src/brand.ts` (names, slogans, hex),
`packages/design-tokens/src/brand-motion.css` (logo transition),
`packages/ui/src/components/VravraLogo.vue`,
`VravraStudioLogo.vue`, `BrandSplash.vue`.

## 1. Identity

| Surface | Name | Slogan (EN) | Slogan (KO) |
| :--- | :--- | :--- | :--- |
| Viewer | `vravra` (브라브라) | Play Your Vibe. Share Your Voice. | 너의 바이브를 플레이하다, 너의 목소리를 공유하다 |
| Studio | `vravra studio` | Empower Your Vision. Scale Your Vibe. | 크리에이터의 모든 것, vravra 스튜디오 |

Keywords: **Vibe** (raw sensory energy) · **Voice** (realtime talk and
reaction) · **Vivid** (crisp video immersion).
Philosophy: **Content-First Canvas** — chrome stays restrained so
thumbnails and the player own the screen; brand color is spent only on
key actions (subscribe, live, upload, hero metrics).

## 2. Color tokens

| Token | Value | Use |
| :--- | :--- | :--- |
| `--color-brand-crimson` | `#FF2E4C` | vravra primary: subscribe/play CTAs, Shorts, active states |
| `--color-brand-crimson-hover` | `#E0223E` | Primary hover |
| `--color-brand-crimson-glow` | `rgba(255,46,76,.38)` | Dark-mode logo glow |
| `--color-studio-cobalt` | `#2563EB` | Studio accents: charts, upload progress, studio CTAs |
| `--color-studio-cobalt-hover` | `#1D4ED8` | Studio hover |
| `--color-studio-cobalt-glow` | `rgba(37,99,235,.35)` | Dark-mode studio logo glow |
| `--color-live-signal` | `#FF3B30` | LIVE badges, streaming avatar rings |
| `--color-premiere-amber` | `#FF9F1C` | Premiere signals |
| `--color-success` / `--color-warning` | `#10B981` / `#F59E0B` | Status semantics |

| Surface | Light (daytime) | Dark (default cinema) |
| :--- | :--- | :--- |
| `--bg-canvas` | `#FFFFFF` | `#0A0A0C` |
| `--bg-surface` | `#F4F4F6` | `#141418` |
| `--bg-elevated` | `#EAECEF` | `#1F1F26` |
| `--border-hairline` | `#E5E7EB` | `#26262E` |
| `--text-primary` | `#0F0F0F` | `#F4F4F6` |
| `--text-secondary` | `#606060` | `#9CA3AF` |
| `--text-muted` | `#909090` | `#6B7280` |

Legacy `--gvp-*` variables remain as aliases resolving to these tokens.
Never hard-code hex in components.

## 3. Typography & media

- Sans: Plus Jakarta Sans, Pretendard, system fallback.
- Data mono: JetBrains Mono + `tabular-nums` for counts, durations,
  timestamps, CTR, revenue.
- Type scale: Display 32/700/1.2/-0.02em · Watch title 20/700/1.35/
  -0.015em · Section 18/600/1.4 · Card 15/600/1.4 (2-line clamp) ·
  Body 14/400/1.5 · Meta 12/500/1.3 (`.vravra-*` classes in tokens.css).
- Standard thumbnails `aspect-video (16:9)`, Shorts `aspect-[9/16]`
  (CLS = 0). Card radius `12px` → `4px` on hover preview.

## 4. Logos

- **VravraLogo** (`size` sm/md/lg, `animate`, `showTagline`): crimson
  geometric `vravra` wordmark, left play triangle, symmetric 5-bar
  equalizers knocked out inside both `v` valleys. Dark-mode glow on
  `.dark`. Clear space ≥ 0.4x logo height on all sides.
- **VravraStudioLogo**: crimson `v` stroke mark with micro-wave plus 4
  ascending cobalt data bars; two-tone `vravra` + `studio` wordmark.
- **favicon.svg / icon-512.svg**: condensed play + `v` + wave square;
  wired via `<link rel="icon">` and `site.webmanifest` in every app.

## 5. Wipe transition (850ms) & splash

1. 0–180ms: play-triangle pop-in.
2. 120–720ms: wordmark `clip-path` wipe left → right
   (`cubic-bezier(0.22, 1, 0.36, 1)`).
3. 350–850ms: wave bars `scaleY(1.35)` sequential pulse.
4. GNB logos animate on mount; wave bars react to hover rhythm.
5. `BrandSplash`: first session per tab shows a `#0A0A0C` overlay,
   fades out 250ms after the wipe; revisits skip (non-blocking).
6. `prefers-reduced-motion` disables all brand motion.

## 6. Component states & layers

- Primary CTA: crimson fill, white bold, hover `#E0223E`,
  active `scale(0.97)`; subscribed → elevated `구독중 ▾` pill.
- Studio CTA: cobalt fill, hover `#1D4ED8`.
- Category chips: elevated idle, inverted `--text-primary` active.
- Z-order: chips `z-10`, LNB `z-30`, GNB `z-40`, dropdowns `z-50`,
  modals + splash `z-60`.
