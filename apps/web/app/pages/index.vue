<script setup lang="ts">
import type { Video } from '@gvp/shared-types'
import { VideoProcessingStatus, VideoType, VideoVisibility } from '@gvp/shared-types'
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const api = useApi()

useSeoMeta({
  title: computed(() =>
    appStore.locale === 'ko'
      ? '홈 — 맞춤형 4K 비디오, 쇼츠 & 실시간 스트리밍'
      : 'Home — Personalized 4K Videos, Shorts & Live Streams'
  ),
  description:
    'Discover personalized 4K HDR videos, Shorts, and real-time live streams on StreamVerse Global Video Platform.'
})

const categories = [
  { id: 'all', labelKo: '전체', labelEn: 'All' },
  { id: 'tech', labelKo: '기술 & 엔지니어링', labelEn: 'Tech & Engineering' },
  { id: 'gaming', labelKo: '게이밍 & e스포츠', labelEn: 'Gaming & Esports' },
  { id: 'music', labelKo: 'K-Pop & 뮤직', labelEn: 'Music & K-Pop' },
  { id: 'cinema', labelKo: '4K 시네마 & 드론', labelEn: '4K Cinema & Drone' },
  { id: 'live', labelKo: '실시간 방송', labelEn: 'Live Now' },
  { id: 'ai', labelKo: 'AI & 딥러닝', labelEn: 'AI & Deep Learning' }
]

const selectedCategory = ref('all')
const activeMenuVideoId = ref<string | null>(null)
const feedbackToast = ref<string | null>(null)

const fallbackVideos: Video[] = [
  {
    id: 'vid_arch_2026',
    channelId: 'ch_tech_kr',
    channel: {
      id: 'ch_tech_kr',
      ownerId: 'usr_1',
      handle: '@gvptech',
      name: 'GVP Tech Engineering',
      avatarUrl: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=120&q=80',
      subscriberCount: 284000,
      videoCount: 142,
      isVerified: true,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z'
    },
    title: 'Building a 10M Concurrent Viewer Global Video Platform with Go 1.24, Nuxt 4 & ClickHouse',
    description: 'Deep-dive into HLS ABR transcoding, NATS JetStream event pipelines, and OpenSearch hybrid ranking.',
    videoType: VideoType.STANDARD,
    durationSeconds: 1845,
    thumbnailUrl: 'https://images.unsplash.com/photo-1518770660439-4636190af475?w=800&q=80',
    visibility: VideoVisibility.PUBLIC,
    status: VideoProcessingStatus.READY,
    viewCount: 412900,
    likeCount: 32400,
    dislikeCount: 120,
    commentCount: 1842,
    tags: ['golang', 'nuxt4', 'system-design'],
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
      avatarUrl: 'https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=120&q=80',
      subscriberCount: 890000,
      videoCount: 310,
      isVerified: true,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z'
    },
    title: '[4K 60fps HDR] Cyberpunk Rainy Night Walk in Gangnam & Hongdae, Seoul 2026',
    description: 'Binaural 3D audio night walk in Seoul captured in 2160p HDR10.',
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
  },
  {
    id: 'vid_esports_finals',
    channelId: 'ch_esports_global',
    channel: {
      id: 'ch_esports_global',
      ownerId: 'usr_3',
      handle: '@esports',
      name: 'Global Esports Arena',
      avatarUrl: 'https://images.unsplash.com/photo-1566492031773-4f4e44671857?w=120&q=80',
      subscriberCount: 2450000,
      videoCount: 980,
      isVerified: true,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z'
    },
    title: 'T1 vs Gen.G — Grand Finals Game 5 Full Highlights (Korean & English Dual Cast)',
    description: 'Unforgettable final teamfight at Baron pit in 1440p 60fps.',
    videoType: VideoType.STANDARD,
    durationSeconds: 1490,
    thumbnailUrl: 'https://images.unsplash.com/photo-1542751371-adc38448a05e?w=800&q=80',
    visibility: VideoVisibility.PUBLIC,
    status: VideoProcessingStatus.READY,
    viewCount: 3480000,
    likeCount: 215000,
    dislikeCount: 1900,
    commentCount: 19400,
    tags: ['esports', 'gaming', 'finals'],
    category: 'gaming',
    isAgeRestricted: false,
    sources: [],
    chapters: [],
    captions: [],
    publishedAt: '2026-10-22T18:30:00Z',
    createdAt: '2026-10-22T17:00:00Z',
    updatedAt: '2026-10-22T18:30:00Z'
  },
  {
    id: 'vid_ai_agents',
    channelId: 'ch_tech_kr',
    channel: {
      id: 'ch_tech_kr',
      ownerId: 'usr_1',
      handle: '@gvptech',
      name: 'GVP Tech Engineering',
      avatarUrl: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=120&q=80',
      subscriberCount: 284000,
      videoCount: 142,
      isVerified: true,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z'
    },
    title: 'Multimodal Video Indexing with Audio-Visual Fingerprints & Vector Embeddings',
    description: 'How our Rights Owner Copyright Engine detects modified clips in under 150ms.',
    videoType: VideoType.STANDARD,
    durationSeconds: 1120,
    thumbnailUrl: 'https://images.unsplash.com/photo-1620712943543-bcc4688e7485?w=800&q=80',
    visibility: VideoVisibility.PUBLIC,
    status: VideoProcessingStatus.READY,
    viewCount: 189400,
    likeCount: 16200,
    dislikeCount: 80,
    commentCount: 740,
    tags: ['ai', 'copyright', 'embeddings'],
    category: 'ai',
    isAgeRestricted: false,
    sources: [],
    chapters: [],
    captions: [],
    publishedAt: '2026-10-23T11:00:00Z',
    createdAt: '2026-10-23T10:00:00Z',
    updatedAt: '2026-10-23T11:00:00Z'
  }
]

