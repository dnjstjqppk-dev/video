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
    <button v-if="dismissible" type="button" class="gvp-banner__close" aria-label="닫기" @click="dismiss">닫기</button>
  </div>
</template>

<style scoped>
.gvp-banner {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  font-size: 0.8125rem;
  border-radius: 0.375rem;
  padding: 0.625rem 0.875rem;
  border: 1px solid;
}
.gvp-banner__dot { width: 0.375rem; height: 0.375rem; border-radius: 50%; margin-top: 0.375rem; flex-shrink: 0; background: currentColor; }
.gvp-banner__content { flex: 1; line-height: 1.5; }
.gvp-banner__content :deep(a) { color: inherit; font-weight: 600; text-decoration: underline; }
.gvp-banner--info { background: #EFF6FF; border-color: #BFDBFE; color: #1D4ED8; }
.gvp-banner--info .gvp-banner__content { color: #1E3A8A; }
.gvp-banner--warning { background: #FFFBEB; border-color: #FDE68A; color: #B45309; }
.gvp-banner--warning .gvp-banner__content { color: #78350F; }
.gvp-banner--success { background: #F0FDF4; border-color: #BBF7D0; color: #15803D; }
.gvp-banner--success .gvp-banner__content { color: #14532D; }
.gvp-banner--danger { background: #FEF2F2; border-color: #FECACA; color: #B91C1C; }
.gvp-banner--danger .gvp-banner__content { color: #7F1D1D; }
.gvp-banner__close {
  background: transparent; border: none; color: inherit;
  cursor: pointer; padding: 0.125rem 0.375rem; font-size: 0.75rem; font-weight: 500;
}
</style>
