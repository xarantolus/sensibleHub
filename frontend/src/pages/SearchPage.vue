<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRouter } from 'vue-router'

import { useSearch, useSongIndex } from '@/api/queries'
import QueryView from '@/components/QueryView.vue'
import SongGrid from '@/components/SongGrid.vue'

const props = defineProps<{ query: string }>()

const router = useRouter()
const search = useSearch(() => props.query)
const { resolve } = useSongIndex()

const trimmed = computed(() => props.query.trim())

watch(
  () => search.data.value?.songIds,
  (ids) => {
    const [only, ...rest] = ids ?? []
    if (only !== undefined && rest.length === 0 && trimmed.value !== '') {
      void router.replace({ name: 'song', params: { id: only } })
    }
  },
  { immediate: true },
)
</script>

<template>
  <h1 class="page-title">
    Search
  </h1>
  <p
    v-if="trimmed === ''"
    class="notification has-text-centered"
  >
    Type something into the search box.
  </p>
  <QueryView
    v-else
    :data="search.data.value"
    :error="search.error.value"
    :is-pending="search.isPending.value"
    :refetch="search.refetch"
  >
    <template #default="{ data }">
      <p
        v-if="data.songIds.length === 0"
        class="notification has-text-centered"
      >
        No results for “{{ trimmed }}”
      </p>
      <SongGrid
        v-else
        :songs="resolve(data.songIds)"
      />
    </template>
  </QueryView>
</template>
