<script setup lang="ts">
import { onErrorCaptured, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { errorMessage } from '@/lib/notify'

const error = ref<unknown>(null)
const route = useRoute()

onErrorCaptured((err) => {
  console.error(err)
  error.value = err
  return false
})

watch(
  () => route.fullPath,
  () => {
    error.value = null
  },
)
</script>

<template>
  <div
    v-if="error"
    class="notification is-danger"
    role="alert"
  >
    <p class="has-text-weight-semibold">
      This page crashed: {{ errorMessage(error) }}
    </p>
    <button
      type="button"
      class="button is-small mt-3"
      @click="error = null"
    >
      Try again
    </button>
  </div>
  <slot v-else />
</template>
