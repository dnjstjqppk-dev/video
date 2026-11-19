<script setup lang="ts">
import type { Video } from '@gvp/shared-types'
import { formatSubscriberCount, formatViewCount } from '@gvp/ui'

const props = defineProps<{
  analytics: { views: number; watchTimeHours: number; subscribers: number; estimatedRevenue: number }
  videos: Video[]
  currency: 'KRW' | 'USD'
}>()
const topVideo = computed(() => [...props.videos].sort((a, b) => b.viewCount - a.viewCount)[0])
const revenue = computed(() => props.currency === 'KRW'
  ? `₩${props.analytics.estimatedRevenue.toLocaleString('ko-KR')}`
  : `$${(props.analytics.estimatedRevenue / 1350).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`)
</script>

<template>
  <section class="flex flex-col rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5" aria-label="채널 분석 요약">
    <h2 class="text-sm font-semibold text-[var(--gvp-text)]">채널 분석</h2>
    <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">현재 구독자 수</p>
    <p class="text-3xl font-semibold text-[var(--gvp-text)] tabular-nums">{{ formatSubscriberCount(analytics.subscribers) }}</p>
    <dl class="mt-4 space-y-2 border-t border-[var(--gvp-border)] pt-3 text-xs tabular-nums">
      <div class="flex justify-between"><dt class="text-[var(--gvp-text-secondary)]">조회수 (28일)</dt><dd class="font-semibold text-[var(--gvp-text)]">{{ formatViewCount(analytics.views) }}</dd></div>
      <div class="flex justify-between"><dt class="text-[var(--gvp-text-secondary)]">시청 시간 (28일)</dt><dd class="font-semibold text-[var(--gvp-text)]">{{ analytics.watchTimeHours.toLocaleString() }}시간</dd></div>
      <div class="flex justify-between"><dt class="text-[var(--gvp-text-secondary)]">예상 수익 (28일)</dt><dd class="font-semibold text-[var(--gvp-text)]">{{ revenue }}</dd></div>
    </dl>
    <div v-if="topVideo" class="mt-3 flex items-center gap-3 rounded-lg bg-[var(--gvp-canvas)] p-2.5">
      <img :src="topVideo.thumbnailUrl" :alt="topVideo.title" class="h-11 w-20 shrink-0 rounded-md object-cover" loading="lazy" />
      <div class="min-w-0">
        <p class="text-[11px] text-[var(--gvp-text-secondary)]">인기 동영상</p>
        <p class="truncate text-xs font-semibold text-[var(--gvp-text)]">{{ topVideo.title }}</p>
      </div>
    </div>
    <NuxtLink to="/analytics" class="mt-3 rounded-md px-2 py-1 text-xs font-medium text-[var(--gvp-brand-hover)] hover:bg-[var(--gvp-elevated)]">채널 분석으로 이동</NuxtLink>
  </section>
</template>