const { data: feedData, pending, error, refresh } = await useAsyncData(
  () => `home-feed-${selectedCategory.value}`,
  async () => {
    try {
      const res = await api.getHomeFeed(1, 24, selectedCategory.value)
      return res.items?.length ? res.items : fallbackVideos
    } catch {
      return fallbackVideos
    }
  },
  { watch: [selectedCategory] }
)

const { data: feedAd } = await useAsyncData('home-feed-ad', async () => {
  try {
    const res = await api.getAdDecision({
      slot_id: 'home_feed_banner',
      user_id: appStore.currentUser?.id,
      category: 'Technology',
      platform: 'WEB'
    })
    return res
  } catch {
    return null
  }
})

function onFeedAdClick() {
  if (!feedAd.value?.creative) return
  if (feedAd.value.campaign_id && feedAd.value.decision_id) {
    api.recordAdEvent({
      event_id: `evt_clk_${Math.random().toString(36).slice(2, 9)}_${Date.now()}`,
      decision_id: feedAd.value.decision_id,
      campaign_id: feedAd.value.campaign_id,
      event_type: 'click'
    }).catch(() => {})
  }
  if (feedAd.value.creative.destinationUrl) {
    window.open(feedAd.value.creative.destinationUrl, '_blank')
  }
}

onMounted(() => {
  if (feedAd.value?.ad_available && feedAd.value?.campaign_id && feedAd.value?.decision_id) {
    api.recordAdEvent({
      event_id: `evt_imp_${Math.random().toString(36).slice(2, 9)}_${Date.now()}`,
      decision_id: feedAd.value.decision_id,
      campaign_id: feedAd.value.campaign_id,
      event_type: 'impression'
    }).catch(() => {})
  }
})

const visibleVideos = computed(() => {
  const list = feedData.value || fallbackVideos
  return list.filter(v => {
    if (appStore.hiddenVideoIds.has(v.id)) return false
    if (appStore.blockedChannelIds.has(v.channelId)) return false
    if (selectedCategory.value !== 'all' && selectedCategory.value !== 'live') {
      return v.category === selectedCategory.value
    }
    return true
  })
})

