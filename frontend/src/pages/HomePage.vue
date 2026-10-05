<script setup lang="ts">
import { computed } from 'vue'

import { useHome, useSongIndex } from '@/api/queries'
import QueryView from '@/components/QueryView.vue'
import SongGrid from '@/components/SongGrid.vue'
import { usePlayer } from '@/stores/player'

const home = useHome()
const { resolve } = useSongIndex()
const player = usePlayer()

const songs = computed(() => resolve(home.data.value?.songIds))
</script>

<template>
  <QueryView
    :data="home.data.value"
    :error="home.error.value"
    :is-pending="home.isPending.value"
    :refetch="home.refetch"
  >
    <template #default="{ data }">
      <div
        class="is-flex is-align-items-center is-justify-content-space-between is-flex-wrap-wrap mb-4"
        style="gap: 0.75rem"
      >
        <h1 class="page-title mb-0">
          {{ data.today ? 'Added today' : 'Newest songs' }}
        </h1>
        <button
          v-if="songs.length > 0"
          type="button"
          class="button is-primary"
          @click="player.playSongs(data.songIds)"
        >
          Play all
        </button>
      </div>
      <SongGrid
        v-if="songs.length > 0"
        :songs="songs"
      />
      <div
        v-else
        class="notification has-text-centered"
      >
        <p class="mb-3">
          There are no songs yet.
        </p>
        <RouterLink
          to="/add"
          class="button is-primary"
        >
          Add your first songs
        </RouterLink>
      </div>
    </template>
  </QueryView>
</template>
