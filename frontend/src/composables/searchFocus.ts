import { ref } from 'vue'

const requests = ref(0)

export function requestSearchFocus(): void {
  requests.value++
}

export function useSearchFocusRequests() {
  return requests
}
