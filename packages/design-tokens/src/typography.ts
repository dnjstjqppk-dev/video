/** Typography: Pretendard + Plus Jakarta Sans / Inter, tabular-nums metrics. */
export const typography = {
  fontSans:
    "'Pretendard Variable', Pretendard, 'Plus Jakarta Sans', Inter, -apple-system, BlinkMacSystemFont, system-ui, sans-serif",
  fontMono:
    "'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
  scale: {
    display: '1.75rem',
    title: '1.25rem',
    heading: '1rem',
    body: '0.875rem',
    caption: '0.75rem',
    micro: '0.6875rem',
  },
  weight: { regular: 400, medium: 500, semibold: 600, bold: 700 },
  lineHeight: { tight: 1.25, base: 1.5, relaxed: 1.65 },
} as const

export const spacing = {
  xs: '0.25rem',
  sm: '0.5rem',
  md: '1rem',
  lg: '1.5rem',
  xl: '2rem',
  '2xl': '2.5rem',
  '3xl': '3rem',
} as const

export const breakpoints = {
  wide: 1600,
  desktop: 1200,
  tablet: 768,
  drawerFull: 240,
  drawerMini: 72,
} as const

export type TypographyTokens = typeof typography
