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
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
  border-radius: 0.5rem;
  overflow: hidden;
}
.gvp-card--pad > .gvp-card__body { padding: 1rem; }
.gvp-card--hover { transition: border-color 150ms ease, background-color 150ms ease; }
.gvp-card--hover:hover { border-color: #CBD5E1; background: #F8FAFC; }
.gvp-card__header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #E2E8F0;
  background: #F8FAFC;
  font-weight: 600;
  font-size: 0.875rem;
  color: #0F172A;
}
.gvp-card__footer {
  padding: 0.75rem 1rem;
  border-top: 1px solid #E2E8F0;
  background: #F8FAFC;
}
</style>
