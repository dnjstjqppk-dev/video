<script setup lang="ts">
import { formatRelativeTime, formatTimecode, formatViewCount } from '../format'
import type { VideoCardProps } from '../index'

withDefaults(defineProps<VideoCardProps & {
  locale?: string
  progressPct?: number
  selected?: boolean
}>(), { locale: 'ko', progressPct: 0, selected: false })
</script>

<template>
  <article class="gvp-vc" :aria-selected="selected">
    <div class="gvp-vc__thumb">
      <img :src="thumbnailUrl" :alt="title" loading="lazy" />
      <span v-if="isLive" class="gvp-vc__live"><span class="gvp-live-dot" />LIVE</span>
      <span v-else class="gvp-vc__dur tabular-nums">{{ formatTimecode(durationSeconds) }}</span>
      <div v-if="progressPct > 0" class="gvp-vc__prog"><span :style="{ width: `${progressPct}%` }" /></div>
      <div class="gvp-vc__quick"><slot name="quick-actions" /></div>
    </div>
    <div class="gvp-vc__meta">
      <slot name="avatar" />
      <div class="gvp-vc__text">
        <h3 class="clamp-2">{{ title }}</h3>
        <p>{{ channelName }} <slot name="badge" /></p>
        <p class="tabular-nums">{{ formatViewCount(viewCount, locale) }} · {{ formatRelativeTime(publishedAt, locale) }}</p>
      </div>
      <slot name="menu" />
    </div>
  </article>
</template>

<style scoped>
.gvp-vc { display: flex; flex-direction: column; gap: 0.625rem; min-width: 0; }
.gvp-vc__thumb { position: relative; aspect-ratio: 16/9; border-radius: 12px; overflow: hidden; background: var(--gvp-elevated); }
.gvp-vc__thumb img { width: 100%; height: 100%; object-fit: cover; }
.gvp-vc__dur { position: absolute; right: 6px; bottom: 6px; font-size: 0.6875rem; padding: 1px 6px; border-radius: 6px; background: rgba(0,0,0,0.82); color: #fff; }
.gvp-vc__live { position: absolute; right: 6px; bottom: 6px; display: inline-flex; align-items: center; gap: 4px; font-size: 0.6875rem; font-weight: 700; padding: 2px 8px 2px 6px; border-radius: 6px; background: var(--gvp-live); color: #fff; }
.gvp-live-dot { width: 6px; height: 6px; border-radius: 50%; background: #fff; }
.gvp-vc__prog { position: absolute; left: 0; right: 0; bottom: 0; height: 3px; background: rgba(255,255,255,0.25); }
.gvp-vc__prog span { display: block; height: 100%; background: var(--gvp-brand); }
.gvp-vc__quick { position: absolute; top: 6px; right: 6px; display: none; gap: 4px; }
.gvp-vc:hover .gvp-vc__quick { display: inline-flex; }
.gvp-vc__meta { display: flex; gap: 0.625rem; }
.gvp-vc__text { flex: 1; min-width: 0; }
.gvp-vc__text h3 { font-size: 0.875rem; font-weight: 600; color: var(--gvp-text); margin: 0; }
.gvp-vc__text p { font-size: 0.75rem; color: var(--gvp-text-secondary); margin: 2px 0 0; }
</style>
