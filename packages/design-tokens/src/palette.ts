/** StreamVerse brand palette — single source of truth.
 * User-spec Electric Indigo primary + Cyan pulse + Rose live.
 * Theme-dependent surfaces live in themes.ts. */
export const brand = {
  primary: '#6366F1',
  primaryHover: '#4F46E5',
  primaryActive: '#4338CA',
  primarySoft: 'rgba(99, 102, 241, 0.12)',
  accent: '#06B6D4',
  accentHover: '#0891B2',
  accentSoft: 'rgba(6, 182, 212, 0.12)',
  live: '#F43F5E',
  liveHover: '#E11D48',
  liveSoft: 'rgba(244, 63, 94, 0.12)',
} as const

export const semantic = {
  success: '#22C55E',
  successSoft: 'rgba(34, 197, 94, 0.12)',
  warning: '#F59E0B',
  warningSoft: 'rgba(245, 158, 11, 0.12)',
  danger: '#EF4444',
  dangerSoft: 'rgba(239, 68, 68, 0.12)',
  info: '#06B6D4',
  infoSoft: 'rgba(6, 182, 212, 0.12)',
} as const

export const chart = {
  series: ['#6366F1', '#06B6D4', '#F43F5E', '#22C55E', '#F59E0B'],
  grid: 'rgba(148, 163, 184, 0.16)',
} as const

export type BrandTokens = typeof brand
export type SemanticTokens = typeof semantic