const shortsShelf = [
  {
    id: 'short_01',
    title: '60초 만에 이해하는 HLS ABR 화질 자동 전환 원리 ⚡',
    views: '1.4M',
    thumbnailUrl: 'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=500&q=80'
  },
  {
    id: 'short_02',
    title: '성수동 웨이팅 1시간 화덕피자 오픈런 🍕',
    views: '890K',
    thumbnailUrl: 'https://images.unsplash.com/photo-1513104890138-7c749659a591?w=500&q=80'
  },
  {
    id: 'short_03',
    title: 'Unreal Engine 5.5 실시간 레이트레이싱 데모',
    views: '2.1M',
    thumbnailUrl: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=500&q=80'
  },
  {
    id: 'short_04',
    title: '서울 한강 드론 라이트쇼 하이라이트 ✨',
    views: '3.2M',
    thumbnailUrl: 'https://images.unsplash.com/photo-1509198397868-475647b2a1e5?w=500&q=80'
  },
  {
    id: 'short_05',
    title: 'Go 1.24 vs Rust 비디오 세그먼트 파서 벤치마크',
    views: '640K',
    thumbnailUrl: 'https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=500&q=80'
  }
]

const liveStreamsShelf = [
  {
    id: 'live_main_01',
    title: '[LIVE] 2026 Global Architecture Summit — Keynote & Live Q&A',
    channelName: 'GVP Tech Engineering',
    viewers: 14820,
    thumbnailUrl: 'https://images.unsplash.com/photo-1540575467063-178a50c2df87?w=800&q=80'
  },
  {
    id: 'live_lofi_247',
    title: '[24/7 LIVE] Cyberpunk Seoul Lo-Fi Synthwave Radio for Coding & Focus',
    channelName: 'Seoul 4K Cinema',
    viewers: 6390,
    thumbnailUrl: 'https://images.unsplash.com/photo-1511671782779-c97d3d27a1d4?w=800&q=80'
  }
]

function formatDuration(sec: number): string {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${String(s).padStart(2, '0')}`
}

function formatViews(count: number): string {
  if (appStore.locale === 'ko') {
    if (count >= 10000) return `${(count / 10000).toFixed(1)}만회`
    return `${count.toLocaleString()}회`
  }
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M views`
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K views`
  return `${count} views`
}

async function handleFeedFeedback(video: Video, action: 'not_interested' | 'dont_recommend_channel') {
  activeMenuVideoId.value = null
  if (action === 'not_interested') {
    appStore.hiddenVideoIds.add(video.id)
    feedbackToast.value = appStore.locale === 'ko'
      ? `"${video.title.slice(0, 28)}..." 동영상을 관심 없음으로 표시했습니다.`
      : `Marked "${video.title.slice(0, 28)}..." as Not Interested.`
  } else {
    appStore.blockedChannelIds.add(video.channelId)
    feedbackToast.value = appStore.locale === 'ko'
      ? `${video.channel?.name || video.channelId} 채널을 추천하지 않도록 설정했습니다.`
      : `We won't recommend videos from ${video.channel?.name || video.channelId}.`
  }
  try {
    await api.sendFeedFeedback(video.id, action)
  } catch {
    // Non-blocking feedback call
  }
  setTimeout(() => {
    feedbackToast.value = null
  }, 4000)
}
</script>

