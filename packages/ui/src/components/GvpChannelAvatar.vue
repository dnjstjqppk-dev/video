<script setup lang="ts">
withDefaults(defineProps<{
  src?: string
  name: string
  sizePx?: number
}>(), { sizePx: 36 })
</script>

<template>
  <span class="gvp-av" :style="{ width: `${sizePx}px`, height: `${sizePx}px` }" :title="name">
    <img v-if="src" :src="src" :alt="name" loading="lazy" />
    <span v-else aria-hidden="true">{{ name.trim().charAt(0).toUpperCase() || '?' }}</span>
  </span>
</template>

<style scoped>
.gvp-av { flex-shrink: 0; display: inline-flex; align-items: center; justify-content: center; border-radius: 50%; overflow: hidden; background: var(--gvp-brand-soft); color: var(--gvp-brand); font-weight: 700; font-size: 0.8125rem; }
.gvp-av img { width: 100%; height: 100%; object-fit: cover; }
</style>
