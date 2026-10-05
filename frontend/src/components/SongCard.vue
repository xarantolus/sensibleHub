<script setup lang="ts">
import type { SongSummary } from '@/api/schema'
import { usePlayer } from '@/stores/player'

import CoverImage from './CoverImage.vue'

const props = defineProps<{ song: SongSummary; queue?: readonly string[] }>()
const player = usePlayer()

function play(): void {
  const ids = props.queue ?? [props.song.id]
  player.playSongs(ids, Math.max(0, ids.indexOf(props.song.id)))
}
</script>

<template>
  <article
    class="song-card"
    :class="{ 'is-current': player.currentId === song.id }"
  >
    <RouterLink
      :to="{ name: 'song', params: { id: song.id } }"
      class="song-card-link"
    >
      <CoverImage :song="song" />
      <div class="song-card-text">
        <p
          class="song-card-title"
          :title="song.title"
        >
          {{ song.title }}
        </p>
        <p
          v-if="song.artist"
          class="song-card-artist"
          :title="song.artist"
        >
          {{ song.artist }}
        </p>
      </div>
    </RouterLink>
    <button
      type="button"
      class="song-card-play button is-primary is-rounded"
      :aria-label="`Play ${song.title}`"
      @click="play"
    >
      ▶
    </button>
  </article>
</template>

<style scoped>
.song-card {
  position: relative;
  min-width: 0;
}

.song-card-link {
  display: block;
  color: inherit;
}

.song-card-text {
  padding-top: 0.4rem;
  min-width: 0;
}

.song-card-title,
.song-card-artist {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.song-card-title {
  font-weight: 600;
}

.song-card-artist {
  font-size: 0.875rem;
  color: var(--bulma-text-weak);
}

.song-card.is-current .song-card-title {
  color: var(--bulma-primary);
}

.song-card-play {
  position: absolute;
  right: 0.5rem;
  top: calc(100% - 4.5rem);
  width: 2.75rem;
  height: 2.75rem;
  padding: 0;
  opacity: 0;
  transition: opacity 0.15s;
}

.song-card:hover .song-card-play,
.song-card-play:focus-visible {
  opacity: 1;
}

@media (hover: none) {
  .song-card-play {
    opacity: 0.9;
  }
}
</style>
