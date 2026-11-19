/** vravra brand constants (SSOT). No hardcoded brand strings in components. */
export const BRAND = {
  name: 'vravra',
  nameKo: '브라브라',
  sloganEn: 'Play Your Vibe. Share Your Voice.',
  sloganKo: '너의 바이브를 플레이하다, 너의 목소리를 공유하다',
  keywords: ['Vibe', 'Voice', 'Vivid'],
  studio: {
    name: 'vravra studio',
    sloganEn: 'Empower Your Vision. Scale Your Vibe.',
    sloganKo: '크리에이터의 모든 것, vravra 스튜디오',
  },
  colors: {
    crimson: '#FF2E4C',
    crimsonHover: '#E0223E',
    cobalt: '#2563EB',
    cobaltHover: '#1D4ED8',
    live: '#FF3B30',
    premiere: '#FF9F1C',
  },
} as const

export type Brand = typeof BRAND
