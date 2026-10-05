import { describe, expect, it } from 'vitest'

import { parseStatsForNerds, progressPercent, shouldDismiss } from './playerUi'

describe('shouldDismiss', () => {
  it('closes after a long slow drag', () => {
    expect(shouldDismiss(140, 1500)).toBe(true)
  })
  it('closes after a short fast flick', () => {
    expect(shouldDismiss(60, 80)).toBe(true)
  })
  it('snaps back after a short slow drag or a tiny twitch', () => {
    expect(shouldDismiss(60, 600)).toBe(false)
    expect(shouldDismiss(10, 5)).toBe(false)
  })
  it('never closes when dragged up', () => {
    expect(shouldDismiss(-200, 50)).toBe(false)
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
