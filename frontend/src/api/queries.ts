import { useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/vue-query'
import { computed, shallowRef, toValue, type MaybeRefOrGetter, type ShallowRef } from 'vue'

import { call, callVoid, client, formData } from './client'
import { bumpCoverRevision } from './media'
import type { paths, SongEditBody, SongSummary } from './schema'

export type ListingKind = paths['/api/v1/listings/{kind}']['get']['parameters']['path']['kind']

export const keys = {
  songs: ['songs'] as const,
  song: (id: string) => ['song', id] as const,
  listing: (kind: ListingKind) => ['listing', kind] as const,
  home: ['home'] as const,
  search: (q: string) => ['search', q] as const,
  album: (artist: string, album: string) => ['album', artist, album] as const,
  artist: (name: string) => ['artist', name] as const,
  downloads: ['downloads'] as const,
  analysis: ['analysis'] as const,
}

/** Queries derived from the song list; they are refetched whenever a song changes. */
export const derivedKeys = [['song'], ['listing'], ['home'], ['search'], ['album'], ['artist']] as const

export function invalidateDerived(qc: QueryClient): void {
  for (const queryKey of derivedKeys) {
    void qc.invalidateQueries({ queryKey })
  }
}

export function useSongs() {
  return useQuery({
    queryKey: keys.songs,
    queryFn: ({ signal }) => call(client.GET('/api/v1/songs', { signal })),
    staleTime: Infinity,
  })
}

const songIndexes = new WeakMap<QueryClient, ShallowRef<ReadonlyMap<string, SongSummary>>>()

function buildIndex(songs: readonly SongSummary[] | undefined): ReadonlyMap<string, SongSummary> {
  return new Map((songs ?? []).map((s) => [s.id, s]))
}

/**
 * One index for the whole app, built from the raw cached list whenever it
 * changes. Building it per component from the reactive query data wrapped
 * every song in a proxy on each live update.
 */
function sharedSongIndex(qc: QueryClient): ShallowRef<ReadonlyMap<string, SongSummary>> {
  let index = songIndexes.get(qc)
  if (index === undefined) {
    const ref = shallowRef(buildIndex(qc.getQueryData<readonly SongSummary[]>(keys.songs)))
    qc.getQueryCache().subscribe((event) => {
      if ((event.type === 'updated' || event.type === 'added') && event.query.queryKey[0] === keys.songs[0]) {
        ref.value = buildIndex(event.query.state.data as readonly SongSummary[] | undefined)
      }
    })
    songIndexes.set(qc, ref)
    index = ref
  }
  return index
}

/** All songs by ID. Views receive IDs from the API and resolve them here. */
export function useSongIndex() {
  const songs = useSongs()
  const index = sharedSongIndex(useQueryClient())
  return {
    songs,
    index,
    resolve: (ids: readonly string[] | undefined): SongSummary[] => {
      const map = index.value
      return (ids ?? []).flatMap((id) => map.get(id) ?? [])
    },
  }
}

export function useSong(id: MaybeRefOrGetter<string>) {
  return useQuery({
    queryKey: computed(() => keys.song(toValue(id))),
    queryFn: ({ signal }) =>
      call(client.GET('/api/v1/songs/{id}', { params: { path: { id: toValue(id) } }, signal })),
  })
}

export function useListing(kind: MaybeRefOrGetter<ListingKind>) {
  return useQuery({
    queryKey: computed(() => keys.listing(toValue(kind))),
    queryFn: ({ signal }) =>
      call(client.GET('/api/v1/listings/{kind}', { params: { path: { kind: toValue(kind) } }, signal })),
  })
}

export function useHome() {
  return useQuery({
    queryKey: keys.home,
    queryFn: ({ signal }) => call(client.GET('/api/v1/home', { signal })),
  })
}

export function useSearch(query: MaybeRefOrGetter<string>, limit = 0) {
  return useQuery({
    queryKey: computed(() => keys.search(toValue(query).trim())),
    queryFn: ({ signal }) =>
      call(client.GET('/api/v1/search', { params: { query: { q: toValue(query).trim(), limit } }, signal })),
    enabled: computed(() => toValue(query).trim() !== ''),
    placeholderData: (prev) => prev,
  })
}

export function useAlbum(artist: MaybeRefOrGetter<string>, album: MaybeRefOrGetter<string>) {
  return useQuery({
    queryKey: computed(() => keys.album(toValue(artist), toValue(album))),
    queryFn: ({ signal }) =>
      call(
        client.GET('/api/v1/albums/{artist}/{album}', {
          params: { path: { artist: toValue(artist), album: toValue(album) } },
          signal,
        }),
      ),
  })
}

export function useArtist(name: MaybeRefOrGetter<string>) {
  return useQuery({
    queryKey: computed(() => keys.artist(toValue(name))),
    queryFn: ({ signal }) =>
      call(client.GET('/api/v1/artists/{artist}', { params: { path: { artist: toValue(name) } }, signal })),
  })
}

export function useDownloads() {
  return useQuery({
    queryKey: keys.downloads,
    queryFn: ({ signal }) => call(client.GET('/api/v1/downloads', { signal })),
  })
}

export function useAnalysisStatus() {
  return useQuery({
    queryKey: keys.analysis,
    queryFn: ({ signal }) => call(client.GET('/api/v1/analysis', { signal })),
  })
}

export async function fetchRandomSong(): Promise<SongSummary> {
  return call(client.GET('/api/v1/songs/random'))
}

function replaceSong(qc: QueryClient, song: SongSummary): void {
  qc.setQueryData<readonly SongSummary[]>(keys.songs, (old) => {
    if (old === undefined) {
      return old
    }
    return old.some((s) => s.id === song.id) ? old.map((s) => (s.id === song.id ? song : s)) : [...old, song]
  })
}

export function removeSong(qc: QueryClient, id: string): void {
  qc.setQueryData<readonly SongSummary[]>(keys.songs, (old) => old?.filter((s) => s.id !== id))
  qc.removeQueries({ queryKey: keys.song(id) })
}

/** Fields that listings, search and grouping depend on. */
function groupingKey(s: SongSummary): string {
  return JSON.stringify([s.title, s.artist, s.album, s.year, s.sync, s.added, s.cover?.size])
}

/**
 * Stores a changed song. Derived queries are only refetched when something
 * they depend on changed; a finished analysis, for example, only refreshes
 * that song's details.
 */
export function upsertSong(qc: QueryClient, song: SongSummary): void {
  const old = qc.getQueryData<readonly SongSummary[]>(keys.songs)?.find((s) => s.id === song.id)
  if (old !== undefined && old.lastEdit !== song.lastEdit) {
    bumpCoverRevision(song.id)
  }
  replaceSong(qc, song)
  if (old !== undefined && groupingKey(old) === groupingKey(song)) {
    void qc.invalidateQueries({ queryKey: keys.song(song.id) })
    return
  }
  invalidateDerived(qc)
}

export function useUpdateSong() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, edit }: { id: string; edit: SongEditBody }) =>
      call(client.PUT('/api/v1/songs/{id}', { params: { path: { id } }, body: edit })),
    onMutate: async ({ id, edit }) => {
      await qc.cancelQueries({ queryKey: keys.songs })
      const previous = qc.getQueryData<readonly SongSummary[]>(keys.songs)
      const current = previous?.find((s) => s.id === id)
      if (current !== undefined) {
        const { year, start, end, ...rest } = edit
        const optimistic: SongSummary = { ...current, ...rest, playback: { start, end } }
        replaceSong(qc, year === undefined ? withoutYear(optimistic) : { ...optimistic, year })
      }
      return { previous }
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous !== undefined) {
        qc.setQueryData(keys.songs, ctx.previous)
      }
    },
    onSuccess: (song) => {
      upsertSong(qc, song)
    },
  })
}

