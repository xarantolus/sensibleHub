<script setup lang="ts">
import AppNavbar from '@/components/AppNavbar.vue'
import ConnectionBanner from '@/components/ConnectionBanner.vue'
import ErrorBoundary from '@/components/ErrorBoundary.vue'
import { useShortcuts } from '@/composables/useShortcuts'

useShortcuts()
</script>

<template>
  <AppNavbar />
  <ConnectionBanner />
  <main class="section app-main">
    <div class="container">
      <ErrorBoundary>
        <RouterView />
      </ErrorBoundary>
    </div>
  </main>
</template>

<style scoped>
.app-main {
  padding-top: 1.5rem;
  padding-bottom: calc(6rem + env(safe-area-inset-bottom));
}

@media (max-width: 768px) {
  .app-main {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>
