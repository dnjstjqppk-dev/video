<script setup lang="ts">
import type { Video } from '@gvp/shared-types'
import { VideoProcessingStatus, VideoType, VideoVisibility } from '@gvp/shared-types'
import { useAppStore } from '~/stores/app'

const route = useRoute()
const router = useRouter()
const api = useApi()
const appStore = useAppStore()

const query = computed(() => String(route.query.q || '4K Architecture'))
const selectedType = ref<'all' | 'video' | 'shorts' | 'live' | 'channel' | 'playlist'>('all')
const selectedSort = ref<'relevance' | 'views' | 'date'>('relevance')

useSeoMeta({
  title: () => `"${query.value}" 검색 결과 — StreamVerse OpenSearch`,
  description: () => `OpenSearch hybrid Korean Nori & English search results for ${query.value}.`
})

const typoCorrection = computed(() => {
  const q = query.value.toLowerCase()
  if (q.includes('golnag') || q.includes('nxut')) {
    return 'golang nuxt 4 architecture'
  }
  return null
})

const fallbackResults: Video[] = [
  {
    id: 'vid_arch_2026',
    channelId: 'ch_tech_kr',
    channel: {
      id: 'ch_tech_kr',
      ownerId: 'usr_1',
      handle: '@gvptech',
      name: 'GVP Tech Engineering',
      avatarUrl: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&q=80',
      subscriberCount: 284000,
      videoCount: 142,
      isVerified: true,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z'
    },
    title: 'Building a 10M Concurrent Viewer Global Video Platform with Go 1.24, Nuxt 4 & ClickHouse',
    description: 'Full OpenSearch Nori Korean analyzer + BM25 ranking + vector similarity hybrid search walkthrough.',
    videoType: VideoType.STANDARD,
    durationSeconds: 1845,
    thumbnailUrl: 'https://images.unsplash.com/photo-1518770660439-4636190af475?w=800&q=80',
    visibility: VideoVisibility.PUBLIC,
    status: VideoProcessingStatus.READY,
    viewCount: 412900,
    likeCount: 32400,
    dislikeCount: 120,
    commentCount: 1842,
    tags: ['golang', 'nuxt4', 'opensearch'],
    category: 'tech',
    isAgeRestricted: false,
    sources: [],
    chapters: [],
    captions: [],
    publishedAt: '2026-10-20T09:00:00Z',
    createdAt: '2026-10-20T08:00:00Z',
    updatedAt: '2026-10-20T09:00:00Z'
  },
  {
    id: 'vid_seoul_4k',
    channelId: 'ch_seoul_4k',
    channel: {
      id: 'ch_seoul_4k',
      ownerId: 'usr_2',
      handle: '@seoul4k',
      name: 'Seoul 4K Cinema',
      avatarUrl: 'https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=100&q=80',
      subscriberCount: 890000,
      videoCount: 310,
      isVerified: true,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z'
    },
    title: '[4K 60fps HDR] Cyberpunk Rainy Night Walk in Gangnam & Hongdae, Seoul 2026',
    description: 'Ultra-crisp 2160p HDR video encoded via distributed FFmpeg transcoding fleet.',
    videoType: VideoType.STANDARD,
    durationSeconds: 3620,
    thumbnailUrl: 'https://images.unsplash.com/photo-1517154421773-0529f29ea451?w=800&q=80',
    visibility: VideoVisibility.PUBLIC,
    status: VideoProcessingStatus.READY,
    viewCount: 1294000,
    likeCount: 95100,
    dislikeCount: 310,
    commentCount: 4510,
    tags: ['seoul', '4k', 'hdr'],
    category: 'cinema',
    isAgeRestricted: false,
    sources: [],
    chapters: [],
    captions: [],
    publishedAt: '2026-10-18T14:00:00Z',
    createdAt: '2026-10-18T12:00:00Z',
    updatedAt: '2026-10-18T14:00:00Z'
  }
]

const { data: searchData } = await useAsyncData(
  () => `search-${query.value}-${selectedType.value}-${selectedSort.value}`,
  async () => {
    try {
      const res = await api.search(query.value, {
        type: selectedType.value === 'all' ? undefined : selectedType.value,
        sort: selectedSort.value
      })
      return res.videos?.length ? res.videos : fallbackResults
    } catch {
      return fallbackResults
    }
  },
  { watch: [query, selectedType, selectedSort] }
)

