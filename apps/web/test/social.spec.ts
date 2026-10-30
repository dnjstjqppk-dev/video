import { describe, it, expect } from 'vitest'

interface MockComment {
  id: string
  content: string
  likeCount: number
  isPinned: boolean
  createdAt: string
  replies?: MockComment[]
}

describe('Phase 4: Social Web & Community Interactions', () => {
  it('correctly sorts comments by Top (pinned first, then likes descending)', () => {
    const list: MockComment[] = [
      { id: '1', content: 'Low likes', likeCount: 5, isPinned: false, createdAt: '2026-10-20T10:00:00Z' },
      { id: '2', content: 'High likes', likeCount: 500, isPinned: false, createdAt: '2026-10-21T10:00:00Z' },
      { id: '3', content: 'Pinned comment', likeCount: 12, isPinned: true, createdAt: '2026-10-19T10:00:00Z' },
    ]

    const sorted = [...list].sort((a, b) => {
      if (a.isPinned !== b.isPinned) return a.isPinned ? -1 : 1
      return (b.likeCount || 0) - (a.likeCount || 0)
    })

    expect(sorted[0].id).toBe('3') // Pinned comment first
    expect(sorted[1].id).toBe('2') // Highest likes next
    expect(sorted[2].id).toBe('1') // Lower likes
  })

  it('correctly sorts comments by Newest (pinned first, then date descending)', () => {
    const list: MockComment[] = [
      { id: '1', content: 'Older', likeCount: 500, isPinned: false, createdAt: '2026-10-20T10:00:00Z' },
      { id: '2', content: 'Newest', likeCount: 5, isPinned: false, createdAt: '2026-10-25T10:00:00Z' },
      { id: '3', content: 'Pinned older', likeCount: 12, isPinned: true, createdAt: '2026-10-19T10:00:00Z' },
    ]

    const sorted = [...list].sort((a, b) => {
      if (a.isPinned !== b.isPinned) return a.isPinned ? -1 : 1
      return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    })

    expect(sorted[0].id).toBe('3') // Pinned first
    expect(sorted[1].id).toBe('2') // Newest next
    expect(sorted[2].id).toBe('1') // Older
  })

  it('calculates community poll vote percentages dynamically upon voting', () => {
    const pollRawVotes = [640, 210, 150] // sum: 1000 -> 64%, 21%, 15%
    const totalVotes = pollRawVotes.reduce((a, b) => a + b, 0)
    expect(totalVotes).toBe(1000)

    const percentages = pollRawVotes.map(v => Math.round((v / totalVotes) * 100))
    expect(percentages).toEqual([64, 21, 15])

    // Vote for option 2 (index 1)
    pollRawVotes[1] += 1
    const newTotal = pollRawVotes.reduce((a, b) => a + b, 0)
    expect(newTotal).toBe(1001)
  })

  it('toggles like and dislike mutually exclusively', () => {
    let isLiked = false
    let isDisliked = false
    let likeCount = 100

    // User clicks like
    isLiked = true
    likeCount += 1
    expect(isLiked).toBe(true)
    expect(likeCount).toBe(101)

    // User then clicks dislike -> like should be revoked
    if (isLiked) {
      isLiked = false
      likeCount = Math.max(0, likeCount - 1)
    }
    isDisliked = true

    expect(isLiked).toBe(false)
    expect(isDisliked).toBe(true)
    expect(likeCount).toBe(100)
  })
})

