/** Motion: 180ms cubic-bezier(0.2,0,0,1) default, CLS-safe. */
export const motion = {
  fast: '120ms',
  base: '180ms',
  slow: '280ms',
  ease: 'cubic-bezier(0.2, 0, 0, 1)',
  hoverPreviewDelayMs: 400,
  dropdown: '180ms cubic-bezier(0.2, 0, 0, 1)',
} as const

export const radius = {
  sm: '0.375rem',
  md: '0.5rem',
  lg: '0.75rem',
  xl: '1rem',
  thumb: '12px',
  pill: '9999px',
} as const

export const shadow = {
  sm: '0 1px 2px 0 rgba(2, 6, 18, 0.35)',
  md: '0 8px 24px -8px rgba(2, 6, 18, 0.45)',
  lg: '0 20px 48px -12px rgba(2, 6, 18, 0.55)',
  glow: '0 0 0 3px rgba(99, 102, 241, 0.22)',
} as const

export const zIndex = {
  header: 40,
  drawer: 45,
  dropdown: 50,
  modal: 60,
  toast: 70,
} as const

export type MotionTokens = typeof motion
