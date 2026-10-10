import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'

import { call, client, isApiError } from '@/api/client'
import { isSkip, needsRefill, refillCount, refillExclude, refillSeed, restartThreshold } from '@/lib/playerLogic'
import { reportPlay } from '@/lib/playReports'
import { shuffled } from '@/lib/shuffle'

const storageKey = 'sh-player-v2'
const maxHistory = 200

export type PlayerStatus = 'idle' | 'loading' | 'playing' | 'paused' | 'waiting-network'

/** The two lists of upcoming songs: what the listener queued, and what the server suggests after it. */
export type UpNextList = 'queue' | 'autoplay'

interface Persisted {
  queue: string[]
  autoplay: string[]
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

type Ids = string | readonly string[]

function asList(ids: Ids): string[] {
  return typeof ids === 'string' ? [ids] : [...new Set(ids)]
}

/**
 * What plays next: first the listener's queue, then autoplay, which the server
 * keeps filled so playback never ends.
 */
export const usePlayer = defineStore('player', () => {
  const saved = load()

  const queue = ref<string[]>(saved.queue ?? [])
  const autoplay = ref<string[]>(saved.autoplay ?? [])
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
  const refilling = ref(false)

  const hasSong = computed(() => currentId.value !== undefined)
  /** Everything that will play after the current song, in order. */
  const upcoming = computed(() => [...queue.value, ...autoplay.value])

  watch([queue, autoplay, history, currentId, volume], persist, { deep: true })

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
      autoplay: autoplay.value,
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

  /**
   * How the current song was left: it 'ended', the listener chose to move on
   * (a skip if that happened early), or something else replaced it ('neutral',
   * e.g. starting an album), which only counts if it was listened to properly.
   */
  type Leave = 'ended' | 'moved-on' | 'neutral'

  function reportLeaving(how: Leave): void {
    const id = currentId.value
    if (id === undefined) {
      return
    }
    const listened = position.value
    const early = isSkip(listened, duration.value)
    if (how === 'neutral' && early) {
      return
    }
    reportPlay({ songId: id, at: new Date().toISOString(), listened, skipped: how === 'moved-on' && early })
  }

  function setCurrent(id: string | undefined, how: Leave = 'neutral'): void {
    if (currentId.value !== id) {
      reportLeaving(how)
    }
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

  function without(list: readonly string[], ids: readonly string[]): string[] {
    const drop = new Set(ids)
    return list.filter((id) => !drop.has(id))
  }

  function removeEverywhere(ids: readonly string[]): void {
    queue.value = without(queue.value, ids)
    autoplay.value = without(autoplay.value, ids)
  }

  /**
   * Plays the first song now and queues the rest right after it, ahead of
   * anything already queued. Autoplay is rebuilt to follow the new songs.
   */
  function playNow(ids: Ids): void {
    const list = asList(ids)
    const [first, ...rest] = list
    if (first === undefined) {
      return
    }
    removeEverywhere(list)
    queue.value = [...rest, ...queue.value]
    autoplay.value = []
    playing.value = true
    setCurrent(first, 'neutral')
  }

  /** Puts songs at the front of the queue. Starts playing if nothing is loaded. */
  function playNext(ids: Ids): void {
    const list = asList(ids)
    removeEverywhere(list)
    queue.value = [...list, ...queue.value]
    startIfIdle()
  }

  /** Puts songs at the end of the queue, which is still before autoplay. */
  function enqueue(ids: Ids): void {
    const list = asList(ids)
    removeEverywhere(list)
    queue.value = [...queue.value, ...list]
    startIfIdle()
  }

  function startIfIdle(): void {
    if (currentId.value === undefined) {
      playing.value = true
      void next()
    }
  }

  function playSong(id: string): void {
    playNow(id)
  }

  /** Plays id now and fills autoplay with songs that fit it. */
  function startRadio(id: string): void {
    removeEverywhere([id])
    autoplay.value = []
    playing.value = true
    setCurrent(id, 'neutral')
  }

  function listRef(list: UpNextList) {
    return list === 'queue' ? queue : autoplay
  }

  function removeAt(list: UpNextList, index: number): void {
    const l = listRef(list)
    l.value = l.value.filter((_, i) => i !== index)
    void refill()
  }

  /** Moves a song within or between the two lists. */
  function move(from: UpNextList, fromIndex: number, to: UpNextList, toIndex: number): void {
    const source = [...listRef(from).value]
    const [id] = source.splice(fromIndex, 1)
    if (id === undefined) {
      return
    }
    listRef(from).value = source
    const target = [...listRef(to).value]
    target.splice(Math.min(Math.max(toIndex, 0), target.length), 0, id)
    listRef(to).value = target
    void refill()
  }

  /** Plays a song from either list now, keeping everything around it. */
  function playFrom(list: UpNextList, index: number): void {
    const id = listRef(list).value[index]
    if (id === undefined) {
      return
    }
    removeAt(list, index)
    playing.value = true
    setCurrent(id, 'moved-on')
  }

  function clear(list: UpNextList): void {
    listRef(list).value = []
    void refill()
  }

  /**
   * Moves to the next song; when nothing is queued it waits for suggestions.
   * `ended` is for the audio engine when a song finished and `failed` when it
   * could not be played; the default is the listener moving on, which counts as
   * a skip when it happens early.
   */
  async function next(reason: 'user' | 'ended' | 'failed' = 'user'): Promise<void> {
    if (upcoming.value.length === 0) {
      await refill()
    }
    const fromQueue = queue.value.length > 0
    const head = fromQueue ? queue.value[0] : autoplay.value[0]
    if (head === undefined) {
      if (reason === 'ended') {
        reportLeaving('ended')
      }
      playing.value = false
      status.value = currentId.value === undefined ? 'idle' : 'paused'
      return
    }
    if (fromQueue) {
      queue.value = queue.value.slice(1)
    } else {
      autoplay.value = autoplay.value.slice(1)
    }
    setCurrent(head, reason === 'ended' ? 'ended' : reason === 'failed' ? 'neutral' : 'moved-on')
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
      queue.value = [currentId.value, ...without(queue.value, [currentId.value])]
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
      startIfIdle()
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

  /** Ends the session: nothing plays and up next is emptied, which hides the player. */
  function close(): void {
    reportLeaving('neutral')
    queue.value = []
    autoplay.value = []
    expanded.value = false
    stop()
  }

  /** Asks the server for more autoplay songs when it runs low. */
  async function refill(): Promise<void> {
    if (refilling.value || !needsRefill(autoplay.value.length)) {
      return
    }
    refilling.value = true
    try {
      const seed = refillSeed(currentId.value, upcoming.value)
      const exclude = refillExclude(currentId.value, upcoming.value, history.value)
      const res = await call(
        client.GET('/api/v1/player/next', {
          params: {
            query: { count: refillCount(autoplay.value.length), exclude, ...(seed === undefined ? {} : { current: seed }) },
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
      const exclude = new Set(refillExclude(currentId.value, upcoming.value, history.value))
      const offline = (await offlineCandidates()).filter((id) => !exclude.has(id))
      addSuggestions(shuffled(offline).slice(0, refillCount(autoplay.value.length)), {})
    } finally {
      refilling.value = false
    }
  }

  function addSuggestions(ids: string[], factors: Record<string, Record<string, number>>): void {
    const known = new Set([...upcoming.value, currentId.value])
    autoplay.value = [...autoplay.value, ...ids.filter((id) => !known.has(id))]
    suggestionFactors.value = { ...suggestionFactors.value, ...factors }
  }

  /** Drops a song that no longer exists from everywhere. */
  function forget(id: string): void {
    removeEverywhere([id])
    history.value = history.value.filter((h) => h !== id)
    if (currentId.value === id) {
      void next('failed')
    }
  }

  return {
    queue,
    autoplay,
    upcoming,
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
    refilling,
    hasSong,
    playNow,
    playSong,
    playNext,
    enqueue,
    startRadio,
    removeAt,
    move,
    playFrom,
    clear,
    next,
    previous,
    toggle,
    stop,
    close,
    refill,
    forget,
  }
})