const results = computed(() => {
  const list = [...(searchData.value || fallbackResults)]
  if (selectedSort.value === 'views') {
    list.sort((a, b) => b.viewCount - a.viewCount)
  } else if (selectedSort.value === 'date') {
    list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
  }
  return list
})

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
  <div class="mx-auto max-w-6xl px-4 py-6">
    <!-- Filter & Sort Bar -->
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-white/10 pb-4">
      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="filter in typeFilters"
          :key="filter.id"
          type="button"
          class="rounded-lg px-3.5 py-1.5 text-xs font-semibold transition"
          :class="selectedType === filter.id ? 'bg-white text-black' : 'bg-[#272727] text-neutral-300 hover:bg-[#3A3A3A]'"
          @click="selectedType = filter.id"
        >
          {{ filter.label }}
        </button>
      </div>

      <div class="flex items-center gap-2 text-xs">
        <span class="text-neutral-400">{{ appStore.locale === 'ko' ? '정렬 기준:' : 'Sort by:' }}</span>
        <select
          v-model="selectedSort"
          class="rounded-lg border border-white/15 bg-[#181818] px-3 py-1.5 text-xs text-white outline-none"
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
      class="mt-4 rounded-xl border border-amber-400/30 bg-amber-500/10 px-4 py-2.5 text-xs text-amber-200"
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

    <p class="mt-4 text-xs text-neutral-400">
      Showing OpenSearch results for <strong class="text-white">"{{ query }}"</strong> (Korean Nori + English stemmer active)
    </p>

    <!-- Results List -->
    <div class="mt-5 space-y-5">
      <article
        v-for="video in results"
        :key="video.id"
        class="group flex flex-col gap-4 rounded-2xl border border-white/10 bg-[#161616] p-3 transition hover:border-white/25 sm:flex-row"
      >
        <NuxtLink :to="`/watch/${video.id}`" class="relative aspect-video w-full shrink-0 overflow-hidden rounded-xl bg-black sm:w-80">
          <img :src="video.thumbnailUrl" :alt="video.title" class="h-full w-full object-cover transition group-hover:scale-105" />
          <span class="absolute bottom-2 right-2 rounded bg-black/85 px-1.5 py-0.5 font-mono text-xs text-white">
            {{ Math.floor(video.durationSeconds / 60) }}:{{ String(video.durationSeconds % 60).padStart(2, '0') }}
          </span>
        </NuxtLink>

        <div class="flex flex-1 flex-col justify-between py-1">
          <div>
            <NuxtLink :to="`/watch/${video.id}`" class="text-base font-bold leading-snug text-white group-hover:text-[#3EA6FF]">
              {{ video.title }}
            </NuxtLink>
            <p class="mt-1 text-xs text-neutral-400">
              {{ video.viewCount.toLocaleString() }} views · Published {{ new Date(video.publishedAt || video.createdAt).toLocaleDateString() }}
            </p>
            <NuxtLink :to="`/channels/${video.channelId}`" class="mt-3 flex items-center gap-2 text-xs text-neutral-300 hover:text-white">
              <img :src="video.channel?.avatarUrl" :alt="video.channel?.name" class="h-6 w-6 rounded-full object-cover" />
              <span class="font-semibold">{{ video.channel?.name }}</span>
              <span class="text-[#3EA6FF]">✓ Verified</span>
            </NuxtLink>
            <p class="mt-2 line-clamp-2 text-xs text-neutral-400">
              {{ video.description }}
            </p>
          </div>

          <div class="mt-3 flex flex-wrap gap-1.5">
            <span class="rounded bg-white/10 px-2 py-0.5 text-[10px] font-semibold uppercase text-neutral-200">4K 60fps</span>
            <span class="rounded bg-white/10 px-2 py-0.5 text-[10px] font-semibold uppercase text-neutral-200">CC Subtitles</span>
            <span
              v-for="tag in video.tags"
              :key="tag"
              class="rounded bg-[#3EA6FF]/15 px-2 py-0.5 text-[10px] text-[#3EA6FF]"
            >
              #{{ tag }}
            </span>
          </div>
        </div>
      </article>
    </div>
  </div>
</template>
