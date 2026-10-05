import { createAsyncStoragePersister } from '@tanstack/query-async-storage-persister'
import { persistQueryClient } from '@tanstack/query-persist-client-core'
import type { QueryClient } from '@tanstack/vue-query'
import { del, get, set } from 'idb-keyval'
import { Workbox } from 'workbox-window'

import { startLookahead } from './lookahead'

const persistedQueries = new Set(['songs', 'song', 'listing', 'home', 'album', 'artist'])
const maxAge = 30 * 24 * 60 * 60 * 1000

/**
 * Makes the app usable without a connection: the library survives reloads
 * (IndexedDB), the app shell and fetched media are served by the service
 * worker, and upcoming songs are downloaded ahead of time.
 */
export function startOfflineSupport(qc: QueryClient): void {
  if (typeof indexedDB !== 'undefined') {
    const [, restored] = persistQueryClient({
      queryClient: qc,
      persister: createAsyncStoragePersister({
        storage: { getItem: (k) => get<string>(k).then((v) => v ?? null), setItem: set, removeItem: del },
        key: 'sh-query-cache',
        throttleTime: 2000,
      }),
      maxAge,
      buster: 'v1',
      dehydrateOptions: {
        shouldDehydrateQuery: (q) => q.state.status === 'success' && persistedQueries.has(String(q.queryKey[0])),
      },
    })
    void restored.catch(() => undefined)
  }

  if (import.meta.env.PROD && 'serviceWorker' in navigator && window.isSecureContext) {
    void new Workbox('/sw.js', { scope: '/' }).register()
  }

  startLookahead(qc)
}
