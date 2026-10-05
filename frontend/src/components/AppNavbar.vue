<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { fetchRandomSong, useDownloads } from '@/api/queries'
import { useSettings } from '@/composables/useSettings'
import { notifyError } from '@/lib/notify'

import SearchBox from './SearchBox.vue'

const router = useRouter()
const route = useRoute()
const downloads = useDownloads()
const settings = useSettings()

const menuOpen = ref(false)
const moreOpen = ref(false)
const more = useTemplateRef('more')
const loadingRandom = ref(false)

function close(): void {
  menuOpen.value = false
  moreOpen.value = false
}

watch(() => route.fullPath, close)

function onDocumentClick(ev: MouseEvent): void {
  if (moreOpen.value && ev.target instanceof Node && more.value?.contains(ev.target) !== true) {
    moreOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
})

async function randomSong(): Promise<void> {
  loadingRandom.value = true
  try {
    const song = await fetchRandomSong()
    close()
    await router.push({ name: 'song', params: { id: song.id } })
  } catch (err) {
    notifyError(err, 'Could not pick a random song')
  } finally {
    loadingRandom.value = false
  }
}
</script>

<template>
  <nav
    class="navbar app-navbar"
    aria-label="main navigation"
  >
    <div class="navbar-brand">
      <RouterLink
        to="/"
        class="navbar-item"
        aria-label="sensibleHub home"
      >
        <img
          src="/fav/safari-pinned-tab.svg"
          alt=""
          width="28"
          height="28"
        >
      </RouterLink>
      <button
        type="button"
        class="navbar-burger"
        :class="{ 'is-active': menuOpen }"
        aria-label="menu"
        :aria-expanded="menuOpen"
        @click="menuOpen = !menuOpen"
      >
        <span aria-hidden="true" />
        <span aria-hidden="true" />
        <span aria-hidden="true" />
        <span aria-hidden="true" />
      </button>
    </div>

    <div
      class="navbar-menu"
      :class="{ 'is-active': menuOpen }"
    >
      <div class="navbar-start">
        <RouterLink
          to="/songs"
          class="navbar-item"
          active-class="is-active"
        >
          Songs
        </RouterLink>
        <RouterLink
          to="/artists"
          class="navbar-item"
          active-class="is-active"
        >
          Artists
        </RouterLink>
        <RouterLink
          to="/years"
          class="navbar-item"
          active-class="is-active"
        >
          Years
        </RouterLink>
        <div
          ref="more"
          class="navbar-item has-dropdown"
          :class="{ 'is-active': moreOpen }"
        >
          <button
            type="button"
            class="navbar-link"
            :aria-expanded="moreOpen"
            @click="moreOpen = !moreOpen"
          >
            More
          </button>
          <div class="navbar-dropdown">
            <button
              type="button"
              class="navbar-item more-item"
              :disabled="loadingRandom"
              @click="randomSong"
            >
              Random song
            </button>
            <hr class="navbar-divider">
            <RouterLink
              to="/incomplete"
              class="navbar-item"
            >
              Incomplete
            </RouterLink>
            <RouterLink
              to="/unsynced"
              class="navbar-item"
            >
              Unsynced
            </RouterLink>
            <RouterLink
              to="/edits"
              class="navbar-item"
            >
              Recently edited
            </RouterLink>
            <RouterLink
              to="/added"
              class="navbar-item"
            >
              Date added
            </RouterLink>
            <hr class="navbar-divider">
            <label class="navbar-item more-item">
              <input
                v-model="settings.statsForNerds"
                type="checkbox"
                class="mr-2"
              >
              Stats for nerds
            </label>
            <a
              class="navbar-item"
              href="https://github.com/xarantolus/sensibleHub"
              target="_blank"
              rel="noopener noreferrer"
            >
              GitHub
            </a>
          </div>
        </div>
      </div>

      <div class="navbar-end">
        <div class="navbar-item navbar-search">
          <SearchBox />
        </div>
        <div class="navbar-item">
          <RouterLink
            to="/add"
            class="button is-primary is-fullwidth-mobile"
          >
            Add
          </RouterLink>
        </div>
      </div>
    </div>

    <progress
      v-if="downloads.data.value?.running"
      class="progress is-small is-primary navbar-progress"
      max="100"
    />
  </nav>
</template>

<style scoped>
.more-item {
  background: none;
  border: 0;
  width: 100%;
  text-align: left;
  cursor: pointer;
}

.navbar-search {
  min-width: 16rem;
}

.navbar-progress {
  position: absolute;
  left: 0;
  right: 0;
  bottom: -0.5rem;
  margin: 0;
  height: 0.5rem;
  border-radius: 0;
}

.navbar-link {
  background: none;
  border: 0;
  cursor: pointer;
}

.navbar-burger {
  background: none;
  border: 0;
  cursor: pointer;
}

@media (max-width: 1023px) {
  .navbar-menu {
    max-height: calc(100vh - var(--bulma-navbar-height));
    overflow-y: auto;
  }

  .navbar-search {
    min-width: 0;
  }

  .is-fullwidth-mobile {
    width: 100%;
  }
}
</style>
