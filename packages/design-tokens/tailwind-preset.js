/** Shared Tailwind preset: vravra spec tokens as utilities.
 * Apps extend this preset; values resolve to tokens.css CSS variables. */
module.exports = {
  theme: {
    extend: {
      colors: {
        brand: {
          crimson: 'var(--color-brand-crimson)',
          hover: 'var(--color-brand-crimson-hover)',
          glow: 'var(--color-brand-crimson-glow)',
        },
        studio: {
          cobalt: 'var(--color-studio-cobalt)',
          hover: 'var(--color-studio-cobalt-hover)',
          glow: 'var(--color-studio-cobalt-glow)',
        },
        live: 'var(--color-live-signal)',
        premiere: 'var(--color-premiere-amber)',
        canvas: 'var(--bg-canvas)',
        surface: 'var(--bg-surface)',
        elevated: 'var(--bg-elevated)',
        hairline: 'var(--border-hairline)',
      },
      textColor: {
        primary: 'var(--text-primary)',
        secondary: 'var(--text-secondary)',
        muted: 'var(--text-muted)',
      },
      fontFamily: {
        sans: ['Plus Jakarta Sans', 'Pretendard Variable', 'Pretendard', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      fontSize: {
        display: ['2rem', { lineHeight: '1.2', letterSpacing: '-0.02em', fontWeight: '700' }],
        'watch-title': ['1.25rem', { lineHeight: '1.35', letterSpacing: '-0.015em', fontWeight: '700' }],
        'section-title': ['1.125rem', { lineHeight: '1.4', fontWeight: '600' }],
        'card-title': ['0.9375rem', { lineHeight: '1.4', fontWeight: '600' }],
      },
      zIndex: {
        60: '60',
      },
    },
  },
}
