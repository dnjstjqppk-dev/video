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
  subscribed?: boolean
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
    videoUrl: 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8'
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
    videoUrl: 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8'
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
    videoUrl: 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8'
  }
])

const activeIndex = ref(0)
const loopEnabled = ref(true)
const isMuted = ref(false)
const showCommentsDrawer = ref(false)

const currentShort = computed(() => shortsList.value[activeIndex.value] || shortsList.value[0])
const nextShort = computed(() => shortsList.value[(activeIndex.value + 1) % shortsList.value.length])

async function sendShortsTelemetry(clip: ShortClip) {
  try {
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
  activeIndex.value = (activeIndex.value + 1) % shortsList.value.length
  sendShortsTelemetry(currentShort.value)
}

function goPrev() {
  activeIndex.value = (activeIndex.value - 1 + shortsList.value.length) % shortsList.value.length
  sendShortsTelemetry(currentShort.value)
}

function toggleLike(clip: ShortClip) {
  clip.liked = !clip.liked
  clip.likes += clip.liked ? 1 : -1
}

function toggleSubscribe(clip: ShortClip) {
  clip.subscribed = !clip.subscribed
}

onMounted(() => {
  sendShortsTelemetry(currentShort.value)
})
</script>

<template>
  <div class="flex h-[calc(100vh-3.5rem)] w-full items-center justify-center bg-[#090909] px-4">
    <!-- Preload Link for Next Short Poster/Segment -->
    <Head>
      <link v-if="nextShort" rel="preload" as="image" :href="nextShort.posterUrl" />
    </Head>

    <div class="relative flex items-center gap-4">
      <!-- Vertical 9:16 Viewport Container -->
      <div class="relative h-[82vh] w-[46vh] max-w-[420px] overflow-hidden rounded-3xl border border-white/15 bg-black shadow-2xl">
        <img
          :src="currentShort.posterUrl"
          :alt="currentShort.title"
          class="h-full w-full object-cover"
        />

        <!-- Top Controls Overlay -->
        <div class="absolute inset-x-4 top-4 flex items-center justify-between">
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
        <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/95 via-black/60 to-transparent p-5 pr-16 text-white">
          <div class="flex items-center gap-2.5">
            <NuxtLink :to="`/channels/${currentShort.channelId}`">
              <img :src="currentShort.channelAvatar" :alt="currentShort.channelHandle" class="h-9 w-9 rounded-full object-cover ring-2 ring-[#FF0033]" />
            </NuxtLink>
            <NuxtLink :to="`/channels/${currentShort.channelId}`" class="text-sm font-bold hover:underline">
              {{ currentShort.channelHandle }}
            </NuxtLink>
            <button
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
          <p class="mt-2 flex items-center gap-1.5 text-xs text-neutral-300">
            <span>🎵</span>
            <span class="truncate">{{ currentShort.audioTitle }}</span>
          </p>
        </div>

        <!-- Right Vertical Engagement Rail -->
        <div class="absolute bottom-6 right-3 flex flex-col items-center gap-4 text-white">
          <button
            type="button"
            class="flex flex-col items-center gap-1"
            @click="toggleLike(currentShort)"
          >
            <span
              class="flex h-11 w-11 items-center justify-center rounded-full text-lg transition"
              :class="currentShort.liked ? 'bg-[#FF0033] text-white' : 'bg-black/60 hover:bg-white/20'"
            >
              👍
            </span>
            <span class="text-xs font-semibold">{{ (currentShort.likes / 1000).toFixed(1) }}K</span>
          </button>

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

          <button
            type="button"
            class="flex flex-col items-center gap-1"
          >
            <span class="flex h-11 w-11 items-center justify-center rounded-full bg-black/60 text-lg hover:bg-white/20">
              ↗️
            </span>
            <span class="text-xs font-semibold">Share</span>
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

      <!-- Optional Shorts Comments Drawer -->
      <div
        v-if="showCommentsDrawer"
        class="hidden h-[82vh] w-80 flex-col rounded-3xl border border-white/15 bg-[#181818] p-4 text-white lg:flex"
      >
        <div class="flex items-center justify-between border-b border-white/10 pb-3">
          <h3 class="text-sm font-bold">Shorts Comments ({{ currentShort.comments }})</h3>
          <button type="button" class="text-xs text-neutral-400 hover:text-white" @click="showCommentsDrawer = false">
            ✕
          </button>
        </div>
        <div class="mt-3 flex-1 space-y-3 overflow-y-auto text-xs">
          <div class="rounded-xl bg-white/5 p-3">
            <p class="font-semibold text-[#3EA6FF]">@dev_minho</p>
            <p class="mt-1 text-neutral-200">오디오/비디오 프리로딩 덕분에 스와이프 딜레이가 0ms네요!</p>
          </div>
          <div class="rounded-xl bg-white/5 p-3">
            <p class="font-semibold text-[#3EA6FF]">@cloud_architect</p>
            <p class="mt-1 text-neutral-200">Separate Shorts view telemetry event firing cleanly.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
