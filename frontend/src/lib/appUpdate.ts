import type { Router } from 'vue-router'

/**
 * Switches to a newly installed app version without cutting off music: at once
 * when nothing plays, otherwise on the next page change made while paused.
 */
export function loadNewVersionWhenIdle(router: Router, isPlaying: () => boolean): void {
  if (!isPlaying()) {
    window.location.reload()
    return
  }
  router.beforeEach((to) => {
    if (isPlaying()) {
      return true
    }
    window.location.assign(to.fullPath)
    return false
  })
}
