<script setup lang="ts">
import { ref, watch } from 'vue'

import type { SongDetail, SongSummary } from '@/api/schema'

import SongCover from './SongCover.vue'
import SongFields from './SongFields.vue'

const props = defineProps<{ song: SongSummary; detail?: SongDetail }>()

/** A new cover picked but not saved yet; the Save button uploads it. */
const coverFile = ref<File>()

watch(
  () => props.song.id,
  () => {
    coverFile.value = undefined
  },
)
</script>

<template>
  <div class="columns song-editor">
    <div class="column is-5-tablet is-4-desktop">
      <SongCover
        v-model:file="coverFile"
        :song="song"
        :loading="detail === undefined"
      />
    </div>
    <div class="column">
      <SongFields
        v-model:cover-file="coverFile"
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
