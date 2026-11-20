<script setup lang="ts">
import type { SeriesDetail } from '@gvp/shared-types'
import { formatTimecode, formatViewCount } from '@gvp/ui'

const props = defineProps<{ detail: SeriesDetail; currentVideoId: string; locale: string }>()
const expanded = ref(false)
const visibleEpisodes = computed(() => expanded.value ? props.detail.episodes : props.detail.episodes.slice(0, 5))
</script>

<template>
  <section aria-label="시리즈 회차" class="overflow-hidden rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
    <div class="border-b border-[var(--gvp-border)] p-3">
      <p class="text-[10px] font-bold uppercase tracking-wider text-[var(--gvp-text-secondary)]">
        {{ locale === 'ko' ? '시리즈' : 'Series' }} · {{ detail.episodes.length }}{{ locale === 'ko' ? '화' : ' eps' }}
      </p>
      <p class="mt-0.5 truncate text-sm font-bold text-[var(--gvp-text)]">{{ detail.series.title }}</p>
    </div>
    <ol class="max-h-[420px] divide-y divide-[var(--gvp-border)] overflow-y-auto">
      <li v-for="ep in visibleEpisodes" :key="ep.videoId">
        <NuxtLink
          :to="`/watch/${ep.videoId}`"
          class="flex items-center gap-3 p-2.5 transition hover:bg-[var(--gvp-elevated)]"
          :class="ep.videoId === currentVideoId ? 'bg-[var(--gvp-elevated)]' : ''"
          :aria-current="ep.videoId === currentVideoId ? 'true' : undefined"
        >
          <span class="w-10 shrink-0 text-center text-xs font-bold tabular-nums text-[var(--gvp-text-secondary)]">
            S{{ ep.seasonNo }}E{{ ep.episodeNo }}
          </span>
          <span class="relative h-12 w-20 shrink-0 overflow-hidden rounded-md bg-[var(--gvp-elevated)]">
            <img :src="ep.thumbnailUrl" :alt="ep.title" class="h-full w-full object-cover" loading="lazy" />
            <span v-if="ep.videoId === currentVideoId" class="absolute inset-0 flex items-center justify-center bg-black/50 text-white">▶</span>
          </span>
          <span class="min-w-0 flex-1">
            <span class="line-clamp-2 block text-xs font-medium text-[var(--gvp-text)]">{{ ep.title }}</span>
            <span class="mt-0.5 block text-[11px] tabular-nums text-[var(--gvp-text-secondary)]">{{ formatViewCount(ep.viewCount, locale) }} · {{ formatTimecode(ep.durationSeconds) }}</span>
          </span>
          <AiBadge v-if="ep.aiGenerated" :tool="ep.aiTool" />
        </NuxtLink>
      </li>
    </ol>
    <button
      v-if="detail.episodes.length > 5"
      type="button"
      class="w-full border-t border-[var(--gvp-border)] py-2 text-xs font-semibold text-[var(--gvp-text-secondary)] hover:text-[var(--gvp-text)]"
      @click="expanded = !expanded"
    >
      {{ expanded ? (locale === 'ko' ? '접기' : 'Show less') : (locale === 'ko' ? `전체 ${detail.episodes.length}화 보기` : `Show all ${detail.episodes.length}`) }}
    </button>
  </section>
</template>
