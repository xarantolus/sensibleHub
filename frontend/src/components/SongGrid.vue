<script setup lang="ts">
import type { SongSummary } from '@/api/schema'

import SongCard from './SongCard.vue'

defineProps<{ songs: readonly SongSummary[] }>()
</script>

<template>
  <div class="song-grid">
    <SongCard
      v-for="song in songs"
      :key="song.id"
      :song="song"
    />
  </div>
</template>

<style scoped>
.song-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(9.5rem, 1fr));
  gap: 1.25rem 1rem;
}

.song-grid > * {
  content-visibility: auto;
  contain-intrinsic-size: auto 13rem;
}

@media (max-width: 480px) {
  .song-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1rem 0.75rem;
  }
}
</style>
