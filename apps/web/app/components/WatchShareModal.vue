<script setup lang="ts">
import { formatTimecode } from '@gvp/ui'

const props = defineProps<{ videoId: string; startSeconds: number }>()
const emit = defineEmits<{ (e: 'close'): void }>()
const includeTime = ref(props.startSeconds > 0)
const copied = ref(false)
const link = computed(() => {
  const base = import.meta.client ? location.origin : ''
  const t = includeTime.value && props.startSeconds > 0 ? `?t=${props.startSeconds}` : ''
  return `${base}/watch/${props.videoId}${t}`
})

async function copy() {
  try { await navigator.clipboard.writeText(link.value) } catch { /* clipboard denied */ }
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

function onKey(e: KeyboardEvent) { if (e.key === 'Escape') emit('close') }
onMounted(() => window.addEventListener('keydown', onKey))
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-[var(--gvp-scrim)] p-4" @mousedown.self="emit('close')">
    <div class="w-full max-w-md rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5 text-[var(--gvp-text)]" role="dialog" aria-modal="true" aria-label="공유">
      <div class="flex items-center justify-between">
        <h3 class="text-base font-bold">공유</h3>
        <button type="button" aria-label="닫기" class="rounded-full p-1.5 text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]" @click="emit('close')">✕</button>
      </div>
      <div class="mt-4 flex items-center gap-2 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-2.5">
        <span class="tabular-nums min-w-0 flex-1 truncate font-mono text-xs text-[var(--gvp-text-secondary)]">{{ link }}</span>
        <button type="button" class="shrink-0 rounded-md bg-[var(--gvp-brand)] px-3 py-1.5 text-xs font-semibold text-white hover:bg-[var(--gvp-brand-hover)]" @click="copy">
          {{ copied ? '복사됨!' : '복사' }}
        </button>
      </div>
      <label class="mt-3 flex cursor-pointer items-center gap-2 text-xs text-[var(--gvp-text)]">
        <input v-model="includeTime" type="checkbox" class="accent-[var(--gvp-brand)]" />
        <span class="tabular-nums">시작 시간 포함 ({{ formatTimecode(startSeconds) }})</span>
      </label>
    </div>
  </div>
</template>
