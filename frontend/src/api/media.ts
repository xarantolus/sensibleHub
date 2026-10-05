import type { SongSummary } from './schema'

type SongRef = Pick<SongSummary, 'id' | 'lastEdit'>

const enc = encodeURIComponent

/** The version parameter changes on every edit so browsers never show a replaced cover. */
export function coverUrl(song: SongRef, size: 'small' | 'full' = 'small'): string {
  const v = enc(song.lastEdit)
  return size === 'small'
    ? `/media/songs/${enc(song.id)}/cover?size=small&v=${v}`
    : `/media/songs/${enc(song.id)}/cover?v=${v}`
}

export function audioUrl(song: SongRef): string {
  return `/media/songs/${enc(song.id)}/audio?v=${enc(song.lastEdit)}`
}

export function mp3Url(song: SongRef): string {
  return `/media/songs/${enc(song.id)}/mp3?v=${enc(song.lastEdit)}`
}

export const placeholderCover = '/fav/android-chrome-512x512.png'
