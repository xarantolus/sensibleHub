<script setup lang="ts">
import { computed } from 'vue'

import { mp3Url } from '@/api/media'
import type { SongSummary } from '@/api/schema'
import { formatDuration } from '@/lib/format'
import { usePlayer } from '@/stores/player'

import CoverImage from './CoverImage.vue'

const props = defineProps<{ song: SongSummary; sourceUrl?: string; imported?: boolean }>()
const player = usePlayer()

const length = computed(() => formatDuration(props.song.playback.end - props.song.playback.start))
</script>

<template>
  <section class="hero-song columns is-vcentered">
    <div class="column is-5-tablet is-4-desktop">
      <CoverImage
        :song="song"
        size="full"
        eager
      />
    </div>
    <div class="column">
      <h1 class="title is-2 hero-title">
        {{ song.title }}
      </h1>
      <p class="subtitle is-5">
        <RouterLink
          v-if="song.artist"
          :to="{ name: 'artist', params: { artist: song.artist } }"
        >
          {{ song.artist }}
        </RouterLink>
        <template v-if="song.artist && song.album">
          ·
        </template>
        <RouterLink
          v-if="song.artist && song.album"
          :to="{ name: 'album', params: { artist: song.artist, album: song.album } }"
        >
          {{ song.album }}
        </RouterLink>
        <span v-else-if="song.album">{{ song.album }}</span>
      </p>
      <p class="has-text-grey mb-4">
        <span v-if="song.year">{{ song.year }} · </span>{{ length }}
      </p>
      <div class="buttons">
        <button
          type="button"
          class="button is-primary is-medium"
          @click="player.playSong(song.id)"
        >
          Play
        </button>
        <button
          type="button"
          class="button"
          @click="player.playNext(song.id)"
        >
          Play next
        </button>
        <button
          type="button"
          class="button"
          @click="player.enqueue(song.id)"
        >
          Add to queue
        </button>
      </div>
      <div class="buttons are-small">
        <a
          class="button is-light"
          :href="mp3Url(song)"
          download
        >Download MP3</a>
        <a
          v-if="sourceUrl && !imported"
          class="button is-light"
          :href="sourceUrl"
          target="_blank"
          rel="noopener noreferrer"
        >Source</a>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero-title {
  overflow-wrap: anywhere;
}
</style>
