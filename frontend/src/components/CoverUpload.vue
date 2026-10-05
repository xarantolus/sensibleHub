<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'

defineProps<{ busy?: boolean; disabled?: boolean; disabledReason?: string; label?: string }>()
const emit = defineEmits<{ select: [file: File] }>()

const accept = 'image/png,image/jpeg,image/webp,image/gif'
const preview = ref<string>()
const file = ref<File>()

function revoke(): void {
  if (preview.value !== undefined) {
    URL.revokeObjectURL(preview.value)
    preview.value = undefined
  }
}

function onChange(event: Event): void {
  revoke()
  const chosen = (event.target as HTMLInputElement).files?.[0]
  file.value = chosen
  if (chosen !== undefined) {
    preview.value = URL.createObjectURL(chosen)
  }
}

function submit(): void {
  if (file.value !== undefined) {
    emit('select', file.value)
  }
}

function clear(): void {
  revoke()
  file.value = undefined
}

defineExpose({ clear })

onBeforeUnmount(revoke)
</script>

<template>
  <div class="cover-upload">
    <div class="file">
      <label class="file-label">
        <input
          class="file-input"
          type="file"
          :accept="accept"
          :disabled="busy || disabled"
          @change="onChange"
        >
        <span class="file-cta">
          <span class="file-label">{{ label ?? 'Choose image' }}</span>
        </span>
        <span
          v-if="file"
          class="file-name"
        >{{ file.name }}</span>
      </label>
    </div>
    <div
      v-if="preview"
      class="cover-upload-preview mt-3"
    >
      <figure class="image is-128x128">
        <img
          :src="preview"
          alt="Selected cover preview"
          class="cover-upload-img"
        >
      </figure>
      <button
        type="button"
        class="button is-primary mt-3"
        :class="{ 'is-loading': busy }"
        :disabled="busy || disabled"
        :title="disabled ? disabledReason : undefined"
        @click="submit"
      >
        Upload
      </button>
    </div>
  </div>
</template>

<style scoped>
.cover-upload-img {
  object-fit: cover;
  width: 100%;
  height: 100%;
  border-radius: var(--bulma-radius);
}
</style>
