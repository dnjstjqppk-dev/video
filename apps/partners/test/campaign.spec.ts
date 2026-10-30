import { describe, it, expect } from 'vitest'

describe('Partner & Advertiser Engine Utilities', () => {
  it('calculates campaign budget burn pacing percentage', () => {
    const calculatePacing = (spent: number, totalBudget: number): number => {
      if (totalBudget <= 0) return 0
      return Math.min(100, Math.round((spent / totalBudget) * 100))
    }

    expect(calculatePacing(250, 1000)).toBe(25)
    expect(calculatePacing(800, 1000)).toBe(80)
    expect(calculatePacing(1200, 1000)).toBe(100)
    expect(calculatePacing(0, 500)).toBe(0)
  })

  it('determines effective eCPM from gross spend and valid impressions', () => {
    const calculateECPM = (spendDollars: number, impressions: number): number => {
      if (impressions <= 0) return 0
      return Math.round((spendDollars / impressions) * 1000 * 100) / 100
    }

    expect(calculateECPM(50, 10000)).toBe(5)
    expect(calculateECPM(120, 20000)).toBe(6)
    expect(calculateECPM(0, 1000)).toBe(0)
  })
})

