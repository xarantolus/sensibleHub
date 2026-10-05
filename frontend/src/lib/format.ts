export function formatDuration(seconds: number): string {
  const total = Math.max(0, Math.floor(seconds))
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  const ss = String(s).padStart(2, '0')
  if (h > 0) {
    return `${String(h)}:${String(m).padStart(2, '0')}:${ss}`
  }
  return `${String(m)}:${ss}`
}
