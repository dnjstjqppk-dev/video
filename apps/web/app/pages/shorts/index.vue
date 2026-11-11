<script setup lang="ts">
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
const api = useApi()

useSeoMeta({
  title: 'Shorts — Vertical Fullscreen Feed',
  description: 'Swipe through high-energy vertical Shorts with instant preloading on StreamVerse.'
})

interface ShortClip {
  id: string
  channelId: string
  channelHandle: string
  channelAvatar: string
  title: string
  audioTitle: string
  likes: number
  comments: number
  posterUrl: string
  videoUrl: string
  liked?: boolean
  disliked?: boolean
  subscribed?: boolean
  relatedVideoId?: string
  relatedVideoTitle?: string
  isSponsored?: boolean
  ctaText?: string
  destinationUrl?: string
  campaignId?: string
  decisionId?: string
}

const shortsList = ref<ShortClip[]>([
  {
    id: 'short_01',
    channelId: 'ch_tech_kr',
    channelHandle: '@gvptech',
    channelAvatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&q=80',
    title: '60초 만에 이해하는 HLS ABR 화질 자동 전환 & R2 멀티파트 업로드 #golang #nuxt4 #streaming',
    audioTitle: 'Original Sound — GVP Tech Engineering',
    likes: 48200,
    comments: 912,
    posterUrl: 'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=720&q=80',
    videoUrl: 'https://media.gvp.media/hls/short_01/master.m3u8',
    relatedVideoId: 'vid_arch_2026',
    relatedVideoTitle: 'Building a 10M Concurrent Viewer Global Video Platform'
  },
  {
    id: 'short_02',
    channelId: 'ch_seoul_4k',
    channelHandle: '@seoul4k',
    channelAvatar: 'https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=100&q=80',
    title: '서울 한강 드론 라이트쇼 1000대 동시 비행 하이라이트 #seoul #drone #4k',
    audioTitle: 'Cyberpunk Horizon — Synthwave Mix',
    likes: 129400,
    comments: 2410,
    posterUrl: 'https://images.unsplash.com/photo-1509198397868-475647b2a1e5?w=720&q=80',
    videoUrl: 'https://media.gvp.media/hls/short_02/master.m3u8',
    relatedVideoId: 'vid_seoul_4k',
    relatedVideoTitle: '[4K 60fps HDR] Cyberpunk Rainy Night Walk in Gangnam & Hongdae'
  },
  {
    id: 'short_03',
    channelId: 'ch_esports_global',
    channelHandle: '@esports',
    channelAvatar: 'https://images.unsplash.com/photo-1566492031773-4f4e44671857?w=100&q=80',
    title: '0.1초 반응속도 페이커 아지르 슈퍼토스 클러치 순간 #esports #t1 #gaming',
    audioTitle: 'Arena Roar — Live Broadcast Audio',
    likes: 312000,
    comments: 5830,
    posterUrl: 'https://images.unsplash.com/photo-1542751371-adc38448a05e?w=720&q=80',
    videoUrl: 'https://media.gvp.media/hls/short_03/master.m3u8',
    relatedVideoId: 'vid_esports_finals',
    relatedVideoTitle: 'T1 vs Gen.G — Grand Finals Game 5 Full Highlights'
  }
])

onMounted(async () => {
  try {
    const adDecision = await api.getAdDecision({
      slot_id: 'shorts_sponsored',
      user_id: appStore.currentUser?.id,
      category: 'Technology',
      platform: 'WEB'
    })
    if (adDecision.ad_available && adDecision.creative) {
      const sponsoredClip: ShortClip = {
        id: 'ad_' + (adDecision.decision_id || 'sponsored'),
        channelId: 'ch_sponsor',
        channelHandle: '@' + (adDecision.creative.advertiserName?.toLowerCase().replace(/\s+/g, '') || 'sponsor'),
        channelAvatar: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=100&q=80',
        title: `[Sponsored] ${adDecision.creative.title} — ${adDecision.creative.category}`,
        audioTitle: 'Promoted Sound — ' + adDecision.creative.advertiserName,
        likes: 15400,
        comments: 320,
        posterUrl: adDecision.creative.mediaUrl,
        videoUrl: adDecision.creative.mediaUrl,
        isSponsored: true,
        ctaText: adDecision.creative.callToAction || 'Learn More',
        destinationUrl: adDecision.creative.destinationUrl,
        campaignId: adDecision.campaign_id,
        decisionId: adDecision.decision_id
      }
      shortsList.value.splice(1, 0, sponsoredClip)
    }
  } catch {
    // Non-blocking ad injection
  }
})

