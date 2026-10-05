<script setup lang="ts">
import { computed } from 'vue'

import { useSong, useSongIndex } from '@/api/queries'
import AnalysisStats from '@/components/AnalysisStats.vue'
import { usePlayer } from '@/stores/player'

const props = defineProps<{ id: string }>()

const player = usePlayer()
const { index } = useSongIndex()
const song = useSong(() => props.id)

const factorNames = computed(() => {
  const names = new Set<string>()
  for (const id of player.queue) {
    for (const name of Object.keys(player.suggestionFactors[id] ?? {})) {
      names.add(name)
    }
  }
  return [...names]
})

const factorRows = computed(() =>
  player.queue.flatMap((id) => {
    const factors = player.suggestionFactors[id]
    return factors === undefined ? [] : [{ id, title: index.value.get(id)?.title ?? id, factors }]
  }),
)
</script>

<template>
  <section class="box nerd">
    <h2 class="title is-5">
      Stats for nerds
    </h2>
    <AnalysisStats
      v-if="song.data.value"
      :analysis="song.data.value.analysis"
      :listening="song.data.value.listening"
    />
    <p
      v-else
      class="has-text-grey"
    >
      Loading…
    </p>

    <template v-if="factorRows.length > 0">
      <h3 class="title is-6 mt-4">
        Why these songs
      </h3>
      <div class="table-container">
        <table class="table is-narrow is-fullwidth factors">
          <thead>
            <tr>
              <th>Song</th>
              <th
                v-for="name in factorNames"
                :key="name"
              >
                {{ name }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in factorRows"
              :key="row.id"
            >
              <td class="factor-title">
                {{ row.title }}
              </td>
              <td
                v-for="name in factorNames"
                :key="name"
              >
                {{ row.factors[name]?.toFixed(2) ?? '–' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </section>
</template>

<style scoped>
.factors {
  font-size: 0.8rem;
  font-variant-numeric: tabular-nums;
  background: transparent;
}

.factor-title {
  max-width: 9rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
