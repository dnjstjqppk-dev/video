import { describe, it, expect } from 'vitest'

describe('Admin Governance & Safety Utilities', () => {
  it('evaluates 3-strikes suspension rule', () => {
    const isAccountSuspended = (copyrightStrikes: number, communityStrikes: number): boolean => {
      return copyrightStrikes >= 3 || communityStrikes >= 3
    }

    expect(isAccountSuspended(0, 0)).toBe(false)
    expect(isAccountSuspended(2, 1)).toBe(false)
    expect(isAccountSuspended(3, 0)).toBe(true)
    expect(isAccountSuspended(0, 3)).toBe(true)
    expect(isAccountSuspended(4, 4)).toBe(true)
  })

  it('determines live stream eligibility from community strikes', () => {
    const canLiveStream = (communityStrikes: number, isVerified: boolean): boolean => {
      return isVerified && communityStrikes === 0
    }

    expect(canLiveStream(0, true)).toBe(true)
    expect(canLiveStream(1, true)).toBe(false)
    expect(canLiveStream(0, false)).toBe(false)
  })
})
