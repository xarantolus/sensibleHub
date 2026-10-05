import { callVoid, client, isApiError } from '@/api/client'
import type { PlayReport } from '@/api/schema'

const storageKey = 'sh-play-reports-v1'
const maxPending = 1000
const flushDelay = 3000

let pending: PlayReport[] = load()
let timer: ReturnType<typeof setTimeout> | undefined
let flushing = false

function load(): PlayReport[] {
  try {
    const raw = localStorage.getItem(storageKey)
    return raw === null ? [] : (JSON.parse(raw) as PlayReport[])
  } catch {
    return []
  }
}

function save(): void {
  try {
    localStorage.setItem(storageKey, JSON.stringify(pending))
  } catch {
    // Without storage, reports made offline are lost on reload; suggestions just learn slower.
  }
}

/**
 * Records that a song stopped playing. Reports are sent in batches and kept
 * across reloads while offline, so skips on the go still count.
 */
export function reportPlay(report: PlayReport): void {
  pending = [...pending, report].slice(-maxPending)
  save()
  scheduleFlush()
}

function scheduleFlush(): void {
  clearTimeout(timer)
  timer = setTimeout(() => void flush(), flushDelay)
}

export async function flush(): Promise<void> {
  if (flushing || pending.length === 0 || !navigator.onLine) {
    return
  }
  flushing = true
  const batch = pending
  try {
    await callVoid(client.POST('/api/v1/player/plays', { body: { plays: batch } }))
    pending = pending.slice(batch.length)
    save()
  } catch (err) {
    // Network trouble: keep the reports for later. Anything the server rejects
    // would be rejected again, so it is dropped.
    if (!isApiError(err, 'network') && !(isApiError(err) && err.status >= 500)) {
      pending = pending.slice(batch.length)
      save()
    }
  } finally {
    flushing = false
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('online', () => void flush())
  window.addEventListener('pagehide', () => void flush())
  if (pending.length > 0) {
    scheduleFlush()
  }
}
