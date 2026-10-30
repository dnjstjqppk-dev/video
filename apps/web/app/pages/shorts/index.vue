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
    title: '60초 만에 이해하는 HLS ABR 화질 자동 전환 & R2 멀티파트 업로드 ⚡ #golang #nuxt4 #streaming',
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
    title: '서울 한강 드론 라이트쇼 1000대 동시 비행 하이라이트 ✨ #seoul #drone #4k',
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
    title: '0.1초 반응속도 페이커 아지르 슈퍼토스 클러치 순간 🔥 #esports #t1 #gaming',
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
  { id: '1', user: '@dev_minho', text: '오디오/비디오 프리로딩 덕분에 스와이프 딜레이가 0ms네요! ⚡', time: '2시간 전', likes: 42 },
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
  <div class="flex h-[calc(100vh-3.5rem)] w-full items-center justify-center bg-[#090909] px-4">
    <!-- Preload Link for Next Short Poster/Segment -->
    <Head>
      <link v-if="nextShort" rel="preload" as="image" :href="nextShort.posterUrl" />
    </Head>

    <div v-if="currentShort" class="relative flex items-center gap-4">
      <!-- Vertical 9:16 Viewport Container -->
      <div
        class="relative h-[82vh] w-[46vh] max-w-[420px] select-none overflow-hidden rounded-3xl border border-white/15 bg-black shadow-2xl cursor-pointer"
        @click="togglePlayPause"
      >
        <img
          :src="currentShort.posterUrl"
          :alt="currentShort.title"
          class="h-full w-full object-cover"
        />

        <!-- Tap-to-Pause / Resume Animated HUD Center Overlay -->
        <transition enter-active-class="transition duration-200 ease-out" enter-from-class="scale-75 opacity-0" enter-to-class="scale-100 opacity-100" leave-active-class="transition duration-300 ease-in" leave-from-class="scale-100 opacity-100" leave-to-class="scale-125 opacity-0">
          <div
            v-if="showPlayPauseOverlay"
            class="pointer-events-none absolute inset-0 flex items-center justify-center"
          >
            <div class="flex h-20 w-20 items-center justify-center rounded-full bg-black/65 text-4xl text-white backdrop-blur">
              {{ isPlaying ? '▶' : '❚❚' }}
            </div>
          </div>
        </transition>

        <!-- Top Controls Overlay -->
        <div class="absolute inset-x-4 top-4 flex items-center justify-between" @click.stop>
          <span class="rounded-full bg-black/60 px-3 py-1 text-xs font-bold text-white backdrop-blur">
            SHORTS · {{ activeIndex + 1 }} / {{ shortsList.length }}
          </span>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="rounded-full bg-black/60 px-2.5 py-1 text-xs font-semibold text-white backdrop-blur"
              @click="loopEnabled = !loopEnabled"
            >
              {{ loopEnabled ? '🔁 Loop ON' : '➡️ Auto-Next' }}
            </button>
            <button
              type="button"
              class="rounded-full bg-black/60 p-2 text-white backdrop-blur"
              @click="isMuted = !isMuted"
            >
              {{ isMuted ? '🔇' : '🔊' }}
            </button>
          </div>
        </div>

        <!-- Bottom Metadata Overlay -->
        <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/95 via-black/65 to-transparent p-5 pr-16 text-white" @click.stop>
          <div class="flex items-center gap-2.5">
            <span v-if="currentShort.isSponsored" class="rounded bg-amber-400 px-2 py-0.5 text-[10px] font-black uppercase text-black">
              Sponsored
            </span>
            <NuxtLink :to="`/channels/${currentShort.channelId}`">
              <img :src="currentShort.channelAvatar" :alt="currentShort.channelHandle" class="h-9 w-9 rounded-full object-cover ring-2 ring-[#FF0033]" />
            </NuxtLink>
            <NuxtLink :to="`/channels/${currentShort.channelId}`" class="text-sm font-bold hover:underline">
              {{ currentShort.channelHandle }}
            </NuxtLink>
            <button
              v-if="!currentShort.isSponsored"
              type="button"
              class="rounded-full px-3 py-1 text-xs font-bold transition"
              :class="currentShort.subscribed ? 'bg-white/20 text-white' : 'bg-[#FF0033] text-white'"
              @click="toggleSubscribe(currentShort)"
            >
              {{ currentShort.subscribed ? 'Subscribed' : 'Subscribe' }}
            </button>
          </div>

          <p class="mt-3 text-sm font-medium leading-snug">
            {{ currentShort.title }}
          </p>

          <!-- Sponsored CTA Button -->
          <div v-if="currentShort.isSponsored && currentShort.destinationUrl" class="mt-2.5">
            <a
              :href="currentShort.destinationUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white shadow transition hover:bg-blue-500"
              @click="api.recordAdEvent({ decision_id: currentShort.decisionId || '', campaign_id: currentShort.campaignId || '', event_type: 'click' })"
            >
              <span>{{ currentShort.ctaText || 'Learn More' }}</span>
              <span>↗</span>
            </a>
          </div>

          <!-- Related Long-form VOD Pill -->
          <div v-if="currentShort.relatedVideoId" class="mt-2.5">
            <NuxtLink
              :to="`/watch/${currentShort.relatedVideoId}`"
              class="inline-flex max-w-[280px] items-center gap-1.5 truncate rounded-full bg-white/15 px-3 py-1 text-xs font-semibold text-white backdrop-blur transition hover:bg-white/25"
            >
              <span>▶</span>
              <span class="truncate">{{ currentShort.relatedVideoTitle }}</span>
              <span class="text-neutral-400">›</span>
            </NuxtLink>
          </div>

          <!-- Remix / Sound Attribution Pill with Spinning Disc -->
          <div
            class="mt-2.5 flex cursor-pointer items-center gap-2 group"
            @click="showSoundModal = !showSoundModal"
          >
            <div class="relative flex h-5 w-5 animate-spin items-center justify-center rounded-full bg-black/80 text-[10px] ring-1 ring-white/30" style="animation-duration: 4s;">
              💿
            </div>
            <div class="flex items-center gap-1.5 overflow-hidden text-xs text-neutral-300 transition group-hover:text-white">
              <span>♫</span>
              <span class="max-w-[220px] truncate">{{ currentShort.audioTitle }}</span>
            </div>
          </div>
        </div>

        <!-- Scrubbing Progress Bar at Bottom -->
        <div class="absolute inset-x-0 bottom-0 h-1 bg-white/20 transition-all hover:h-2">
          <div class="h-full bg-[#FF0033]" :style="{ width: `${playbackProgress}%` }" />
        </div>

        <!-- Right Vertical Engagement Rail -->
        <div class="absolute bottom-6 right-3 flex flex-col items-center gap-3.5 text-white" @click.stop>
          <!-- Like Button -->
          <button
            type="button"
            class="flex flex-col items-center gap-1"
            @click="toggleLike(currentShort)"
          >
            <span
              class="flex h-11 w-11 items-center justify-center rounded-full text-lg transition"
              :class="currentShort.liked ? 'bg-[#FF0033] text-white shadow-lg' : 'bg-black/60 hover:bg-white/20'"
            >
              👍
            </span>
            <span class="text-xs font-semibold">{{ (currentShort.likes / 1000).toFixed(1) }}K</span>
          </button>

          <!-- Dislike Button -->
          <button
            type="button"
            class="flex flex-col items-center gap-1"
            @click="toggleDislike(currentShort)"
          >
            <span
              class="flex h-11 w-11 items-center justify-center rounded-full text-lg transition"
              :class="currentShort.disliked ? 'bg-[#3EA6FF] text-white shadow-lg' : 'bg-black/60 hover:bg-white/20'"
            >
              👎
            </span>
            <span class="text-xs font-semibold">{{ appStore.locale === 'ko' ? '싫어요' : 'Dislike' }}</span>
          </button>

          <!-- Comments Drawer Button -->
          <button
            type="button"
            class="flex flex-col items-center gap-1"
            @click="showCommentsDrawer = !showCommentsDrawer"
          >
            <span class="flex h-11 w-11 items-center justify-center rounded-full bg-black/60 text-lg hover:bg-white/20">
              💬
            </span>
            <span class="text-xs font-semibold">{{ currentShort.comments }}</span>
          </button>

          <!-- Share Button -->
          <button
            type="button"
            class="flex flex-col items-center gap-1"
            @click="handleShare(currentShort)"
          >
            <span class="flex h-11 w-11 items-center justify-center rounded-full bg-black/60 text-lg hover:bg-white/20">
              ↗️
            </span>
            <span class="text-xs font-semibold">{{ shareCopied ? 'Copied!' : 'Share' }}</span>
          </button>

          <!-- Remix / Audio Button -->
          <button
            type="button"
            class="flex flex-col items-center gap-1"
            @click="showSoundModal = !showSoundModal"
          >
            <span class="flex h-11 w-11 items-center justify-center rounded-full bg-black/60 text-lg hover:bg-white/20">
              ✂️
            </span>
            <span class="text-xs font-semibold">Remix</span>
          </button>
        </div>
      </div>

      <!-- Up/Down Swipe Navigation Buttons -->
      <div class="flex flex-col gap-3">
        <button
          type="button"
          class="flex h-12 w-12 items-center justify-center rounded-full bg-[#272727] text-white shadow transition hover:bg-[#3A3A3A]"
          aria-label="Previous Short"
          @click="goPrev"
        >
          ↑
        </button>
        <button
          type="button"
          class="flex h-12 w-12 items-center justify-center rounded-full bg-[#272727] text-white shadow transition hover:bg-[#3A3A3A]"
          aria-label="Next Short"
          @click="goNext"
        >
          ↓
        </button>
      </div>

      <!-- Interactive Shorts Comments Drawer (Side Sheet) -->
      <div
        v-if="showCommentsDrawer"
        class="hidden h-[82vh] w-84 flex-col rounded-3xl border border-white/15 bg-[#181818] p-4 text-white lg:flex"
      >
        <div class="flex items-center justify-between border-b border-white/10 pb-3">
          <h3 class="text-sm font-bold">Shorts Comments ({{ currentShort.comments }})</h3>
          <button type="button" class="text-xs text-neutral-400 hover:text-white" @click="showCommentsDrawer = false">
            ✕
          </button>
        </div>

        <!-- Comments Feed -->
        <div class="mt-3 flex-1 space-y-3 overflow-y-auto pr-1 text-xs">
          <div
            v-for="cmt in shortComments"
            :key="cmt.id"
            class="rounded-xl bg-white/5 p-3"
          >
            <div class="flex items-center justify-between">
              <span class="font-semibold text-[#3EA6FF]">{{ cmt.user }}</span>
              <span class="text-[10px] text-neutral-400">{{ cmt.time }}</span>
            </div>
            <p class="mt-1 text-neutral-200">{{ cmt.text }}</p>
            <div class="mt-2 flex items-center gap-3 text-neutral-400">
              <button type="button" class="flex items-center gap-1 hover:text-white" @click="cmt.likes++">
                <span>👍</span>
                <span>{{ cmt.likes }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Add Comment Input Box -->
        <form class="mt-3 flex items-center gap-2 border-t border-white/10 pt-3" @submit.prevent="addShortComment">
          <input
            v-model="newShortComment"
            type="text"
            :placeholder="appStore.locale === 'ko' ? '댓글 추가...' : 'Add comment...'"
            class="flex-1 rounded-xl border border-white/15 bg-black/40 px-3 py-2 text-xs text-white outline-none focus:border-[#3EA6FF]"
          />
          <button
            type="submit"
            class="rounded-xl bg-[#3EA6FF] px-3.5 py-2 text-xs font-bold text-black transition hover:bg-[#65B8FF]"
          >
            {{ appStore.locale === 'ko' ? '게시' : 'Post' }}
          </button>
        </form>
      </div>

      <!-- Sound Attribution / Remix Sheet Modal -->
      <div
        v-if="showSoundModal"
        class="hidden h-[82vh] w-80 flex-col rounded-3xl border border-white/15 bg-[#181818] p-5 text-white lg:flex"
      >
        <div class="flex items-center justify-between border-b border-white/10 pb-3">
          <h3 class="text-sm font-bold">{{ appStore.locale === 'ko' ? '오디오 상세' : 'Audio Details' }}</h3>
          <button type="button" class="text-xs text-neutral-400 hover:text-white" @click="showSoundModal = false">
            ✕
          </button>
        </div>
        <div class="mt-4 flex flex-col items-center text-center">
          <div class="flex h-20 w-20 animate-spin items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500 text-3xl shadow-lg" style="animation-duration: 6s;">
            💿
          </div>
          <h4 class="mt-4 text-sm font-bold text-white">{{ currentShort.audioTitle }}</h4>
          <p class="mt-1 text-xs text-neutral-400">14.8K {{ appStore.locale === 'ko' ? '개의 Shorts에서 사용됨' : 'Shorts created' }}</p>
          <button
            type="button"
            class="mt-5 w-full rounded-full bg-white py-2 text-xs font-bold text-black shadow transition hover:bg-neutral-200"
            @click="showSoundModal = false"
          >
            {{ appStore.locale === 'ko' ? '이 사운드로 만들기' : 'Use this sound' }}
          </button>
        </div>
      </div>
    </div>
    <p v-else role="status" class="text-sm text-neutral-400">
      {{ appStore.locale === 'ko' ? '표시할 Shorts가 없습니다.' : 'No Shorts available.' }}
    </p>
  </div>
</template>
