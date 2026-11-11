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
  <div class="gvp-skeleton animate-pulse" aria-hidden="true">
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
.gvp-skeleton { display: flex; flex-direction: column; gap: 0.625rem; }
.gvp-skeleton__thumb {
  aspect-ratio: 16 / 9;
  border-radius: 0.375rem;
  background: #F1F5F9;
}
.gvp-skeleton__row { display: flex; gap: 0.75rem; }
.gvp-skeleton__avatar {
  width: 2.25rem; height: 2.25rem; border-radius: 50%; flex-shrink: 0;
  background: #F1F5F9;
}
.gvp-skeleton__lines { flex: 1; display: flex; flex-direction: column; gap: 0.5rem; }
.gvp-skeleton__line {
  height: 0.75rem;
  border-radius: 0.375rem;
  background: #F1F5F9;
}
</style>
