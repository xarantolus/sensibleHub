import { describe, expect, it } from 'vitest'

import { evictionOrder } from './budget'

describe('evictionOrder', () => {
  const entries = {
    old: { size: 40, used: 1 },
    mid: { size: 40, used: 2 },
    playing: { size: 40, used: 0 },
    recent: { size: 40, used: 3 },
  }

  it('evicts nothing while within budget', () => {
    expect(evictionOrder(entries, new Set(), 10, 200)).toEqual([])
  })

  it('evicts least recently used first until the new file fits', () => {
    expect(evictionOrder(entries, new Set(), 50, 160)).toEqual(['playing', 'old'])
  })

  it('never evicts pinned files, even when they are the oldest', () => {
    expect(evictionOrder(entries, new Set(['playing']), 50, 160)).toEqual(['old', 'mid'])
  })

  it('evicts everything unpinned when the new file alone is too big', () => {
    expect(evictionOrder(entries, new Set(['playing']), 500, 100)).toEqual(['old', 'mid', 'recent'])
  })
})
