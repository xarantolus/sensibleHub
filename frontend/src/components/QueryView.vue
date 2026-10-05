<script setup lang="ts" generic="T extends object">
import ErrorState from './ErrorState.vue'

defineProps<{
  data: T | undefined
  error: unknown
  isPending: boolean
  refetch: () => unknown
}>()

defineSlots<{
  default: (props: { data: T }) => unknown
  loading?: () => unknown
}>()
</script>

<template>
  <slot
    v-if="data !== undefined"
    :data="data"
  />
  <ErrorState
    v-else-if="error"
    :error="error"
    :retry="refetch"
  />
  <template v-else-if="isPending">
    <slot name="loading">
      <div class="is-flex is-justify-content-center py-6">
        <span
          class="loader"
          aria-label="Loading"
        />
      </div>
    </slot>
  </template>
</template>
