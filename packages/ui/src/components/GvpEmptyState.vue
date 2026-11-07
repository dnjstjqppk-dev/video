<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  actionLabel?: string
}>()
defineEmits<{ (e: 'action'): void }>()
</script>

<template>
  <div class="gvp-empty" role="status">
    <div class="gvp-empty__glyph" aria-hidden="true">
      <slot name="icon">◌</slot>
    </div>
    <p class="gvp-empty__title">{{ title }}</p>
    <p v-if="description" class="gvp-empty__desc">{{ description }}</p>
    <button
      v-if="actionLabel"
      type="button"
      class="gvp-empty__action"
      @click="$emit('action')"
    >
      {{ actionLabel }}
    </button>
  </div>
</template>

<style scoped>
.gvp-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.5rem;
  padding: 3rem 1.5rem;
  background: var(--gvp-surface-1);
  border: 1px dashed var(--gvp-border-strong);
  border-radius: var(--gvp-radius-lg);
}
.gvp-empty__glyph {
  font-size: 2rem;
  color: var(--gvp-text-faint);
  line-height: 1;
}
.gvp-empty__title { font-weight: 700; font-size: 1rem; color: var(--gvp-text); margin: 0; }
.gvp-empty__desc { font-size: 0.8125rem; color: var(--gvp-text-muted); margin: 0; max-width: 28rem; }
.gvp-empty__action {
  margin-top: 0.75rem;
  font-family: var(--gvp-font-sans);
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--gvp-brand-contrast);
  background: var(--gvp-brand);
  border: 1px solid transparent;
  border-radius: var(--gvp-radius-pill);
  padding: 0.5rem 1.25rem;
  cursor: pointer;
}
.gvp-empty__action:hover { background: var(--gvp-brand-hover); }
</style>