const activeIndex = ref(0)
const loopEnabled = ref(true)
const isMuted = ref(false)
const showCommentsDrawer = ref(false)
const isPlaying = ref(true)
const showPlayPauseOverlay = ref(false)
const shareCopied = ref(false)
const showSoundModal = ref(false)
const playbackProgress = ref(38) // current progress %
const newShortComment = ref('')

const shortComments = ref([
  { id: '1', user: '@dev_minho', text: '오디오/비디오 프리로딩 덕분에 스와이프 딜레이가 0ms네요!', time: '2시간 전', likes: 42 },
  { id: '2', user: '@cloud_architect', text: 'Separate Shorts view telemetry event firing cleanly.', time: '5시간 전', likes: 19 },
  { id: '3', user: '@frontend_pro', text: '9:16 가상화 뷰포트랑 햅틱 스와이프 감각 완벽하네요.', time: '1일 전', likes: 7 }
])

const currentShort = computed(() => shortsList.value[activeIndex.value] || shortsList.value[0])
const nextShort = computed(() => shortsList.value[(activeIndex.value + 1) % shortsList.value.length])

function togglePlayPause() {
  isPlaying.value = !isPlaying.value
  showPlayPauseOverlay.value = true
  setTimeout(() => {
    showPlayPauseOverlay.value = false
  }, 650)
}

function handleShare(clip: ShortClip) {
  const url = import.meta.client ? `${window.location.origin}/shorts#${clip.id}` : `https://gvp.media/shorts#${clip.id}`
  if (import.meta.client && navigator.clipboard) {
    navigator.clipboard.writeText(url)
  }
  shareCopied.value = true
  setTimeout(() => {
    shareCopied.value = false
  }, 2500)
}

function addShortComment() {
  const text = newShortComment.value.trim()
  if (!text) return
  shortComments.value.unshift({
    id: `sc_${Date.now()}`,
    user: `@${appStore.currentUser?.username || 'viewer'}`,
    text,
    time: '방금 전',
    likes: 0
  })
  if (currentShort.value) {
    currentShort.value.comments += 1
  }
  newShortComment.value = ''
}

async function sendShortsTelemetry(clip: ShortClip | undefined) {
  if (!clip) return
  try {
    if (clip.isSponsored && clip.decisionId && clip.campaignId) {
      await api.recordAdEvent({
        event_id: `evt_imp_${clip.id}_${Date.now()}`,
        decision_id: clip.decisionId,
        campaign_id: clip.campaignId,
        event_type: 'impression'
      })
    }
    await api.recordAnalyticsEvent({
      event_id: `evt_short_${Date.now()}`,
      event_type: 'valid_view',
      occurred_at: new Date().toISOString(),
      user_id: appStore.currentUser?.id,
      session_id: appStore.sessionId,
      video_id: clip.id,
      channel_id: clip.channelId,
      device_type: 'web_mobile',
      client_version: 'web-shorts-4.0.0',
      playback_session_id: `shorts_sess_${clip.id}_${Date.now()}`,
      watch_duration_seconds: 15,
      current_position_seconds: 15,
      video_duration_seconds: 60
    })
  } catch {
    // Non-blocking
  }
}

function goNext() {
  if (!shortsList.value.length) return
  activeIndex.value = (activeIndex.value + 1) % shortsList.value.length
  playbackProgress.value = 0
  sendShortsTelemetry(currentShort.value)
}

function goPrev() {
  if (!shortsList.value.length) return
  activeIndex.value = (activeIndex.value - 1 + shortsList.value.length) % shortsList.value.length
  playbackProgress.value = 0
  sendShortsTelemetry(currentShort.value)
}

function toggleLike(clip: ShortClip) {
  clip.liked = !clip.liked
  if (clip.liked && clip.disliked) {
    clip.disliked = false
  }
  clip.likes += clip.liked ? 1 : -1
}

function toggleDislike(clip: ShortClip) {
  clip.disliked = !clip.disliked
  if (clip.disliked && clip.liked) {
    clip.liked = false
    clip.likes = Math.max(0, clip.likes - 1)
  }
}

function toggleSubscribe(clip: ShortClip) {
  clip.subscribed = !clip.subscribed
}

function onKeydown(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
  if (e.key === 'ArrowDown' || e.key === 'j') {
    e.preventDefault()
    goNext()
  } else if (e.key === 'ArrowUp' || e.key === 'k') {
    e.preventDefault()
    goPrev()
  } else if (e.key === 'm') {
    isMuted.value = !isMuted.value
  } else if (e.key === ' ') {
    e.preventDefault()
    togglePlayPause()
  }
}

