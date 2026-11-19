/** True OLED Dark cinema (default) + daytime Light themes.
 * Keys map 1:1 to spec vars in tokens.css. */
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
  canvas: '#0A0A0C',
  surface: '#141418',
  elevated: '#1F1F26',
  overlay: 'rgba(0, 0, 0, 0.62)',
  borderSubtle: '#26262E',
  borderStrong: '#3A3A44',
  textPrimary: '#F4F4F6',
  textSecondary: '#9CA3AF',
  textMuted: '#6B7280',
  textInverse: '#0F0F0F',
  focusRing: '#FF6B81',
}

export const lightTheme: SurfaceTheme = {
  canvas: '#FFFFFF',
  surface: '#F4F4F6',
  elevated: '#EAECEF',
  overlay: 'rgba(15, 15, 16, 0.45)',
  borderSubtle: '#E5E7EB',
  borderStrong: '#D4D4D8',
  textPrimary: '#0F0F0F',
  textSecondary: '#606060',
  textMuted: '#909090',
  textInverse: '#FFFFFF',
  focusRing: '#E0223E',
}

export const themes = { dark: darkTheme, light: lightTheme } as const
export type ThemeName = keyof typeof themes
