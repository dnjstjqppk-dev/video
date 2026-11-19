import { brand, semantic } from './palette'
import { lightTheme } from './themes'
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
    canvas: lightTheme.canvas, s1: lightTheme.surface, s2: lightTheme.elevated,
    overlay: lightTheme.overlay, border: lightTheme.borderSubtle, borderStrong: lightTheme.borderStrong,
  },
  text: { primary: lightTheme.textPrimary, secondary: lightTheme.textSecondary, muted: lightTheme.textMuted, inverse: '#FFFFFF' },
  spacing, radius, shadow, motion,
  typography: { fontSans: typography.fontSans, fontMono: typography.fontMono },
} as const

export type DesignTokens = typeof tokens
