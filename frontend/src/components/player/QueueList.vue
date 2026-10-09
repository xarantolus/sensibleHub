<script setup lang="ts">
import Sortable from 'sortablejs'
import { onBeforeUnmount, onMounted } from 'vue'

import { useSongIndex } from '@/api/queries'
import CoverImage from '@/components/CoverImage.vue'
import { usePlayer, type UpNextList } from '@/stores/player'

import PlayerIcon from './PlayerIcon.vue'

const player = usePlayer()
const { index } = useSongIndex()

const listEls = new Map<UpNextList, HTMLElement>()

function setListEl(name: UpNextList, el: unknown): void {
  if (el instanceof HTMLElement) {
    listEls.set(name, el)
  }
}

const lists: { name: UpNextList; title: string }[] = [
  { name: 'queue', title: 'Queue' },
  { name: 'autoplay', title: 'Autoplay' },
]

function isSlim(name: UpNextList): boolean {
  return name === 'queue' && player.queue.length === 0
}

function listOf(el: HTMLElement): UpNextList {
  return el.dataset.list === 'autoplay' ? 'autoplay' : 'queue'
}

let sortables: Sortable[] = []

onMounted(() => {
  sortables = [...listEls.values()].map((el) =>
    Sortable.create(el, {
      group: 'up-next',
      handle: '.drag-handle',
      animation: 150,
      ghostClass: 'is-ghost',
      onEnd: (e) => {
        const { item, from, to, oldIndex, newIndex } = e
        if (oldIndex === undefined || newIndex === undefined) {
          return
        }
        // Undo Sortable's DOM move; the store change re-renders the lists.
        item.remove()
        from.insertBefore(item, from.children[oldIndex] ?? null)
        if (from !== to || oldIndex !== newIndex) {
          player.move(listOf(from), oldIndex, listOf(to), newIndex)
        }
      },
    }),
  )
})

onBeforeUnmount(() => {
  for (const s of sortables) {
    s.destroy()
  }
})
</script>

<template>
  <section aria-label="Up next">
    <template
      v-for="list in lists"
      :key="list.name"
    >
      <div
        v-if="!isSlim(list.name)"
        class="queue-head"
      >
        <h2 class="title is-5 mb-0">
          {{ list.title }}
        </h2>
        <button
          type="button"
          class="button is-small is-text"
          :disabled="player[list.name].length === 0"
          @click="player.clear(list.name)"
        >
          {{ list.name === 'queue' ? 'Clear' : 'Refresh' }}
        </button>
      </div>

      <ol
        :ref="(el) => setListEl(list.name, el)"
        class="queue"
        :class="{ 'is-slim': isSlim(list.name) }"
        :data-list="list.name"
      >
        <li
          v-for="(id, i) in player[list.name]"
          :key="id"
          class="row"
        >
          <span
            class="drag-handle"
            :aria-label="`Drag to reorder ${index.get(id)?.title ?? 'song'}`"
            title="Drag to reorder"
          >
            <svg
              viewBox="0 0 24 24"
              aria-hidden="true"
            ><path
              fill="currentColor"
              d="M9 5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m0 7a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m-1.5 8.5a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3M18 5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m-1.5 8.5a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3m1.5 5.5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"
            /></svg>
          </span>
          <template v-if="index.get(id)">
            <div class="row-cover">
              <CoverImage :song="index.get(id)!" />
            </div>
            <div class="row-info">
              <span
                class="row-title"
                :title="index.get(id)!.title"
              >{{ index.get(id)!.title }}</span>
              <span class="row-sub">{{ index.get(id)!.artist }}</span>
            </div>
          </template>
          <template v-else>
            <div class="row-cover" />
            <div class="row-info">
              <span class="row-sub">Unavailable song</span>
            </div>
          </template>
          <div class="row-actions">
            <button
              type="button"
              class="sh-icon-button"
              aria-label="Play now"
              title="Play now"
              @click="player.playFrom(list.name, i)"
            >
              <PlayerIcon name="play" />
            </button>
            <button
              type="button"
              class="sh-icon-button"
              aria-label="Remove"
              title="Remove"
              @click="player.removeAt(list.name, i)"
            >
              <PlayerIcon name="close" />
            </button>
          </div>
        </li>
      </ol>
    </template>
  </section>
</template>

<style scoped>
.queue-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1rem;
}

.queue-head:first-child {
  margin-top: 0;
}

.queue {
  list-style: none;
  margin: 0;
  min-height: 2.5rem;
  border-radius: var(--bulma-radius);
}

.queue:empty {
  border: 1px dashed var(--bulma-border);
}

.queue.is-slim {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 2.5rem;
  margin-bottom: 0.5rem;
  border: 1px dashed var(--bulma-border);
  font-size: 0.8rem;
  color: var(--bulma-text-weak);
}

.queue.is-slim::before {
  content: 'Queue';
}

.row {
  display: grid;
  grid-template-columns: 1.5rem 2.5rem minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0;
  border-bottom: 1px solid var(--bulma-border-weak);
  background: inherit;
}

.drag-handle {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 2.5rem;
  color: var(--bulma-text-weak);
  cursor: grab;
  touch-action: none;
}

.drag-handle svg {
  width: 1.25rem;
  height: 1.25rem;
}

.row.is-ghost {
  opacity: 0.4;
  background: var(--bulma-scheme-main-ter);
}

.row-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.row-title,
.row-sub {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row-title {
  font-weight: 600;
  color: var(--bulma-text-strong);
}

.row-sub {
  font-size: 0.8rem;
  color: var(--bulma-text-weak);
}

.row-actions {
  display: flex;
}

.row-actions .sh-icon-button {
  width: 2.25rem;
  height: 2.25rem;
  font-size: 1.25rem;
}
</style>
