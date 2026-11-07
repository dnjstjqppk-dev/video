<script setup lang="ts">
withDefaults(
  defineProps<{
    lines?: number
    avatar?: boolean
    aspectVideo?: boolean
  }>(),
  { lines: 3, avatar: false, aspectVideo: false }
)
</script>

<template>
  <div class="gvp-skeleton" aria-hidden="true">
    <div v-if="aspectVideo" class="gvp-skeleton__thumb" />
    <div class="gvp-skeleton__row">
      <div v-if="avatar" class="gvp-skeleton__avatar" />
      <div class="gvp-skeleton__lines">
        <div v-for="n in lines" :key="n" class="gvp-skeleton__line" :style="{ width: n === lines ? '55%' : '100%' }" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.gvp-skeleton { display: flex; flex-direction: column; gap: 0.75rem; }
.gvp-skeleton__thumb {
  aspect-ratio: 16 / 9;
  border-radius: var(--gvp-radius-md);
  background: linear-gradient(100deg, var(--gvp-surface-2) 40%, var(--gvp-surface-3) 50%, var(--gvp-surface-2) 60%);
  background-size: 200% 100%;
  animation: gvp-shimmer 1.4s ease infinite;
}
.gvp-skeleton__row { display: flex; gap: 0.75rem; }
.gvp-skeleton__avatar {
  width: 2.25rem; height: 2.25rem; border-radius: 50%; flex-shrink: 0;
  background: var(--gvp-surface-2);
}
.gvp-skeleton__lines { flex: 1; display: flex; flex-direction: column; gap: 0.5rem; }
.gvp-skeleton__line {
  height: 0.75rem;
  border-radius: var(--gvp-radius-pill);
  background: linear-gradient(100deg, var(--gvp-surface-2) 40%, var(--gvp-surface-3) 50%, var(--gvp-surface-2) 60%);
  background-size: 200% 100%;
  animation: gvp-shimmer 1.4s ease infinite;
}
@keyframes gvp-shimmer { to { background-position: -200% 0; } }
@media (prefers-reduced-motion: reduce) {
  .gvp-skeleton__thumb, .gvp-skeleton__line { animation: none; }
}
</style>
