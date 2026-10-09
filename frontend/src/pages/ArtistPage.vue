<script setup lang="ts">
import { computed } from 'vue'

import { useArtist, useSetArtistSync, useSongIndex } from '@/api/queries'
import QueryView from '@/components/QueryView.vue'
import SongActions from '@/components/SongActions.vue'
import SongGrid from '@/components/SongGrid.vue'
import { useOnline } from '@/composables/useOnline'
import { formatDuration } from '@/lib/format'
import { shuffled } from '@/lib/shuffle'
import { usePlayer } from '@/stores/player'

const props = defineProps<{ artist: string }>()

const query = useArtist(() => props.artist)
const { resolve } = useSongIndex()
const player = usePlayer()
const online = useOnline()
const setSync = useSetArtistSync()

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

const shuffledIds = computed(() => shuffled(allIds.value))

function toggleSync(sync: boolean): void {
  setSync.mutate({ artist: props.artist, sync })
}

function radio(): void {
  const [id] = shuffledIds.value
  if (id !== undefined) {
    player.startRadio(id)
  }
}
</script>

<template>
  <div>
    <div>
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
            {{ allIds.length }} {{ allIds.length === 1 ? 'song' : 'songs' }} · {{ formatDuration(data.playTime) }}<template v-if="years">
              · {{ years }}
            </template>
          </p>
          <div class="buttons">
            <SongActions
              :ids="allIds"
              label="Play all"
            />
            <SongActions
              :ids="shuffledIds"
              label="Shuffle"
              :primary="false"
            />
            <button
              type="button"
              class="button"
              :disabled="allIds.length === 0"
              @click="radio"
            >
              Radio
            </button>
            <span class="artist-sync">
              <o-switch
                :model-value="setSync.isPending.value ? setSync.variables.value?.sync : data.sync"
                :disabled="!online || setSync.isPending.value"
                :title="online ? undefined : 'You are offline'"
                @update:model-value="toggleSync"
              >
                Sync
              </o-switch>
            </span>
          </div>

          <section
            v-for="album in data.albums"
            :key="album.title"
            class="mt-6"
          >
            <h2 class="title is-4">
              <RouterLink
                v-if="album.title && data.name"
                :to="{ name: 'album', params: { artist: data.name, album: album.title } }"
              >
                {{ album.title }}
              </RouterLink>
              <template v-else>
                Other songs
              </template>
            </h2>
            <SongGrid
              :songs="resolve(album.songIds)"
              show-year
            />
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
  </div>
</template>

<style scoped>
.artist-sync {
  display: inline-flex;
  align-items: center;
  height: var(--bulma-control-height);
  margin-left: 0.25rem;
}
</style>
