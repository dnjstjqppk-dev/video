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
  font-weight: 700;
  border: 1px solid transparent;
  border-radius: var(--gvp-radius-pill);
  cursor: pointer;
  transition: background-color var(--gvp-motion-base) var(--gvp-ease), color var(--gvp-motion-base) var(--gvp-ease), border-color var(--gvp-motion-base) var(--gvp-ease), opacity var(--gvp-motion-fast) var(--gvp-ease);
  white-space: nowrap;
}
.gvp-btn--sm { font-size: 0.75rem; padding: 0.375rem 0.875rem; }
.gvp-btn--md { font-size: 0.875rem; padding: 0.625rem 1.5rem; }
.gvp-btn--lg { font-size: 1rem; padding: 0.75rem 2rem; }
.gvp-btn--block { width: 100%; }
.gvp-btn--primary { background: var(--gvp-brand); color: var(--gvp-brand-contrast); }
.gvp-btn--primary:hover:not(:disabled) { background: var(--gvp-brand-hover); }
.gvp-btn--primary:active:not(:disabled) { background: var(--gvp-brand-active); }
.gvp-btn--secondary { background: var(--gvp-surface-3); color: var(--gvp-text); }
.gvp-btn--secondary:hover:not(:disabled) { background: #333336; }
.gvp-btn--link { background: var(--gvp-link); color: #04121f; }
.gvp-btn--link:hover:not(:disabled) { background: var(--gvp-link-hover); }
.gvp-btn--danger { background: transparent; color: var(--gvp-danger); border-color: var(--gvp-danger); }
.gvp-btn--danger:hover:not(:disabled) { background: var(--gvp-danger-soft); }
.gvp-btn--ghost { background: transparent; color: var(--gvp-text-secondary); }
.gvp-btn--ghost:hover:not(:disabled) { background: rgba(255, 255, 255, 0.08); color: var(--gvp-text); }
.gvp-btn:disabled { cursor: not-allowed; opacity: 0.55; }
.gvp-btn__spinner {
  width: 1em; height: 1em;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: gvp-spin 0.7s linear infinite;
}
@keyframes gvp-spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) {
  .gvp-btn__spinner { animation: none; border-top-color: currentColor; opacity: 0.6; }
}
</style>
