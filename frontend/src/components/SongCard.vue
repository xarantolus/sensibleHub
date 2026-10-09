<script setup lang="ts">
import type { SongSummary } from '@/api/schema'
import { usePlayer } from '@/stores/player'

import CoverImage from './CoverImage.vue'
import SongActions from './SongActions.vue'

defineProps<{ song: SongSummary }>()
const player = usePlayer()
</script>

<template>
  <article
    class="song-card"
    :class="{ 'is-current': player.currentId === song.id }"
  >
    <div class="song-card-cover">
      <RouterLink
        :to="{ name: 'song', params: { id: song.id } }"
        class="song-card-link"
        tabindex="-1"
        aria-hidden="true"
      >
        <CoverImage :song="song" />
      </RouterLink>
      <div class="song-card-play">
        <SongActions
          variant="icon"
          :ids="[song.id]"
          :radio-id="song.id"
          :label="`Play ${song.title}`"
        />
      </div>
    </div>
    <RouterLink
      :to="{ name: 'song', params: { id: song.id } }"
      class="song-card-link"
    >
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
  </article>
</template>

<style scoped>
.song-card {
  min-width: 0;
}

.song-card-cover {
  position: relative;
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
  right: 0.4rem;
  bottom: 0.4rem;
  opacity: 0;
  transition: opacity 0.15s;
}

@media (max-width: 480px) {
  .song-card-play :deep(.song-actions-icon) {
    width: 2.1rem;
    height: 2.1rem;
    font-size: 0.85rem;
  }

  .song-card-title {
    font-size: 0.875rem;
  }

  .song-card-artist {
    font-size: 0.75rem;
  }
}

.song-card:hover .song-card-play,
.song-card-play:focus-within {
  opacity: 1;
}

.song-card-play :deep(.button:not(.is-primary)) {
  background: var(--bulma-scheme-main);
  box-shadow: var(--bulma-shadow);
}

@media (hover: none) {
  .song-card-play {
    opacity: 0.9;
  }
}
</style>
