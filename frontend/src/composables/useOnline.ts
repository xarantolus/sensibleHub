import { readonly, ref } from 'vue'

const online = ref(typeof navigator === 'undefined' ? true : navigator.onLine)

if (typeof window !== 'undefined') {
  window.addEventListener('online', () => {
    online.value = true
  })
  window.addEventListener('offline', () => {
    online.value = false
  })
}

/** Whether the device reports a network connection. Use it to disable actions that need the server. */
export function useOnline() {
  return readonly(online)
}
