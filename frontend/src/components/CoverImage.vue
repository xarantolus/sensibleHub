<script setup lang="ts">
import { computed, ref } from 'vue'

import { coverUrl, placeholderCover } from '@/api/media'
import type { SongSummary } from '@/api/schema'

const props = withDefaults(
  defineProps<{
    song: Pick<SongSummary, 'id' | 'cover' | 'title'>
    size?: 'small' | 'full'
    eager?: boolean
  }>(),
  { size: 'small', eager: false },
)

const url = computed(() => (props.song.cover === undefined ? placeholderCover : coverUrl(props.song, props.size)))
const failedUrl = ref<string>()
const src = computed(() => (failedUrl.value === url.value ? placeholderCover : url.value))
</script>

<template>
  <figure
    class="image is-square cover-image"
    :style="{ backgroundColor: song.cover?.color }"
  >
    <img
      :src="src"
      :alt="song.title"
      :loading="eager ? 'eager' : 'lazy'"
      decoding="async"
      @error="failedUrl = url"
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
