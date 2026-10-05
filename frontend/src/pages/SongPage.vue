<script setup lang="ts">
import { computed } from 'vue'

import { useSong, useSongIndex } from '@/api/queries'
import AnalysisStats from '@/components/AnalysisStats.vue'
import QueryView from '@/components/QueryView.vue'
import SongEditPanel from '@/components/SongEditPanel.vue'
import SongGrid from '@/components/SongGrid.vue'
import SongHero from '@/components/SongHero.vue'
import { useSettings } from '@/composables/useSettings'

const props = defineProps<{ id: string }>()

const settings = useSettings()
const query = useSong(() => props.id)
const { index, resolve } = useSongIndex()

const summary = computed(() => index.value.get(props.id))
</script>

<template>
  <section class="section">
    <div class="container">
      <QueryView
        :data="query.data.value"
        :error="query.error.value"
        :is-pending="query.isPending.value"
        :refetch="query.refetch"
      >
        <template #loading>
          <SongHero
            v-if="summary"
            :song="summary"
          />
          <div
            v-else
            class="is-flex is-justify-content-center py-6"
          >
            <span
              class="loader"
              aria-label="Loading"
            />
          </div>
        </template>
        <template #default="{ data }">
          <SongHero
            :song="data"
            :source-url="data.sourceUrl"
            :imported="data.imported"
          />
          <section
            v-if="settings.statsForNerds"
            class="box mt-5"
          >
            <h2 class="title is-5">
              Stats for nerds
            </h2>
            <AnalysisStats :analysis="data.analysis" />
          </section>
          <SongEditPanel
            :song="data"
            class="mt-5"
          />
          <template v-if="resolve(data.related).length > 0">
            <h2 class="title is-4 mt-6">
              Similar songs
            </h2>
            <SongGrid :songs="resolve(data.related)" />
          </template>
        </template>
      </QueryView>
    </div>
  </section>
</template>
