<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    labelledBy?: string
    maxWidth?: string
  }>(),
  { labelledBy: 'gvp-modal-title', maxWidth: '28rem' }
)
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void }>()

const dialogRef = ref<HTMLDivElement | null>(null)
let lastFocused: HTMLElement | null = null

function close() {
  emit('update:modelValue', false)
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.stopPropagation()
    close()
    return
  }
  if (e.key !== 'Tab' || !dialogRef.value) return
  const focusables = dialogRef.value.querySelectorAll<HTMLElement>(
    'a[href], button:not(:disabled), input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex]:not([tabindex="-1"])'
  )
  if (focusables.length === 0) return
  const first = focusables[0]
  const last = focusables[focusables.length - 1]
  if (e.shiftKey && document.activeElement === first) {
    e.preventDefault()
    last?.focus()
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault()
    first?.focus()
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      lastFocused = document.activeElement as HTMLElement | null
      document.addEventListener('keydown', onKeydown, true)
      requestAnimationFrame(() => {
        dialogRef.value?.querySelector<HTMLElement>('button, input, select, textarea, a[href]')?.focus()
      })
    } else {
      document.removeEventListener('keydown', onKeydown, true)
      lastFocused?.focus?.()
    }
  }
)

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown, true)
})

onMounted(() => {
  if (props.modelValue) document.addEventListener('keydown', onKeydown, true)
})
</script>

<template>
  <Teleport to="body">
    <div v-if="modelValue" class="gvp-modal__backdrop" @mousedown.self="close">
      <div
        ref="dialogRef"
        class="gvp-modal"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="labelledBy"
        :style="{ maxWidth }"
      >
        <div class="gvp-modal__head">
          <h2 :id="labelledBy" class="gvp-modal__title">{{ title }}</h2>
          <button type="button" class="gvp-modal__close" aria-label="닫기" @click="close">닫기</button>
        </div>
        <div class="gvp-modal__body">
          <slot />
        </div>
        <div v-if="$slots.footer" class="gvp-modal__footer">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.gvp-modal__backdrop {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: rgba(15, 23, 42, 0.45);
}
.gvp-modal {
  width: 100%;
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
  border-radius: 0.5rem;
  box-shadow: 0 4px 6px -1px rgba(15, 23, 42, 0.08);
  overflow: hidden;
}
.gvp-modal__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid #E2E8F0;
  background: #F8FAFC;
}
.gvp-modal__title {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #0F172A;
}
.gvp-modal__close {
  background: transparent;
  border: none;
  color: #475569;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 0.25rem;
}
.gvp-modal__close:hover { background: #E2E8F0; color: #0F172A; }
.gvp-modal__body { padding: 1.25rem; color: #0F172A; font-size: 0.875rem; }
.gvp-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  border-top: 1px solid #E2E8F0;
  background: #F8FAFC;
}
</style>
