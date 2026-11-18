<script setup lang="ts">
withDefaults(defineProps<{ lines?: number }>(), { lines: 2 })
</script>

<template>
  <div class="gvp-vs" aria-hidden="true">
    <div class="gvp-shimmer gvp-vs__thumb" />
    <div class="gvp-vs__row">
      <div class="gvp-shimmer gvp-vs__av" />
      <div class="gvp-vs__tx">
        <div class="gvp-shimmer gvp-vs__t" />
        <div v-if="lines > 1" class="gvp-shimmer gvp-vs__s" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.gvp-vs { display: flex; flex-direction: column; gap: 10px; }
.gvp-vs__thumb { aspect-ratio: 16/9; }
.gvp-vs__row { display: flex; gap: 10px; }
.gvp-vs__av { width: 36px; height: 36px; border-radius: 50% !important; flex-shrink: 0; }
.gvp-vs__tx { flex: 1; display: flex; flex-direction: column; gap: 6px; padding-top: 4px; }
.gvp-vs__t { height: 14px; width: 88%; border-radius: 6px !important; }
.gvp-vs__s { height: 12px; width: 55%; border-radius: 6px !important; }
</style>