<template>
  <div class="mx-auto max-w-[1720px] px-4 py-4 sm:px-6">
    <!-- Category Filter Chips -->
    <div class="sticky top-14 z-30 -mx-4 mb-6 flex items-center gap-2 overflow-x-auto border-b border-white/10 bg-[#0F0F0F]/95 px-4 py-3 backdrop-blur sm:-mx-6 sm:px-6">
      <button
        v-for="cat in categories"
        :key="cat.id"
        type="button"
        class="shrink-0 rounded-lg px-3.5 py-1.5 text-xs font-semibold transition"
        :class="
          selectedCategory === cat.id
            ? 'bg-white text-black shadow'
            : 'bg-[#272727] text-neutral-200 hover:bg-[#3A3A3A]'
        "
        @click="selectedCategory = cat.id"
      >
        {{ appStore.locale === 'ko' ? cat.labelKo : cat.labelEn }}
      </button>
    </div>

    <!-- Toast Notification for Feed Feedback -->
    <div
      v-if="feedbackToast"
      class="mb-4 flex items-center justify-between rounded-xl border border-[#3EA6FF]/40 bg-[#182330] px-4 py-3 text-sm text-white"
    >
      <span>{{ feedbackToast }}</span>
      <button
        type="button"
        class="text-xs font-semibold text-[#3EA6FF] hover:underline"
        @click="appStore.hiddenVideoIds.clear(); appStore.blockedChannelIds.clear(); feedbackToast = null"
      >
        {{ appStore.locale === 'ko' ? '실행 취소 (Undo)' : 'Undo' }}
      </button>
    </div>

    <!-- Sponsored Native Ad Slot Banner (Real Ad Decision Loop - ARTICLE 7) -->
    <section
      v-if="feedAd?.ad_available && feedAd?.creative"
      class="mb-8 overflow-hidden rounded-2xl border border-amber-500/30 bg-gradient-to-r from-[#1A1A2E] via-[#16213E] to-[#0F3460] p-5 shadow-xl transition hover:border-amber-400/50"
    >
      <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
        <div class="flex items-center gap-4">
          <img
            v-if="feedAd.creative.mediaUrl"
            :src="feedAd.creative.mediaUrl"
            :alt="feedAd.creative.title"
            class="h-16 w-16 shrink-0 rounded-xl object-cover shadow border border-white/10"
          />
          <div>
            <div class="mb-1.5 flex items-center gap-2">
              <span class="rounded bg-amber-400 px-2 py-0.5 text-[10px] font-extrabold uppercase text-black">
                Sponsored Ad
              </span>
              <span class="text-xs text-neutral-300">
                {{ feedAd.creative.advertiserName }} · Verified Ad Flight
              </span>
            </div>
            <h2 class="text-lg font-bold text-white sm:text-xl">
              {{ feedAd.creative.title }}
            </h2>
            <p class="mt-1 text-xs text-neutral-300 sm:text-sm">
              {{ feedAd.creative.category }} · {{ feedAd.creative.callToAction }}
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="onFeedAdClick"
          class="shrink-0 rounded-full bg-blue-600 px-6 py-2.5 text-xs font-bold text-white shadow-lg transition hover:bg-blue-500"
        >
          {{ feedAd.creative.callToAction || 'Visit Site →' }}
        </button>
      </div>
    </section>
    <section v-else class="mb-8 overflow-hidden rounded-2xl border border-white/10 bg-gradient-to-r from-[#1A1A2E] via-[#16213E] to-[#0F3460] p-5 shadow-lg">
      <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
        <div>
          <div class="mb-1.5 flex items-center gap-2">
            <span class="rounded bg-amber-400 px-2 py-0.5 text-[10px] font-extrabold uppercase text-black">
              Sponsored Ad
            </span>
            <span class="text-xs text-neutral-300">StreamVerse Partner Network · CPM Verified</span>
          </div>
          <h2 class="text-lg font-bold text-white sm:text-xl">
            {{ appStore.locale === 'ko' ? 'StreamVerse Premium — 광고 없는 4K HDR 스트리밍 & 오프라인 저장' : 'StreamVerse Premium — Ad-Free 4K HDR & Offline Downloads' }}
          </h2>
          <p class="mt-1 text-xs text-neutral-300 sm:text-sm">
            {{ appStore.locale === 'ko' ? '백그라운드 재생과 무손실 공간 음향을 지금 바로 경험하세요.' : 'Experience background playback, spatial audio, and zero interruptions.' }}
          </p>
        </div>
        <NuxtLink
          to="/premium"
          class="shrink-0 rounded-full bg-[#FF0033] px-5 py-2.5 text-xs font-bold text-white shadow transition hover:bg-[#D9002B]"
        >
          {{ appStore.locale === 'ko' ? '프리미엄 시작하기 →' : 'Try Premium Free →' }}
        </NuxtLink>
      </div>
    </section>

    <!-- Loading Skeleton State -->
    <div v-if="pending" class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <div v-for="i in 8" :key="i" class="animate-pulse">
        <div class="aspect-video w-full rounded-2xl bg-white/10" />
        <div class="mt-3 flex gap-3">
          <div class="h-9 w-9 shrink-0 rounded-full bg-white/10" />
          <div class="flex-1 space-y-2">
            <div class="h-4 w-4/5 rounded bg-white/10" />
            <div class="h-3 w-1/2 rounded bg-white/10" />
          </div>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="rounded-2xl border border-red-500/30 bg-red-950/20 p-8 text-center">
      <p class="text-base font-semibold text-red-300">
        {{ appStore.locale === 'ko' ? '피드를 불러오는 중 오류가 발생했습니다.' : 'Failed to load the home video feed.' }}
      </p>
      <button
        type="button"
        class="mt-3 rounded-full bg-[#FF0033] px-4 py-2 text-xs font-bold text-white"
        @click="refresh()"
      >
        {{ appStore.locale === 'ko' ? '다시 시도' : 'Retry' }}
      </button>
    </div>

    <!-- Empty State -->
    <div v-else-if="visibleVideos.length === 0" class="rounded-2xl border border-white/10 bg-[#181818] p-12 text-center">
      <p class="text-base font-semibold text-neutral-300">
        {{ appStore.locale === 'ko' ? '선택한 카테고리에 표시할 동영상이 없습니다.' : 'No videos match the selected category.' }}
      </p>
      <button
        type="button"
        class="mt-4 rounded-full bg-white/10 px-4 py-2 text-xs font-semibold text-white hover:bg-white/20"
        @click="selectedCategory = 'all'; appStore.hiddenVideoIds.clear(); appStore.blockedChannelIds.clear()"
      >
        {{ appStore.locale === 'ko' ? '필터 초기화' : 'Reset Filters' }}
      </button>
    </div>

    <!-- Main Video Grid -->
    <section v-else class="grid grid-cols-1 gap-x-5 gap-y-8 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <article
        v-for="video in visibleVideos"
        :key="video.id"
        class="group relative flex flex-col"
      >
        <NuxtLink :to="`/watch/${video.id}`" class="relative aspect-video w-full overflow-hidden rounded-2xl bg-[#1F1F1F]">
          <img
            :src="video.thumbnailUrl"
            :alt="video.title"
            class="h-full w-full object-cover transition duration-300 group-hover:scale-105"
            loading="lazy"
          />
          <span class="absolute bottom-2 right-2 rounded bg-black/85 px-1.5 py-0.5 font-mono text-xs font-semibold text-white">
            {{ formatDuration(video.durationSeconds) }}
          </span>
          <!-- Watch progress bar if previously started -->
          <div
            v-if="appStore.getWatchProgress(video.id) > 0"
            class="absolute inset-x-0 bottom-0 h-1 bg-white/30"
          >
            <div
              class="h-full bg-[#FF0033]"
              :style="{ width: `${Math.min(100, (appStore.getWatchProgress(video.id) / video.durationSeconds) * 100)}%` }"
            />
          </div>
        </NuxtLink>

        <div class="mt-3 flex items-start justify-between gap-3">
          <NuxtLink :to="`/channels/${video.channelId}`" class="shrink-0">
            <img
              :src="video.channel?.avatarUrl || 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&q=80'"
              :alt="video.channel?.name"
              class="h-9 w-9 rounded-full object-cover"
            />
          </NuxtLink>

          <div class="min-w-0 flex-1">
            <NuxtLink :to="`/watch/${video.id}`" class="line-clamp-2 text-sm font-semibold leading-snug group-hover:text-[#3EA6FF]">
              {{ video.title }}
            </NuxtLink>
            <NuxtLink :to="`/channels/${video.channelId}`" class="mt-1 flex items-center gap-1 text-xs text-neutral-400 hover:text-white">
              <span>{{ video.channel?.name || 'StreamVerse Creator' }}</span>
              <svg v-if="video.channel?.isVerified" class="h-3.5 w-3.5 text-[#3EA6FF]" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10 10-4.5 10-10S17.5 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
              </svg>
            </NuxtLink>
            <p class="text-xs text-neutral-400">
              {{ formatViews(video.viewCount) }} · 4K HDR
            </p>
          </div>

          <!-- Context Menu ("Not interested" / "Don't recommend channel") -->
          <div class="relative">
            <button
              type="button"
              class="rounded-full p-1.5 text-neutral-400 hover:bg-white/10 hover:text-white"
              aria-label="Video options"
              @click="activeMenuVideoId = activeMenuVideoId === video.id ? null : video.id"
            >
              <svg class="h-5 w-5 fill-current" viewBox="0 0 24 24">
                <path d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z" />
              </svg>
            </button>

            <div
              v-if="activeMenuVideoId === video.id"
              class="absolute right-0 top-8 z-30 w-56 rounded-xl border border-white/15 bg-[#212121] py-1.5 text-xs text-white shadow-2xl"
            >
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3.5 py-2 text-left hover:bg-white/10"
                @click="handleFeedFeedback(video, 'not_interested')"
              >
                <span>🚫</span>
                <span>{{ appStore.locale === 'ko' ? '관심 없음 (Not interested)' : 'Not interested' }}</span>
              </button>
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3.5 py-2 text-left hover:bg-white/10"
                @click="handleFeedFeedback(video, 'dont_recommend_channel')"
              >
                <span>⛔</span>
                <span>{{ appStore.locale === 'ko' ? '채널 추천 안함 (Don\'t recommend)' : "Don't recommend channel" }}</span>
              </button>
            </div>
          </div>
        </div>
      </article>
    </section>

    <!-- Live Now Shelf -->
    <section class="mt-12 border-t border-white/10 pt-8">
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <span class="flex h-3 w-3 rounded-full bg-[#FF0033] ring-4 ring-[#FF0033]/25" />
          <h2 class="text-lg font-bold">
            {{ appStore.locale === 'ko' ? '지금 라이브 중 (Live Now)' : 'Live Now — Ultra-Low Latency HLS' }}
          </h2>
        </div>
        <NuxtLink to="/live/live_main_01" class="text-xs font-semibold text-[#3EA6FF] hover:underline">
          {{ appStore.locale === 'ko' ? '전체 보기 →' : 'View All Live →' }}
        </NuxtLink>
      </div>

      <div class="grid grid-cols-1 gap-5 md:grid-cols-2">
        <NuxtLink
          v-for="stream in liveStreamsShelf"
          :key="stream.id"
          :to="`/live/${stream.id}`"
          class="group flex flex-col overflow-hidden rounded-2xl border border-white/10 bg-[#161616] transition hover:border-[#FF0033]/50 sm:flex-row"
        >
          <div class="relative aspect-video w-full sm:w-64 shrink-0 overflow-hidden">
            <img :src="stream.thumbnailUrl" :alt="stream.title" class="h-full w-full object-cover transition group-hover:scale-105" />
            <span class="absolute left-2.5 top-2.5 rounded bg-[#FF0033] px-2 py-0.5 text-[11px] font-bold uppercase text-white">
              LIVE · {{ stream.viewers.toLocaleString() }} watching
            </span>
          </div>
          <div class="flex flex-col justify-center p-4">
            <h3 class="font-bold leading-snug group-hover:text-[#3EA6FF]">{{ stream.title }}</h3>
            <p class="mt-1.5 text-xs text-neutral-400">{{ stream.channelName }}</p>
            <span class="mt-3 inline-flex w-fit items-center gap-1.5 rounded-full bg-white/10 px-3 py-1 text-[11px] font-semibold text-white">
              💬 SuperChat & DVR Enabled
            </span>
          </div>
        </NuxtLink>
      </div>
    </section>

    <!-- Shorts Shelf -->
    <section class="mt-12 border-t border-white/10 pt-8">
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="rounded-lg bg-[#FF0033] px-2 py-0.5 text-xs font-extrabold text-white">SHORTS</span>
          <h2 class="text-lg font-bold">
            {{ appStore.locale === 'ko' ? '인기 쇼츠 (Trending Shorts)' : 'Trending Vertical Shorts' }}
          </h2>
        </div>
        <NuxtLink to="/shorts" class="text-xs font-semibold text-[#3EA6FF] hover:underline">
          {{ appStore.locale === 'ko' ? '쇼츠 플레이어 열기 →' : 'Open Fullscreen Shorts →' }}
        </NuxtLink>
      </div>

      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-5">
        <NuxtLink
          v-for="shortItem in shortsShelf"
          :key="shortItem.id"
          to="/shorts"
          class="group relative aspect-[9/16] overflow-hidden rounded-2xl bg-[#1F1F1F]"
        >
          <img
            :src="shortItem.thumbnailUrl"
            :alt="shortItem.title"
            class="h-full w-full object-cover transition duration-300 group-hover:scale-105"
            loading="lazy"
          />
          <div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/20 to-transparent" />
          <div class="absolute inset-x-3 bottom-3 text-white">
            <p class="line-clamp-2 text-xs font-bold leading-snug">{{ shortItem.title }}</p>
            <p class="mt-1 text-[11px] text-neutral-300">{{ shortItem.views }} views</p>
          </div>
        </NuxtLink>
      </div>
    </section>
  </div>
</template>
