import { useQueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { getActivePinia } from 'pinia'
import { createApp, readonly, ref, type Component } from 'vue'

interface DocumentPictureInPicture {
  requestWindow(options?: { width?: number; height?: number }): Promise<Window>
  window: Window | null
}

declare global {
  interface Window {
    documentPictureInPicture?: DocumentPictureInPicture
  }
}

const isOpen = ref(false)

/**
 * Pops a component out into an always-on-top window (Chromium's Document
 * Picture-in-Picture). It shares the player and query state with the page.
 * `supported` is false in other browsers; hide the button there.
 */
export function usePictureInPicture(component: Component) {
  const supported = typeof window !== 'undefined' && window.documentPictureInPicture !== undefined
  const queryClient = useQueryClient()
  const pinia = getActivePinia()

  async function open(): Promise<void> {
    const pip = window.documentPictureInPicture
    if (pip === undefined || pinia === undefined) {
      return
    }
    if (pip.window !== null) {
      pip.window.focus()
      return
    }

    const win = await pip.requestWindow({ width: 380, height: 140 })
    for (const node of document.head.querySelectorAll('style, link[rel="stylesheet"]')) {
      win.document.head.append(node.cloneNode(true))
    }
    win.document.documentElement.className = document.documentElement.className
    const root = win.document.createElement('div')
    win.document.body.append(root)

    const app = createApp(component, { compact: true })
    app.use(pinia).use(VueQueryPlugin, { queryClient }).mount(root)
    isOpen.value = true

    win.addEventListener('pagehide', () => {
      app.unmount()
      isOpen.value = false
    })
  }

  return { supported, isOpen: readonly(isOpen), open }
}
