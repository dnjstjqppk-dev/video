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
.gvp-btn--primary { background: var(--gvp-brand); color: #FFFFFF; border-color: var(--gvp-brand); }
.gvp-btn--primary:hover:not(:disabled) { background: var(--gvp-brand-hover); border-color: var(--gvp-brand-hover); }
.gvp-btn--secondary { background: var(--gvp-surface); color: var(--gvp-text); border-color: var(--gvp-border-strong); }
.gvp-btn--secondary:hover:not(:disabled) { background: var(--gvp-elevated); }
.gvp-btn--link { background: var(--gvp-brand); color: #FFFFFF; border-color: var(--gvp-brand); }
.gvp-btn--link:hover:not(:disabled) { background: var(--gvp-brand-hover); border-color: var(--gvp-brand-hover); }
.gvp-btn--danger { background: var(--gvp-live-soft); color: var(--gvp-live); border-color: transparent; }
.gvp-btn--danger:hover:not(:disabled) { background: var(--gvp-live); color: #fff; }
.gvp-btn--ghost { background: transparent; color: var(--gvp-text-secondary); }
.gvp-btn--ghost:hover:not(:disabled) { background: var(--gvp-elevated); color: var(--gvp-text); }
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