function withoutYear(song: SongSummary): SongSummary {
  const copy = { ...song }
  delete copy.year
  return copy
}

export function useDeleteSong() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => callVoid(client.DELETE('/api/v1/songs/{id}', { params: { path: { id } } })),
    onSuccess: (_data, id) => {
      removeSong(qc, id)
      invalidateDerived(qc)
    },
  })
}

export function useSetCover() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) =>
      call(
        client.PUT('/api/v1/songs/{id}/cover', {
          params: { path: { id } },
          body: { cover: '' },
          bodySerializer: () => formData({ cover: file }),
        }),
      ),
    onSuccess: (song) => {
      upsertSong(qc, song)
    },
  })
}

export function useDeleteCover() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => call(client.DELETE('/api/v1/songs/{id}/cover', { params: { path: { id } } })),
    onSuccess: (song) => {
      upsertSong(qc, song)
    },
  })
}

export function useSetAlbumCover() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ artist, album, file }: { artist: string; album: string; file: File }) =>
      callVoid(
        client.PUT('/api/v1/albums/{artist}/{album}/cover', {
          params: { path: { artist, album } },
          body: { cover: '' },
          bodySerializer: () => formData({ cover: file }),
        }),
      ),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: keys.songs })
      invalidateDerived(qc)
    },
  })
}

export function useEnqueueDownload() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (query: string) => call(client.POST('/api/v1/downloads', { body: { query } })),
    onSuccess: (status) => {
      qc.setQueryData(keys.downloads, status)
    },
  })
}

export function useAbortDownload() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => callVoid(client.DELETE('/api/v1/downloads/current')),
    onSettled: () => {
      void qc.invalidateQueries({ queryKey: keys.downloads })
    },
  })
}
