export const swipeCloseDistance = 80

export function isSwipeDown(dx: number, dy: number): boolean {
  return dy >= swipeCloseDistance && dy > Math.abs(dx) * 1.5
}

export function progressPercent(position: number, duration: number): number {
  if (!(duration > 0)) {
    return 0
  }
  return Math.min(100, Math.max(0, (position / duration) * 100))
}

export function parseStatsForNerds(raw: string | null): boolean {
  if (raw === null) {
    return false
  }
  try {
    const parsed: unknown = JSON.parse(raw)
    return typeof parsed === 'object' && parsed !== null && 'statsForNerds' in parsed && parsed.statsForNerds === true
  } catch {
    return false
  }
}

export function moveTarget(index: number, delta: -1 | 1, length: number): number | undefined {
  const to = index + delta
  return to < 0 || to >= length ? undefined : to
}
