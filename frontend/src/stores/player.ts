import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'

import { call, client, isApiError } from '@/api/client'
import { needsRefill, refillCount, refillExclude, refillSeed, restartThreshold } from '@/lib/playerLogic'
import { shuffled } from '@/lib/shuffle'

const storageKey = 'sh-player-v1'
const maxHistory = 200

export type PlayerStatus = 'idle' | 'loading' | 'playing' | 'paused' | 'waiting-network'

interface Persisted {
  queue: string[]
  history: string[]
  currentId?: string
  position: number
  volume: number
}

function load(): Partial<Persisted> {
  try {
    const raw = localStorage.getItem(storageKey)
    return raw === null ? {} : (JSON.parse(raw) as Partial<Persisted>)
  } catch {
    return {}
  }
}

/** Songs that can play without the network, e.g. from the audio cache; set by the offline module. */
let offlineCandidates: () => Promise<string[]> = () => Promise.resolve([])

export function setOfflineCandidates(source: () => Promise<string[]>): void {
  offlineCandidates = source
}

/**
 * The play queue. The server decides what plays once the queue runs low;
 * this store holds what is queued and asks for more.
 */
export const usePlayer = defineStore('player', () => {
  const saved = load()

  const queue = ref<string[]>(saved.queue ?? [])
  const history = ref<string[]>(saved.history ?? [])
  const currentId = ref<string | undefined>(saved.currentId)
  /** Seconds into the trimmed song; written by the audio engine. */
  const position = ref(saved.position ?? 0)
  const duration = ref(0)
  const volume = ref(saved.volume ?? 1)

  /** Whether the user wants audio to play; the engine follows this. */
  const playing = ref(false)
  const status = ref<PlayerStatus>(currentId.value === undefined ? 'idle' : 'paused')
  const expanded = ref(false)
  /** Bumped to make the engine restart the current song from its beginning. */
  const restartToken = ref(0)
  /** A position (seconds into the song) the engine should jump to. */
  const seekRequest = ref<number | undefined>()
  /** Why the server suggested a song, for the stats-for-nerds panel. */
  const suggestionFactors = ref<Record<string, Record<string, number>>>({})
  /** Songs the server suggested, as opposed to ones the user queued. */
  const suggested = ref<Set<string>>(new Set())
  const refilling = ref(false)

  const hasSong = computed(() => currentId.value !== undefined)

  watch(
    [queue, history, currentId, volume],
    () => {
      persist()
    },
    { deep: true },
  )

  let lastPositionSave = 0
  watch(position, () => {
    const now = Date.now()
    if (now - lastPositionSave > 5000) {
      lastPositionSave = now
      persist()
    }
  })

  function persist(): void {
    const data: Persisted = {
      queue: queue.value,
      history: history.value.slice(-maxHistory),
      position: position.value,
      volume: volume.value,
      ...(currentId.value === undefined ? {} : { currentId: currentId.value }),
    }
    try {
      localStorage.setItem(storageKey, JSON.stringify(data))
    } catch {
      // Storage full or disabled: the player still works, it just won't survive a reload.
    }
  }

  function setCurrent(id: string | undefined): void {
    if (currentId.value !== undefined && currentId.value !== id) {
      history.value.push(currentId.value)
      if (history.value.length > maxHistory) {
        history.value.splice(0, history.value.length - maxHistory)
      }
    }
    if (currentId.value === id) {
      restartToken.value++
    }
    currentId.value = id
    position.value = 0
    duration.value = 0
    void refill()
  }

  /** Replaces the queue with ids and starts playing ids[start]. */
  function playSongs(ids: readonly string[], start = 0): void {
    const first = ids[start]
    if (first === undefined) {
      return
    }
    queue.value = ids.slice(start + 1)
    suggested.value = new Set()
    playing.value = true
    setCurrent(first)
  }

  function playSong(id: string): void {
    playSongs([id])
  }

  /** Plays id now and continues with server suggestions that fit it. */
  function startRadio(id: string): void {
    playSongs([id])
  }

  function playNext(id: string): void {
    queue.value = [id, ...queue.value.filter((q) => q !== id)]
    suggested.value.delete(id)
    if (currentId.value === undefined) {
      playing.value = true
      void next()
    }
  }

  function enqueue(id: string): void {
    const userQueued = queue.value.filter((q) => q !== id && !suggested.value.has(q))
    const rest = queue.value.filter((q) => q !== id && suggested.value.has(q))
    queue.value = [...userQueued, id, ...rest]
    suggested.value.delete(id)
    if (currentId.value === undefined) {
      playing.value = true
      void next()
    }
  }

  function removeFromQueue(index: number): void {
    queue.value = queue.value.filter((_, i) => i !== index)
    void refill()
  }

  function moveInQueue(from: number, to: number): void {
    const q = [...queue.value]
    const [item] = q.splice(from, 1)
    if (item !== undefined) {
      q.splice(to, 0, item)
      queue.value = q
    }
  }

  function clearQueue(): void {
    queue.value = []
    suggested.value = new Set()
    void refill()
  }

  /** Skips to the next song; when nothing is queued it waits for suggestions. */
  async function next(): Promise<void> {
    if (queue.value.length === 0) {
      await refill()
    }
    const [head, ...rest] = queue.value
    if (head === undefined) {
      playing.value = false
      status.value = currentId.value === undefined ? 'idle' : 'paused'
      return
    }
    queue.value = rest
    suggested.value.delete(head)
    setCurrent(head)
  }

  /** Restarts the song when it has played a while, else goes back one song. */
  function previous(): void {
    const prev = history.value.at(-1)
    if (position.value > restartThreshold || prev === undefined) {
      restartToken.value++
      return
    }
    history.value.pop()
    if (currentId.value !== undefined) {
      queue.value = [currentId.value, ...queue.value]
    }
    currentId.value = prev
    position.value = 0
    playing.value = true
  }

  function seekTo(seconds: number): void {
    seekRequest.value = Math.max(0, seconds)
  }

  function seekBy(seconds: number): void {
    seekTo(position.value + seconds)
  }

  function toggle(): void {
    if (currentId.value === undefined) {
      void next()
      playing.value = true
      return
    }
    playing.value = !playing.value
  }

  function stop(): void {
    playing.value = false
    if (currentId.value !== undefined) {
      history.value.push(currentId.value)
    }
    currentId.value = undefined
    position.value = 0
    status.value = 'idle'
  }

  /** Asks the server for more songs when the queue runs low. */
  async function refill(): Promise<void> {
    if (refilling.value || !needsRefill(queue.value.length)) {
      return
    }
    refilling.value = true
    try {
      const seed = refillSeed(currentId.value, queue.value)
      const exclude = refillExclude(currentId.value, queue.value, history.value)
      const res = await call(
        client.GET('/api/v1/player/next', {
          params: {
            query: { count: refillCount, exclude, ...(seed === undefined ? {} : { current: seed }) },
          },
        }),
      )
      addSuggestions(
        res.songs.map((s) => s.id),
        Object.fromEntries(res.songs.map((s) => [s.id, s.factors])),
      )
    } catch (err) {
      if (!isApiError(err, 'network')) {
        throw err
      }
      const exclude = new Set(refillExclude(currentId.value, queue.value, history.value))
      const offline = (await offlineCandidates()).filter((id) => !exclude.has(id))
      addSuggestions(shuffled(offline).slice(0, refillCount), {})
    } finally {
      refilling.value = false
    }
  }

  function addSuggestions(ids: string[], factors: Record<string, Record<string, number>>): void {
    const known = new Set([...queue.value, currentId.value])
    const fresh = ids.filter((id) => !known.has(id))
    queue.value = [...queue.value, ...fresh]
    for (const id of fresh) {
      suggested.value.add(id)
    }
    suggestionFactors.value = { ...suggestionFactors.value, ...factors }
  }

  /** Drops a song that no longer exists from everywhere. */
  function forget(id: string): void {
    queue.value = queue.value.filter((q) => q !== id)
    history.value = history.value.filter((h) => h !== id)
    if (currentId.value === id) {
      void next()
    }
  }

  return {
    queue,
    history,
    currentId,
    position,
    duration,
    volume,
    playing,
    status,
    expanded,
    restartToken,
    seekRequest,
    seekTo,
    seekBy,
    suggestionFactors,
    suggested,
    refilling,
    hasSong,
    playSongs,
    playSong,
    startRadio,
    playNext,
    enqueue,
    removeFromQueue,
    moveInQueue,
    clearQueue,
    next,
    previous,
    toggle,
    stop,
    refill,
    forget,
  }
})
