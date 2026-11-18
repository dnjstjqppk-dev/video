/** Deep Obsidian Dark (default) + Pure Crisp Light themes.
 * Keys map 1:1 to CSS vars in gvp-themes.css. */
export interface SurfaceTheme {
  canvas: string
  surface: string
  elevated: string
  overlay: string
  borderSubtle: string
  borderStrong: string
  textPrimary: string
  textSecondary: string
  textMuted: string
  textInverse: string
  focusRing: string
}

export const darkTheme: SurfaceTheme = {
  canvas: '#0B0D12',
  surface: '#131722',
  elevated: '#1C2232',
  overlay: 'rgba(3, 5, 10, 0.62)',
  borderSubtle: 'rgba(255, 255, 255, 0.08)',
  borderStrong: 'rgba(255, 255, 255, 0.16)',
  textPrimary: '#F8FAFC',
  textSecondary: '#94A3B8',
  textMuted: '#64748B',
  textInverse: '#0F172A',
  focusRing: '#818CF8',
}

export const lightTheme: SurfaceTheme = {
  canvas: '#F8FAFC',
  surface: '#FFFFFF',
  elevated: '#F1F5F9',
  overlay: 'rgba(15, 23, 42, 0.45)',
  borderSubtle: '#E2E8F0',
  borderStrong: '#CBD5E1',
  textPrimary: '#0F172A',
  textSecondary: '#475569',
  textMuted: '#94A3B8',
  textInverse: '#FFFFFF',
  focusRing: '#4F46E5',
}

export const themes = { dark: darkTheme, light: lightTheme } as const
export type ThemeName = keyof typeof themes
