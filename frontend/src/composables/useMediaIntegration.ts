import { onBeforeUnmount, watch, watchEffect, type Ref } from 'vue'

import { coverUrl } from '@/api/media'
import type { SongSummary } from '@/api/schema'
import { usePlayer } from '@/stores/player'

const previewSize = 240
const defaultThemeColor = '#baffda'
const seekStep = 10

interface AudioSession {
  type: string
}

/**
 * Connects the player to the operating system: lock screen and notification
 * controls with artwork, media keys, the tab title and the browser's theme
 * colour. Every API is optional and skipped where the browser lacks it.
 */
export function useMediaIntegration(audio: Readonly<Ref<HTMLAudioElement | null>>, song: Readonly<Ref<SongSummary | undefined>>): void {
  const player = usePlayer()
  const session = 'mediaSession' in navigator ? navigator.mediaSession : undefined

  const audioSession = (navigator as Navigator & { audioSession?: AudioSession }).audioSession
  if (audioSession !== undefined) {
    audioSession.type = 'playback'
  }

  if (session !== undefined) {
    const handlers: [MediaSessionAction, MediaSessionActionHandler][] = [
      ['play', () => (player.playing = true)],
      ['pause', () => (player.playing = false)],
      ['stop', () => {
        player.stop()
      }],
      ['previoustrack', () => {
        player.previous()
      }],
      ['nexttrack', () => void player.next()],
      ['seekbackward', (d) => {
        player.seekBy(-(d.seekOffset ?? seekStep))
      }],
      ['seekforward', (d) => {
        player.seekBy(d.seekOffset ?? seekStep)
      }],
      ['seekto', (d) => {
        if (d.seekTime !== undefined) {
          player.seekTo(d.seekTime)
        }
      }],
    ]
    for (const [action, handler] of handlers) {
      try {
        session.setActionHandler(action, handler)
      } catch {
        // The browser does not support this action.
      }
    }
    onBeforeUnmount(() => {
      for (const [action] of handlers) {
        try {
          session.setActionHandler(action, null)
        } catch {
          // Not supported, nothing to remove.
        }
      }
    })

    watch(
      song,
      (s) => {
        if (s === undefined) {
          session.metadata = null
          return
        }
        const artwork: MediaImage[] =
          s.cover === undefined
            ? [{ src: '/fav/android-chrome-512x512.png', sizes: '512x512', type: 'image/png' }]
            : [
                { src: coverUrl(s, 'small'), sizes: `${String(previewSize)}x${String(previewSize)}`, type: 'image/jpeg' },
                { src: coverUrl(s, 'full'), sizes: `${String(s.cover.size)}x${String(s.cover.size)}` },
              ]
        session.metadata = new MediaMetadata({ title: s.title, artist: s.artist ?? '', album: s.album ?? '', artwork })
      },
      { immediate: true },
    )

    watchEffect(() => {
      session.playbackState = player.currentId === undefined ? 'none' : player.playing ? 'playing' : 'paused'
    })

    watch([() => Math.floor(player.position), () => player.duration], ([position, duration]) => {
      if (duration <= 0 || !('setPositionState' in session)) {
        return
      }
      try {
        session.setPositionState({
          duration,
          position: Math.min(position, duration),
          playbackRate: audio.value?.playbackRate ?? 1,
        })
      } catch {
        // Invalid state while a song is still loading; the next update fixes it.
      }
    })
  }

  const baseTitle = document.title
  watchEffect(() => {
    const s = song.value
    document.title = s !== undefined && player.playing ? `▶ ${s.title}${s.artist ? ` · ${s.artist}` : ''}` : baseTitle
  })

  const themeMeta = ensureThemeMeta()
  watchEffect(() => {
    themeMeta.content = song.value?.cover?.color ?? defaultThemeColor
  })
}

function ensureThemeMeta(): HTMLMetaElement {
  let meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')
  if (meta === null) {
    meta = document.createElement('meta')
    meta.name = 'theme-color'
    document.head.append(meta)
  }
  return meta
}
