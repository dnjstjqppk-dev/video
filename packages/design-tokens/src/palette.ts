/** vravra brand palette — single source of truth.
 * Vivid Crimson primary + Studio Cobalt + Signal live + Premiere amber.
 * Theme-dependent surfaces live in themes.ts. */
export const brand = {
  primary: '#FF2E4C',
  primaryHover: '#E0223E',
  primaryActive: '#C81E3A',
  primarySoft: 'rgba(255, 46, 76, 0.12)',
  primaryGlow: 'rgba(255, 46, 76, 0.38)',
  accent: '#2563EB',
  accentHover: '#1D4ED8',
  accentSoft: 'rgba(37, 99, 235, 0.12)',
  accentGlow: 'rgba(37, 99, 235, 0.35)',
  live: '#FF3B30',
  liveHover: '#D92D20',
  liveSoft: 'rgba(255, 59, 48, 0.14)',
  premiere: '#FF9F1C',
} as const

export const studio = {
  cobalt: '#2563EB',
  cobaltHover: '#1D4ED8',
  cobaltSoft: 'rgba(37, 99, 235, 0.12)',
} as const

export const semantic = {
  success: '#10B981',
  successSoft: 'rgba(16, 185, 129, 0.12)',
  warning: '#F59E0B',
  warningSoft: 'rgba(245, 158, 11, 0.12)',
  danger: '#EF4444',
  dangerSoft: 'rgba(239, 68, 68, 0.12)',
  info: '#2563EB',
  infoSoft: 'rgba(37, 99, 235, 0.12)',
} as const

export const chart = {
  series: ['#FF2E4C', '#2563EB', '#FF9F1C', '#10B981', '#9CA3AF'],
  grid: 'rgba(148, 163, 184, 0.16)',
} as const

export type BrandTokens = typeof brand
export type SemanticTokens = typeof semantic
