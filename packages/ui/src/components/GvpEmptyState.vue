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
  gap: 0.375rem;
  padding: 2rem 1.5rem;
  background: #F8FAFC;
  border: 1px dashed #E2E8F0;
  border-radius: 0.5rem;
}
.gvp-empty__title { font-weight: 600; font-size: 0.875rem; color: #0F172A; margin: 0; }
.gvp-empty__desc { font-size: 0.8125rem; color: #64748B; margin: 0; max-width: 28rem; }
.gvp-empty__action {
  margin-top: 0.5rem;
  font-family: var(--gvp-font-sans);
  font-size: 0.8125rem;
  font-weight: 500;
  color: #FFFFFF;
  background: #0F172A;
  border: 1px solid #0F172A;
  border-radius: 0.375rem;
  height: 2.25rem;
  padding: 0 1rem;
  cursor: pointer;
}
.gvp-empty__action:hover { background: #1E293B; }
</style>
