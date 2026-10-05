import type { QueryClient } from '@tanstack/vue-query'
import { readonly, ref } from 'vue'

import { invalidateDerived, keys, removeSong, upsertSong } from './queries'
import type { DownloadStatus, paths } from './schema'

type ServerEvent = paths['/api/v1/events']['get']['responses'][200]['content']['text/event-stream'][number]
type EventName = ServerEvent['event']
type EventData<N extends EventName> = Extract<ServerEvent, { event: N }>['data']

export type ConnectionState = 'connecting' | 'open' | 'reconnecting'

const state = ref<ConnectionState>('connecting')
export const connectionState = readonly(state)

type Listener<N extends EventName> = (data: EventData<N>) => void
const extraListeners = new Map<EventName, Set<(data: unknown) => void>>()

/** Subscribe to one server event type in addition to the cache updates done here. */
export function onServerEvent<N extends EventName>(name: N, listener: Listener<N>): () => void {
  let set = extraListeners.get(name)
  if (set === undefined) {
    set = new Set()
    extraListeners.set(name, set)
  }
  const untyped = listener as (data: unknown) => void
  set.add(untyped)
  return () => set.delete(untyped)
}

const maxRetryDelay = 30_000

/**
 * Keeps the query cache in sync with the server. After any interruption all
 * queries are refetched, because events sent while disconnected are lost.
 */
export function startServerEvents(qc: QueryClient): () => void {
  let source: EventSource | undefined
  let retryTimer: ReturnType<typeof setTimeout> | undefined
  let retryDelay = 1000
  let interrupted = false
  let stopped = false

  const handle = <N extends EventName>(es: EventSource, name: N, apply: Listener<N>) => {
    es.addEventListener(name, (ev: MessageEvent<string>) => {
      const data = JSON.parse(ev.data) as EventData<N>
      apply(data)
      extraListeners.get(name)?.forEach((l) => {
        l(data)
      })
    })
  }

  const setDownloads = (patch: Partial<DownloadStatus>) => {
    qc.setQueryData<DownloadStatus>(keys.downloads, (old) => ({ running: false, queued: 0, ...old, ...patch }))
  }

  const connect = () => {
    if (stopped) {
      return
    }
    const es = new EventSource('/api/v1/events')
    source = es

    es.onopen = () => {
      state.value = 'open'
      retryDelay = 1000
      if (interrupted) {
        interrupted = false
        void qc.invalidateQueries()
      }
    }

    es.onerror = () => {
      state.value = 'reconnecting'
      interrupted = true
      if (es.readyState === EventSource.CLOSED) {
        scheduleReconnect()
      }
    }

    handle(es, 'songAdded', ({ song }) => {
      upsertSong(qc, song)
    })
    handle(es, 'songUpdated', ({ song }) => {
      upsertSong(qc, song)
    })
    handle(es, 'songDeleted', ({ id }) => {
      removeSong(qc, id)
      invalidateDerived(qc)
    })
    handle(es, 'downloadStarted', () => {
      setDownloads({ running: true })
      void qc.invalidateQueries({ queryKey: keys.downloads })
    })
    handle(es, 'downloadFinished', ({ error }) => {
      setDownloads(error === undefined ? { running: false } : { running: false, lastError: error })
      void qc.invalidateQueries({ queryKey: keys.downloads })
    })
  }

  const scheduleReconnect = () => {
    source?.close()
    clearTimeout(retryTimer)
    retryTimer = setTimeout(connect, retryDelay)
    retryDelay = Math.min(retryDelay * 2, maxRetryDelay)
  }

  const onOnline = () => {
    if (source?.readyState !== EventSource.OPEN) {
      retryDelay = 1000
      scheduleReconnect()
    }
  }
  window.addEventListener('online', onOnline)

  connect()

  return () => {
    stopped = true
    clearTimeout(retryTimer)
    window.removeEventListener('online', onOnline)
    source?.close()
  }
}
