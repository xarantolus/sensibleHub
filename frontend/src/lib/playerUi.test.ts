import { describe, expect, it } from 'vitest'

import { isSwipeDown, parseStatsForNerds, progressPercent } from './playerUi'

describe('isSwipeDown', () => {
  it('accepts a long mostly vertical drag', () => {
    expect(isSwipeDown(10, 120)).toBe(true)
  })
  it('rejects short drags', () => {
    expect(isSwipeDown(0, 40)).toBe(false)
  })
  it('rejects diagonal and upward drags', () => {
    expect(isSwipeDown(100, 110)).toBe(false)
    expect(isSwipeDown(0, -200)).toBe(false)
  })
})

describe('progressPercent', () => {
  it('clamps and handles unknown duration', () => {
    expect(progressPercent(30, 120)).toBe(25)
    expect(progressPercent(5, 0)).toBe(0)
    expect(progressPercent(500, 100)).toBe(100)
    expect(progressPercent(-5, 100)).toBe(0)
  })
})

describe('parseStatsForNerds', () => {
  it('reads the stored flag', () => {
    expect(parseStatsForNerds('{"statsForNerds":true}')).toBe(true)
    expect(parseStatsForNerds('{"statsForNerds":false}')).toBe(false)
  })
  it('falls back to false for missing or broken data', () => {
    expect(parseStatsForNerds(null)).toBe(false)
    expect(parseStatsForNerds('nope')).toBe(false)
    expect(parseStatsForNerds('{"statsForNerds":"yes"}')).toBe(false)
    expect(parseStatsForNerds('3')).toBe(false)
  })
})
