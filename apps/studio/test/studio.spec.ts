import { describe, it, expect } from 'vitest'

describe('Creator Studio Utilities', () => {
  it('calculates creator revenue split according to 55/45 platform rule', () => {
    const calculateCreatorRevenue = (grossAdRevenue: number, splitRatio = 0.55): number => {
      return Math.round(grossAdRevenue * splitRatio * 100) / 100
    }

    expect(calculateCreatorRevenue(1000)).toBe(550)
    expect(calculateCreatorRevenue(2500)).toBe(1375)
    expect(calculateCreatorRevenue(0)).toBe(0)
  })

  it('validates video upload formats and metadata constraints', () => {
    const isValidUpload = (meta: { title: string; duration: number; mimeType: string }): boolean => {
      const allowedMimes = ['video/mp4', 'video/quicktime', 'video/webm']
      return (
        meta.title.trim().length > 0 &&
        meta.title.length <= 100 &&
        meta.duration > 0 &&
        meta.duration <= 43200 && // 12 hours max
        allowedMimes.includes(meta.mimeType)
      )
    }

    expect(isValidUpload({ title: 'My Video', duration: 120, mimeType: 'video/mp4' })).toBe(true)
    expect(isValidUpload({ title: '', duration: 120, mimeType: 'video/mp4' })).toBe(false)
    expect(isValidUpload({ title: 'Too Long', duration: 50000, mimeType: 'video/mp4' })).toBe(false)
    expect(isValidUpload({ title: 'Invalid Mime', duration: 60, mimeType: 'image/png' })).toBe(false)
  })
})

