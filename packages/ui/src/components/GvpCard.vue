<script setup lang="ts">
withDefaults(
  defineProps<{ padded?: boolean; hoverable?: boolean }>(),
  { padded: true, hoverable: false }
)
</script>

<template>
  <section class="gvp-card" :class="{ 'gvp-card--pad': padded, 'gvp-card--hover': hoverable }">
    <header v-if="$slots.header" class="gvp-card__header">
      <slot name="header" />
    </header>
    <div class="gvp-card__body">
      <slot />
    </div>
    <footer v-if="$slots.footer" class="gvp-card__footer">
      <slot name="footer" />
    </footer>
  </section>
</template>

<style scoped>
.gvp-card {
  background: var(--gvp-surface-1);
  border: 1px solid var(--gvp-border);
  border-radius: var(--gvp-radius-lg);
  box-shadow: var(--gvp-shadow-sm);
  overflow: hidden;
}
.gvp-card--pad > .gvp-card__body { padding: 1.25rem; }
.gvp-card--hover { transition: border-color var(--gvp-motion-base) var(--gvp-ease), box-shadow var(--gvp-motion-base) var(--gvp-ease); }
.gvp-card--hover:hover { border-color: var(--gvp-border-strong); box-shadow: var(--gvp-shadow-md); }
.gvp-card__header {
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid var(--gvp-border);
  font-weight: 700;
  font-size: 0.9375rem;
}
.gvp-card__footer {
  padding: 0.875rem 1.25rem;
  border-top: 1px solid var(--gvp-border);
}
</style>
