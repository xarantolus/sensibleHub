<script setup lang="ts">
import { computed, ref, useTemplateRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import { useSearch, useSongIndex } from '@/api/queries'
import { useDebounced } from '@/composables/useDebounced'
import { useSearchFocusRequests } from '@/composables/searchFocus'

interface Suggestion {
  label: string
  value: string
  artist: string
}

const router = useRouter()
const input = ref('')
const debounced = useDebounced(input, 150)
const search = useSearch(debounced, 5)
const { resolve } = useSongIndex()
const box = useTemplateRef('box')
const focusRequests = useSearchFocusRequests()

const suggestions = computed<Suggestion[]>(() =>
  input.value.trim() === '' || debounced.value.trim() === ''
    ? []
    : resolve(search.data.value?.songIds).map((s) => ({ label: s.title, value: s.id, artist: s.artist ?? '' })),
)

let picked = false

function onSelect(value: string | undefined): void {
  if (value === undefined) {
    return
  }
  picked = true
  setTimeout(() => {
    picked = false
  })
  input.value = ''
  void router.push({ name: 'song', params: { id: value } })
  box.value?.blur()
}

function onEnter(): void {
  const q = input.value.trim()
  if (picked || q === '') {
    return
  }
  void router.push({ name: 'search', query: { q } })
  box.value?.blur()
}

watch(focusRequests, () => {
  box.value?.focus()
})

defineExpose({
  focus: () => {
    box.value?.focus()
  },
})
</script>

<template>
  <o-autocomplete
    ref="box"
    v-model:input="input"
    :options="suggestions"
    backend-filtering
    :debounce="0"
    max-height="none"
    clear-on-select
    expanded
    rounded
    placeholder="Search songs"
    autocomplete="off"
    @select="onSelect"
    @keydown.enter="onEnter"
  >
    <template #option="{ option }">
      <span class="suggestion">
        <span class="suggestion-title">{{ option.item.label }}</span>
        <span
          v-if="option.item.artist"
          class="suggestion-artist"
        >{{ option.item.artist }}</span>
      </span>
    </template>
    <template
      v-if="input.trim() !== ''"
      #empty
    >
      No matching songs
    </template>
  </o-autocomplete>
</template>

<style scoped>
.suggestion {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.suggestion-title,
.suggestion-artist {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.suggestion-title {
  font-weight: 600;
}

.suggestion-artist {
  font-size: 0.8125rem;
  color: var(--bulma-text-weak);
}
</style>
