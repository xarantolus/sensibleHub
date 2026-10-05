<script setup lang="ts">
import { computed } from 'vue'

import { useHome, useSongIndex } from '@/api/queries'
import QueryView from '@/components/QueryView.vue'
import SongActions from '@/components/SongActions.vue'
import SongGrid from '@/components/SongGrid.vue'

const home = useHome()
const { resolve } = useSongIndex()

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
        <SongActions
          v-if="songs.length > 0"
          :ids="data.songIds"
          label="Play all"
        />
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
