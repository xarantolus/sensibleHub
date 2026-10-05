export interface CacheEntry {
  size: number
  used: number
}

/**
 * Picks cached files to delete so that adding `incoming` bytes stays within
 * `budget`: least recently used first, never anything pinned (playing or queued).
 */
export function evictionOrder(
  entries: Readonly<Record<string, CacheEntry>>,
  pinned: ReadonlySet<string>,
  incoming: number,
  budget: number,
): string[] {
  let total = incoming
  for (const e of Object.values(entries)) {
    total += e.size
  }
  const victims: string[] = []
  const candidates = Object.entries(entries)
    .filter(([url]) => !pinned.has(url))
    .sort(([, a], [, b]) => a.used - b.used)
  for (const [url, e] of candidates) {
    if (total <= budget) {
      break
    }
    victims.push(url)
    total -= e.size
  }
  return victims
}
