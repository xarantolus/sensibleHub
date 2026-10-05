/// <reference lib="webworker" />
import { CacheableResponsePlugin } from 'workbox-cacheable-response'
import { clientsClaim } from 'workbox-core'
import { ExpirationPlugin } from 'workbox-expiration'
import { createHandlerBoundToURL, precacheAndRoute } from 'workbox-precaching'
import { RangeRequestsPlugin } from 'workbox-range-requests'
import { NavigationRoute, registerRoute } from 'workbox-routing'
import { NetworkFirst, StaleWhileRevalidate } from 'workbox-strategies'

import { audioCacheName, coverCacheName } from '../offline/cacheNames'

declare const self: ServiceWorkerGlobalScope & { __WB_MANIFEST: (string | { url: string; revision: string | null })[] }

void self.skipWaiting()
clientsClaim()

precacheAndRoute(self.__WB_MANIFEST)

registerRoute(
  new NavigationRoute(createHandlerBoundToURL('/index.html'), {
    denylist: [/^\/api\//, /^\/media\//],
  }),
)

// Audio is cached as whole files by the page's look-ahead downloader and only
// used when the network fails. Chrome aborts playback when one song's range
// requests switch from the network to sliced cache responses mid-stream, so
// online playback never touches the cache; when the connection drops mid-song,
// the player reloads the song and then reads it from the cache from the start.
registerRoute(
  ({ url }) => url.pathname.startsWith('/media/songs/') && /\/(audio|mp3)$/.test(url.pathname),
  new NetworkFirst({
    cacheName: audioCacheName,
    plugins: [new CacheableResponsePlugin({ statuses: [200] }), new RangeRequestsPlugin()],
  }),
)

registerRoute(
  ({ url }) => url.pathname.startsWith('/media/songs/') && url.pathname.endsWith('/cover'),
  new StaleWhileRevalidate({
    cacheName: coverCacheName,
    plugins: [new CacheableResponsePlugin({ statuses: [200] }), new ExpirationPlugin({ maxEntries: 2000 })],
  }),
)
