<script setup lang="ts">
import { computed } from 'vue'

import { useListing, useSongIndex, type ListingKind } from '@/api/queries'
import QueryView from '@/components/QueryView.vue'
import SongGrid from '@/components/SongGrid.vue'

const props = defineProps<{ kind: ListingKind; title: string }>()

const listing = useListing(() => props.kind)
const { resolve } = useSongIndex()

const jumpBarMinGroups = 4

const groups = computed(() =>
  (listing.data.value ?? []).map((group, i) => ({ ...group, anchor: `group-${String(i)}`, songs: resolve(group.songIds) })),
)

function jump(anchor: string): void {
  document.getElementById(anchor)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
</script>

<template>
  <h1 class="page-title">
    {{ title }}
  </h1>
  <QueryView
    :data="listing.data.value"
    :error="listing.error.value"
    :is-pending="listing.isPending.value"
    :refetch="listing.refetch"
  >
    <template #default="{ data }">
      <p
        v-if="data.length === 0"
        class="notification has-text-centered"
      >
        Nothing here.
      </p>
      <template v-else>
        <nav
          v-if="data.length >= jumpBarMinGroups"
          class="jump-bar"
          aria-label="Jump to group"
        >
          <button
            v-for="g in groups"
            :key="g.anchor"
            type="button"
            class="button is-small is-light"
            @click="jump(g.anchor)"
          >
            {{ g.title }}
          </button>
        </nav>
        <section
          v-for="g in groups"
          :id="g.anchor"
          :key="g.anchor"
          class="listing-group"
        >
          <h2 class="title is-4 mb-1">
            <RouterLink
              v-if="g.link"
              :to="g.link"
            >
              {{ g.title }}
            </RouterLink>
            <template v-else>
              {{ g.title }}
            </template>
          </h2>
          <p
            v-if="g.description"
            class="has-text-grey mb-3"
          >
            {{ g.description }}
          </p>
          <SongGrid :songs="g.songs" />
        </section>
      </template>
    </template>
  </QueryView>
</template>

<style scoped>
.jump-bar {
  position: sticky;
  top: var(--bulma-navbar-height);
  z-index: 10;
  display: flex;
  gap: 0.375rem;
  overflow-x: auto;
  padding: 0.5rem 0;
  margin-bottom: 1rem;
  background: var(--bulma-scheme-main);
  scrollbar-width: thin;
}

.jump-bar .button {
  flex: none;
}

.listing-group {
  margin-bottom: 2rem;
  scroll-margin-top: calc(var(--bulma-navbar-height) + 3.5rem);
}
</style>
