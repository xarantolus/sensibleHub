<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { isApiError } from '@/api/client'
import { audioUrl } from '@/api/media'
import {
  useDeleteCover,
  useDeleteSong,
  useSetCover,
  useUpdateSong,
} from '@/api/queries'
import type { SongDetail } from '@/api/schema'
import { notify } from '@/lib/notify'
import { useOnline } from '@/composables/useOnline'

import ConfirmModal from './ConfirmModal.vue'
import CoverUpload from './CoverUpload.vue'

const props = defineProps<{ song: SongDetail }>()

const router = useRouter()
const online = useOnline()
const offlineTitle = computed(() => (online.value ? undefined : 'You are offline'))

const update = useUpdateSong()
const setCover = useSetCover()
const deleteCover = useDeleteCover()
const deleteSong = useDeleteSong()

const open = ref(false)
const title = ref('')
const artist = ref('')
const album = ref('')
const year = ref('')
const start = ref('0')
const end = ref('0')
const sync = ref(false)
const errors = ref<Record<string, string>>({})
const confirmCover = ref(false)
const confirmSong = ref(false)
const upload = ref<InstanceType<typeof CoverUpload>>()

function reset(): void {
  const s = props.song
  title.value = s.title
  artist.value = s.artist ?? ''
  album.value = s.album ?? ''
  year.value = s.year === undefined ? '' : String(s.year)
  start.value = String(s.playback.start)
  end.value = String(s.playback.end)
  sync.value = s.sync
  errors.value = {}
}

watch(() => props.song.lastEdit, reset, { immediate: true })

const yearValue = computed(() => (year.value.trim() === '' ? undefined : Number(year.value)))
const startValue = computed(() => (start.value.trim() === '' ? Number.NaN : Number(start.value)))
const endValue = computed(() => (end.value.trim() === '' ? Number.NaN : Number(end.value)))

const localErrors = computed(() => {
  const out: Record<string, string> = {}
  if (title.value.trim() === '') {
    out.title = 'A title is required'
  }
  if (yearValue.value !== undefined && !Number.isInteger(yearValue.value)) {
    out.year = 'The year must be a whole number'
  }
  if (!(startValue.value >= 0 && startValue.value <= props.song.duration)) {
    out.start = `Must be between 0 and ${String(props.song.duration)}`
  }
  if (!(endValue.value >= 0 && endValue.value <= props.song.duration)) {
    out.end = `Must be between 0 and ${String(props.song.duration)}`
  } else if (startValue.value >= endValue.value) {
    out.end = 'The end must be after the start'
  }
  return out
})

function message(field: string): string | undefined {
  return errors.value[field] ?? localErrors.value[field]
}

function variant(field: string): string | undefined {
  return message(field) === undefined ? undefined : 'danger'
}

function save(): void {
  errors.value = {}
  if (Object.keys(localErrors.value).length > 0) {
    errors.value = localErrors.value
    return
  }
  const edit = {
    title: title.value.trim(),
    artist: artist.value.trim(),
    album: album.value.trim(),
    start: startValue.value,
    end: endValue.value,
    sync: sync.value,
    ...(yearValue.value === undefined ? {} : { year: yearValue.value }),
  }
  update.mutate(
    { id: props.song.id, edit },
    {
      onSuccess: () => {
        notify('Song saved', 'success')
      },
      onError: (err) => {
        if (isApiError(err)) {
          errors.value = { ...err.fieldErrors }
        }
      },
    },
  )
}

function changeCover(file: File): void {
  setCover.mutate(
    { id: props.song.id, file },
    {
      onSuccess: () => {
        upload.value?.clear()
        notify('Cover updated', 'success')
      },
    },
  )
}

function removeCover(): void {
  deleteCover.mutate(props.song.id, {
    onSuccess: () => {
      confirmCover.value = false
      notify('Cover removed', 'success')
    },
  })
}

