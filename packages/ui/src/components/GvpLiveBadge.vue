<script setup lang="ts">
withDefaults(defineProps<{ label?: string; pulse?: boolean }>(), { label: 'LIVE', pulse: true })
</script>

<template>
  <span class="gvp-live tabular-nums" role="status" :aria-label="label">
    <span v-if="pulse" class="gvp-live__dot gvp-live-dot" aria-hidden="true" />
    {{ label }}
  </span>
</template>

<style scoped>
.gvp-live { display: inline-flex; align-items: center; gap: 5px; font-size: 0.6875rem; font-weight: 700; letter-spacing: 0.04em; padding: 2px 8px; border-radius: 6px; background: var(--gvp-live); color: #fff; }
.gvp-live__dot { width: 6px; height: 6px; border-radius: 50%; background: #fff; }
</style>
