<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useTemplateRef } from 'vue'

import { usePlayer } from '@/stores/player'

const props = withDefaults(
  defineProps<{
    /** The songs to act on, in play order. */
    ids: readonly string[]
    /** Song to start a radio from; empty hides that option. */
    radioId?: string
    /** `icon`: round button on song cards; `small`: compact list button; `button`: labelled page button. */
    variant?: 'icon' | 'small' | 'button'
    label?: string
    primary?: boolean
  }>(),
  { variant: 'button', label: 'Play', primary: true, radioId: '' },
)

const player = usePlayer()
const trigger = useTemplateRef<HTMLButtonElement>('trigger')
const open = ref(false)
const pos = ref({ top: 0, left: 0, up: false, right: false })

const iconOnly = computed(() => props.variant !== 'button')
const buttonClass = computed(() => {
  switch (props.variant) {
    case 'icon':
      return ['song-actions-icon', 'button', 'is-rounded']
    case 'small':
      return ['button', 'is-small']
    case 'button':
      return ['button', { 'is-primary': props.primary }]
  }
})

/** Nothing playing: the button just plays. Otherwise it opens the menu. */
function onClick(): void {
  if (!player.hasSong) {
    player.playNow(props.ids)
    return
  }
  if (open.value) {
    close()
    return
  }
  const r = trigger.value?.getBoundingClientRect()
  if (r === undefined) {
    return
  }
  const up = r.bottom > window.innerHeight * 0.6
  const right = r.left > window.innerWidth / 2
  pos.value = { top: up ? r.top : r.bottom, left: right ? r.right : r.left, up, right }
  open.value = true
  window.addEventListener('pointerdown', onOutside, true)
  window.addEventListener('keydown', onKey)
  window.addEventListener('scroll', close, true)
  window.addEventListener('resize', close)
}

function close(): void {
  open.value = false
  window.removeEventListener('pointerdown', onOutside, true)
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('scroll', close, true)
  window.removeEventListener('resize', close)
}

function onOutside(e: PointerEvent): void {
  const t = e.target
  if (t instanceof Node && !trigger.value?.contains(t) && !(t instanceof Element && t.closest('.song-actions-menu'))) {
    close()
  }
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') {
    e.stopPropagation()
    close()
    trigger.value?.focus()
  }
}

function run(action: () => void): void {
  action()
  close()
}

onBeforeUnmount(close)
</script>

<template>
  <button
    ref="trigger"
    type="button"
    :class="[buttonClass, { 'is-primary': variant === 'icon' && !player.hasSong }]"
    :aria-label="iconOnly ? (player.hasSong ? `More options: ${label}` : label) : undefined"
    :title="iconOnly ? (player.hasSong ? 'More options' : label) : undefined"
    :aria-haspopup="player.hasSong ? 'menu' : undefined"
    :aria-expanded="player.hasSong ? open : undefined"
    :disabled="ids.length === 0"
    @click.stop.prevent="onClick"
  >
    <template v-if="iconOnly">
      {{ player.hasSong ? '⋮' : '▶' }}
    </template>
    <template v-else>
      {{ label }}{{ player.hasSong ? ' ▾' : '' }}
    </template>
  </button>
  <Teleport
    v-if="open"
    to="body"
  >
    <div
      class="dropdown is-active song-actions-menu"
      :class="{ 'is-up': pos.up, 'is-right': pos.right }"
      :style="{ top: `${pos.top}px`, left: `${pos.left}px` }"
    >
      <div
        class="dropdown-menu"
        role="menu"
      >
        <div class="dropdown-content">
          <button
            type="button"
            class="dropdown-item"
            role="menuitem"
            @click="run(() => player.playNow(ids))"
          >
            Play now
          </button>
          <button
            type="button"
            class="dropdown-item"
            role="menuitem"
            @click="run(() => player.playNext(ids))"
          >
            Play next
          </button>
          <button
            type="button"
            class="dropdown-item"
            role="menuitem"
            @click="run(() => player.enqueue(ids))"
          >
            Add to queue
          </button>
          <button
            v-if="radioId"
            type="button"
            class="dropdown-item"
            role="menuitem"
            @click="run(() => player.startRadio(radioId))"
          >
            Start radio
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.song-actions-icon {
  width: 2.75rem;
  height: 2.75rem;
  padding: 0;
  font-size: 1.1rem;
}
</style>

<style>
.dropdown.song-actions-menu {
  position: fixed;
  z-index: 60;
  width: 0;
  height: 0;
}

.song-actions-menu .dropdown-item {
  width: 100%;
  border: 0;
  background: none;
  text-align: left;
  cursor: pointer;
  font: inherit;
  color: var(--bulma-text);
}

.song-actions-menu .dropdown-item:hover {
  background: var(--bulma-scheme-main-ter);
}

.dropdown.song-actions-menu .dropdown-menu {
  position: absolute;
  min-width: 11rem;
}
</style>
