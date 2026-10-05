<script setup lang="ts">
import type { SongDetail, SongSummary } from '@/api/schema'

import SongCover from './SongCover.vue'
import SongFields from './SongFields.vue'

defineProps<{ song: SongSummary; detail?: SongDetail }>()
</script>

<template>
  <div class="columns song-editor">
    <div class="column is-5-tablet is-4-desktop">
      <SongCover
        :song="song"
        :loading="detail === undefined"
      />
    </div>
    <div class="column">
      <SongFields
        v-bind="detail === undefined ? { song } : { song, detail }"
      />
    </div>
  </div>
</template>

<style scoped>
.song-editor > .column {
  min-width: 0;
}
</style>
