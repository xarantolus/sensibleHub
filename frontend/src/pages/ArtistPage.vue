<script setup lang="ts">
import { computed } from 'vue'

import { useArtist, useSongIndex } from '@/api/queries'
import QueryView from '@/components/QueryView.vue'
import SongGrid from '@/components/SongGrid.vue'
import { formatDuration } from '@/lib/format'
import { shuffled } from '@/lib/shuffle'
import { usePlayer } from '@/stores/player'

const props = defineProps<{ artist: string }>()

const query = useArtist(() => props.artist)
const { resolve } = useSongIndex()
const player = usePlayer()

const allIds = computed(() => query.data.value?.albums.flatMap((a) => a.songIds) ?? [])

const years = computed(() => {
  const { yearStart, yearEnd } = query.data.value ?? {}
  if (yearStart === undefined && yearEnd === undefined) {
    return undefined
  }
  const from = yearStart ?? yearEnd
  const to = yearEnd ?? yearStart
  return from === to ? String(from) : `${String(from)}–${String(to)}`
})

function playAll(): void {
  player.playSongs(allIds.value)
}

function shuffle(): void {
  player.playSongs(shuffled(allIds.value))
}
</script>

<template>
  <section class="section">
    <div class="container">
      <QueryView
        :data="query.data.value"
        :error="query.error.value"
        :is-pending="query.isPending.value"
        :refetch="query.refetch"
      >
        <template #default="{ data }">
          <h1 class="title is-2">
            {{ data.name }}
          </h1>
          <p class="subtitle is-6 has-text-grey">
            {{ formatDuration(data.playTime) }}<template v-if="years">
              · {{ years }}
            </template>
          </p>
          <div class="buttons">
            <button
              type="button"
              class="button is-primary"
              :disabled="allIds.length === 0"
              @click="playAll"
            >
              Play all
            </button>
            <button
              type="button"
              class="button"
              :disabled="allIds.length === 0"
              @click="shuffle"
            >
              Shuffle
            </button>
          </div>

          <section
            v-for="album in data.albums"
            :key="album.title"
            class="mt-6"
          >
            <h2 class="title is-4">
              <RouterLink :to="{ name: 'album', params: { artist: data.name, album: album.title } }">
                {{ album.title }}
              </RouterLink>
            </h2>
            <SongGrid :songs="resolve(album.songIds)" />
          </section>

          <section
            v-if="data.featured.length > 0"
            class="mt-6"
          >
            <h2 class="title is-4">
              Featured on
            </h2>
            <SongGrid :songs="resolve(data.featured)" />
          </section>
        </template>
      </QueryView>
    </div>
  </section>
</template>