function removeSong(): void {
  deleteSong.mutate(props.song.id, {
    onSuccess: () => {
      confirmSong.value = false
      notify('Song deleted', 'success')
      void router.replace('/songs')
    },
  })
}
</script>

<template>
  <o-collapse
    v-model:open="open"
    label="Edit details"
    class="box"
  >
    <form
      class="pt-4"
      novalidate
      @submit.prevent="save"
    >
      <o-field
        label="Title"
        :variant="variant('title')"
        :message="message('title')"
      >
        <input
          v-model="title"
          class="input"
          type="text"
          required
        >
      </o-field>
      <div class="columns is-multiline">
        <div class="column is-12-mobile is-4-tablet">
          <o-field
            label="Artist"
            :variant="variant('artist')"
            :message="message('artist')"
          >
            <input
              v-model="artist"
              class="input"
              type="text"
            >
          </o-field>
        </div>
        <div class="column is-12-mobile is-5-tablet">
          <o-field
            label="Album"
            :variant="variant('album')"
            :message="message('album')"
          >
            <input
              v-model="album"
              class="input"
              type="text"
            >
          </o-field>
        </div>
        <div class="column is-12-mobile is-3-tablet">
          <o-field
            label="Year"
            :variant="variant('year')"
            :message="message('year')"
          >
            <input
              v-model="year"
              class="input"
              type="number"
              inputmode="numeric"
              step="1"
            >
          </o-field>
        </div>
      </div>

      <div class="columns">
        <div class="column">
          <o-field
            label="Trim start (seconds)"
            :variant="variant('start')"
            :message="message('start')"
          >
            <input
              v-model="start"
              class="input"
              type="number"
              min="0"
              :max="song.duration"
              step="0.1"
            >
          </o-field>
        </div>
        <div class="column">
          <o-field
            label="Trim end (seconds)"
            :variant="variant('end')"
            :message="message('end')"
          >
            <input
              v-model="end"
              class="input"
              type="number"
              min="0"
              :max="song.duration"
              step="0.1"
            >
          </o-field>
        </div>
      </div>
      <audio
        class="audio-preview mb-4"
        controls
        preload="none"
        :src="audioUrl(song)"
      />

      <o-field>
        <o-switch v-model="sync">
          Sync to clients
        </o-switch>
      </o-field>

      <button
        type="submit"
        class="button is-primary"
        :class="{ 'is-loading': update.isPending.value }"
        :disabled="!online || update.isPending.value"
        :title="offlineTitle"
      >
        Save
      </button>
    </form>

    <hr>
    <h3 class="title is-5">
      Cover
    </h3>
    <p
      v-if="song.cover"
      class="has-text-grey mb-3"
    >
      Cover: {{ song.cover.size }}×{{ song.cover.size }}px
    </p>
    <CoverUpload
      ref="upload"
      :busy="setCover.isPending.value"
      :disabled="!online"
      disabled-reason="You are offline"
      label="Change cover"
      @select="changeCover"
    />
    <button
      v-if="song.cover"
      type="button"
      class="button is-danger is-light mt-3"
      :disabled="!online"
      :title="offlineTitle"
      @click="confirmCover = true"
    >
      Remove cover
    </button>

    <hr>
    <button
      type="button"
      class="button is-danger"
      :disabled="!online"
      :title="offlineTitle"
      @click="confirmSong = true"
    >
      Delete song
    </button>

    <ConfirmModal
      v-model:active="confirmCover"
      title="Remove cover?"
      confirm-label="Remove"
      :busy="deleteCover.isPending.value"
      @confirm="removeCover"
    >
      <p>The cover of "{{ song.title }}" will be removed.</p>
    </ConfirmModal>
    <ConfirmModal
      v-model:active="confirmSong"
      title="Delete song?"
      confirm-label="Delete"
      :busy="deleteSong.isPending.value"
      @confirm="removeSong"
    >
      <p>"{{ song.title }}" will be deleted permanently.</p>
    </ConfirmModal>
  </o-collapse>
</template>

<style scoped>
.audio-preview {
  width: 100%;
}
</style>
