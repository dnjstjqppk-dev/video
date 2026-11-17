<script setup lang="ts">
import type { Video } from '@gvp/shared-types'
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

const { data: feedData, pending, error, refresh } = await useAsyncData(
  () => `home-feed-${selectedCategory.value}`,
  async () => {
    const res = await api.getHomeFeed(1, 24, selectedCategory.value)
    return res.items || []
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
  const list = feedData.value || []
  return list.filter(v => {
    if (appStore.hiddenVideoIds.has(v.id)) return false
    if (appStore.blockedChannelIds.has(v.channelId)) return false
    if (selectedCategory.value !== 'all' && selectedCategory.value !== 'live') {
      return v.category === selectedCategory.value
    }
    return true
  })
})

const { data: shortsData } = await useAsyncData('home-shorts', async () => {
  try {
    const res = await api.getShortsFeed(1, 10)
    return res.items || []
  } catch {
    return []
  }
})

const shortsShelf = computed(() => shortsData.value || [])

const { data: liveData } = await useAsyncData('home-live', async () => {
  try {
    const res = await api.listLiveStreams(1, 4)
    return res.items || []
  } catch {
    return []
  }
})

const liveStreamsShelf = computed(() => liveData.value || [])

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
  <div class="mx-auto max-w-[1720px] bg-white px-4 py-4 text-slate-900 sm:px-6">
    <!-- Category Filter Chips -->
    <div class="sticky top-14 z-30 -mx-4 mb-6 flex items-center gap-2 overflow-x-auto border-b border-slate-200 bg-white px-4 py-3 sm:-mx-6 sm:px-6">
      <button
        v-for="cat in categories"
        :key="cat.id"
        type="button"
        class="shrink-0 rounded-md px-3.5 py-1.5 text-xs font-semibold transition"
        :class="
          selectedCategory === cat.id
            ? 'bg-slate-900 text-white'
            : 'border border-slate-200 bg-slate-100 text-slate-700 hover:bg-slate-50'
        "
        @click="selectedCategory = cat.id"
      >
        {{ appStore.locale === 'ko' ? cat.labelKo : cat.labelEn }}
      </button>
    </div>

    <!-- Toast Notification for Feed Feedback -->
    <div
      v-if="feedbackToast"
      class="mb-4 flex items-center justify-between rounded-lg border border-blue-200 bg-blue-50 px-4 py-3 text-sm text-blue-700"
    >
      <span>{{ feedbackToast }}</span>
      <button
        type="button"
        class="text-xs font-semibold text-blue-700 hover:underline"
        @click="appStore.hiddenVideoIds.clear(); appStore.blockedChannelIds.clear(); feedbackToast = null"
      >
        {{ appStore.locale === 'ko' ? '실행 취소 (Undo)' : 'Undo' }}
      </button>
    </div>

    <!-- Sponsored Native Ad Slot Banner (Real Ad Decision Loop - ARTICLE 7) -->
    <section
      v-if="feedAd?.ad_available && feedAd?.creative"
      class="mb-8 overflow-hidden rounded-lg border border-slate-200 bg-slate-50 p-5 transition hover:bg-white"
    >
      <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
        <div class="flex items-center gap-4">
          <img
            v-if="feedAd.creative.mediaUrl"
            :src="api.resolveUrl(feedAd.creative.mediaUrl)"
            :alt="feedAd.creative.title"
            class="h-16 w-16 shrink-0 rounded-lg border border-slate-200 object-cover"
          />
          <div>
            <div class="mb-1.5 flex items-center gap-2">
              <span class="rounded-md border border-amber-200 bg-amber-50 px-2 py-0.5 text-[10px] font-bold uppercase text-amber-700">
                Sponsored Ad
              </span>
              <span class="text-xs text-slate-500">
                {{ feedAd.creative.advertiserName }} · Verified Ad Flight
              </span>
            </div>
            <h2 class="text-lg font-bold text-slate-900 sm:text-xl">
              {{ feedAd.creative.title }}
            </h2>
            <p class="mt-1 text-xs text-slate-600 sm:text-sm">
              {{ feedAd.creative.category }} · {{ feedAd.creative.callToAction }}
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="onFeedAdClick"
          class="shrink-0 rounded-md bg-blue-600 px-5 py-2.5 text-xs font-semibold text-white transition hover:bg-blue-700"
        >
          {{ feedAd.creative.callToAction || 'Visit Site →' }}
        </button>
      </div>
    </section>
    <section v-else class="mb-8 overflow-hidden rounded-lg border border-slate-200 bg-slate-50 p-5">
      <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
        <div>
          <div class="mb-1.5 flex items-center gap-2">
            <span class="rounded-md border border-amber-200 bg-amber-50 px-2 py-0.5 text-[10px] font-bold uppercase text-amber-700">
              Sponsored Ad
            </span>
            <span class="text-xs text-slate-500">StreamVerse Partner Network · CPM Verified</span>
          </div>
          <h2 class="text-lg font-bold text-slate-900 sm:text-xl">
            {{ appStore.locale === 'ko' ? 'StreamVerse Premium — 광고 없는 4K HDR 스트리밍 & 오프라인 저장' : 'StreamVerse Premium — Ad-Free 4K HDR & Offline Downloads' }}
          </h2>
          <p class="mt-1 text-xs text-slate-600 sm:text-sm">
            {{ appStore.locale === 'ko' ? '백그라운드 재생과 무손실 공간 음향을 지금 바로 경험하세요.' : 'Experience background playback, spatial audio, and zero interruptions.' }}
          </p>
        </div>
        <NuxtLink
          to="/premium"
          class="shrink-0 rounded-md bg-slate-900 px-5 py-2.5 text-xs font-semibold text-white transition hover:bg-slate-800"
        >
          {{ appStore.locale === 'ko' ? '프리미엄 시작하기 →' : 'Try Premium Free →' }}
        </NuxtLink>
      </div>
    </section>

    <!-- Loading Skeleton State -->
    <div v-if="pending" class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <div v-for="i in 8" :key="i" class="animate-pulse">
        <div class="aspect-video w-full rounded-lg border border-slate-200 bg-slate-100" />
        <div class="mt-3 flex gap-3">
          <div class="h-9 w-9 shrink-0 rounded-full bg-slate-100" />
          <div class="flex-1 space-y-2">
            <div class="h-4 w-4/5 rounded-md bg-slate-100" />
            <div class="h-3 w-1/2 rounded-md bg-slate-100" />
          </div>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="rounded-lg border border-red-200 bg-red-50 p-8 text-center">
      <p class="text-base font-semibold text-red-700">
        {{ appStore.locale === 'ko' ? '피드를 불러오는 중 오류가 발생했습니다.' : 'Failed to load the home video feed.' }}
      </p>
      <button
        type="button"
        class="mt-3 rounded-md bg-slate-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-slate-800"
        @click="refresh()"
      >
        {{ appStore.locale === 'ko' ? '다시 시도' : 'Retry' }}
      </button>
    </div>

    <!-- Empty State -->
    <div v-else-if="visibleVideos.length === 0" class="rounded-lg border border-slate-200 bg-slate-50 p-12 text-center">
      <p class="text-base font-semibold text-slate-600">
        {{ appStore.locale === 'ko' ? '선택한 카테고리에 표시할 동영상이 없습니다.' : 'No videos match the selected category.' }}
      </p>
      <button
        type="button"
        class="mt-4 rounded-md border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700 transition hover:bg-slate-50"
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
        <NuxtLink :to="`/watch/${video.id}`" class="relative aspect-video w-full overflow-hidden rounded-lg border border-slate-200 bg-slate-100">
          <img
            :src="video.thumbnailUrl"
            :alt="video.title"
            class="h-full w-full object-cover transition duration-200"
            loading="lazy"
          />
          <span class="absolute bottom-2 right-2 rounded-md border border-slate-200 bg-white/95 px-1.5 py-0.5 font-mono text-xs font-semibold text-slate-900 tabular-nums">
            {{ formatDuration(video.durationSeconds) }}
          </span>
          <!-- Watch progress bar if previously started -->
          <div
            v-if="appStore.getWatchProgress(video.id) > 0"
            class="absolute inset-x-0 bottom-0 h-1 bg-slate-200"
          >
            <div
              class="h-full bg-blue-600"
              :style="{ width: `${Math.min(100, (appStore.getWatchProgress(video.id) / video.durationSeconds) * 100)}%` }"
            />
          </div>
        </NuxtLink>

        <div class="mt-3 flex items-start justify-between gap-3">
          <NuxtLink :to="`/channels/${video.channelId}`" class="shrink-0">
            <img
              :src="video.channel?.avatarUrl || 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&q=80'"
              :alt="video.channel?.name"
              class="h-9 w-9 rounded-full border border-slate-200 object-cover"
            />
          </NuxtLink>

          <div class="min-w-0 flex-1">
            <NuxtLink :to="`/watch/${video.id}`" class="line-clamp-2 text-sm font-semibold leading-snug text-slate-900 group-hover:text-blue-600">
              {{ video.title }}
            </NuxtLink>
            <NuxtLink :to="`/channels/${video.channelId}`" class="mt-1 flex items-center gap-1 text-xs text-slate-600 hover:text-slate-900">
              <span>{{ video.channel?.name || 'StreamVerse Creator' }}</span>
              <svg v-if="video.channel?.isVerified" class="h-3.5 w-3.5 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10 10-4.5 10-10S17.5 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
              </svg>
            </NuxtLink>
            <p class="text-xs text-slate-500 tabular-nums">
              {{ formatViews(video.viewCount) }} · 4K HDR
            </p>
          </div>

          <!-- Context Menu ("Not interested" / "Don't recommend channel") -->
          <div class="relative">
            <button
              type="button"
              class="rounded-md p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-900"
              aria-label="Video options"
              @click="activeMenuVideoId = activeMenuVideoId === video.id ? null : video.id"
            >
              <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24">
                <path d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z" />
              </svg>
            </button>

            <div
              v-if="activeMenuVideoId === video.id"
              class="absolute right-0 top-8 z-30 w-56 rounded-lg border border-slate-200 bg-white py-1.5 text-xs text-slate-700 shadow-sm"
            >
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3.5 py-2 text-left hover:bg-slate-50"
                @click="handleFeedFeedback(video, 'not_interested')"
              >
                <svg class="h-4 w-4 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                </svg>
                <span>{{ appStore.locale === 'ko' ? '관심 없음 (Not interested)' : 'Not interested' }}</span>
              </button>
              <button
                type="button"
                class="flex w-full items-center gap-2 px-3.5 py-2 text-left hover:bg-slate-50"
                @click="handleFeedFeedback(video, 'dont_recommend_channel')"
              >
                <svg class="h-4 w-4 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 12H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span>{{ appStore.locale === 'ko' ? '채널 추천 안함 (Don\'t recommend)' : "Don't recommend channel" }}</span>
              </button>
            </div>
          </div>
        </div>
      </article>
    </section>

    <!-- Live Now Shelf -->
    <section class="mt-12 border-t border-slate-200 pt-8">
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <span class="flex h-2.5 w-2.5 rounded-full bg-red-600" />
          <h2 class="text-lg font-bold text-slate-900">
            {{ appStore.locale === 'ko' ? '지금 라이브 중 (Live Now)' : 'Live Now — Ultra-Low Latency HLS' }}
          </h2>
        </div>
        <NuxtLink to="/live/live_main_01" class="text-xs font-semibold text-blue-700 hover:underline">
          {{ appStore.locale === 'ko' ? '전체 보기 →' : 'View All Live →' }}
        </NuxtLink>
      </div>

      <div v-if="liveStreamsShelf.length" class="grid grid-cols-1 gap-5 md:grid-cols-2">
        <NuxtLink
          v-for="stream in liveStreamsShelf"
          :key="stream.id"
          :to="`/live/${stream.id}`"
          class="group flex flex-col overflow-hidden rounded-lg border border-slate-200 bg-white transition hover:bg-slate-50 sm:flex-row"
        >
          <div class="relative aspect-video w-full shrink-0 overflow-hidden bg-slate-100 sm:w-64">
            <img :src="stream.thumbnailUrl" :alt="stream.title" class="h-full w-full object-cover" />
            <span class="absolute left-2.5 top-2.5 rounded-md border border-red-200 bg-red-50 px-2 py-0.5 text-[11px] font-bold uppercase text-red-700 tabular-nums">
              LIVE · {{ stream.viewerCount.toLocaleString() }} watching
            </span>
          </div>
          <div class="flex flex-col justify-center p-4">
            <h3 class="font-bold leading-snug text-slate-900 group-hover:text-blue-600">{{ stream.title }}</h3>
            <span class="mt-3 inline-flex w-fit items-center gap-1.5 rounded-md border border-slate-200 bg-slate-50 px-2.5 py-1 text-[11px] font-semibold text-slate-700">
              <svg class="h-3.5 w-3.5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 10h.01M12 10h.01M16 10h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <span>SuperChat & DVR Enabled</span>
            </span>
          </div>
        </NuxtLink>
      </div>
    </section>

    <!-- Shorts Shelf -->
    <section class="mt-12 border-t border-slate-200 pt-8">
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="rounded-md border border-blue-200 bg-blue-50 px-2 py-0.5 text-xs font-bold text-blue-700">SHORTS</span>
          <h2 class="text-lg font-bold text-slate-900">
            {{ appStore.locale === 'ko' ? '인기 쇼츠 (Trending Shorts)' : 'Trending Vertical Shorts' }}
          </h2>
        </div>
        <NuxtLink to="/shorts" class="text-xs font-semibold text-blue-700 hover:underline">
          {{ appStore.locale === 'ko' ? '쇼츠 플레이어 열기 →' : 'Open Fullscreen Shorts →' }}
        </NuxtLink>
      </div>

      <div v-if="shortsShelf.length" class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-5">
        <NuxtLink
          v-for="shortItem in shortsShelf"
          :key="shortItem.id"
          :to="`/watch/${shortItem.id}`"
          class="group flex flex-col overflow-hidden rounded-lg border border-slate-200 bg-white transition hover:bg-slate-50"
        >
          <div class="relative aspect-[9/16] w-full overflow-hidden bg-slate-100">
            <img
              :src="shortItem.thumbnailUrl"
              :alt="shortItem.title"
              class="h-full w-full object-cover"
              loading="lazy"
            />
          </div>
          <div class="border-t border-slate-200 bg-white p-3">
            <p class="line-clamp-2 text-xs font-semibold leading-snug text-slate-900 group-hover:text-blue-600">{{ shortItem.title }}</p>
            <p class="mt-1 text-[11px] text-slate-500 tabular-nums">{{ formatViews(shortItem.viewCount) }}</p>
          </div>
        </NuxtLink>
      </div>
      <p v-else class="rounded-lg border border-slate-200 bg-slate-50 p-6 text-center text-xs text-slate-500">
        {{ appStore.locale === 'ko' ? '표시할 쇼츠가 없습니다.' : 'No Shorts to show yet.' }}
      </p>
    </section>
  </div>
</template>
