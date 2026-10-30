import { describe, it, expect } from 'vitest'

describe('Web Platform Utilities', () => {
  it('formats video duration correctly', () => {
    const formatDuration = (seconds: number): string => {
      const h = Math.floor(seconds / 3600)
      const m = Math.floor((seconds % 3600) / 60)
      const s = seconds % 60
      if (h > 0) {
        return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
      }
      return `${m}:${s.toString().padStart(2, '0')}`
    }

    expect(formatDuration(45)).toBe('0:45')
    expect(formatDuration(125)).toBe('2:05')
    expect(formatDuration(3665)).toBe('1:01:05')
  })

  it('formats view counts into human-readable strings', () => {
    const formatViews = (views: number): string => {
      if (views >= 1_000_000) {
        return `${(views / 1_000_000).toFixed(1)}M`
      }
      if (views >= 1_000) {
        return `${(views / 1_000).toFixed(1)}K`
      }
      return views.toString()
    }

    expect(formatViews(950)).toBe('950')
    expect(formatViews(14200)).toBe('14.2K')
    expect(formatViews(3500000)).toBe('3.5M')
  })
})
