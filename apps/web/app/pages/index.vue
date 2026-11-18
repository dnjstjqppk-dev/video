<script setup lang="ts">
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
  { id: 'music', labelKo: '음악', labelEn: 'Music' },
  { id: 'gaming', labelKo: '게임', labelEn: 'Gaming' },
  { id: 'news', labelKo: '뉴스', labelEn: 'News' },
  { id: 'live', labelKo: '라이브', labelEn: 'Live' },
  { id: 'tech', labelKo: '테크·개발', labelEn: 'Tech & Dev' },
  { id: 'finance', labelKo: '경제·재테크', labelEn: 'Finance' },
  { id: 'travel', labelKo: '여행·브이로그', labelEn: 'Travel & Vlog' },
  { id: 'recent', labelKo: '최근에 업로드된 동영상', labelEn: 'Recently uploaded' },
  { id: 'watched', labelKo: '감상한 동영상', labelEn: 'Watched' }
]

const selectedCategory = ref('all')
const feedbackToast = ref<string | null>(null)
const library = useLibraryStore()
const chipRail = ref<HTMLElement | null>(null)

function scrollChips(dir: 1 | -1) {
  chipRail.value?.scrollBy({ left: dir * 320, behavior: 'smooth' })
}

function onCardToast(msg: string) {
  feedbackToast.value = msg
  setTimeout(() => { feedbackToast.value = null }, 4000)
}

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
  const cat = selectedCategory.value
  if (cat === 'watched') {
    library.hydrate()
    return library.history.filter(v => !appStore.hiddenVideoIds.has(v.id) && !appStore.blockedChannelIds.has(v.channelId))
  }
  const list = (feedData.value || []).filter(v => {
    if (appStore.hiddenVideoIds.has(v.id)) return false
    if (appStore.blockedChannelIds.has(v.channelId)) return false
    if (cat !== 'all' && cat !== 'live' && cat !== 'recent') return v.category === cat
    return true
  })
  if (cat === 'recent') {
    return [...list].sort((a, b) => +new Date(b.publishedAt || b.createdAt) - +new Date(a.publishedAt || a.createdAt))
  }
  return list
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

function formatViews(count: number): string {
  if (appStore.locale === 'ko') {
    if (count >= 10000) return `${(count / 10000).toFixed(1)}만회`
    return `${count.toLocaleString()}회`
  }
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M views`
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K views`
  return `${count} views`
}

</script>

