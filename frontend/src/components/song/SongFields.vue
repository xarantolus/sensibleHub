<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { isApiError } from '@/api/client'
import { audioUrl, mp3Url } from '@/api/media'
import { useDeleteSong, useUpdateSong } from '@/api/queries'
import type { SongDetail, SongEditBody, SongSummary } from '@/api/schema'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { useOnline } from '@/composables/useOnline'
import { formatDuration } from '@/lib/format'
import { notify } from '@/lib/notify'
import { usePlayer } from '@/stores/player'

const props = defineProps<{ song: SongSummary; detail?: SongDetail }>()

const router = useRouter()
const player = usePlayer()
const online = useOnline()
const update = useUpdateSong()
const deleteSong = useDeleteSong()

const title = ref('')
const artist = ref('')
const album = ref('')
const year = ref('')
const start = ref('0')
const end = ref('0')
const sync = ref(false)
const errors = ref<Record<string, string>>({})
const failure = ref<string>()
const confirmSong = ref(false)

const locked = computed(() => props.detail === undefined)
const readOnly = computed(() => locked.value || !online.value)
const offlineTitle = computed(() => (online.value ? undefined : 'You are offline'))

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
  failure.value = undefined
}

watch(() => props.song.lastEdit, reset, { immediate: true })

const yearValue = computed(() => (year.value.trim() === '' ? undefined : Number(year.value)))
const startValue = computed(() => (start.value.trim() === '' ? Number.NaN : Number(start.value)))
const endValue = computed(() => (end.value.trim() === '' ? Number.NaN : Number(end.value)))

const dirty = computed(
  () =>
    title.value.trim() !== props.song.title ||
    artist.value.trim() !== (props.song.artist ?? '') ||
    album.value.trim() !== (props.song.album ?? '') ||
    yearValue.value !== props.song.year ||
    startValue.value !== props.song.playback.start ||
    endValue.value !== props.song.playback.end ||
    sync.value !== props.song.sync,
)

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
  return errors.value[field]
}

