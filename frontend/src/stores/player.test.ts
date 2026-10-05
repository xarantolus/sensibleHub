import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { usePlayer } from './player'

vi.mock('@/lib/playReports', () => ({ reportPlay: vi.fn() }))

beforeEach(() => {
  const data = new Map<string, string>()
  vi.stubGlobal('localStorage', {
    getItem: (k: string) => data.get(k) ?? null,
    setItem: (k: string, v: string) => data.set(k, v),
    removeItem: (k: string) => data.delete(k),
  })
  setActivePinia(createPinia())
  vi.stubGlobal('fetch', () => Promise.reject(new TypeError('offline in tests')))
})

function withAutoplay(...ids: string[]) {
  const p = usePlayer()
  p.autoplay = ids
  return p
}

describe('up next', () => {
  it('play all starts the first song and queues the rest', () => {
    const p = usePlayer()
    p.playNow(['a', 'b', 'c'])
    expect(p.currentId).toBe('a')
    expect(p.queue).toEqual(['b', 'c'])
  })

  it('add to queue appends after queued songs but before autoplay', () => {
    const p = withAutoplay('x', 'y')
    p.currentId = 'now'
    p.enqueue('a')
    p.enqueue(['b', 'c'])
    expect(p.upcoming).toEqual(['a', 'b', 'c', 'x', 'y'])
  })

  it('play next goes to the front of the queue', () => {
    const p = withAutoplay('x')
    p.currentId = 'now'
    p.enqueue('a')
    p.playNext('b')
    expect(p.upcoming).toEqual(['b', 'a', 'x'])
  })

  it('queueing a suggested song moves it out of autoplay', () => {
    const p = withAutoplay('x', 'y')
    p.currentId = 'now'
    p.enqueue('y')
    expect(p.queue).toEqual(['y'])
    expect(p.autoplay).toEqual(['x'])
  })

  it('plays the queue before autoplay', async () => {
    const p = withAutoplay('x')
    p.currentId = 'now'
    p.enqueue('a')
    await p.next()
    expect(p.currentId).toBe('a')
    await p.next()
    expect(p.currentId).toBe('x')
  })

  it('starts playing when queueing while idle', () => {
    const p = usePlayer()
    p.enqueue(['a', 'b'])
    expect(p.currentId).toBe('a')
    expect(p.queue).toEqual(['b'])
  })

  it('moves songs within and between the lists', () => {
    const p = withAutoplay('x', 'y', 'z')
    p.currentId = 'now'
    p.enqueue(['a', 'b'])

    p.move('autoplay', 1, 'queue', 1)
    expect(p.queue).toEqual(['a', 'y', 'b'])
    expect(p.autoplay).toEqual(['x', 'z'])

    p.move('queue', 0, 'autoplay', 2)
    expect(p.queue).toEqual(['y', 'b'])
    expect(p.autoplay).toEqual(['x', 'z', 'a'])

    p.move('queue', 1, 'queue', 0)
    expect(p.queue).toEqual(['b', 'y'])
  })

  it('play now from a list keeps the other songs', () => {
    const p = withAutoplay('x', 'y')
    p.currentId = 'now'
    p.enqueue('a')
    p.playFrom('autoplay', 1)
    expect(p.currentId).toBe('y')
    expect(p.upcoming).toEqual(['a', 'x'])
  })
})
