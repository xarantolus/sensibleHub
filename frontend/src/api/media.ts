import { reactive } from 'vue'

import type { SongSummary } from './schema'

type SongRef = Pick<SongSummary, 'id'>

const enc = encodeURIComponent

/**
 * Covers are revalidated by the browser (no-cache + ETag), but within a page it
 * reuses an image it already shows for the same URL. When a song changes, its
 * revision bumps a URL fragment so the new cover is fetched; fragments never
 * reach the server.
 */
const coverRevision = reactive(new Map<string, number>())

export function bumpCoverRevision(id: string): void {
  coverRevision.set(id, (coverRevision.get(id) ?? 0) + 1)
}

export function coverUrl(song: SongRef, size: 'small' | 'full' = 'small'): string {
  const rev = coverRevision.get(song.id)
  const base = size === 'small' ? `/media/songs/${enc(song.id)}/cover?size=small` : `/media/songs/${enc(song.id)}/cover`
  return rev === undefined ? base : `${base}#${String(rev)}`
}

export function audioUrl(song: SongRef): string {
  return `/media/songs/${enc(song.id)}/audio`
}

export function mp3Url(song: SongRef): string {
  return `/media/songs/${enc(song.id)}/mp3`
}

export const placeholderCover = '/fav/android-chrome-512x512.png'
