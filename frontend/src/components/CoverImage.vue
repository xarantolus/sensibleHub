<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { coverUrl, placeholderCover } from '@/api/media'
import type { SongSummary } from '@/api/schema'

const props = withDefaults(
  defineProps<{
    song: Pick<SongSummary, 'id' | 'lastEdit' | 'cover' | 'title'>
    size?: 'small' | 'full'
    eager?: boolean
  }>(),
  { size: 'small', eager: false },
)

const failed = ref(false)
watch(
  () => props.song.lastEdit,
  () => {
    failed.value = false
  },
)

const src = computed(() =>
  props.song.cover === undefined || failed.value ? placeholderCover : coverUrl(props.song, props.size),
)
</script>

<template>
  <figure
    class="image is-square cover-image"
    :style="{ backgroundColor: song.cover?.color }"
  >
    <img
      :src="src"
      :alt="`Cover of ${song.title}`"
      :loading="eager ? 'eager' : 'lazy'"
      decoding="async"
      @error="failed = true"
    >
  </figure>
</template>

<style scoped>
.cover-image {
  border-radius: var(--bulma-radius);
  overflow: hidden;
  background-color: var(--bulma-background);
}

.cover-image img {
  object-fit: cover;
}
</style>
