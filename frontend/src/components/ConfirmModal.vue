<script setup lang="ts">
defineProps<{
  title: string
  confirmLabel: string
  busy?: boolean
}>()
const active = defineModel<boolean>('active', { required: true })
const emit = defineEmits<{ confirm: [] }>()
</script>

<template>
  <o-modal
    v-model:active="active"
    alert
    close-on-escape
    close-on-outside
  >
    <div class="box">
      <h2 class="title is-5">
        {{ title }}
      </h2>
      <div class="content">
        <slot />
      </div>
      <div class="buttons is-right">
        <button
          type="button"
          class="button"
          :disabled="busy"
          @click="active = false"
        >
          Cancel
        </button>
        <button
          type="button"
          class="button is-danger"
          :class="{ 'is-loading': busy }"
          @click="emit('confirm')"
        >
          {{ confirmLabel }}
        </button>
      </div>
    </div>
  </o-modal>
</template>
