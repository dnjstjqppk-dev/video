export const tokens = {
  colors: {
    primary: {
      50: '#fef2f2',
      100: '#fee2e2',
      200: '#fecaca',
      300: '#fca5a5',
      400: '#f87171',
      500: '#ef4444',
      600: '#dc2626', // Brand Red
      700: '#b91c1c',
      800: '#991b1b',
      900: '#7f1d1d',
      950: '#450a0a',
    },
    neutral: {
      50: '#fafafa',
      100: '#f4f4f5',
      200: '#e4e4e7',
      300: '#d4d4d8',
      400: '#a1a1aa',
      500: '#71717a',
      600: '#52525b',
      700: '#3f3f46',
      800: '#27272a',
      900: '#18181b',
      950: '#09090b', // Deep Dark Canvas
    },
    accent: {
      cyan: '#06b6d4',
      purple: '#a855f7',
      green: '#22c55e',
      amber: '#f59e0b',
    }
  },
  breakpoints: {
    xs: '360px',
    sm: '390px',
    md: '768px',
    lg: '1024px',
    xl: '1440px',
    '2xl': '1920px'
  },
  typography: {
    fontSans: 'Pretendard, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", sans-serif',
    fontMono: '"JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace'
  }
} as const;

export type DesignTokens = typeof tokens;

