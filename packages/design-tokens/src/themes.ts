/** Daytime Light (default) + True OLED Dark cinema themes.
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

export const lightTheme: SurfaceTheme = {
  canvas: '#FFFFFF',
  surface: '#F8F9FA',
  elevated: '#F2F4F7',
  overlay: 'rgba(15, 15, 16, 0.45)',
  borderSubtle: '#E5E7EB',
  borderStrong: '#D4D4D8',
  textPrimary: '#0F0F0F',
  textSecondary: '#606060',
  textMuted: '#9AA0A6',
  textInverse: '#FFFFFF',
  focusRing: '#E02643',
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

export const themes = { light: lightTheme, dark: darkTheme } as const
export type ThemeName = keyof typeof themes
