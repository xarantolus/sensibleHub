import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

/**
 * The play queue. The server decides what plays after the queue runs out;
 * this store only holds what the user (or the server) put into it.
 */
export const usePlayer = defineStore('player', () => {
  const queue = ref<string[]>([])
  const history = ref<string[]>([])
  const currentId = ref<string | undefined>()
  const playing = ref(false)
  const expanded = ref(false)

  const hasSong = computed(() => currentId.value !== undefined)

  /** Replaces the queue with ids and starts playing ids[start]. */
  function playSongs(ids: readonly string[], start = 0): void {
    const first = ids[start]
    if (first === undefined) {
      return
    }
    if (currentId.value !== undefined) {
      history.value.push(currentId.value)
    }
    currentId.value = first
    queue.value = ids.slice(start + 1)
    playing.value = true
  }

  function playSong(id: string): void {
    playSongs([id])
  }

  function playNext(id: string): void {
    queue.value = [id, ...queue.value.filter((q) => q !== id)]
    if (currentId.value === undefined) {
      next()
    }
  }

  function enqueue(id: string): void {
    queue.value = [...queue.value.filter((q) => q !== id), id]
    if (currentId.value === undefined) {
      next()
    }
  }

  function next(): void {
    const [head, ...rest] = queue.value
    if (head === undefined) {
      return
    }
    if (currentId.value !== undefined) {
      history.value.push(currentId.value)
    }
    currentId.value = head
    queue.value = rest
    playing.value = true
  }

  function previous(): void {
    const prev = history.value.pop()
    if (prev === undefined) {
      return
    }
    if (currentId.value !== undefined) {
      queue.value = [currentId.value, ...queue.value]
    }
    currentId.value = prev
    playing.value = true
  }

  return { queue, history, currentId, playing, expanded, hasSong, playSongs, playSong, playNext, enqueue, next, previous }
})
