<script setup lang="ts">
import { computed } from 'vue'

import type { Analysis, Listening } from '@/api/schema'

const props = defineProps<{ analysis: Analysis; listening?: Listening }>()

const listeningRows = computed(() => {
  const l = props.listening
  if (l === undefined) {
    return []
  }
  const when = (iso: string | undefined) => (iso === undefined ? 'never' : new Date(iso).toLocaleString())
  return [
    ['Plays (decaying)', l.plays.toFixed(2)],
    ['Skips (decaying)', l.skips.toFixed(2)],
    ['Last played', when(l.lastPlayed)],
    ['Last skipped', when(l.lastSkipped)],
  ] as const
})

function fixed(value: number | undefined, digits: number, unit = ''): string {
  return value === undefined ? '–' : `${value.toFixed(digits)}${unit}`
}

const rows = computed(() => {
  const a = props.analysis
  return [
    ['Key', a.key === undefined ? '–' : `${a.key}${a.camelot === undefined ? '' : ` (${a.camelot})`}`],
    ['Key strength', fixed(a.keyStrength, 2)],
    ['BPM', fixed(a.bpm, 1)],
    ['Beat strength', fixed(a.beatStrength, 2)],
    ['Loudness', fixed(a.loudnessLufs, 1, ' LUFS')],
    ['Loudness range', fixed(a.loudnessRange, 1, ' LU')],
    ['Energy', fixed(a.energyDb, 1, ' dB')],
    ['Onset rate', fixed(a.onsetRate, 2, ' /s')],
    ['Centroid', fixed(a.centroid, 0, ' Hz')],
    ['Flatness', fixed(a.flatness, 3)],
  ] as const
})
</script>

<template>
  <p
    v-if="analysis.status === 'pending'"
    class="has-text-grey"
  >
    Analysis pending
  </p>
  <p
    v-else-if="analysis.status === 'failed'"
    class="has-text-danger"
  >
    Analysis failed<template v-if="analysis.error">
      : {{ analysis.error }}
    </template>
  </p>
  <dl
    v-else
    class="stats"
  >
    <template
      v-for="[label, value] in rows"
      :key="label"
    >
      <dt class="has-text-grey">
        {{ label }}
      </dt>
      <dd>{{ value }}</dd>
    </template>
  </dl>
  <dl
    v-if="listeningRows.length > 0"
    class="stats mt-3"
  >
    <template
      v-for="[label, value] in listeningRows"
      :key="label"
    >
      <dt class="has-text-grey">
        {{ label }}
      </dt>
      <dd>{{ value }}</dd>
    </template>
  </dl>
</template>

<style scoped>
.stats {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.15rem 1rem;
  font-size: 0.875rem;
  font-variant-numeric: tabular-nums;
}

dd {
  text-align: right;
}
</style>
