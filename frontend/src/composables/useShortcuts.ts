import { onBeforeUnmount, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { usePlayer } from '@/stores/player'

import { requestSearchFocus } from './searchFocus'
import { isInteractiveTarget, isTypingTarget, shortcutDestination } from './shortcutKeys'

export function useShortcuts(): void {
  const router = useRouter()
  const player = usePlayer()

  function go(path: string): void {
    if (router.currentRoute.value.path !== path) {
      void router.push(path)
    }
  }

  function onKeydown(ev: KeyboardEvent): void {
    if (ev.ctrlKey || ev.metaKey || ev.altKey) {
      return
    }

    if (ev.key === 'Escape') {
      if (isTypingTarget(ev.target) && ev.target instanceof HTMLElement) {
        ev.target.blur()
      } else {
        go('/')
      }
      ev.preventDefault()
      return
    }

    if (isTypingTarget(ev.target)) {
      return
    }

    if (ev.shiftKey && (ev.key === 'ArrowRight' || ev.key === 'ArrowLeft')) {
      ev.preventDefault()
      if (ev.key === 'ArrowRight') {
        player.next()
      } else {
        player.previous()
      }
      return
    }

    if (ev.shiftKey) {
      return
    }

    if (ev.key === ' ') {
      if (player.hasSong && !isInteractiveTarget(ev.target)) {
        ev.preventDefault()
        player.playing = !player.playing
      }
      return
    }

    if (ev.key === '/') {
      ev.preventDefault()
      requestSearchFocus()
      return
    }

    const destination = shortcutDestination(ev.key)
    if (destination !== undefined) {
      ev.preventDefault()
      go(destination)
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', onKeydown)
  })
  onBeforeUnmount(() => {
    window.removeEventListener('keydown', onKeydown)
  })
}
