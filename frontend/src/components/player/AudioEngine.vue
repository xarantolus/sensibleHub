<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'

import { audioUrl, mp3Url } from '@/api/media'
import { useSongIndex } from '@/api/queries'
import { useMediaIntegration } from '@/composables/useMediaIntegration'
import { classifyMediaError, failureAction, fileTime, loudnessGain, retryDelay, songTime } from '@/lib/playerLogic'
import { notify } from '@/lib/notify'
import { usePlayer } from '@/stores/player'

const player = usePlayer()
const { index, songs } = useSongIndex()
const audio = useTemplateRef<HTMLAudioElement>('audio')

const song = computed(() => (player.currentId === undefined ? undefined : index.value.get(player.currentId)))
const useMp3 = ref(false)
const src = computed(() => {
  const s = song.value
  if (s === undefined) {
    return undefined
  }
  return useMp3.value ? mp3Url(s) : audioUrl(s)
})

let resumeAt = player.position
let retryAttempt = 0
let retryTimer: ReturnType<typeof setTimeout> | undefined

watch(
  () => player.currentId,
  (id, old) => {
    if (id !== old) {
      useMp3.value = false
      resumeAt = 0
      retryAttempt = 0
      clearTimeout(retryTimer)
    }
  },
)

watch([() => player.currentId, songs.data], () => {
  const id = player.currentId
  if (id !== undefined && songs.data.value !== undefined && !index.value.has(id)) {
    player.forget(id)
  }
})

watch(
  () => player.restartToken,
  () => {
    const el = audio.value
    if (el !== null && song.value !== undefined) {
      el.currentTime = song.value.playback.start
      player.position = 0
    }
  },
)

watch(
  () => player.seekRequest,
  (t) => {
    const el = audio.value
    if (t === undefined || el === null || song.value === undefined) {
      return
    }
    el.currentTime = fileTime(t, song.value.playback)
    player.seekRequest = undefined
  },
)

const volume = computed(() => player.volume * loudnessGain(song.value?.loudness))
watch(
  volume,
  (v) => {
    if (audio.value !== null) {
      audio.value.volume = Math.min(1, Math.max(0, v))
    }
  },
  { immediate: true, flush: 'post' },
)

watch(
  () => player.playing,
  (playing) => {
    if (playing) {
      void startPlayback()
    } else {
      audio.value?.pause()
    }
  },
)

async function startPlayback(): Promise<void> {
  const el = audio.value
  if (el === null || src.value === undefined) {
    return
  }
  try {
    await el.play()
  } catch (err) {
    if (err instanceof DOMException && err.name === 'NotAllowedError') {
      player.playing = false
      player.status = 'paused'
    } else if (!(err instanceof DOMException && err.name === 'AbortError')) {
      throw err
    }
  }
}

function onLoadedMetadata(): void {
  const el = audio.value
  const s = song.value
  if (el === null || s === undefined) {
    return
  }
  player.duration = s.playback.end - s.playback.start
  el.currentTime = fileTime(resumeAt, s.playback)
  el.volume = Math.min(1, Math.max(0, volume.value))
  if (player.playing) {
    void startPlayback()
  }
}

function onTimeUpdate(): void {
  const el = audio.value
  const s = song.value
  if (el === null || s === undefined) {
    return
  }
  player.position = songTime(el.currentTime, s.playback)
  resumeAt = player.position
  const trimmedEnd = s.playback.end < el.duration - 0.25
  if (trimmedEnd && el.currentTime >= s.playback.end - 0.15 && !el.paused) {
    void player.next()
  }
}

function onError(): void {
  const el = audio.value
  if (el === null || src.value === undefined) {
    return
  }
  const action = failureAction(classifyMediaError(el.error?.code), useMp3.value)
  switch (action) {
    case 'ignore':
      return
    case 'try-mp3':
      useMp3.value = true
      return
    case 'retry-later':
      player.status = 'waiting-network'
      scheduleRetry()
      return
    case 'skip':
      notify(`Cannot play ${song.value?.title ?? 'this song'}, skipping it`, 'warning')
      void player.next()
  }
}

function scheduleRetry(): void {
  clearTimeout(retryTimer)
  retryTimer = setTimeout(retryNow, retryDelay(retryAttempt++))
}

function retryNow(): void {
  clearTimeout(retryTimer)
  const el = audio.value
  if (el === null || player.status !== 'waiting-network') {
    return
  }
  if (!navigator.onLine) {
    scheduleRetry()
    return
  }
  el.load()
}

function onOnline(): void {
  if (player.status === 'waiting-network') {
    retryAttempt = 0
    retryNow()
  }
}

function onPlaying(): void {
  retryAttempt = 0
  player.status = 'playing'
  player.playing = true
}

function onPause(): void {
  const el = audio.value
  if (el !== null && !el.ended && player.status !== 'waiting-network') {
    player.status = 'paused'
    player.playing = false
  }
}

function onWaiting(): void {
  player.status = 'loading'
}

function onEnded(): void {
  void player.next()
}

onMounted(() => {
  window.addEventListener('online', onOnline)
})
onBeforeUnmount(() => {
  window.removeEventListener('online', onOnline)
  clearTimeout(retryTimer)
})

useMediaIntegration(audio, song)
</script>

<template>
  <audio
    ref="audio"
    :src="src"
    preload="auto"
    @loadedmetadata="onLoadedMetadata"
    @timeupdate="onTimeUpdate"
    @error="onError"
    @playing="onPlaying"
    @pause="onPause"
    @waiting="onWaiting"
    @ended="onEnded"
  />
</template>
