<script setup lang="ts">
import { computed } from 'vue'

import { useAlbum, useSetAlbumCover, useSongIndex } from '@/api/queries'
import type { SongSummary } from '@/api/schema'
import CoverImage from '@/components/CoverImage.vue'
import CoverUpload from '@/components/CoverUpload.vue'
import QueryView from '@/components/QueryView.vue'
import SongActions from '@/components/SongActions.vue'
import { useOnline } from '@/composables/useOnline'
import { formatDuration } from '@/lib/format'
import { notify } from '@/lib/notify'
import { shuffled } from '@/lib/shuffle'
import { usePlayer } from '@/stores/player'

const props = defineProps<{ artist: string; album: string }>()

const query = useAlbum(() => props.artist, () => props.album)
const { resolve } = useSongIndex()
const setCover = useSetAlbumCover()
const player = usePlayer()
const online = useOnline()

const songs = computed(() => resolve(query.data.value?.songIds))
const total = computed(() => songs.value.reduce((sum, s) => sum + s.playback.end - s.playback.start, 0))

function length(song: SongSummary): string {
  return formatDuration(song.playback.end - song.playback.start)
}

const ids = computed(() => songs.value.map((s) => s.id))
const shuffledIds = computed(() => shuffled(ids.value))

function radio(): void {
  const [id] = shuffled(songs.value.map((s) => s.id))
  if (id !== undefined) {
    player.startRadio(id)
  }
}

function changeCover(file: File): void {
  setCover.mutate(
    { artist: props.artist, album: props.album, file },
    {
      onSuccess: () => {
        notify('Cover set for all songs', 'success')
      },
    },
  )
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
          <div class="columns is-vcentered">
            <div class="column is-4-tablet is-3-desktop">
              <CoverImage
                v-if="songs[0]"
                :song="songs[0]"
                size="full"
                eager
              />
            </div>
            <div class="column">
              <h1 class="title is-2">
                {{ data.title }}
              </h1>
              <p class="subtitle is-5">
                <RouterLink :to="{ name: 'artist', params: { artist: data.artist } }">
                  {{ data.artist }}
                </RouterLink>
              </p>
              <p class="has-text-grey mb-4">
                {{ songs.length }} {{ songs.length === 1 ? 'song' : 'songs' }} · {{ formatDuration(total) }}
              </p>
              <div class="buttons">
                <SongActions
                  :ids="ids"
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
                  :disabled="songs.length === 0"
                  @click="radio"
                >
                  Radio
                </button>
              </div>
            </div>
          </div>

          <ol class="track-list mt-5">
            <li
              v-for="(song, i) in songs"
              :key="song.id"
              class="track"
            >
              <span class="track-no has-text-grey">{{ i + 1 }}</span>
              <RouterLink
                :to="{ name: 'song', params: { id: song.id } }"
                class="track-title"
                :title="song.title"
              >
                {{ song.title }}
              </RouterLink>
              <span class="track-length has-text-grey">{{ length(song) }}</span>
              <div class="track-actions">
                <SongActions
                  variant="small"
                  :ids="[song.id]"
                  :radio-id="song.id"
                  :label="`Play ${song.title}`"
                />
              </div>
            </li>
          </ol>

          <o-collapse
            :open="false"
            label="Set cover for all songs"
            class="box mt-5"
          >
            <div class="pt-4">
              <CoverUpload
                :busy="setCover.isPending.value"
                :disabled="!online"
                disabled-reason="You are offline"
                label="Choose cover"
                @select="changeCover"
              />
            </div>
          </o-collapse>
        </template>
      </QueryView>
    </div>
  </section>
</template>

<style scoped>
.track-list {
  list-style: none;
  margin: 0;
}

.track {
  display: grid;
  grid-template-columns: 2rem minmax(0, 1fr) auto auto;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--bulma-border-weak);
}

.track-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.track-no,
.track-length {
  font-variant-numeric: tabular-nums;
}

.track-actions {
  margin-bottom: 0;
  flex-wrap: nowrap;
}
</style>
