<script setup lang="ts">
import type { Video } from '@gvp/shared-types'
import { formatViewCount } from '@gvp/ui'

const props = defineProps<{ videos: Video[] }>()
const latest = computed(() => [...props.videos].sort((a, b) => +new Date(b.publishedAt || b.createdAt) - +new Date(a.publishedAt || a.createdAt))[0])
const avgViews = computed(() => props.videos.length ? props.videos.reduce((s, v) => s + v.viewCount, 0) / props.videos.length : 0)
const rank = computed(() => latest.value ? [...props.videos].sort((a, b) => b.viewCount - a.viewCount).findIndex(v => v.id === latest.value!.id) + 1 : 0)
const deltaPct = computed(() => {
  if (!latest.value || avgViews.value <= 0) return 0
  return Math.round((latest.value.viewCount - avgViews.value) / avgViews.value * 100)
})
</script>

<template>
  <section class="flex flex-col rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5" aria-label="최신 동영상 실적">
    <h2 class="text-sm font-semibold text-[var(--gvp-text)]">최신 동영상 실적</h2>
    <div v-if="!latest" class="py-6 text-center text-xs text-[var(--gvp-text-secondary)]">
      <p>아직 업로드한 동영상이 없습니다.</p>
      <NuxtLink to="/upload" class="mt-2 inline-block font-medium underline">첫 동영상 업로드하기</NuxtLink>
    </div>
    <div v-else class="mt-3">
      <img :src="latest.thumbnailUrl" :alt="latest.title" class="aspect-video w-full rounded-lg border border-[var(--gvp-border)] object-cover" loading="lazy" />
      <p class="mt-3 line-clamp-2 text-sm font-semibold text-[var(--gvp-text)]">{{ latest.title }}</p>
      <div class="mt-2 flex items-center gap-3 text-xs text-[var(--gvp-text-secondary)] tabular-nums">
        <span>조회수 {{ formatViewCount(latest.viewCount) }}</span>
        <span>댓글 {{ latest.commentCount.toLocaleString() }}</span>
        <span>좋아요 {{ latest.likeCount.toLocaleString() }}</span>
      </div>
      <p class="mt-2 rounded-lg bg-[var(--gvp-canvas)] px-3 py-2 text-xs text-[var(--gvp-text-secondary)] tabular-nums">
        최근 {{ videos.length }}개 중 {{ rank }}위 · 채널 평균 대비 {{ deltaPct >= 0 ? '+' : '' }}{{ deltaPct }}%
      </p>
      <div class="mt-3 flex gap-2 text-xs font-medium">
        <NuxtLink to="/analytics" class="rounded-md px-2 py-1 text-[var(--gvp-brand-hover)] hover:bg-[var(--gvp-elevated)]">분석 보기</NuxtLink>
        <NuxtLink to="/comments" class="rounded-md px-2 py-1 text-[var(--gvp-brand-hover)] hover:bg-[var(--gvp-elevated)]">댓글 보기</NuxtLink>
      </div>
    </div>
  </section>
</template>
