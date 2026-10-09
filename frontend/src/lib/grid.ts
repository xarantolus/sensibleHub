const maxJumpTargets = 26

/**
 * Labels for the jump bar. Short listings get one target per group; long ones
 * are condensed to decades (for years) or first letters, pointing at the first
 * group of each.
 */
export function jumpTargets(groups: readonly { key: string; title: string }[]): { key: string; label: string }[] {
  if (groups.length <= maxJumpTargets) {
    return groups.map((g) => ({ key: g.key, label: g.title }))
  }
  const years = groups.every((g) => /^\d{4}$/.test(g.title))
  const seen = new Set<string>()
  const out: { key: string; label: string }[] = []
  for (const g of groups) {
    const label = years ? `${g.title.slice(0, 3)}0s` : (g.title.trim()[0]?.toUpperCase() ?? '#')
    if (!seen.has(label)) {
      seen.add(label)
      out.push({ key: g.key, label })
    }
  }
  return out
}

export function chunk<T>(items: readonly T[], size: number): T[][] {
  const out: T[][] = []
  const n = Math.max(1, Math.floor(size))
  for (let i = 0; i < items.length; i += n) {
    out.push(items.slice(i, i + n))
  }
  return out
}
