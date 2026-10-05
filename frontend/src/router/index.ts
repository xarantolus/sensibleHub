import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import type { ListingKind } from '@/api/queries'

const listings: { path: string; kind: ListingKind; title: string }[] = [
  { path: '/songs', kind: 'title', title: 'Songs' },
  { path: '/artists', kind: 'artist', title: 'Artists' },
  { path: '/years', kind: 'year', title: 'Years' },
  { path: '/incomplete', kind: 'incomplete', title: 'Incomplete' },
  { path: '/unsynced', kind: 'unsynced', title: 'Unsynced' },
  { path: '/edits', kind: 'edited', title: 'Recently edited' },
  { path: '/added', kind: 'added', title: 'Date added' },
]

export const listingRoutes = listings

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: () => import('@/pages/HomePage.vue'), meta: { title: 'Sensible Hub' } },
  { path: '/add', name: 'add', component: () => import('@/pages/AddPage.vue'), meta: { title: 'Add songs' } },
  {
    path: '/search',
    name: 'search',
    component: () => import('@/pages/SearchPage.vue'),
    props: (route) => ({ query: typeof route.query.q === 'string' ? route.query.q : '' }),
    meta: { title: 'Search' },
  },
  {
    path: '/song/:id',
    name: 'song',
    component: () => import('@/pages/SongPage.vue'),
    props: true,
  },
  {
    path: '/album/:artist/:album',
    name: 'album',
    component: () => import('@/pages/AlbumPage.vue'),
    props: true,
  },
  {
    path: '/artist/:artist',
    name: 'artist',
    component: () => import('@/pages/ArtistPage.vue'),
    props: true,
  },
  ...listings.map(
    (l): RouteRecordRaw => ({
      path: l.path,
      name: `listing-${l.kind}`,
      component: () => import('@/pages/ListingPage.vue'),
      props: { kind: l.kind, title: l.title },
      meta: { title: l.title },
    }),
  ),
  { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/pages/NotFoundPage.vue') },
]

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
  }
}

export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: (_to, _from, saved) => saved ?? { top: 0 },
})

router.afterEach((to) => {
  if (to.meta.title !== undefined) {
    document.title = to.meta.title
  }
})
