<script setup lang="ts">
withDefaults(
  defineProps<{ tone?: 'info' | 'warning' | 'success' | 'danger'; dismissible?: boolean }>(),
  { tone: 'info', dismissible: false }
)
const emit = defineEmits<{ (e: 'dismiss'): void }>()
const visible = defineModel<boolean>('visible', { default: true })
function dismiss() {
  visible.value = false
  emit('dismiss')
}
</script>

<template>
  <div v-if="visible" class="gvp-banner" :class="`gvp-banner--${tone}`" role="status">
    <span class="gvp-banner__dot" aria-hidden="true" />
    <div class="gvp-banner__content">
      <slot />
    </div>
    <button v-if="dismissible" type="button" class="gvp-banner__close" aria-label="닫기" @click="dismiss">✕</button>
  </div>
</template>

<style scoped>
.gvp-banner {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  font-size: 0.8125rem;
  border-radius: var(--gvp-radius-md);
  padding: 0.75rem 1rem;
  border: 1px solid;
}
.gvp-banner__dot { width: 0.5rem; height: 0.5rem; border-radius: 50%; margin-top: 0.25rem; flex-shrink: 0; background: currentColor; }
.gvp-banner__content { flex: 1; line-height: 1.5; }
.gvp-banner__content :deep(a) { color: inherit; font-weight: 700; text-decoration: underline; }
.gvp-banner--info { background: var(--gvp-info-soft); border-color: rgba(62, 166, 255, 0.35); color: var(--gvp-info); }
.gvp-banner--info .gvp-banner__content { color: var(--gvp-text-secondary); }
.gvp-banner--warning { background: var(--gvp-warning-soft); border-color: rgba(245, 158, 11, 0.35); color: var(--gvp-warning); }
.gvp-banner--warning .gvp-banner__content { color: var(--gvp-text-secondary); }
.gvp-banner--success { background: var(--gvp-success-soft); border-color: rgba(34, 197, 94, 0.35); color: var(--gvp-success); }
.gvp-banner--success .gvp-banner__content { color: var(--gvp-text-secondary); }
.gvp-banner--danger { background: var(--gvp-danger-soft); border-color: rgba(239, 68, 68, 0.35); color: var(--gvp-danger); }
.gvp-banner--danger .gvp-banner__content { color: var(--gvp-text-secondary); }
.gvp-banner__close {
  background: transparent; border: none; color: inherit;
  cursor: pointer; padding: 0.125rem 0.25rem; font-size: 0.75rem;
}
</style>
