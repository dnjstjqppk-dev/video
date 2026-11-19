<script setup lang="ts">
import type { Video } from '@gvp/shared-types'
import { VideoQuality } from '@gvp/shared-types'
import { formatRelativeTime, formatTimecode, formatViewCount } from '@gvp/ui'
import { useAppStore } from '~/stores/app'

const route = useRoute()
const router = useRouter()
const api = useApi()
const appStore = useAppStore()

const query = computed(() => String(route.query.q || '4K Architecture'))
const selectedType = ref<'all' | 'video' | 'shorts' | 'live' | 'channel' | 'playlist'>('all')
const selectedSort = ref<'relevance' | 'views' | 'date'>('relevance')

useSeoMeta({
  title: () => `"${query.value}" 검색 결과 — vravra 브라브라`,
  description: () => `OpenSearch hybrid Korean Nori & English search results for ${query.value} on vravra.`
})

const typoCorrection = computed(() => {
  const q = query.value.toLowerCase()
  if (q.includes('golnag') || q.includes('nxut')) {
    return 'golang nuxt 4 architecture'
  }
  return null
})

const { data: searchData, pending: searchPending, error: searchError } = await useAsyncData(
  () => `search-${query.value}-${selectedType.value}-${selectedSort.value}`,
  async () => {
    const res = await api.search(query.value, {
      type: selectedType.value === 'all' ? undefined : selectedType.value,
      sort: selectedSort.value
    })
    return res.videos || []
  },
  { watch: [query, selectedType, selectedSort] }
)

const results = computed(() => {
  const list = [...(searchData.value || [])]
  if (selectedSort.value === 'views') {
    list.sort((a, b) => b.viewCount - a.viewCount)
  } else if (selectedSort.value === 'date') {
    list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
  }
  return list
})

function hasCaptions(video: Video): boolean {
  return (video.captions?.length || 0) > 0
}

function has4K(video: Video): boolean {
  return (video.sources || []).some(s => s.quality === VideoQuality.P2160 || (s.height || 0) >= 2160)
}

const typeFilters: Array<{ id: 'all' | 'video' | 'shorts' | 'live' | 'channel' | 'playlist'; label: string }> = [
  { id: 'all', label: 'All Types' },
  { id: 'video', label: 'Videos' },
  { id: 'shorts', label: 'Shorts' },
  { id: 'live', label: 'Live' },
  { id: 'channel', label: 'Channels' },
  { id: 'playlist', label: 'Playlists' }
]
</script>

