<script setup lang="ts">
import type { Channel } from '@gvp/shared-types'
import { useAppStore } from '~/stores/app'

const route = useRoute()
const channelId = computed(() => String(route.params.id || 'ch_tech_kr'))
const api = useApi()
const appStore = useAppStore()

const { data: channelData, pending: channelPending, error: channelError } = await useAsyncData(`channel-${channelId.value}`, async () => {
  return await api.getChannel(channelId.value)
})

// Template v-if guards on channelData; the assertion keeps bindings typed.
const channel = computed(() => channelData.value as Channel)

useSeoMeta({
  title: () => channel.value ? `${channel.value.name} (${channel.value.handle}) — Channel` : 'Channel — StreamVerse',
  description: () => channel.value?.description || ''
})

const activeTab = ref<'videos' | 'shorts' | 'live' | 'playlists' | 'community'>('videos')
const isSubscribed = ref(false)
const pollSelection = ref<number | null>(null)
// Local-only widget: no fabricated base votes.
const pollRawVotes = ref([0, 0, 0])

const totalPollVotes = computed(() => pollRawVotes.value.reduce((a, b) => a + b, 0))
const pollPercentages = computed(() => {
  const tot = totalPollVotes.value || 1
  return pollRawVotes.value.map(v => Math.round((v / tot) * 100))
})

function votePoll(index: number) {
  if (pollSelection.value !== null || !Number.isInteger(index) || pollRawVotes.value[index] === undefined) return
  pollSelection.value = index
  pollRawVotes.value[index] += 1
}

function formatClock(totalSeconds: number): string {
  const sec = Math.max(0, Math.floor(totalSeconds || 0))
  return `${Math.floor(sec / 60)}:${String(sec % 60).padStart(2, '0')}`
}

const { data: uploadsData } = await useAsyncData(`channel-videos-${channelId.value}`, async () => {
  try {
    const res = await api.getChannelVideos(channelId.value, 1, 24)
    return res.items || []
  } catch {
    return []
  }
})
const channelVideos = computed(() => (uploadsData.value || []).filter(v => v.videoType !== 'shorts'))
const channelShorts = computed(() => (uploadsData.value || []).filter(v => v.videoType === 'shorts'))

const { data: liveData } = await useAsyncData(`channel-live-${channelId.value}`, async () => {
  try {
    const res = await api.listLiveStreams(1, 20)
    return (res.items || []).filter(s => s.channelId === channelId.value)
  } catch {
    return []
  }
})
const channelLive = computed(() => liveData.value || [])

</script>

