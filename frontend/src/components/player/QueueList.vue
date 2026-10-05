<script setup lang="ts">
import { computed } from 'vue'

import { useSongIndex } from '@/api/queries'
import CoverImage from '@/components/CoverImage.vue'
import { moveTarget } from '@/lib/playerUi'
import { usePlayer } from '@/stores/player'

import PlayerIcon from './PlayerIcon.vue'

const player = usePlayer()
const { index } = useSongIndex()

const rows = computed(() =>
  player.queue.flatMap((id, i) => {
    const song = index.value.get(id)
    return song === undefined ? [] : [{ i, song }]
  }),
)

function playNow(i: number): void {
  player.moveInQueue(i, 0)
  void player.next()
}

function move(i: number, delta: -1 | 1): void {
  const to = moveTarget(i, delta, player.queue.length)
  if (to !== undefined) {
    player.moveInQueue(i, to)
  }
}
</script>

<template>
  <section aria-label="Up next">
    <div class="queue-head">
      <h2 class="title is-5 mb-0">
        Up next
      </h2>
      <button
        type="button"
        class="button is-small is-light"
        :disabled="player.queue.length === 0"
        @click="player.clearQueue()"
      >
        Clear
      </button>
    </div>

    <p
      v-if="player.refilling"
      class="has-text-grey queue-note"
    >
      Finding more songs…
    </p>
    <p
      v-else-if="rows.length === 0"
      class="has-text-grey queue-note"
    >
      Nothing queued.
    </p>

    <ol class="queue">
      <li
        v-for="{ i, song } in rows"
        :key="`${i}-${song.id}`"
        class="row"
      >
        <div class="row-cover">
          <CoverImage :song="song" />
        </div>
        <div class="row-info">
          <span
            class="row-title"
            :title="song.title"
          >{{ song.title }}</span>
          <span class="row-sub">
            <span
              v-if="player.suggested.has(song.id)"
              class="tag is-light is-primary"
            >Suggested</span>
            {{ song.artist }}
          </span>
        </div>
        <div class="row-actions">
          <button
            type="button"
            class="sh-icon-button"
            :aria-label="`Play ${song.title} now`"
            title="Play now"
            @click="playNow(i)"
          >
            <PlayerIcon name="play" />
          </button>
          <button
            type="button"
            class="sh-icon-button"
            :aria-label="`Move ${song.title} up`"
            title="Move up"
            :disabled="i === 0"
            @click="move(i, -1)"
          >
            <PlayerIcon name="up" />
          </button>
          <button
            type="button"
            class="sh-icon-button"
            :aria-label="`Move ${song.title} down`"
            title="Move down"
            :disabled="i === player.queue.length - 1"
            @click="move(i, 1)"
          >
            <PlayerIcon name="down" />
          </button>
          <button
            type="button"
            class="sh-icon-button"
            :aria-label="`Remove ${song.title} from queue`"
            title="Remove"
            @click="player.removeFromQueue(i)"
          >
            <PlayerIcon name="close" />
          </button>
        </div>
      </li>
    </ol>
  </section>
</template>

<style scoped>
.queue-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.queue-note {
  padding: 0.5rem 0;
}

.queue {
  list-style: none;
  margin: 0;
}

.row {
  display: grid;
  grid-template-columns: 2.5rem minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0;
  border-bottom: 1px solid var(--bulma-border-weak);
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

@media (max-width: 480px) {
  .row-actions .sh-icon-button {
    width: 2rem;
  }
}
</style>
