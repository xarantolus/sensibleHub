<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import { useSongIndex } from '@/api/queries'
import CoverImage from '@/components/CoverImage.vue'
import { useSettings } from '@/composables/useSettings'
import { formatDuration } from '@/lib/format'
import { shouldDismiss } from '@/lib/playerUi'
import { usePlayer } from '@/stores/player'

import PlayerIcon from './PlayerIcon.vue'
import PlayerNerdStats from './PlayerNerdStats.vue'
import QueueList from './QueueList.vue'

const player = usePlayer()
const settings = useSettings()
const { index } = useSongIndex()

const song = computed(() => (player.currentId === undefined ? undefined : index.value.get(player.currentId)))
const scrub = ref<number>()
const shown = computed(() => scrub.value ?? player.position)
const remaining = computed(() => Math.max(0, player.duration - shown.value))

function close(): void {
  player.expanded = false
}

watch(
  () => player.expanded && song.value !== undefined,
  (open) => {
    document.documentElement.classList.toggle('is-clipped', open)
  },
  { immediate: true },
)

watch(song, (s) => {
  if (s === undefined) {
    close()
  }
})

function onKeydown(ev: KeyboardEvent): void {
  if (ev.key === 'Escape' && player.expanded) {
    ev.preventDefault()
    ev.stopPropagation()
    close()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown, true)
})
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown, true)
  document.documentElement.classList.remove('is-clipped')
})

/** How far the sheet is dragged down, in pixels; it follows the finger and is kept when closing, so the exit animation continues from there. */
const drag = ref(0)
const dragging = ref(false)
let touchStart: { x: number; y: number; t: number } | undefined

function onTouchStart(ev: TouchEvent): void {
  const t = ev.touches[0]
  touchStart = t === undefined ? undefined : { x: t.clientX, y: t.clientY, t: ev.timeStamp }
  drag.value = 0
}

function onTouchMove(ev: TouchEvent): void {
  const t = ev.touches[0]
  if (touchStart === undefined || t === undefined) {
    return
  }
  const dy = t.clientY - touchStart.y
  if (!dragging.value && Math.abs(t.clientX - touchStart.x) > Math.abs(dy)) {
    touchStart = undefined
    return
  }
  dragging.value = true
  drag.value = Math.max(0, dy)
}

function onTouchEnd(ev: TouchEvent): void {
  const t = ev.changedTouches[0]
  if (touchStart !== undefined && t !== undefined && shouldDismiss(t.clientY - touchStart.y, ev.timeStamp - touchStart.t)) {
    close()
  } else {
    drag.value = 0
  }
  dragging.value = false
  touchStart = undefined
}

function onAfterLeave(): void {
  drag.value = 0
}

function onScrub(ev: Event): void {
  if (ev.target instanceof HTMLInputElement) {
    scrub.value = Number(ev.target.value)
  }
}

function commitScrub(ev: Event): void {
  if (ev.target instanceof HTMLInputElement) {
    player.seekTo(Number(ev.target.value))
  }
  scrub.value = undefined
}
</script>

<template>
  <Transition
    name="sheet"
    :duration="{ enter: 420, leave: 300 }"
    @after-leave="onAfterLeave"
  >
    <div
      v-if="player.expanded && song"
      class="full-backdrop"
      :style="{ '--drag': `${drag}px` }"
      @click.self="close"
    >
      <div
        class="full"
        :class="{ 'is-dragging': dragging }"
        role="dialog"
        aria-modal="true"
        aria-label="Player"
      >
        <header
          class="full-header"
          @touchstart.passive="onTouchStart"
          @touchmove.passive="onTouchMove"
          @touchend.passive="onTouchEnd"
          @touchcancel.passive="onTouchEnd"
        >
          <span
            class="grab-handle"
            aria-hidden="true"
          />
          <button
            type="button"
            class="sh-icon-button"
            aria-label="Close player"
            @click="close"
          >
            <PlayerIcon name="chevron-down" />
          </button>
          <span class="full-header-title has-text-weight-semibold">Now playing</span>
          <span class="header-spacer" />
        </header>

        <div class="full-scroll">
          <div class="full-layout">
            <div class="full-main">
              <div class="full-cover">
                <CoverImage
                  :song="song"
                  size="full"
                  eager
                />
              </div>

              <div class="full-meta">
                <h1 class="title is-4 mb-1">
                  <RouterLink
                    :to="{ name: 'song', params: { id: song.id } }"
                    class="full-title-link"
                    @click="close"
                  >
                    {{ song.title }}
                  </RouterLink>
                </h1>
                <p class="full-links">
                  <RouterLink
                    v-if="song.artist"
                    :to="{ name: 'artist', params: { artist: song.artist } }"
                    @click="close"
                  >
                    {{ song.artist }}
                  </RouterLink>
                  <template v-if="song.artist && song.album">
                    ·
                  </template>
                  <RouterLink
                    v-if="song.artist && song.album"
                    :to="{ name: 'album', params: { artist: song.artist, album: song.album } }"
                    @click="close"
                  >
                    {{ song.album }}
                  </RouterLink>
                  <span v-else-if="song.album">{{ song.album }}</span>
                </p>
                <p
                  v-if="player.status === 'loading' || player.status === 'waiting-network'"
                  class="has-text-primary is-size-7"
                  aria-live="polite"
                >
                  {{ player.status === 'loading' ? 'Loading…' : 'Waiting for connection…' }}
                </p>
              </div>

              <div class="full-seek">
                <input
                  class="sh-range"
                  type="range"
                  min="0"
                  :max="player.duration"
                  step="1"
                  :value="shown"
                  aria-label="Seek"
                  @input="onScrub"
                  @change="commitScrub"
                >
                <div class="full-times has-text-grey">
                  <span>{{ formatDuration(shown) }}</span>
                  <span>-{{ formatDuration(remaining) }}</span>
                </div>
              </div>

              <div class="full-controls">
                <button
                  type="button"
                  class="sh-icon-button big"
                  aria-label="Previous song"
                  @click="player.previous()"
                >
                  <PlayerIcon name="prev" />
                </button>
                <button
                  type="button"
                  class="sh-icon-button big is-primary"
                  :aria-label="player.playing ? 'Pause' : 'Play'"
                  @click="player.toggle()"
                >
                  <PlayerIcon :name="player.playing ? 'pause' : 'play'" />
                </button>
                <button
                  type="button"
                  class="sh-icon-button big"
                  aria-label="Next song"
                  @click="void player.next()"
                >
                  <PlayerIcon name="next" />
                </button>
              </div>

              <label class="full-volume">
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
            </div>

            <div class="full-side">
              <QueueList />
              <PlayerNerdStats
                v-if="settings.statsForNerds && player.currentId"
                :id="player.currentId"
                class="mt-5"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.full-backdrop {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  justify-content: center;
  align-items: stretch;
  background: color-mix(in srgb, var(--bulma-scheme-invert) 55%, transparent);
}

