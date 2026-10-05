<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'

import { coverUrl, placeholderCover } from '@/api/media'
import type { SongSummary } from '@/api/schema'
import { useDeleteCover, useSetCover } from '@/api/queries'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { useOnline } from '@/composables/useOnline'
import { notify } from '@/lib/notify'

const props = defineProps<{ song: SongSummary; loading?: boolean }>()

const online = useOnline()
const setCover = useSetCover()
const deleteCover = useDeleteCover()

const input = ref<HTMLInputElement>()
const file = ref<File>()
const preview = ref<string>()
const failed = ref(false)
const confirmDelete = ref(false)

const locked = computed(() => !online.value || props.loading)
const lockedTitle = computed(() => (online.value ? undefined : 'You are offline'))
const src = computed(() =>
  preview.value ??
  (props.song.cover === undefined || failed.value
    ? placeholderCover
    : coverUrl(props.song, 'full')),
)

function revoke(): void {
  if (preview.value !== undefined) {
    URL.revokeObjectURL(preview.value)
    preview.value = undefined
  }
}

function pick(): void {
  if (!locked.value) {
    input.value?.click()
  }
}

function onChange(event: Event): void {
  const target = event.target as HTMLInputElement
  const chosen = target.files?.[0]
  target.value = ''
  if (chosen === undefined) {
    return
  }
  revoke()
  file.value = chosen
  preview.value = URL.createObjectURL(chosen)
}

function cancel(): void {
  revoke()
  file.value = undefined
}

function save(): void {
  if (file.value === undefined) {
    return
  }
  setCover.mutate(
    { id: props.song.id, file: file.value },
    {
      onSuccess: () => {
        cancel()
        failed.value = false
        notify('Cover updated', 'success')
      },
    },
  )
}

function remove(): void {
  deleteCover.mutate(props.song.id, {
    onSuccess: () => {
      confirmDelete.value = false
      notify('Cover removed', 'success')
    },
  })
}

onBeforeUnmount(revoke)
</script>

<template>
  <div>
    <figure
      class="image is-square song-cover"
      :style="{ backgroundColor: song.cover?.color }"
    >
      <img
        :src="src"
        :alt="`Cover of ${song.title}`"
        @error="failed = true"
      >
      <div
        v-if="file === undefined && !loading"
        class="song-cover-overlay"
        :class="{ 'is-disabled': locked }"
        role="button"
        :tabindex="locked ? -1 : 0"
        :aria-disabled="locked"
        :title="lockedTitle"
        @click="pick"
        @keydown.enter.prevent="pick"
        @keydown.space.prevent="pick"
      >
        <span class="button is-small is-primary">Upload a new cover image</span>
        <button
          v-if="song.cover"
          type="button"
          class="button is-small is-danger"
          :disabled="locked"
          @click.stop="confirmDelete = true"
        >
          Delete current cover
        </button>
      </div>
      <input
        ref="input"
        class="is-hidden"
        type="file"
        accept="image/png,image/jpeg,image/webp,image/gif"
        tabindex="-1"
        @change="onChange"
      >
    </figure>
    <div
      v-if="file !== undefined"
      class="buttons mt-3"
    >
      <button
        type="button"
        class="button is-primary"
        :class="{ 'is-loading': setCover.isPending.value }"
        :disabled="locked"
        :title="lockedTitle"
        @click="save"
      >
        Save cover
      </button>
      <button
        type="button"
        class="button"
        :disabled="setCover.isPending.value"
        @click="cancel"
      >
        Cancel
      </button>
    </div>
    <p
      v-else-if="song.cover"
      class="help"
    >
      The cover image size is {{ song.cover.size }}x{{ song.cover.size }}px
    </p>

    <ConfirmModal
      v-model:active="confirmDelete"
      title="Delete cover?"
      confirm-label="Delete"
      :busy="deleteCover.isPending.value"
      @confirm="remove"
    >
      <p>The cover of "{{ song.title }}" will be deleted.</p>
    </ConfirmModal>
  </div>
</template>

<style scoped>
.song-cover {
  position: relative;
  overflow: hidden;
  border-radius: var(--bulma-radius-large);
  background-color: var(--bulma-background);
}

.song-cover img {
  object-fit: cover;
}

.song-cover-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 1rem;
  background: color-mix(in srgb, var(--bulma-scheme-main) 70%, transparent);
  opacity: 0;
  cursor: pointer;
  transition: opacity 0.15s ease;
}

.song-cover-overlay.is-disabled {
  cursor: not-allowed;
}

.song-cover:hover .song-cover-overlay,
.song-cover-overlay:focus-visible,
.song-cover-overlay:focus-within {
  opacity: 1;
}

@media (hover: none) {
  .song-cover-overlay {
    opacity: 1;
    inset: auto 0 0 0;
    padding: 0.75rem;
  }
}
</style>