<template>
  <div class="mx-auto max-w-6xl bg-[var(--gvp-surface)] px-4 py-6 text-[var(--gvp-text)]">
    <!-- Filter & Sort Bar -->
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-[var(--gvp-border)] pb-4">
      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="filter in typeFilters"
          :key="filter.id"
          type="button"
          class="rounded-full px-3.5 py-1.5 text-xs font-semibold transition"
          :class="selectedType === filter.id ? 'bg-[var(--gvp-text)] text-[var(--gvp-canvas)]' : 'bg-[var(--gvp-elevated)] text-[var(--gvp-text)] hover:bg-[var(--gvp-border-strong)]'"
          @click="selectedType = filter.id"
        >
          {{ filter.label }}
        </button>
      </div>

      <div class="flex items-center gap-2 text-xs">
        <span class="font-medium text-[var(--gvp-text-secondary)]">{{ appStore.locale === 'ko' ? '정렬 기준:' : 'Sort by:' }}</span>
        <select
          v-model="selectedSort"
          class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-3 py-1.5 text-xs font-medium text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-brand)]"
        >
          <option value="relevance">Relevance (BM25 + Vector)</option>
          <option value="views">View Count (High to Low)</option>
          <option value="date">Upload Date (Newest)</option>
        </select>
      </div>
    </div>

    <!-- Typo Correction Banner -->
    <div
      v-if="typoCorrection"
      class="mt-4 rounded-lg border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-4 py-2.5 text-xs text-[var(--gvp-warning)]"
    >
      Did you mean:
      <button
        type="button"
        class="ml-1 font-bold underline"
        @click="router.push({ path: '/search', query: { q: typoCorrection } })"
      >
        {{ typoCorrection }}
      </button>
      ?
    </div>

    <p class="mt-4 text-xs text-[var(--gvp-text-secondary)]">
      Showing OpenSearch results for <strong class="font-semibold text-[var(--gvp-text)]">"{{ query }}"</strong> (Korean Nori + English stemmer active)
    </p>

    <div v-if="searchPending" class="mt-5 space-y-4">
      <div v-for="i in 3" :key="i" class="h-32 animate-pulse rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="searchError" class="mt-5 rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">
        {{ appStore.locale === 'ko' ? '검색 중 오류가 발생했습니다.' : 'Search failed. Please retry.' }}
      </p>
    </div>
    <div v-else-if="results.length === 0" class="mt-5 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-12 text-center">
      <p class="text-sm font-semibold text-[var(--gvp-text-secondary)]">
        {{ appStore.locale === 'ko' ? `"${query}"에 대한 결과가 없습니다.` : `No results for "${query}".` }}
      </p>
    </div>

    <!-- Results List -->
    <div v-else class="mt-5 space-y-4">
      <article
        v-for="video in results"
        :key="video.id"
        class="group flex flex-col gap-4 rounded-xl p-2 transition hover:bg-[var(--gvp-elevated)] sm:flex-row"
      >
        <NuxtLink :to="`/watch/${video.id}`" class="relative aspect-video w-full shrink-0 overflow-hidden rounded-xl bg-[var(--gvp-elevated)] sm:w-80">
          <img :src="video.thumbnailUrl" :alt="video.title" class="h-full w-full object-cover" loading="lazy" />
          <span class="absolute bottom-2 right-2 rounded bg-black/80 px-1 py-0.5 text-xs font-medium text-white tabular-nums">
            {{ formatTimecode(video.durationSeconds) }}
          </span>
        </NuxtLink>

        <div class="flex flex-1 flex-col justify-between py-1">
          <div>
            <NuxtLink :to="`/watch/${video.id}`" class="line-clamp-2 text-base font-semibold leading-snug text-[var(--gvp-text)]">
              {{ video.title }}
            </NuxtLink>
            <p class="mt-1 text-xs text-[var(--gvp-text-secondary)] tabular-nums">
              {{ formatViewCount(video.viewCount, appStore.locale) }} · {{ formatRelativeTime(video.publishedAt || video.createdAt, appStore.locale) }}
            </p>
            <NuxtLink :to="`/channels/${video.channelId}`" class="mt-3 flex items-center gap-2 text-xs text-[var(--gvp-text)]">
              <img :src="video.channel?.avatarUrl" :alt="video.channel?.name" class="h-6 w-6 rounded-full object-cover" loading="lazy" />
              <span class="font-medium">{{ video.channel?.name }}</span>
              <svg v-if="video.channel?.isVerified" class="h-3.5 w-3.5 shrink-0 text-[var(--gvp-text-secondary)]" fill="currentColor" viewBox="0 0 24 24" aria-label="인증된 채널"><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10 10-4.5 10-10S17.5 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" /></svg>
            </NuxtLink>
            <p class="mt-2 line-clamp-2 text-xs text-[var(--gvp-text-secondary)]">
              {{ video.description }}
            </p>
          </div>

          <div v-if="has4K(video) || hasCaptions(video) || (video.tags?.length || 0) > 0" class="mt-3 flex flex-wrap gap-1.5">
            <span v-if="has4K(video)" class="rounded bg-[var(--gvp-elevated)] px-1.5 py-0.5 text-[10px] font-semibold uppercase text-[var(--gvp-text-secondary)] tabular-nums">4K</span>
            <span v-if="hasCaptions(video)" class="rounded bg-[var(--gvp-elevated)] px-1.5 py-0.5 text-[10px] font-semibold uppercase text-[var(--gvp-text-secondary)]">CC</span>
            <span
              v-for="tag in video.tags"
              :key="tag"
              class="rounded bg-[var(--gvp-elevated)] px-1.5 py-0.5 text-[10px] font-medium text-[var(--gvp-text-secondary)]"
            >
              #{{ tag }}
            </span>
          </div>
        </div>
      </article>
    </div>
  </div>
</template>
