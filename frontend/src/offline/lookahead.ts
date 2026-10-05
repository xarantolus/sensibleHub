import type { QueryClient } from '@tanstack/vue-query'
import { watch } from 'vue'

import { audioUrl, coverUrl } from '@/api/media'
import { keys } from '@/api/queries'
import type { SongSummary } from '@/api/schema'
import { setOfflineCandidates, usePlayer } from '@/stores/player'

import { audioCacheName } from './cacheNames'
import { evictionOrder, type CacheEntry } from './budget'

const budgetBytes = 300 * 1024 * 1024
const lookaheadSongs = 3
const metaKey = 'sh-audio-cache-v1'

type Meta = Record<string, CacheEntry>

function loadMeta(): Meta {
  try {
    return JSON.parse(localStorage.getItem(metaKey) ?? '{}') as Meta
  } catch {
    return {}
  }
}

function saveMeta(meta: Meta): void {
  try {
    localStorage.setItem(metaKey, JSON.stringify(meta))
  } catch {
    // Without metadata the budget is only enforced approximately.
  }
}

function songIdFromAudioUrl(url: string): string | undefined {
  return /\/media\/songs\/([^/]+)\/audio/.exec(new URL(url, location.origin).pathname)?.[1]
}

function saveData(): boolean {
  const conn = (navigator as Navigator & { connection?: { saveData?: boolean } }).connection
  return conn?.saveData === true
}

/**
 * Downloads the playing song and the next few queued ones into Cache Storage,
 * where the service worker serves them from when the network is gone. Needs a
 * secure context (HTTPS or localhost); elsewhere it does nothing.
 */
export function startLookahead(qc: QueryClient): void {
  if (!('caches' in window) || !window.isSecureContext) {
    return
  }
  void navigator.storage.persist().catch(() => false)

  const player = usePlayer()
  const songById = (id: string): SongSummary | undefined =>
    qc.getQueryData<readonly SongSummary[]>(keys.songs)?.find((s) => s.id === id)

  setOfflineCandidates(async () => {
    const cache = await caches.open(audioCacheName)
    const requests = await cache.keys()
    return requests.flatMap((r) => songIdFromAudioUrl(r.url) ?? [])
  })

  let requested = 0
  let handled = 0
  let running = false

  const wanted = (): SongSummary[] =>
    [player.currentId, ...player.upcoming.slice(0, lookaheadSongs)].flatMap((id) => (id === undefined ? [] : (songById(id) ?? [])))

  async function run(): Promise<void> {
    requested++
    if (running) {
      return
    }
    running = true
    try {
      while (handled < requested && navigator.onLine && !saveData()) {
        handled = requested
        const songs = wanted()
        const pinned = new Set(songs.map((s) => audioUrl(s)))
        for (const song of songs) {
          await prefetch(song, pinned)
        }
      }
    } finally {
      running = false
    }
  }

  watch(
    () => [player.currentId, ...player.upcoming.slice(0, lookaheadSongs)].join(),
    () => void run(),
    { immediate: true },
  )
  window.addEventListener('online', () => void run())
}

async function prefetch(song: SongSummary, pinned: ReadonlySet<string>): Promise<void> {
  const url = audioUrl(song)
  const cache = await caches.open(audioCacheName)
  let meta = loadMeta()

  if ((await cache.match(url)) !== undefined) {
    const entry = meta[url]
    if (entry !== undefined) {
      entry.used = Date.now()
      saveMeta(meta)
    }
    return
  }

  try {
    const res = await fetch(url)
    if (res.status !== 200) {
      return
    }
    const blob = await res.blob()

    const victims = evictionOrder(meta, pinned, blob.size, budgetBytes)
    for (const victim of victims) {
      await cache.delete(victim)
    }
    meta = Object.fromEntries(Object.entries(meta).filter(([u]) => !victims.includes(u)))

    // Keep the network response's validators (Last-Modified etc.): a song may switch
    // from network to cache mid-playback, and the media stack aborts if they differ.
    const headers = new Headers(res.headers)
    headers.set('Content-Length', String(blob.size))
    headers.delete('Content-Encoding')
    await cache.put(url, new Response(blob, { headers }))
    meta[url] = { size: blob.size, used: Date.now() }
    saveMeta(meta)

    if (song.cover !== undefined) {
      await Promise.allSettled([fetch(coverUrl(song, 'small')), fetch(coverUrl(song, 'full'))])
    }
  } catch {
    // Network dropped mid-download; it is retried on the next queue change or when back online.
  }
}