function save(): void {
  errors.value = {}
  failure.value = undefined
  if (Object.keys(localErrors.value).length > 0) {
    errors.value = localErrors.value
    return
  }
  const edit: SongEditBody = {
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
          if (Object.keys(err.fieldErrors).length === 0) {
            failure.value = err.message
          }
        } else {
          failure.value = 'Saving failed'
        }
      },
    },
  )
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
  <form
    class="song-fields"
    novalidate
    @submit.prevent="save"
  >
    <div class="buttons has-addons mb-4 is-flex-wrap-nowrap action-group">
      <button
        type="button"
        class="button is-primary"
        @click="player.playSong(song.id)"
      >
        Play
      </button>
      <button
        type="button"
        class="button"
        @click="player.playNext(song.id)"
      >
        Play next
      </button>
      <button
        type="button"
        class="button"
        @click="player.enqueue(song.id)"
      >
        Add to queue
      </button>
      <button
        type="button"
        class="button"
        @click="player.startRadio(song.id)"
      >
        Radio
      </button>
    </div>

    <div
      v-if="failure"
      class="notification is-danger is-light"
    >
      {{ failure }}
    </div>

    <div class="field">
      <div class="field has-addons mb-0">
        <div class="control field-label-control">
          <a class="button is-static">Title</a>
        </div>
        <div class="control is-expanded">
          <input
            v-model="title"
            class="input"
            :class="{ 'is-danger': message('title') }"
            type="text"
            placeholder="Title"
            :disabled="readOnly"
          >
        </div>
      </div>
      <p
        v-if="message('title')"
        class="help is-danger"
      >
        {{ message('title') }}
      </p>
    </div>

    <div class="field">
      <div class="field has-addons mb-0">
        <div class="control field-label-control">
          <RouterLink
            v-if="song.artist"
            class="button is-static link-button"
            :to="{ name: 'artist', params: { artist: song.artist } }"
          >
            Artist
          </RouterLink>
          <a
            v-else
            class="button is-static"
          >Artist</a>
        </div>
        <div class="control is-expanded">
          <input
            v-model="artist"
            class="input"
            :class="{ 'is-danger': message('artist') }"
            type="text"
            placeholder="Artist"
            :disabled="readOnly"
          >
        </div>
      </div>
      <p
        v-if="message('artist')"
        class="help is-danger"
      >
        {{ message('artist') }}
      </p>
    </div>

    <div class="field">
      <div class="field has-addons mb-0">
        <div class="control field-label-control">
          <RouterLink
            v-if="song.artist && song.album"
            class="button is-static link-button"
            :to="{ name: 'album', params: { artist: song.artist, album: song.album } }"
          >
            Album
          </RouterLink>
          <a
            v-else
            class="button is-static"
          >Album</a>
        </div>
        <div class="control is-expanded">
          <input
            v-model="album"
            class="input"
            :class="{ 'is-danger': message('album') }"
            type="text"
            placeholder="Album"
            :disabled="readOnly"
          >
        </div>
      </div>
      <p
        v-if="message('album')"
        class="help is-danger"
      >
        {{ message('album') }}
      </p>
    </div>

    <div class="field">
      <div class="field has-addons mb-0">
        <div class="control field-label-control">
          <a class="button is-static">Year</a>
        </div>
        <div class="control is-expanded">
          <input
            v-model="year"
            class="input"
            :class="{ 'is-danger': message('year') }"
            type="number"
            inputmode="numeric"
            step="1"
            placeholder="Year"
            :disabled="readOnly"
          >
        </div>
      </div>
      <p
        v-if="message('year')"
        class="help is-danger"
      >
        {{ message('year') }}
      </p>
    </div>

    <div class="field has-addons">
      <div class="control field-label-control">
        <a class="button is-static">Source</a>
      </div>
      <div class="control is-expanded">
        <span class="input static-value">
          <template v-if="detail">
            <span
              v-if="detail.imported"
              class="ellipsis"
            >{{ detail.sourceUrl }}</span>
            <a
              v-else
              class="ellipsis"
              :href="detail.sourceUrl"
              target="_blank"
              rel="noopener noreferrer"
            >{{ detail.sourceUrl }}</a>
          </template>
        </span>
      </div>
    </div>

    <div class="field has-addons">
      <div class="control field-label-control">
        <a class="button is-static">Duration</a>
      </div>
      <div class="control is-expanded">
        <span class="input static-value">
          <span class="ellipsis">{{ formatDuration(song.duration) }}</span>
        </span>
      </div>
    </div>

    <div class="trim-row">
      <div class="field">
        <div class="field has-addons mb-0">
          <div class="control field-label-control">
            <a class="button is-static">Start</a>
          </div>
          <div class="control is-expanded">
            <input
              v-model="start"
              class="input"
              :class="{ 'is-danger': message('start') }"
              type="number"
              min="0"
              :max="song.duration"
              step="0.1"
              :disabled="readOnly"
            >
          </div>
        </div>
        <p
          v-if="message('start')"
          class="help is-danger"
        >
          {{ message('start') }}
        </p>
      </div>
      <div class="field">
        <div class="field has-addons mb-0">
          <div class="control field-label-control">
            <a class="button is-static">End</a>
          </div>
          <div class="control is-expanded">
            <input
              v-model="end"
              class="input"
              :class="{ 'is-danger': message('end') }"
              type="number"
              min="0"
              :max="song.duration"
              step="0.1"
              :disabled="readOnly"
            >
          </div>
        </div>
        <p
          v-if="message('end')"
          class="help is-danger"
        >
          {{ message('end') }}
        </p>
      </div>
    </div>

    <div
      v-if="detail"
      class="field"
    >
      <audio
        class="audio-preview"
        controls
        preload="none"
        :src="audioUrl(song)"
      />
    </div>

    <div
      v-if="detail"
      class="field"
    >
      <a
        class="button"
        :href="mp3Url(song)"
        download
      >Download MP3</a>
    </div>

    <div class="field">
      <o-switch
        v-model="sync"
        :disabled="readOnly"
      >
        Enable synchronization
      </o-switch>
    </div>

    <div class="form-actions">
      <button
        type="submit"
        class="button is-primary save-song"
        :class="{ 'is-loading': update.isPending.value }"
        :disabled="readOnly || !dirty || update.isPending.value"
        :title="offlineTitle"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-danger delete-song"
        :disabled="readOnly"
        :title="offlineTitle"
        @click="confirmSong = true"
      >
        Delete
      </button>
    </div>

    <ConfirmModal
      v-model:active="confirmSong"
      title="Delete song?"
      confirm-label="Delete"
      :busy="deleteSong.isPending.value"
      @confirm="removeSong"
    >
      <p>"{{ song.title }}" will be deleted permanently.</p>
    </ConfirmModal>
  </form>
</template>

<style scoped>
.song-fields {
  min-width: 0;
}

.field-label-control {
  flex: none;
}

.field-label-control .button {
  width: 6rem;
  justify-content: flex-start;
}

.link-button.is-static {
  pointer-events: auto;
  cursor: pointer;
  color: var(--bulma-link);
}

.control.is-expanded {
  min-width: 0;
}

.control.is-expanded .input {
  width: 100%;
}

.static-value {
  display: flex;
  align-items: center;
  background-color: var(--bulma-scheme-main-bis);
  min-width: 0;
}

.ellipsis {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trim-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  column-gap: 1rem;
}

.trim-row > .field {
  min-width: 0;
}

.audio-preview {
  display: block;
  width: 100%;
}

.form-actions {
  display: flex;
  justify-content: center;
  gap: 2.5%;
}

.form-actions .save-song {
  width: 70%;
}

.form-actions .delete-song {
  width: 27.5%;
}

@media screen and (max-width: 768px) {
  .action-group .button {
    flex: 1 1 auto;
    padding-inline: 0.5rem;
    font-size: var(--bulma-size-small);
  }

  .trim-row {
    grid-template-columns: 1fr;
  }
}
</style>
