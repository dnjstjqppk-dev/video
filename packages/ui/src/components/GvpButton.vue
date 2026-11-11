<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'link' | 'danger' | 'ghost'
    size?: 'sm' | 'md' | 'lg'
    loading?: boolean
    block?: boolean
    type?: 'button' | 'submit' | 'reset'
  }>(),
  { variant: 'primary', size: 'md', loading: false, block: false, type: 'button' }
)
</script>

<template>
  <button
    :type="type"
    class="gvp-btn"
    :class="[`gvp-btn--${variant}`, `gvp-btn--${size}`, { 'gvp-btn--block': block }]"
    :disabled="loading || ($attrs.disabled as boolean | undefined)"
  >
    <span v-if="loading" class="gvp-btn__spinner" aria-hidden="true" />
    <slot />
  </button>
</template>

<style scoped>
.gvp-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-family: var(--gvp-font-sans);
  font-weight: 500;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 150ms ease, color 150ms ease, border-color 150ms ease;
  white-space: nowrap;
}
.gvp-btn--sm { font-size: 0.75rem; height: 2rem; padding: 0 0.75rem; }
.gvp-btn--md { font-size: 0.875rem; height: 2.25rem; padding: 0 1rem; }
.gvp-btn--lg { font-size: 0.875rem; height: 2.5rem; padding: 0 1.25rem; }
.gvp-btn--block { width: 100%; }
.gvp-btn--primary { background: #0F172A; color: #FFFFFF; border-color: #0F172A; }
.gvp-btn--primary:hover:not(:disabled) { background: #1E293B; border-color: #1E293B; }
.gvp-btn--secondary { background: #FFFFFF; color: #334155; border-color: #CBD5E1; }
.gvp-btn--secondary:hover:not(:disabled) { background: #F8FAFC; color: #0F172A; }
.gvp-btn--link { background: #2563EB; color: #FFFFFF; border-color: #2563EB; }
.gvp-btn--link:hover:not(:disabled) { background: #1D4ED8; border-color: #1D4ED8; }
.gvp-btn--danger { background: #FEF2F2; color: #B91C1C; border-color: #FECACA; }
.gvp-btn--danger:hover:not(:disabled) { background: #FEE2E2; }
.gvp-btn--ghost { background: transparent; color: #475569; }
.gvp-btn--ghost:hover:not(:disabled) { background: #F1F5F9; color: #0F172A; }
.gvp-btn:disabled { cursor: not-allowed; opacity: 0.5; }
.gvp-btn__spinner {
  width: 0.875rem; height: 0.875rem;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: gvp-spin 0.7s linear infinite;
}
@keyframes gvp-spin { to { transform: rotate(360deg); } }
</style>
