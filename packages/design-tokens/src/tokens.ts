import { brand, semantic } from './palette'
import { darkTheme } from './themes'
import { motion, radius, shadow } from './motion'
import { typography, spacing } from './typography'

/** Backward-compatible aggregate for existing @gvp/ui primitives. */
export const tokens = {
  colors: {
    brand: { DEFAULT: brand.primary, hover: brand.primaryHover, soft: brand.primarySoft, contrast: '#FFFFFF' },
    link: { DEFAULT: brand.primary, hover: brand.primaryHover, soft: brand.primarySoft },
    semantic: { success: semantic.success, warning: semantic.warning, danger: semantic.danger, info: semantic.info },
    live: brand.live,
  },
  surfaces: {
    canvas: darkTheme.canvas, s1: darkTheme.surface, s2: darkTheme.elevated,
    overlay: darkTheme.overlay, border: darkTheme.borderSubtle, borderStrong: darkTheme.borderStrong,
  },
  text: { primary: darkTheme.textPrimary, secondary: darkTheme.textSecondary, muted: darkTheme.textMuted, inverse: '#FFFFFF' },
  spacing, radius, shadow, motion,
  typography: { fontSans: typography.fontSans, fontMono: typography.fontMono },
} as const

export type DesignTokens = typeof tokens
