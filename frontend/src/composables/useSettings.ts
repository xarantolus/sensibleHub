import { reactive, watch } from 'vue'

import { parseStatsForNerds } from '@/lib/playerUi'

const storageKey = 'sh-settings-v1'

function read(): string | null {
  try {
    return localStorage.getItem(storageKey)
  } catch {
    return null
  }
}

const settings = reactive({ statsForNerds: parseStatsForNerds(read()) })

watch(settings, () => {
  try {
    localStorage.setItem(storageKey, JSON.stringify(settings))
  } catch {
    // Storage disabled: the setting just won't survive a reload.
  }
})

export function useSettings() {
  return settings
}