<template>
  <div class="mx-auto max-w-[1600px] bg-white pb-12 text-slate-900">
    <div v-if="channelPending" class="animate-pulse">
      <div class="h-44 w-full bg-slate-100 sm:h-60" />
      <div class="mx-4 mt-6 h-8 w-1/3 rounded-md bg-slate-100 sm:mx-8" />
    </div>
    <div v-else-if="channelError || !channelData" class="mx-4 mt-10 rounded-lg border border-red-200 bg-red-50 p-12 text-center sm:mx-8" role="alert">
      <p class="text-base font-semibold text-red-700">
        {{ appStore.locale === 'ko' ? '채널을 불러오지 못했습니다.' : 'This channel is unavailable.' }}
      </p>
    </div>
    <template v-else>
    <!-- Channel Banner -->
    <div class="h-44 w-full overflow-hidden border-b border-slate-200 bg-slate-100 sm:h-60">
      <img :src="channel.bannerUrl" :alt="channel.name" class="h-full w-full object-cover" />
    </div>

    <div class="px-4 sm:px-8">
      <!-- Channel Header -->
      <div class="mt-6 flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
        <div class="flex items-center gap-5">
          <img
            :src="channel.avatarUrl"
            :alt="channel.name"
            class="h-20 w-20 rounded-full border-4 border-white object-cover ring-1 ring-slate-200 sm:h-28 sm:w-28"
          />
          <div>
            <div class="flex items-center gap-2">
              <h1 class="text-xl font-extrabold text-slate-900 sm:text-2xl">{{ channel.name }}</h1>
              <span v-if="channel.isVerified" class="rounded-md border border-blue-200 bg-blue-50 px-2 py-0.5 text-xs font-bold text-blue-700">
                Verified
              </span>
            </div>
            <p class="mt-1 text-xs text-slate-500 tabular-nums">
              {{ channel.handle }} · {{ channel.subscriberCount.toLocaleString() }} subscribers · {{ channel.videoCount }} videos
            </p>
            <p class="mt-2 max-w-2xl text-xs text-slate-600">
              {{ channel.description }}
            </p>
          </div>
        </div>

        <button
          type="button"
          class="rounded-md px-6 py-2.5 text-xs font-semibold transition"
          :class="isSubscribed ? 'border border-slate-200 bg-slate-100 text-slate-700 hover:bg-slate-200' : 'bg-slate-900 text-white hover:bg-slate-800'"
          @click="isSubscribed = !isSubscribed"
        >
          {{ isSubscribed ? (appStore.locale === 'ko' ? '구독중' : 'Subscribed') : (appStore.locale === 'ko' ? '구독하기' : 'Subscribe') }}
        </button>
      </div>

      <!-- Navigation Tabs -->
      <div class="mt-8 flex gap-6 border-b border-slate-200 text-sm font-semibold">
        <button
          v-for="tab in (['videos', 'shorts', 'live', 'playlists', 'community'] as const)"
          :key="tab"
          type="button"
          class="border-b-2 pb-3 capitalize transition"
          :class="activeTab === tab ? 'border-slate-900 text-slate-900' : 'border-transparent text-slate-500 hover:text-slate-900'"
          @click="activeTab = tab"
        >
          {{ tab }}
        </button>
      </div>

      <!-- Tab Content: Videos -->
      <p v-if="activeTab === 'videos' && channelVideos.length === 0" class="mt-6 rounded-lg border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
        {{ appStore.locale === 'ko' ? '업로드된 동영상이 없습니다.' : 'No videos uploaded yet.' }}
      </p>
      <div v-if="activeTab === 'videos' && channelVideos.length" class="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        <NuxtLink
          v-for="vid in channelVideos"
          :key="vid.id"
          :to="`/watch/${vid.id}`"
          class="group rounded-lg border border-slate-200 bg-white p-3 transition hover:bg-slate-50"
        >
          <div class="relative aspect-video overflow-hidden rounded-md border border-slate-200 bg-slate-100">
            <img :src="vid.thumbnailUrl" :alt="vid.title" class="h-full w-full object-cover" />
            <span class="absolute bottom-2 right-2 rounded-md border border-slate-200 bg-white/95 px-1.5 py-0.5 font-mono text-xs font-semibold text-slate-900 tabular-nums">
              {{ formatClock(vid.durationSeconds) }}
            </span>
          </div>
          <h3 class="mt-3 font-bold leading-snug text-slate-900 group-hover:text-blue-600">{{ vid.title }}</h3>
          <p class="mt-1 text-xs text-slate-500 tabular-nums">{{ vid.viewCount.toLocaleString() }} views</p>
        </NuxtLink>
      </div>

      <!-- Tab Content: Shorts -->
      <p v-else-if="activeTab === 'shorts' && channelShorts.length === 0" class="mt-6 rounded-lg border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
        {{ appStore.locale === 'ko' ? '업로드된 쇼츠가 없습니다.' : 'No Shorts uploaded yet.' }}
      </p>
      <div v-else-if="activeTab === 'shorts'" class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
        <NuxtLink
          v-for="vid in channelShorts"
          :key="vid.id"
          :to="`/watch/${vid.id}`"
          class="group flex flex-col overflow-hidden rounded-lg border border-slate-200 bg-white transition hover:bg-slate-50"
        >
          <div class="aspect-[9/16] w-full overflow-hidden bg-slate-100">
            <img :src="vid.thumbnailUrl" :alt="vid.title" class="h-full w-full object-cover" />
          </div>
          <div class="border-t border-slate-200 bg-white p-3 text-xs font-bold text-slate-900 tabular-nums group-hover:text-blue-600">
            {{ vid.title }}
          </div>
        </NuxtLink>
      </div>

      <!-- Tab Content: Live -->
      <p v-else-if="activeTab === 'live' && channelLive.length === 0" class="mt-6 rounded-lg border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
        {{ appStore.locale === 'ko' ? '현재 진행 중인 라이브가 없습니다.' : 'This channel is not live right now.' }}
      </p>
      <div v-else-if="activeTab === 'live'" class="mt-6 space-y-4">
        <NuxtLink
          v-for="stream in channelLive"
          :key="stream.id"
          :to="`/live/${stream.id}`"
          class="flex flex-col gap-4 rounded-lg border border-slate-200 bg-white p-4 transition hover:bg-slate-50 sm:flex-row"
        >
          <img :src="stream.thumbnailUrl" :alt="stream.title" class="aspect-video w-full rounded-md border border-slate-200 object-cover sm:w-72" />
          <div>
            <span class="rounded-md border border-red-200 bg-red-50 px-2 py-0.5 text-xs font-bold text-red-700">LIVE NOW</span>
            <h3 class="mt-2 text-lg font-bold text-slate-900">{{ stream.title }}</h3>
            <p class="mt-1 text-xs text-slate-500 tabular-nums">{{ stream.viewerCount.toLocaleString() }} concurrent viewers</p>
          </div>
        </NuxtLink>
      </div>

      <!-- Tab Content: Playlists -->
      <div v-else-if="activeTab === 'playlists'" class="mt-6 rounded-lg border border-slate-200 bg-slate-50 p-8 text-center text-xs text-slate-500">
        {{ appStore.locale === 'ko' ? '재생목록 연동 준비 중입니다.' : 'Playlist browsing is coming soon.' }}
      </div>

      <!-- Tab Content: Community Posts & Polls -->
      <div v-else class="mt-6 max-w-2xl rounded-lg border border-slate-200 bg-white p-5">
        <div class="flex items-center gap-3">
          <img :src="channel.avatarUrl" :alt="channel.name" class="h-10 w-10 rounded-full border border-slate-200 object-cover" />
          <div>
            <p class="text-sm font-bold text-slate-900">{{ channel.name }}</p>
            <p class="text-xs text-slate-500 tabular-nums">2 hours ago · Community Poll</p>
          </div>
        </div>
        <p class="mt-3 text-sm text-slate-700">
          다음 주 기술 딥다이브 영상 주제로 어떤 아키텍처를 가장 보고 싶으신가요? 투표해 주세요!
        </p>
        <div class="mt-4 space-y-2">
          <button
            v-for="(option, idx) in ['AV1 Hardware Encoding on GPU Fleet', 'ClickHouse Real-Time Creator Analytics', 'WebTransport Low-Latency Live Ingest']"
            :key="option"
            type="button"
            class="relative flex w-full items-center justify-between overflow-hidden rounded-md border px-4 py-3 text-left text-xs font-semibold transition"
            :class="pollSelection === idx ? 'border-blue-200 bg-blue-50 text-blue-700' : 'border-slate-200 bg-slate-50 hover:bg-slate-100 text-slate-700'"
            @click="votePoll(idx)"
          >
            <div
              v-if="pollSelection !== null"
              class="absolute inset-y-0 left-0 bg-blue-100/80 transition-all duration-300"
              :style="{ width: `${pollPercentages[idx]}%` }"
            />
            <span class="relative z-10">{{ option }}</span>
            <span v-if="pollSelection !== null" class="relative z-10 font-mono font-bold tabular-nums">{{ pollPercentages[idx] }}%</span>
          </button>
        </div>
        <p v-if="pollSelection !== null" class="mt-2 text-[11px] font-medium text-slate-500 tabular-nums">
          {{ totalPollVotes.toLocaleString() }} {{ appStore.locale === 'ko' ? '표' : 'votes' }}
        </p>
      </div>
    </div>
    </template>
  </div>
</template>
