<script setup lang="ts">
withDefaults(
  defineProps<{
    tone?: 'success' | 'warning' | 'danger' | 'info' | 'neutral' | 'brand'
    size?: 'sm' | 'md'
  }>(),
  { tone: 'neutral', size: 'sm' }
)
</script>

<template>
  <span class="gvp-badge" :class="[`gvp-badge--${tone}`, `gvp-badge--${size}`]">
    <slot />
  </span>
</template>

<style scoped>
.gvp-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-family: var(--gvp-font-sans);
  font-weight: 700;
  border-radius: var(--gvp-radius-pill);
  white-space: nowrap;
  line-height: 1.4;
}
.gvp-badge--sm { font-size: 0.6875rem; padding: 0.125rem 0.625rem; }
.gvp-badge--md { font-size: 0.75rem; padding: 0.25rem 0.875rem; }
.gvp-badge--success { background: var(--gvp-success-soft); color: var(--gvp-success); }
.gvp-badge--warning { background: var(--gvp-warning-soft); color: var(--gvp-warning); }
.gvp-badge--danger { background: var(--gvp-danger-soft); color: var(--gvp-danger); }
.gvp-badge--info { background: var(--gvp-info-soft); color: var(--gvp-info); }
.gvp-badge--brand { background: var(--gvp-brand-soft); color: var(--gvp-brand); }
.gvp-badge--neutral { background: rgba(255, 255, 255, 0.08); color: var(--gvp-text-secondary); }
</style>