.full {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 100%;
  min-height: 0;
  background: var(--bulma-scheme-main);
  color: var(--bulma-text);
  overflow: hidden;
}

.full-header {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: calc(0.5rem + env(safe-area-inset-top)) 0.75rem 0.5rem;
  touch-action: pan-x;
}

.header-spacer {
  width: 2.5rem;
}

.full-scroll {
  position: relative;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 0 1rem calc(1.5rem + env(safe-area-inset-bottom));
}

.full-layout {
  display: grid;
  gap: 1.5rem;
  max-width: 64rem;
  margin: 0 auto;
}

.full-main {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 26rem;
  margin: 0 auto;
}

.full-cover {
  width: min(100%, 55vh);
  margin: 0 auto;
  box-shadow: 0 0.5rem 2rem color-mix(in srgb, var(--bulma-scheme-invert) 30%, transparent);
  border-radius: var(--bulma-radius);
}

.full-meta {
  min-width: 0;
}

.full-title-link {
  color: inherit;
}

.full-title-link:hover {
  text-decoration: underline;
}

.full-meta .title {
  overflow-wrap: anywhere;
}

.full-links a {
  text-decoration: underline;
}

.full-times {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
  font-variant-numeric: tabular-nums;
}

.full-controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1.25rem;
}

.sh-icon-button.big {
  width: 3.25rem;
  height: 3.25rem;
  font-size: 2rem;
}

.sh-icon-button.big.is-primary {
  width: 4.25rem;
  height: 4.25rem;
  font-size: 2.5rem;
}

.full-volume {
  display: none;
  align-items: center;
  gap: 0.5rem;
}

.grab-handle {
  position: absolute;
  top: calc(0.35rem + env(safe-area-inset-top));
  left: 50%;
  width: 2.25rem;
  height: 0.3rem;
  margin-left: -1.125rem;
  border-radius: 1rem;
  background: var(--bulma-text-weak);
  opacity: 0.5;
}

/* The backdrop only fades; the panel moves. Easing follows the iOS sheet curve. */
.full {
  transform: translateY(var(--drag, 0));
  transition: transform 0.3s cubic-bezier(0.32, 0.72, 0, 1);
  border-radius: 1rem 1rem 0 0;
}

.full.is-dragging {
  transition: none;
}

/* Only the dimming fades; fading the backdrop element would make the panel see-through. */
.sheet-enter-active,
.sheet-leave-active {
  transition: background-color 0.3s ease;
}

.sheet-enter-active .full {
  transition: transform 0.42s cubic-bezier(0.32, 0.72, 0, 1);
}

.sheet-leave-active .full {
  transition: transform 0.3s cubic-bezier(0.4, 0, 1, 1);
}

.sheet-enter-from,
.sheet-leave-to {
  background-color: transparent;
}

.sheet-enter-from .full,
.sheet-leave-to .full {
  transform: translateY(100%);
}

@media (min-width: 769px) {
  .full-backdrop {
    padding: 2rem;
  }

  .full {
    max-width: 68rem;
    border-radius: var(--bulma-radius-large);
    box-shadow: var(--bulma-shadow);
  }

  .grab-handle {
    display: none;
  }

  .sheet-enter-active .full,
  .sheet-leave-active .full {
    transition:
      transform 0.3s cubic-bezier(0.2, 0.8, 0.2, 1),
      opacity 0.3s ease;
  }

  .sheet-enter-from .full,
  .sheet-leave-to .full {
    transform: translateY(1.5rem) scale(0.97);
    opacity: 0;
  }

  .full-volume {
    display: flex;
  }

  .full-layout {
    grid-template-columns: minmax(0, 26rem) minmax(0, 1fr);
    align-items: start;
  }

  .full-cover {
    width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .full,
  .sheet-enter-active,
  .sheet-leave-active,
  .sheet-enter-active .full,
  .sheet-leave-active .full {
    transition: none;
  }
}
</style>
