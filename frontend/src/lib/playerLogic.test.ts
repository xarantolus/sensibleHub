import { describe, expect, it } from 'vitest'

import {
  classifyMediaError,
  failureAction,
  fileTime,
  loudnessGain,
  needsRefill,
  refillExclude,
  refillSeed,
  retryDelay,
  songTime,
} from './playerLogic'

describe('refill', () => {
  it('follows the last queued song so suggestions chain', () => {
    expect(refillSeed('cur', ['a', 'b'])).toBe('b')
    expect(refillSeed('cur', [])).toBe('cur')
    expect(refillSeed(undefined, [])).toBeUndefined()
  })

  it('refills only when the queue runs low', () => {
    expect(needsRefill(0)).toBe(true)
    expect(needsRefill(2)).toBe(true)
    expect(needsRefill(3)).toBe(false)
  })

  it('excludes current, queued and recent songs without duplicates', () => {
    const history = Array.from({ length: 80 }, (_, i) => `h${String(i)}`)
    const ex = refillExclude('cur', ['q1', 'h79'], history)
    expect(ex).toContain('cur')
    expect(ex).toContain('q1')
    expect(ex).toContain('h79')
    expect(ex).not.toContain('h0')
    expect(new Set(ex).size).toBe(ex.length)
  })
})

describe('playback failures', () => {
  it('keeps the song and retries on network errors, even after the MP3 fallback', () => {
    expect(failureAction(classifyMediaError(2), false)).toBe('retry-later')
    expect(failureAction(classifyMediaError(2), true)).toBe('retry-later')
  })

  it('falls back to MP3 once for undecodable audio, then skips', () => {
    expect(failureAction(classifyMediaError(4), false)).toBe('try-mp3')
    expect(failureAction(classifyMediaError(3), true)).toBe('skip')
  })

  it('ignores aborted loads, which happen when switching songs', () => {
    expect(failureAction(classifyMediaError(1), false)).toBe('ignore')
  })

  it('backs off exponentially up to 30 seconds', () => {
    expect(retryDelay(0)).toBe(1000)
    expect(retryDelay(3)).toBe(8000)
    expect(retryDelay(20)).toBe(30_000)
  })
})

describe('trimmed time', () => {
  const pb = { start: 10, end: 70 }

  it('maps between file and song positions within the trim', () => {
    expect(songTime(25, pb)).toBe(15)
    expect(songTime(5, pb)).toBe(0)
    expect(songTime(90, pb)).toBe(60)
    expect(fileTime(15, pb)).toBe(25)
    expect(fileTime(-3, pb)).toBe(10)
    expect(fileTime(100, pb)).toBe(70)
  })
})

describe('loudness normalisation', () => {
  it('turns loud songs down to the target and leaves quiet ones at full volume', () => {
    expect(loudnessGain(-10)).toBeCloseTo(0.501, 2)
    expect(loudnessGain(-16)).toBe(1)
    expect(loudnessGain(-25)).toBe(1)
  })

  it('does not change songs without a measurement', () => {
    expect(loudnessGain(undefined)).toBe(1)
    expect(loudnessGain(Number.NEGATIVE_INFINITY)).toBe(1)
  })
})
