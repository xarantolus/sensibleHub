import { describe, expect, it } from 'vitest'

import { chunk, jumpTargets } from './grid'

const groups = (titles: string[]) => titles.map((title, i) => ({ key: String(i), title }))

describe('jumpTargets', () => {
  it('keeps one target per group for short listings', () => {
    expect(jumpTargets(groups(['A', 'B', 'C']))).toEqual([
      { key: '0', label: 'A' },
      { key: '1', label: 'B' },
      { key: '2', label: 'C' },
    ])
  })

  it('condenses many years into decades pointing at the first year of each', () => {
    const years = Array.from({ length: 40 }, (_, i) => String(2026 - i))
    const t = jumpTargets(groups(years))
    expect(t.map((x) => x.label)).toEqual(['2020s', '2010s', '2000s', '1990s', '1980s'])
    expect(t[1]).toEqual({ key: '7', label: '2010s' })
  })

  it('condenses many names into first letters', () => {
    const names = Array.from({ length: 30 }, (_, i) => `${i < 15 ? 'alpha' : 'beta'} ${String(i)}`)
    expect(jumpTargets(groups(names))).toEqual([
      { key: '0', label: 'A' },
      { key: '15', label: 'B' },
    ])
  })
})

describe('chunk', () => {
  it('splits into rows of the given size with a shorter last row', () => {
    expect(chunk([1, 2, 3, 4, 5], 2)).toEqual([[1, 2], [3, 4], [5]])
  })

  it('handles empty input and nonsense sizes', () => {
    expect(chunk([], 3)).toEqual([])
    expect(chunk([1, 2], 0)).toEqual([[1], [2]])
  })
})
