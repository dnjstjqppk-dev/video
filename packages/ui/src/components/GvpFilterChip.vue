<script setup lang="ts">
withDefaults(defineProps<{
  active?: boolean
  count?: number | null
}>(), { active: false, count: null })
defineEmits<{ (e: 'select'): void }>()
</script>

<template>
  <button type="button" class="gvp-chip tabular-nums" :class="{ 'gvp-chip--on': active }"
    :aria-pressed="active" @click="$emit('select')">
    <slot />
    <span v-if="count !== null && count !== undefined" class="gvp-chip__n">{{ count }}</span>
  </button>
</template>

<style scoped>
.gvp-chip { display: inline-flex; align-items: center; gap: 6px; height: 32px; padding: 0 12px; border-radius: 8px; font-size: 0.8125rem; font-weight: 500; white-space: nowrap; background: var(--gvp-elevated); color: var(--gvp-text); border: 1px solid transparent; cursor: pointer; transition: background-color 180ms cubic-bezier(0.2,0,0,1); }
.gvp-chip:hover { border-color: var(--gvp-border-strong); }
.gvp-chip--on { background: var(--gvp-text); color: var(--gvp-canvas); font-weight: 600; }
.gvp-chip__n { font-size: 0.6875rem; padding: 1px 6px; border-radius: 9999px; background: var(--gvp-brand); color: #fff; }
</style>
