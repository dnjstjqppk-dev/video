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
  <span class="gvp-badge tabular-nums" :class="[`gvp-badge--${tone}`, `gvp-badge--${size}`]">
    <slot />
  </span>
</template>

<style scoped>
.gvp-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-family: var(--gvp-font-sans);
  font-weight: 500;
  border-radius: 0.375rem;
  border: 1px solid transparent;
  white-space: nowrap;
  line-height: 1.25;
}
.gvp-badge--sm { font-size: 0.75rem; padding: 0.125rem 0.5rem; }
.gvp-badge--md { font-size: 0.75rem; padding: 0.25rem 0.625rem; }
.gvp-badge--success { background: #F0FDF4; color: #15803D; border-color: #BBF7D0; }
.gvp-badge--warning { background: #FFFBEB; color: #B45309; border-color: #FDE68A; }
.gvp-badge--danger { background: #FEF2F2; color: #B91C1C; border-color: #FECACA; }
.gvp-badge--info { background: #EFF6FF; color: #1D4ED8; border-color: #BFDBFE; }
.gvp-badge--brand { background: #0F172A; color: #FFFFFF; border-color: #0F172A; }
.gvp-badge--neutral { background: #F1F5F9; color: #475569; border-color: #E2E8F0; }
</style>
