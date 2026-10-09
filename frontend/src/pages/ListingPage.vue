<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'

import { useListing, useSongIndex, type ListingKind } from '@/api/queries'
import QueryView from '@/components/QueryView.vue'
import VirtualSongGrid, { type GridGroup } from '@/components/VirtualSongGrid.vue'
import { jumpTargets } from '@/lib/grid'

const props = defineProps<{ kind: ListingKind; title: string }>()

const listing = useListing(() => props.kind)
const { resolve } = useSongIndex()
const grid = useTemplateRef<InstanceType<typeof VirtualSongGrid>>('grid')

const groups = computed<GridGroup[]>(() =>
  (listing.data.value ?? []).map((group, i) => ({
    key: String(i),
    title: group.title,
    ...(group.link === undefined ? {} : { link: group.link }),
    ...(group.description === undefined ? {} : { description: group.description }),
    songs: resolve(group.songIds),
  })),
)

const targets = computed(() => jumpTargets(groups.value.map((g) => ({ key: g.key, title: g.title ?? '' }))))
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
        Nothing here
      </p>
      <template v-else>
        <nav
          v-if="targets.length > 3"
          class="jump-bar"
          aria-label="Jump to"
        >
          <button
            v-for="t in targets"
            :key="t.key"
            type="button"
            class="jump"
            @click="grid?.scrollToGroup(t.key)"
          >
            {{ t.label }}
          </button>
        </nav>
        <VirtualSongGrid
          ref="grid"
          :groups="groups"
        />
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
  flex-wrap: nowrap;
  gap: 0.25rem;
  overflow-x: auto;
  padding: 0.5rem 0;
  margin-bottom: 0.5rem;
  background: var(--bulma-scheme-main);
  border-bottom: 1px solid var(--bulma-border-weak);
  scrollbar-width: none;
}

.jump {
  flex: none;
  min-width: 2rem;
  padding: 0.2rem 0.55rem;
  border: 0;
  border-radius: var(--bulma-radius);
  background: transparent;
  color: var(--bulma-text);
  font: inherit;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
}

.jump:hover {
  background: var(--bulma-scheme-main-ter);
  color: var(--bulma-link-text);
}
</style>
