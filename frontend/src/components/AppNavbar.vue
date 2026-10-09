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
const nav = useTemplateRef('nav')
const more = useTemplateRef('more')
const loadingRandom = ref(false)

function close(): void {
  menuOpen.value = false
  moreOpen.value = false
}

watch(() => route.fullPath, close)

function onDocumentClick(ev: MouseEvent): void {
  if (!(ev.target instanceof Node)) {
    return
  }
  if (moreOpen.value && more.value?.contains(ev.target) !== true) {
    moreOpen.value = false
  }
  if (menuOpen.value && nav.value?.contains(ev.target) !== true) {
    menuOpen.value = false
  }
}

function onKeydown(ev: KeyboardEvent): void {
  if (ev.key === 'Escape' && (menuOpen.value || moreOpen.value)) {
    ev.stopPropagation()
    close()
  }
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  window.addEventListener('keydown', onKeydown, true)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  window.removeEventListener('keydown', onKeydown, true)
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
    ref="nav"
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
          <span class="nav-emoji">🎵</span>Songs
        </RouterLink>
        <RouterLink
          to="/artists"
          class="navbar-item"
          active-class="is-active"
        >
          <span class="nav-emoji">🎙️</span>Artists
        </RouterLink>
        <RouterLink
          to="/years"
          class="navbar-item"
          active-class="is-active"
        >
          <span class="nav-emoji">📅</span>Years
        </RouterLink>
      </div>

      <div class="navbar-end">
        <div
          v-if="downloads.data.value?.running"
          class="navbar-item"
        >
          <progress
            class="progress is-small is-primary navbar-progress"
            max="100"
            aria-label="Downloading"
          />
        </div>

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
          <div class="navbar-dropdown is-right">
            <button
              type="button"
              class="navbar-item more-item"
              :disabled="loadingRandom"
              @click="randomSong"
            >
              <span class="nav-emoji">🔀</span>Random song
            </button>
            <hr class="navbar-divider">
            <RouterLink
              to="/incomplete"
              class="navbar-item"
            >
              <span class="nav-emoji">🏷️</span>Incomplete
            </RouterLink>
            <RouterLink
              to="/unsynced"
              class="navbar-item"
            >
              <span class="nav-emoji">❌</span>Unsynced
            </RouterLink>
            <RouterLink
              to="/edits"
              class="navbar-item"
            >
              <span class="nav-emoji">🖊️</span>Recently edited
            </RouterLink>
            <RouterLink
              to="/added"
              class="navbar-item"
            >
              <span class="nav-emoji">🕙</span>Date added
            </RouterLink>
            <hr class="navbar-divider">
            <label class="navbar-item more-item">
              <span class="nav-emoji"><input
                v-model="settings.statsForNerds"
                type="checkbox"
              ></span>
              Stats for nerds
            </label>
            <a
              class="navbar-item"
              href="https://github.com/xarantolus/sensibleHub"
              target="_blank"
              rel="noopener noreferrer"
            >
              <span class="nav-emoji">⭐</span>GitHub
            </a>
          </div>
        </div>

        <div class="navbar-item navbar-search">
          <SearchBox />
        </div>

        <RouterLink
          to="/add"
          class="navbar-item"
          active-class="is-active"
        >
          <span class="nav-emoji">📝</span>Add
        </RouterLink>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.nav-emoji {
  display: inline-block;
  width: 1.5rem;
  text-align: center;
  margin-right: 0.25rem;
}

.more-item {
  background: none;
  border: 0;
  width: 100%;
  text-align: left;
  cursor: pointer;
  font: inherit;
}

.navbar-search {
  width: 17rem;
}

.navbar-search > :deep(*) {
  width: 100%;
}

.navbar-progress {
  width: 6rem;
  margin: 0;
}

.navbar-link {
  background: none;
  border: 0;
  cursor: pointer;
  font: inherit;
}

.navbar-burger {
  background: none;
  border: 0;
  cursor: pointer;
}

.navbar-link::after {
  transition: transform 0.2s ease;
}

.has-dropdown.is-active > .navbar-link::after {
  transform: rotate(135deg);
  margin-top: -0.125em;
}

@media (min-width: 1024px) {
  .navbar-end {
    align-items: center;
    gap: 0.25rem;
  }

  .navbar-end > .navbar-item,
  .navbar-end > .navbar-item.has-dropdown {
    align-items: center;
  }

  .navbar-end > .navbar-item.has-dropdown {
    padding: 0;
  }

  .has-dropdown.is-active > .navbar-dropdown {
    animation: dropdown-in 0.15s ease-out;
  }
}

@keyframes dropdown-in {
  from {
    opacity: 0;
    transform: translateY(-0.4rem);
  }
}

@media (max-width: 1023px) {
  /* Slides over the page instead of pushing it down; Bulma only toggles display. */
  .navbar-menu {
    display: block;
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    max-height: calc(100vh - var(--bulma-navbar-height));
    overflow-y: auto;
    padding: 0.5rem 0;
    box-shadow: 0 0.75rem 1.5rem rgb(0 0 0 / 0.2);
    visibility: hidden;
    opacity: 0;
    transform: translateY(-0.75rem);
    transition:
      opacity 0.2s ease,
      transform 0.2s ease,
      visibility 0s linear 0.2s;
  }

  .navbar-menu.is-active {
    visibility: visible;
    opacity: 1;
    transform: none;
    transition:
      opacity 0.2s ease,
      transform 0.2s ease;
  }

  .navbar-menu .navbar-item,
  .navbar-menu .navbar-link {
    padding: 0.6rem 1rem;
  }

  .navbar-end {
    margin-top: 0.25rem;
    padding-top: 0.25rem;
    border-top: 1px solid var(--bulma-border-weak);
  }

  .navbar-menu .navbar-item.has-dropdown {
    flex-direction: column;
    align-items: stretch;
    padding: 0;
  }

  .has-dropdown .navbar-link {
    width: 100%;
    text-align: left;
  }

  .has-dropdown .navbar-dropdown {
    box-shadow: none;
    border-top: 0;
    padding: 0;
  }

  .has-dropdown .navbar-dropdown .navbar-item {
    padding-left: 2rem;
  }

  .has-dropdown .navbar-dropdown .navbar-divider {
    margin: 0.25rem 0;
  }

  .has-dropdown:not(.is-active) > .navbar-dropdown {
    display: none;
  }

  .has-dropdown.is-active > .navbar-dropdown {
    animation: dropdown-in 0.15s ease-out;
  }

  .navbar-search {
    width: auto;
  }

  .navbar-progress {
    width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .navbar-menu,
  .navbar-link::after {
    transition: none;
  }

  .has-dropdown.is-active > .navbar-dropdown {
    animation: none;
  }
}
</style>