<template>
  <div class="mx-auto max-w-[1720px] bg-[var(--gvp-surface)] px-4 py-4 text-[var(--gvp-text)] sm:px-6">
    <!-- Category Filter Chips -->
    <div class="sticky top-14 z-30 -mx-4 mb-6 flex items-center gap-1 border-b border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-2 py-3 sm:-mx-6 sm:px-4">
      <button type="button" aria-label="카테고리 왼쪽으로 스크롤" class="hidden h-8 w-8 shrink-0 items-center justify-center rounded-full text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] sm:inline-flex" @click="scrollChips(-1)">‹</button>
      <div ref="chipRail" class="no-scrollbar flex flex-1 items-center gap-2 overflow-x-auto">
        <button
          v-for="cat in categories"
          :key="cat.id"
          type="button"
          :aria-pressed="selectedCategory === cat.id"
          class="h-8 shrink-0 rounded-lg px-3 text-xs font-medium transition"
          :class="
            selectedCategory === cat.id
              ? 'bg-[var(--gvp-text)] text-[var(--gvp-canvas)] font-semibold'
              : 'bg-[var(--gvp-elevated)] text-[var(--gvp-text)] hover:bg-[var(--gvp-border-strong)]'
          "
          @click="selectedCategory = cat.id"
        >
          {{ appStore.locale === 'ko' ? cat.labelKo : cat.labelEn }}
        </button>
      </div>
      <button type="button" aria-label="카테고리 오른쪽으로 스크롤" class="hidden h-8 w-8 shrink-0 items-center justify-center rounded-full text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] sm:inline-flex" @click="scrollChips(1)">›</button>
    </div>

    <!-- Toast Notification for Feed Feedback -->
    <div
      v-if="feedbackToast"
      class="mb-4 flex items-center justify-between rounded-lg border border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] px-4 py-3 text-sm text-[var(--gvp-brand-hover)]"
    >
      <span>{{ feedbackToast }}</span>
      <button
        type="button"
        class="text-xs font-semibold text-[var(--gvp-brand-hover)] hover:underline"
        @click="appStore.hiddenVideoIds.clear(); appStore.blockedChannelIds.clear(); feedbackToast = null"
      >
        {{ appStore.locale === 'ko' ? '실행 취소 (Undo)' : 'Undo' }}
      </button>
    </div>

    <!-- Sponsored Native Ad Slot Banner (Real Ad Decision Loop - ARTICLE 7) -->
    <section
      v-if="feedAd?.ad_available && feedAd?.creative"
      class="mb-8 overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-5 transition hover:bg-[var(--gvp-surface)]"
    >
      <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
        <div class="flex items-center gap-4">
          <img
            v-if="feedAd.creative.mediaUrl"
            :src="api.resolveUrl(feedAd.creative.mediaUrl)"
            :alt="feedAd.creative.title"
            class="h-16 w-16 shrink-0 rounded-lg border border-[var(--gvp-border)] object-cover"
          />
          <div>
            <div class="mb-1.5 flex items-center gap-2">
              <span class="rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-2 py-0.5 text-[10px] font-bold uppercase text-[var(--gvp-warning)]">
                Sponsored Ad
              </span>
              <span class="text-xs text-[var(--gvp-text-secondary)]">
                {{ feedAd.creative.advertiserName }} · Verified Ad Flight
              </span>
            </div>
            <h2 class="text-lg font-bold text-[var(--gvp-text)] sm:text-xl">
              {{ feedAd.creative.title }}
            </h2>
            <p class="mt-1 text-xs text-[var(--gvp-text-secondary)] sm:text-sm">
              {{ feedAd.creative.category }} · {{ feedAd.creative.callToAction }}
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="onFeedAdClick"
          class="shrink-0 rounded-md bg-[var(--gvp-brand)] px-5 py-2.5 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
        >
          {{ feedAd.creative.callToAction || 'Visit Site →' }}
        </button>
      </div>
    </section>
    <section v-else class="mb-8 overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-5">
      <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
        <div>
          <div class="mb-1.5 flex items-center gap-2">
            <span class="rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-2 py-0.5 text-[10px] font-bold uppercase text-[var(--gvp-warning)]">
              Sponsored Ad
            </span>
            <span class="text-xs text-[var(--gvp-text-secondary)]">StreamVerse Partner Network · CPM Verified</span>
          </div>
          <h2 class="text-lg font-bold text-[var(--gvp-text)] sm:text-xl">
            {{ appStore.locale === 'ko' ? 'StreamVerse Premium — 광고 없는 4K HDR 스트리밍 & 오프라인 저장' : 'StreamVerse Premium — Ad-Free 4K HDR & Offline Downloads' }}
          </h2>
          <p class="mt-1 text-xs text-[var(--gvp-text-secondary)] sm:text-sm">
            {{ appStore.locale === 'ko' ? '백그라운드 재생과 무손실 공간 음향을 지금 바로 경험하세요.' : 'Experience background playback, spatial audio, and zero interruptions.' }}
          </p>
        </div>
        <NuxtLink
          to="/premium"
          class="shrink-0 rounded-md bg-[var(--gvp-brand)] px-5 py-2.5 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
        >
          {{ appStore.locale === 'ko' ? '프리미엄 시작하기 →' : 'Try Premium Free →' }}
        </NuxtLink>
      </div>
    </section>

    <!-- Loading Skeleton State -->
    <div v-if="pending" class="grid grid-cols-1 gap-6 md:grid-cols-2 min-[1200px]:grid-cols-3 min-[1600px]:grid-cols-4">
      <div v-for="i in 8" :key="i" class="animate-pulse">
        <div class="aspect-video w-full rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
        <div class="mt-3 flex gap-3">
          <div class="h-9 w-9 shrink-0 rounded-full bg-[var(--gvp-elevated)]" />
          <div class="flex-1 space-y-2">
            <div class="h-4 w-4/5 rounded-md bg-[var(--gvp-elevated)]" />
            <div class="h-3 w-1/2 rounded-md bg-[var(--gvp-elevated)]" />
          </div>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center">
      <p class="text-base font-semibold text-[var(--gvp-live)]">
        {{ appStore.locale === 'ko' ? '피드를 불러오는 중 오류가 발생했습니다.' : 'Failed to load the home video feed.' }}
      </p>
      <button
        type="button"
        class="mt-3 rounded-md bg-[var(--gvp-brand)] px-4 py-2 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
        @click="refresh()"
      >
        {{ appStore.locale === 'ko' ? '다시 시도' : 'Retry' }}
      </button>
    </div>

    <!-- Empty State -->
    <div v-else-if="visibleVideos.length === 0" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-12 text-center">
      <p class="text-base font-semibold text-[var(--gvp-text-secondary)]">
        {{ appStore.locale === 'ko' ? '선택한 카테고리에 표시할 동영상이 없습니다.' : 'No videos match the selected category.' }}
      </p>
      <button
        type="button"
        class="mt-4 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-4 py-2 text-xs font-semibold text-[var(--gvp-text)] transition hover:bg-[var(--gvp-canvas)]"
        @click="selectedCategory = 'all'; appStore.hiddenVideoIds.clear(); appStore.blockedChannelIds.clear()"
      >
        {{ appStore.locale === 'ko' ? '필터 초기화' : 'Reset Filters' }}
      </button>
    </div>

    <!-- Main Video Grid -->
    <section v-else class="grid grid-cols-1 gap-x-5 gap-y-8 md:grid-cols-2 min-[1200px]:grid-cols-3 min-[1600px]:grid-cols-4">
      <FeedVideoCard
        v-for="video in visibleVideos"
        :key="video.id"
        :video="video"
        @toast="onCardToast"
      />
    </section>

    <!-- Live Now Shelf -->
    <section class="mt-12 border-t border-[var(--gvp-border)] pt-8">
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <span class="flex h-2.5 w-2.5 rounded-full bg-[var(--gvp-live)]" />
          <h2 class="text-lg font-bold text-[var(--gvp-text)]">
            {{ appStore.locale === 'ko' ? '지금 라이브 중 (Live Now)' : 'Live Now — Ultra-Low Latency HLS' }}
          </h2>
        </div>
        <NuxtLink to="/live/live_main_01" class="text-xs font-semibold text-[var(--gvp-brand-hover)] hover:underline">
          {{ appStore.locale === 'ko' ? '전체 보기 →' : 'View All Live →' }}
        </NuxtLink>
      </div>

      <div v-if="liveStreamsShelf.length" class="grid grid-cols-1 gap-5 md:grid-cols-2">
        <NuxtLink
          v-for="stream in liveStreamsShelf"
          :key="stream.id"
          :to="`/live/${stream.id}`"
          class="group flex flex-col overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] transition hover:bg-[var(--gvp-canvas)] sm:flex-row"
        >
          <div class="relative aspect-video w-full shrink-0 overflow-hidden bg-[var(--gvp-elevated)] sm:w-64">
            <img :src="stream.thumbnailUrl" :alt="stream.title" class="h-full w-full object-cover" />
            <span class="absolute left-2.5 top-2.5 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-2 py-0.5 text-[11px] font-bold uppercase text-[var(--gvp-live)] tabular-nums">
              LIVE · {{ stream.viewerCount.toLocaleString() }} watching
            </span>
          </div>
          <div class="flex flex-col justify-center p-4">
            <h3 class="font-bold leading-snug text-[var(--gvp-text)] group-hover:text-[var(--gvp-brand)]">{{ stream.title }}</h3>
            <span class="mt-3 inline-flex w-fit items-center gap-1.5 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-2.5 py-1 text-[11px] font-semibold text-[var(--gvp-text)]">
              <svg class="h-3.5 w-3.5 text-[var(--gvp-brand)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 10h.01M12 10h.01M16 10h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <span>SuperChat & DVR Enabled</span>
            </span>
          </div>
        </NuxtLink>
      </div>
    </section>

    <!-- Shorts Shelf -->
    <section class="mt-12 border-t border-[var(--gvp-border)] pt-8">
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="rounded-md border border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] px-2 py-0.5 text-xs font-bold text-[var(--gvp-brand-hover)]">SHORTS</span>
          <h2 class="text-lg font-bold text-[var(--gvp-text)]">
            {{ appStore.locale === 'ko' ? '인기 쇼츠 (Trending Shorts)' : 'Trending Vertical Shorts' }}
          </h2>
        </div>
        <NuxtLink to="/shorts" class="text-xs font-semibold text-[var(--gvp-brand-hover)] hover:underline">
          {{ appStore.locale === 'ko' ? '쇼츠 플레이어 열기 →' : 'Open Fullscreen Shorts →' }}
        </NuxtLink>
      </div>

      <div v-if="shortsShelf.length" class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-5">
        <NuxtLink
          v-for="shortItem in shortsShelf"
          :key="shortItem.id"
          :to="`/watch/${shortItem.id}`"
          class="group flex flex-col overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] transition hover:bg-[var(--gvp-canvas)]"
        >
          <div class="relative aspect-[9/16] w-full overflow-hidden bg-[var(--gvp-elevated)]">
            <img
              :src="shortItem.thumbnailUrl"
              :alt="shortItem.title"
              class="h-full w-full object-cover"
              loading="lazy"
            />
          </div>
          <div class="border-t border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-3">
            <p class="line-clamp-2 text-xs font-semibold leading-snug text-[var(--gvp-text)] group-hover:text-[var(--gvp-brand)]">{{ shortItem.title }}</p>
            <p class="mt-1 text-[11px] text-[var(--gvp-text-secondary)] tabular-nums">{{ formatViews(shortItem.viewCount) }}</p>
          </div>
        </NuxtLink>
      </div>
      <p v-else class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">
        {{ appStore.locale === 'ko' ? '표시할 쇼츠가 없습니다.' : 'No Shorts to show yet.' }}
      </p>
    </section>
  </div>
</template>
