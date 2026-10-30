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

  it('executes second-price auction pricing rule', () => {
    interface AdBid {
      campaignId: string
      cpmBidCents: number
    }

    function runSecondPriceAuction(bids: AdBid[]): { winnerId: string; clearingPriceCents: number } | null {
      if (!bids.length) return null
      const sorted = [...bids].sort((a, b) => b.cpmBidCents - a.cpmBidCents)
      const winner = sorted[0]
      if (!winner) return null
      const secondBid = sorted[1] ? sorted[1].cpmBidCents : Math.round(winner.cpmBidCents * 0.8)
      return {
        winnerId: winner.campaignId,
        clearingPriceCents: secondBid + 1
      }
    }

    const bids: AdBid[] = [
      { campaignId: 'camp_A', cpmBidCents: 500 },
      { campaignId: 'camp_B', cpmBidCents: 850 },
      { campaignId: 'camp_C', cpmBidCents: 320 },
    ]

    const result = runSecondPriceAuction(bids)
    expect(result).not.toBeNull()
    expect(result?.winnerId).toBe('camp_B')
    expect(result?.clearingPriceCents).toBe(501) // 2nd highest ($5.00) + 1 cent
  })

  it('verifies double-entry ledger balance invariant (Debits == Credits)', () => {
    interface LedgerEntry {
      account: string
      direction: 'DEBIT' | 'CREDIT'
      amount: number
    }

    function isLedgerBalanced(entries: LedgerEntry[]): boolean {
      let debits = 0
      let credits = 0
      for (const e of entries) {
        if (e.direction === 'DEBIT') debits += e.amount
        if (e.direction === 'CREDIT') credits += e.amount
      }
      return debits === credits
    }

    const validLedger: LedgerEntry[] = [
      { account: 'platform_escrow', direction: 'DEBIT', amount: 10000 },
      { account: 'creator_payable', direction: 'CREDIT', amount: 7000 },
      { account: 'platform_revenue', direction: 'CREDIT', amount: 3000 },
    ]
    expect(isLedgerBalanced(validLedger)).toBe(true)

    const unbalancedLedger: LedgerEntry[] = [
      { account: 'platform_escrow', direction: 'DEBIT', amount: 10000 },
      { account: 'creator_payable', direction: 'CREDIT', amount: 6000 },
    ]
    expect(isLedgerBalanced(unbalancedLedger)).toBe(false)
  })

  it('evaluates MRC standard video viewability (50% pixels for >= 2.0s)', () => {
    function isMrcViewable(pixelRatio: number, continuousSeconds: number): boolean {
      return pixelRatio >= 0.50 && continuousSeconds >= 2.0
    }

    expect(isMrcViewable(0.8, 2.5)).toBe(true)
    expect(isMrcViewable(0.5, 2.0)).toBe(true)
    expect(isMrcViewable(0.49, 3.0)).toBe(false) // Insufficient pixel area
    expect(isMrcViewable(0.9, 1.8)).toBe(false) // Insufficient duration
  })
})
