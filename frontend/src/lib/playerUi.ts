export const dismissDistance = 120
export const dismissVelocity = 0.5

/**
 * Whether letting go of the player sheet after dragging it `dy` pixels down in
 * `ms` milliseconds should close it: dragged far enough, or flicked down fast.
 */
export function shouldDismiss(dy: number, ms: number): boolean {
  if (dy <= 0) {
    return false
  }
  return dy >= dismissDistance || (dy >= 30 && dy / Math.max(ms, 1) >= dismissVelocity)
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
