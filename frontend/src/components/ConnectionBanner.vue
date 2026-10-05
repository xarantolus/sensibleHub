<script setup lang="ts">
import { useOnline } from '@/composables/useOnline'
import { connectionState } from '@/api/events'

const online = useOnline()
</script>

<template>
  <div
    v-if="!online"
    class="connection-banner has-background-warning-light has-text-warning-dark"
    role="status"
  >
    You're offline. Cached songs keep playing; editing is unavailable.
  </div>
  <div
    v-else-if="connectionState === 'reconnecting'"
    class="connection-banner has-background-warning-light has-text-warning-dark"
    role="status"
  >
    Lost connection to the server, reconnecting…
  </div>
</template>

<style scoped>
.connection-banner {
  padding: 0.4rem 1rem;
  text-align: center;
  font-size: 0.875rem;
}
</style>
