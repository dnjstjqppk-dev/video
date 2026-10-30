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

  it('validates channel handles (@handle format)', () => {
    const isValidHandle = (handle: string): boolean => {
      return /^@[a-zA-Z0-9_]{3,30}$/.test(handle)
    }

    expect(isValidHandle('@gvptech')).toBe(true)
    expect(isValidHandle('@creator_2026')).toBe(true)
    expect(isValidHandle('gvptech')).toBe(false) // missing @
    expect(isValidHandle('@ab')).toBe(false) // too short
    expect(isValidHandle('@invalid-symbol!')).toBe(false)
  })

  it('filters comments with automated word filter', () => {
    const blockedWords = ['scam_link', 'free_crypto', 'telegram_bot']
    const isSpam = (text: string): boolean => {
      const lower = text.toLowerCase()
      return blockedWords.some(w => lower.includes(w))
    }

    expect(isSpam('Join my telegram_bot now for profits')).toBe(true)
    expect(isSpam('Check out the new scam_link')).toBe(true)
    expect(isSpam('This HLS ABR tutorial is awesome and informative')).toBe(false)
  })

  it('processes bulk visibility and monetization operations', () => {
    const items = [
      { id: '1', selected: true, visibility: 'PUBLIC', monetized: false },
      { id: '2', selected: false, visibility: 'PUBLIC', monetized: false },
      { id: '3', selected: true, visibility: 'PUBLIC', monetized: false },
    ]

    // Bulk set private
    items.forEach(i => {
      if (i.selected) i.visibility = 'PRIVATE'
    })
    expect(items[0]?.visibility).toBe('PRIVATE')
    expect(items[1]?.visibility).toBe('PUBLIC')
    expect(items[2]?.visibility).toBe('PRIVATE')

    // Bulk enable monetization
    items.forEach(i => {
      if (i.selected) i.monetized = true
    })
    expect(items[0]?.monetized).toBe(true)
    expect(items[1]?.monetized).toBe(false)
    expect(items[2]?.monetized).toBe(true)
  })
})
