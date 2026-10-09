<script setup lang="ts">
import { computed, ref } from 'vue'

import { isApiError } from '@/api/client'
import { useAbortDownload, useAnalysisStatus, useDownloads, useEnqueueDownload, useSongIndex } from '@/api/queries'
import QueryView from '@/components/QueryView.vue'
import SongCard from '@/components/SongCard.vue'
import { useSettings } from '@/composables/useSettings'
import { useOnline } from '@/composables/useOnline'
import { describeFailure } from '@/lib/downloadErrors'
import { notify } from '@/lib/notify'

const downloads = useDownloads()
const analysis = useAnalysisStatus()
const settings = useSettings()
const enqueue = useEnqueueDownload()
const abort = useAbortDownload()
const { songs } = useSongIndex()
const online = useOnline()

const query = ref('')
const inlineError = ref<{ message: string; songId?: string }>()

const newest = computed(() =>
  (songs.data.value ?? []).reduce<(typeof songs.data.value & object)[number] | undefined>(
    (best, s) => (best === undefined || s.added > best.added ? s : best),
    undefined,
  ),
)

function submit(): void {
  const q = query.value.trim()
  if (q === '') {
    return
  }
  inlineError.value = undefined
  enqueue.mutate(q, {
    onSuccess: () => {
      query.value = ''
      notify('Added to the download queue', 'success')
    },
    onError: (err) => {
      if (isApiError(err, 'already_downloaded')) {
        const songId = err.problem?.songId
        inlineError.value =
          songId === undefined
            ? { message: 'This song has already been downloaded.' }
            : { message: 'This song has already been downloaded.', songId }
      } else if (isApiError(err, 'queue_full')) {
        inlineError.value = { message: 'Download queue is full' }
      }
    },
  })
}
</script>

<template>
  <div>
    <div class="add-page">
      <h1 class="title">
        Add songs
      </h1>

      <form
        novalidate
        @submit.prevent="submit"
      >
        <o-field label="Link or search term">
          <input
            v-model="query"
            class="input"
            type="text"
            autocomplete="off"
            autocapitalize="off"
            spellcheck="false"
            required
          >
        </o-field>
        <p class="help mb-3">
          Any site supported by yt-dlp works. Search terms are looked up on YouTube Music.
        </p>
        <p
          v-if="inlineError"
          class="notification is-warning is-light"
          role="alert"
        >
          {{ inlineError.message }}
          <RouterLink
            v-if="inlineError.songId"
            :to="{ name: 'song', params: { id: inlineError.songId } }"
          >
            Open the existing song
          </RouterLink>
        </p>
        <button
          type="submit"
          class="button is-primary"
          :class="{ 'is-loading': enqueue.isPending.value }"
          :disabled="!online || enqueue.isPending.value || query.trim() === ''"
          :title="online ? undefined : 'You are offline'"
        >
          Add
        </button>
      </form>

      <div
        v-if="settings.statsForNerds && analysis.data.value?.running"
        class="mt-4"
      >
        <p class="is-size-7 has-text-grey">
          Analysing songs: {{ analysis.data.value.done }} / {{ analysis.data.value.total }}
        </p>
        <progress
          class="progress is-small is-primary"
          :value="analysis.data.value.done"
          :max="analysis.data.value.total"
        />
      </div>

      <QueryView
        :data="downloads.data.value"
        :error="downloads.error.value"
        :is-pending="downloads.isPending.value"
        :refetch="downloads.refetch"
      >
        <template #default="{ data }">
          <div
            v-if="data.running"
            class="box mt-5"
          >
            <p class="has-text-weight-semibold">
              Downloading
            </p>
            <p class="download-url">
              {{ data.url }}
            </p>
            <p
              v-if="data.queued > 0"
              class="has-text-grey"
            >
              {{ data.queued }} more queued
            </p>
            <progress
              class="progress is-primary mt-3"
              max="100"
            />
            <button
              type="button"
              class="button is-danger is-light is-small"
              :class="{ 'is-loading': abort.isPending.value }"
              :disabled="!online || abort.isPending.value"
              :title="online ? undefined : 'You are offline'"
              @click="abort.mutate()"
            >
              Abort
            </button>
          </div>

          <div
            v-if="data.lastError"
            class="notification is-danger is-light mt-5"
            role="alert"
          >
            <p class="has-text-weight-semibold">
              {{ describeFailure(data.lastError).message }}
            </p>
            <p
              v-if="data.lastError.message"
              class="mt-1"
            >
              {{ data.lastError.message }}
            </p>
            <RouterLink
              v-if="describeFailure(data.lastError).songId"
              :to="{ name: 'song', params: { id: describeFailure(data.lastError).songId } }"
            >
              Open the existing song
            </RouterLink>
            <o-collapse
              v-if="data.lastError.output"
              :open="false"
              label="Downloader output"
              class="mt-3"
            >
              <pre class="downloader-output">{{ data.lastError.output }}</pre>
            </o-collapse>
          </div>
        </template>
      </QueryView>

      <section
        v-if="newest"
        class="mt-6"
      >
        <h2 class="title is-4">
          Newest song
        </h2>
        <div class="newest">
          <SongCard :song="newest" />
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.add-page {
  max-width: 40rem;
}

.download-url {
  overflow-wrap: anywhere;
}

.downloader-output {
  max-height: 20rem;
  overflow: auto;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.newest {
  max-width: 12rem;
}
</style>