onMounted(() => {
  sendShortsTelemetry(currentShort.value)
  if (import.meta.client) {
    window.addEventListener('keydown', onKeydown)
  }
})

onUnmounted(() => {
  if (import.meta.client) {
    window.removeEventListener('keydown', onKeydown)
  }
})
</script>

<template>
  <div class="flex min-h-[calc(100vh-3.5rem)] w-full items-center justify-center bg-slate-50 px-4 py-6 text-slate-900">
    <!-- Preload Link for Next Short Poster/Segment -->
    <Head>
      <link v-if="nextShort" rel="preload" as="image" :href="nextShort.posterUrl" />
    </Head>

    <div v-if="currentShort" class="relative flex items-center gap-4">
      <!-- Vertical 9:16 Viewport Container -->
      <div
        class="relative flex h-[82vh] w-[46vh] max-w-[420px] select-none flex-col overflow-hidden rounded-lg border border-slate-200 bg-white cursor-pointer"
        @click="togglePlayPause"
      >
        <!-- Top Light Utility Bar -->
        <div class="flex items-center justify-between border-b border-slate-200 bg-white px-3.5 py-2.5" @click.stop>
          <span class="rounded-md border border-slate-200 bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-900 tabular-nums">
            SHORTS · {{ activeIndex + 1 }} / {{ shortsList.length }}
          </span>
          <div class="flex items-center gap-1.5">
            <button
              type="button"
              class="rounded-md border border-slate-200 bg-slate-50 px-2.5 py-1 text-xs font-semibold text-slate-700 transition hover:bg-slate-100"
              @click="loopEnabled = !loopEnabled"
            >
              {{ loopEnabled ? 'Loop: ON' : 'Auto-Next' }}
            </button>
            <button
              type="button"
              class="rounded-md border border-slate-200 bg-slate-50 p-1.5 text-slate-700 transition hover:bg-slate-100"
              :aria-label="isMuted ? 'Unmute' : 'Mute'"
              @click="isMuted = !isMuted"
            >
              <svg v-if="isMuted" class="h-4 w-4 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15zM17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" />
              </svg>
              <svg v-else class="h-4 w-4 text-slate-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Video Frame -->
        <div class="relative flex-1 overflow-hidden bg-slate-950">
          <img
            :src="currentShort.posterUrl"
            :alt="currentShort.title"
            class="h-full w-full object-cover"
          />

          <!-- Tap Pause / Resume HUD Center Overlay -->
          <div
            v-if="showPlayPauseOverlay"
            class="pointer-events-none absolute inset-0 flex items-center justify-center"
          >
            <div class="flex h-16 w-16 items-center justify-center rounded-lg border border-slate-200 bg-white/95 text-slate-900 shadow-sm">
              <svg v-if="isPlaying" class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z" />
              </svg>
              <svg v-else class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
              </svg>
            </div>
          </div>
        </div>

        <!-- Scrubbing Progress Bar -->
        <div class="h-1.5 w-full bg-slate-200">
          <div class="h-full bg-blue-600" :style="{ width: `${playbackProgress}%` }" />
        </div>

        <!-- Bottom Metadata Panel (Clean White Surface) -->
        <div class="border-t border-slate-200 bg-white p-4 text-slate-900" @click.stop>
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-2.5 min-w-0">
              <span v-if="currentShort.isSponsored" class="rounded-md border border-amber-200 bg-amber-50 px-2 py-0.5 text-[10px] font-bold uppercase text-amber-700">
                Sponsored
              </span>
              <NuxtLink :to="`/channels/${currentShort.channelId}`" class="shrink-0">
                <img :src="currentShort.channelAvatar" :alt="currentShort.channelHandle" class="h-8 w-8 rounded-full border border-slate-200 object-cover" />
              </NuxtLink>
              <NuxtLink :to="`/channels/${currentShort.channelId}`" class="truncate text-sm font-bold text-slate-900 hover:text-blue-600">
                {{ currentShort.channelHandle }}
              </NuxtLink>
            </div>
            <button
              v-if="!currentShort.isSponsored"
              type="button"
              class="shrink-0 rounded-md px-3 py-1 text-xs font-semibold transition"
              :class="currentShort.subscribed ? 'border border-slate-200 bg-slate-100 text-slate-700' : 'bg-slate-900 text-white hover:bg-slate-800'"
              @click="toggleSubscribe(currentShort)"
            >
              {{ currentShort.subscribed ? 'Subscribed' : 'Subscribe' }}
            </button>
          </div>

          <p class="mt-2.5 line-clamp-2 text-xs font-medium leading-relaxed text-slate-700">
            {{ currentShort.title }}
          </p>

          <!-- Sponsored CTA Button -->
          <div v-if="currentShort.isSponsored && currentShort.destinationUrl" class="mt-2.5">
            <a
              :href="currentShort.destinationUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3.5 py-1.5 text-xs font-semibold text-white transition hover:bg-blue-700"
              @click="api.recordAdEvent({ decision_id: currentShort.decisionId || '', campaign_id: currentShort.campaignId || '', event_type: 'click' })"
            >
              <span>{{ currentShort.ctaText || 'Learn More' }}</span>
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M7 17L17 7M17 7H7M17 7V17" />
              </svg>
            </a>
          </div>

          <!-- Related Long-form VOD Pill -->
          <div v-if="currentShort.relatedVideoId" class="mt-2.5">
            <NuxtLink
              :to="`/watch/${currentShort.relatedVideoId}`"
              class="inline-flex max-w-full items-center gap-1.5 truncate rounded-md border border-slate-200 bg-slate-50 px-2.5 py-1 text-xs font-medium text-slate-700 transition hover:bg-slate-100"
            >
              <svg class="h-3.5 w-3.5 shrink-0 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z" />
              </svg>
              <span class="truncate">{{ currentShort.relatedVideoTitle }}</span>
            </NuxtLink>
          </div>

          <!-- Remix / Sound Attribution Pill -->
          <button
            type="button"
            class="mt-2.5 flex w-full items-center gap-2 rounded-md border border-slate-200 bg-slate-50 px-2.5 py-1.5 text-left text-xs text-slate-600 transition hover:bg-slate-100 hover:text-slate-900"
            @click="showSoundModal = !showSoundModal"
          >
            <svg class="h-4 w-4 shrink-0 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
            </svg>
            <span class="truncate font-medium">{{ currentShort.audioTitle }}</span>
          </button>
        </div>
      </div>

      <!-- Right Vertical Engagement Rail & Navigation Controls -->
      <div class="flex flex-col items-center gap-3" @click.stop>
        <!-- Like Button -->
        <button
          type="button"
          class="flex w-16 flex-col items-center gap-1 rounded-lg border px-2 py-2.5 text-xs font-semibold transition"
          :class="currentShort.liked ? 'border-blue-200 bg-blue-50 text-blue-700' : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'"
          @click="toggleLike(currentShort)"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
          </svg>
          <span class="tabular-nums">{{ (currentShort.likes / 1000).toFixed(1) }}K</span>
        </button>

        <!-- Dislike Button -->
        <button
          type="button"
          class="flex w-16 flex-col items-center gap-1 rounded-lg border px-2 py-2.5 text-xs font-semibold transition"
          :class="currentShort.disliked ? 'border-blue-200 bg-blue-50 text-blue-700' : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'"
          @click="toggleDislike(currentShort)"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M10 14H5.236a2 2 0 01-1.789-2.894l3.5-7A2 2 0 018.736 3h4.018a2 2 0 01.485.06l3.76.94m-7 10v5a2 2 0 002 2h.096c.5 0 .905-.405.905-.904 0-.715.211-1.413.608-2.008L17 13V4m-7 10h2m5-10h2a2 2 0 012 2v6a2 2 0 01-2 2h-2.5" />
          </svg>
          <span>{{ appStore.locale === 'ko' ? '싫어요' : 'Dislike' }}</span>
        </button>

        <!-- Comments Drawer Button -->
        <button
          type="button"
          class="flex w-16 flex-col items-center gap-1 rounded-lg border border-slate-200 bg-white px-2 py-2.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-50"
          @click="showCommentsDrawer = !showCommentsDrawer"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 10h.01M12 10h.01M16 10h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          <span class="tabular-nums">{{ currentShort.comments }}</span>
        </button>

        <!-- Share Button -->
        <button
          type="button"
          class="flex w-16 flex-col items-center gap-1 rounded-lg border border-slate-200 bg-white px-2 py-2.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-50"
          @click="handleShare(currentShort)"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
          </svg>
          <span>{{ shareCopied ? 'Copied!' : 'Share' }}</span>
        </button>

        <!-- Remix / Audio Button -->
        <button
          type="button"
          class="flex w-16 flex-col items-center gap-1 rounded-lg border border-slate-200 bg-white px-2 py-2.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-50"
          @click="showSoundModal = !showSoundModal"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
          </svg>
          <span>Remix</span>
        </button>

        <!-- Up/Down Swipe Navigation Buttons -->
        <div class="mt-2 flex flex-col gap-2">
          <button
            type="button"
            class="flex h-10 w-16 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-700 transition hover:bg-slate-50"
            aria-label="Previous Short"
            @click="goPrev"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
            </svg>
          </button>
          <button
            type="button"
            class="flex h-10 w-16 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-700 transition hover:bg-slate-50"
            aria-label="Next Short"
            @click="goNext"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Interactive Shorts Comments Drawer (Side Sheet) -->
      <div
        v-if="showCommentsDrawer"
        class="hidden h-[82vh] w-80 flex-col rounded-lg border border-slate-200 bg-white p-4 text-slate-900 lg:flex"
      >
        <div class="flex items-center justify-between border-b border-slate-200 pb-3">
          <h3 class="text-sm font-bold text-slate-900 tabular-nums">Shorts Comments ({{ currentShort.comments }})</h3>
          <button type="button" class="rounded-md p-1 text-xs text-slate-500 hover:bg-slate-100 hover:text-slate-900" @click="showCommentsDrawer = false">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Comments Feed -->
        <div class="mt-3 flex-1 space-y-3 overflow-y-auto pr-1 text-xs">
          <div
            v-for="cmt in shortComments"
            :key="cmt.id"
            class="rounded-lg border border-slate-200 bg-slate-50 p-3"
          >
            <div class="flex items-center justify-between">
              <span class="font-semibold text-blue-700">{{ cmt.user }}</span>
              <span class="text-[10px] text-slate-500">{{ cmt.time }}</span>
            </div>
            <p class="mt-1 text-slate-700">{{ cmt.text }}</p>
            <div class="mt-2 flex items-center gap-3 text-slate-500">
              <button type="button" class="flex items-center gap-1 rounded-md px-1.5 py-0.5 hover:bg-slate-200 hover:text-slate-900 tabular-nums" @click="cmt.likes++">
                <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
                </svg>
                <span>{{ cmt.likes }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Add Comment Input Box -->
        <form class="mt-3 flex items-center gap-2 border-t border-slate-200 pt-3" @submit.prevent="addShortComment">
          <input
            v-model="newShortComment"
            type="text"
            :placeholder="appStore.locale === 'ko' ? '댓글 추가...' : 'Add comment...'"
            class="flex-1 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-900 outline-none focus:border-blue-600 focus:bg-white"
          />
          <button
            type="submit"
            class="rounded-md bg-slate-900 px-3.5 py-2 text-xs font-semibold text-white transition hover:bg-slate-800"
          >
            {{ appStore.locale === 'ko' ? '게시' : 'Post' }}
          </button>
        </form>
      </div>

      <!-- Sound Attribution / Remix Sheet Modal -->
      <div
        v-if="showSoundModal"
        class="hidden h-[82vh] w-80 flex-col rounded-lg border border-slate-200 bg-white p-5 text-slate-900 lg:flex"
      >
        <div class="flex items-center justify-between border-b border-slate-200 pb-3">
          <h3 class="text-sm font-bold text-slate-900">{{ appStore.locale === 'ko' ? '오디오 상세' : 'Audio Details' }}</h3>
          <button type="button" class="rounded-md p-1 text-xs text-slate-500 hover:bg-slate-100 hover:text-slate-900" @click="showSoundModal = false">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="mt-6 flex flex-col items-center text-center">
          <div class="flex h-16 w-16 items-center justify-center rounded-lg border border-blue-200 bg-blue-50 text-blue-700">
            <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
            </svg>
          </div>
          <h4 class="mt-4 text-sm font-bold text-slate-900">{{ currentShort.audioTitle }}</h4>
          <p class="mt-1 text-xs text-slate-500 tabular-nums">14.8K {{ appStore.locale === 'ko' ? '개의 Shorts에서 사용됨' : 'Shorts created' }}</p>
          <button
            type="button"
            class="mt-5 w-full rounded-md bg-slate-900 py-2 text-xs font-semibold text-white transition hover:bg-slate-800"
            @click="showSoundModal = false"
          >
            {{ appStore.locale === 'ko' ? '이 사운드로 만들기' : 'Use this sound' }}
          </button>
        </div>
      </div>
    </div>
    <p v-else role="status" class="text-sm text-slate-500">
      {{ appStore.locale === 'ko' ? '표시할 Shorts가 없습니다.' : 'No Shorts available.' }}
    </p>
  </div>
</template>
