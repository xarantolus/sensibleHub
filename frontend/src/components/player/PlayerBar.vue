<script setup lang="ts">
import { computed } from 'vue'

import { useSongIndex } from '@/api/queries'
import CoverImage from '@/components/CoverImage.vue'
import { usePictureInPicture } from '@/composables/usePictureInPicture'
import { formatDuration } from '@/lib/format'
import { progressPercent } from '@/lib/playerUi'
import { usePlayer } from '@/stores/player'

import PlayerBar from './PlayerBar.vue'
import PlayerIcon from './PlayerIcon.vue'

const props = defineProps<{ compact?: boolean }>()

const player = usePlayer()
const { index } = useSongIndex()
const pip = usePictureInPicture(PlayerBar)

const song = computed(() => (player.currentId === undefined ? undefined : index.value.get(player.currentId)))
const percent = computed(() => progressPercent(player.position, player.duration))
const hint = computed(() => {
  if (player.status === 'loading') {
    return 'Loading…'
  }
  if (player.status === 'waiting-network') {
    return 'Waiting for connection…'
  }
  return undefined
})

function expand(): void {
  if (!props.compact) {
    player.expanded = true
  }
}

function seek(ev: Event): void {
  if (ev.target instanceof HTMLInputElement) {
    player.seekTo(Number(ev.target.value))
  }
}
</script>

<template>
  <div
    v-if="player.hasSong && song && (compact || !player.expanded)"
    class="player-bar"
    :class="{ 'is-compact': compact }"
    @click="expand"
  >
    <div
      v-if="!compact"
      class="bar-progress"
      aria-hidden="true"
    >
      <div
        class="bar-progress-fill"
        :style="{ width: `${percent}%` }"
      />
    </div>

    <button
      type="button"
      class="bar-cover"
      aria-label="Open player"
      @click.stop="expand"
    >
      <CoverImage
        :song="song"
        eager
      />
    </button>

    <div class="bar-info">
      <component
        :is="compact ? 'span' : 'RouterLink'"
        v-bind="compact ? {} : { to: { name: 'song', params: { id: song.id } } }"
        class="bar-title"
        :title="song.title"
        @click.stop
      >
        {{ song.title }}
      </component>
      <span
        class="bar-sub"
        :class="{ 'has-text-primary': hint }"
        aria-live="polite"
      >{{ hint ?? song.artist ?? '' }}</span>
    </div>

    <div
      class="bar-controls"
      @click.stop
    >
      <button
        type="button"
        class="sh-icon-button"
        aria-label="Previous song"
        @click="player.previous()"
      >
        <PlayerIcon name="prev" />
      </button>
      <button
        type="button"
        class="sh-icon-button is-primary"
        :aria-label="player.playing ? 'Pause' : 'Play'"
        @click="player.toggle()"
      >
        <PlayerIcon :name="player.playing ? 'pause' : 'play'" />
      </button>
      <button
        type="button"
        class="sh-icon-button"
        aria-label="Next song"
        @click="void player.next()"
      >
        <PlayerIcon name="next" />
      </button>
    </div>

    <div
      v-if="compact"
      class="bar-seek"
    >
      <input
        class="sh-range"
        type="range"
        min="0"
        :max="player.duration"
        step="1"
        :value="player.position"
        aria-label="Seek"
        @change="seek"
      >
    </div>

    <div
      class="bar-extra"
      @click.stop
    >
      <span class="bar-time has-text-grey">
        {{ formatDuration(player.position) }} / {{ formatDuration(player.duration) }}
      </span>
      <template v-if="!compact">
        <label class="bar-volume">
          <PlayerIcon name="volume" />
          <input
            v-model.number="player.volume"
            class="sh-range"
            type="range"
            min="0"
            max="1"
            step="0.05"
            aria-label="Volume"
          >
        </label>
        <button
          v-if="pip.supported"
          type="button"
          class="sh-icon-button"
          aria-label="Pop out player"
          title="Pop out player"
          @click="void pip.open()"
        >
          <PlayerIcon name="pip" />
        </button>
      </template>
    </div>
  </div>
</template>

<style scoped>
.player-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 40;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem calc(0.5rem + env(safe-area-inset-bottom));
  background: var(--bulma-scheme-main-bis);
  border-top: 1px solid var(--bulma-border-weak);
  color: var(--bulma-text);
  cursor: pointer;
}

.bar-progress {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--bulma-border-weak);
}

.bar-progress-fill {
  height: 100%;
  background: var(--bulma-primary);
}

.bar-cover {
  width: 3rem;
  padding: 0;
  border: 0;
  background: none;
  cursor: pointer;
}

.bar-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.bar-title,
.bar-sub {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bar-title {
  font-weight: 600;
  color: var(--bulma-text-strong);
}

.bar-sub {
  font-size: 0.8rem;
  color: var(--bulma-text-weak);
}

.bar-controls {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  cursor: default;
}

.bar-extra {
  display: none;
  align-items: center;
  gap: 0.75rem;
  cursor: default;
}

.bar-time {
  font-size: 0.8rem;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.bar-volume {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  width: 8rem;
}

@media (max-width: 768px) {
  .player-bar:not(.is-compact) {
    grid-template-columns: auto minmax(0, 1fr);
    grid-template-areas:
      'cover info'
      'controls controls';
    row-gap: 0.25rem;
  }

  .player-bar:not(.is-compact) .bar-cover {
    grid-area: cover;
    width: 2.75rem;
  }

  .player-bar:not(.is-compact) .bar-info {
    grid-area: info;
  }

  .player-bar:not(.is-compact) .bar-controls {
    grid-area: controls;
    justify-content: center;
    gap: 1.5rem;
  }
}

@media (min-width: 769px) {
  .player-bar:not(.is-compact) {
    grid-template-columns: auto minmax(0, 1fr) auto minmax(0, 1fr);
    padding-left: 1.5rem;
    padding-right: 1.5rem;
  }

  .player-bar:not(.is-compact) .bar-extra {
    display: flex;
    justify-content: flex-end;
  }
}

.is-compact {
  position: static;
  min-height: 100vh;
  grid-template-columns: 5.5rem minmax(0, 1fr);
  grid-template-areas:
    'cover info'
    'cover controls'
    'seek seek'
    'extra extra';
  align-content: center;
  padding: 0.75rem;
  border-top: 0;
  cursor: default;
}

.is-compact .bar-cover {
  grid-area: cover;
  width: 5.5rem;
}

.is-compact .bar-info {
  grid-area: info;
}

.is-compact .bar-controls {
  grid-area: controls;
}

.is-compact .bar-seek {
  grid-area: seek;
}

.is-compact .bar-extra {
  grid-area: extra;
  display: flex;
  justify-content: center;
}
</style>
