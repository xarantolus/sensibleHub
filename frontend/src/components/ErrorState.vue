<script setup lang="ts">
import { computed } from 'vue'

import { isApiError } from '@/api/client'
import { errorMessage } from '@/lib/notify'

const props = defineProps<{ error: unknown; retry?: () => unknown }>()

const offline = computed(() => isApiError(props.error, 'network'))
const notFound = computed(() => isApiError(props.error, 'not_found'))
</script>

<template>
  <div
    class="notification"
    :class="notFound ? 'is-warning' : 'is-danger'"
    role="alert"
  >
    <p
      v-if="notFound"
      class="has-text-weight-semibold"
    >
      Not found
    </p>
    <p
      v-else-if="offline"
      class="has-text-weight-semibold"
    >
      Server unreachable
    </p>
    <p
      v-else
      class="has-text-weight-semibold"
    >
      Loading failed: {{ errorMessage(error) }}
    </p>
    <div class="buttons mt-3">
      <button
        v-if="retry && !notFound"
        type="button"
        class="button is-small"
        @click="retry()"
      >
        Try again
      </button>
      <RouterLink
        v-if="notFound"
        to="/"
        class="button is-small"
      >
        Go home
      </RouterLink>
    </div>
  </div>
</template>
