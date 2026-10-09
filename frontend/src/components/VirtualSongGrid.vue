<script setup lang="ts">
import { useWindowVirtualizer } from '@tanstack/vue-virtual'
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef, type ComponentPublicInstance } from 'vue'

import type { SongSummary } from '@/api/schema'
import { chunk } from '@/lib/grid'

import SongCard from './SongCard.vue'

export interface GridGroup {
  key: string
  title?: string
  link?: string
  description?: string
  songs: readonly SongSummary[]
}

const props = defineProps<{ groups: readonly GridGroup[] }>()

const minCardWidth = 136
const gap = 16
const mobileBreakpoint = 480
const mobileColumns = 3

/**
 * Renders only the rows near the viewport: a library can have thousands of
 * songs, and building a card for each is what made big listings slow.
 */
const root = useTemplateRef<HTMLElement>('root')
const width = ref(0)
let observer: ResizeObserver | undefined

onMounted(() => {
  const el = root.value
  if (el === null) {
    return
  }
  width.value = el.clientWidth
  observer = new ResizeObserver(([entry]) => {
    if (entry !== undefined) {
      width.value = entry.contentRect.width
    }
  })
  observer.observe(el)
})
onBeforeUnmount(() => observer?.disconnect())

const columns = computed(() => {
  if (width.value <= mobileBreakpoint) {
    return mobileColumns
  }
  return Math.max(2, Math.floor((width.value + gap) / (minCardWidth + gap)))
})

/** A row is either a group header or one line of cards. */
interface Row {
  key: string
  header?: GridGroup
  songs: readonly SongSummary[]
}

const rows = computed<Row[]>(() =>
  props.groups.flatMap((g): Row[] => [
    ...(g.title === undefined ? [] : [{ key: `h-${g.key}`, header: g, songs: [] }]),
    ...chunk(g.songs, columns.value).map((songs, i) => ({ key: `${g.key}-${String(i)}`, songs })),
  ]),
)

const stickyOffset = 120

const cardRowHeight = computed(() => {
  const cardWidth = (width.value - gap * (columns.value - 1)) / Math.max(columns.value, 1)
  return cardWidth + 64
})

const scrollMargin = ref(0)
const virtualizer = useWindowVirtualizer(
  computed(() => ({
    count: rows.value.length,
    estimateSize: (i: number) => (rows.value[i]?.header === undefined ? cardRowHeight.value : 64),
    overscan: 3,
    scrollMargin: scrollMargin.value,
    scrollPaddingStart: stickyOffset,
    getItemKey: (i: number) => rows.value[i]?.key ?? i,
  })),
)

onMounted(() => {
  scrollMargin.value = root.value?.offsetTop ?? 0
})

const items = computed(() => virtualizer.value.getVirtualItems())

function measure(el: Element | ComponentPublicInstance | null): void {
  if (el instanceof HTMLElement) {
    virtualizer.value.measureElement(el)
  }
}

/** Scrolls so that the group's header sits right below the sticky bars. */
function scrollToGroup(key: string): void {
  const index = rows.value.findIndex((r) => r.header?.key === key)
  if (index !== -1) {
    virtualizer.value.scrollToIndex(index, { align: 'start' })
  }
}

defineExpose({ scrollToGroup })
</script>

<template>
  <div
    ref="root"
    class="virtual-grid"
    :style="{ height: `${virtualizer.getTotalSize()}px` }"
  >
    <div
      v-for="item in items"
      :key="String(item.key)"
      :ref="measure"
      :data-index="item.index"
      class="virtual-row"
      :style="{ transform: `translateY(${item.start - virtualizer.options.scrollMargin}px)` }"
    >
      <div
        v-if="rows[item.index]?.header"
        class="group-header"
      >
        <h2 class="title is-4 mb-0">
          <RouterLink
            v-if="rows[item.index]?.header?.link"
            :to="rows[item.index]?.header?.link ?? ''"
          >
            {{ rows[item.index]?.header?.title }}
          </RouterLink>
          <template v-else>
            {{ rows[item.index]?.header?.title }}
          </template>
        </h2>
        <span
          v-if="rows[item.index]?.header?.description"
          class="group-count"
        >
          {{ rows[item.index]?.header?.description }}
        </span>
      </div>
      <div
        v-else
        class="song-row"
        :style="{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }"
      >
        <SongCard
          v-for="song in rows[item.index]?.songs"
          :key="song.id"
          :song="song"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.virtual-grid {
  position: relative;
  width: 100%;
}

.virtual-row {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
}

.song-row {
  display: grid;
  gap: 1rem;
  padding-bottom: 1.25rem;
}

@media (max-width: 480px) {
  .song-row {
    gap: 0.625rem;
    padding-bottom: 1rem;
  }
}

.group-header {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  padding: 1.25rem 0 0.75rem;
}

.group-count {
  color: var(--bulma-text-weak);
  font-size: 0.875rem;
}
</style>
