<script setup lang="ts">
import type { Channel } from '@gvp/shared-types'
import { useAppStore } from '~/stores/app'

const route = useRoute()
const channelId = computed(() => String(route.params.id || 'ch_tech_kr'))
const api = useApi()
const appStore = useAppStore()

const fallbackChannel: Channel = {
  id: channelId.value,
  ownerId: 'usr_1',
  handle: '@gvptech',
  name: 'GVP Tech Engineering',
  description:
    'Official engineering channel of StreamVerse Global Video Platform. Sharing deep dives on Go 1.24, Nuxt 4, Flutter 3, OpenSearch, and distributed HLS transcoding.',
  avatarUrl: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&q=80',
  bannerUrl: 'https://images.unsplash.com/photo-1518770660439-4636190af475?w=1600&q=80',
  subscriberCount: 284000,
  videoCount: 142,
  isVerified: true,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2026-10-20T00:00:00Z'
}

const { data: channelData } = await useAsyncData(`channel-${channelId.value}`, async () => {
  try {
    return await api.getChannel(channelId.value)
  } catch {
    return fallbackChannel
  }
})

const channel = computed(() => channelData.value || fallbackChannel)

useSeoMeta({
  title: () => `${channel.value.name} (${channel.value.handle}) — Channel`,
  description: () => channel.value.description || ''
})

const activeTab = ref<'videos' | 'shorts' | 'live' | 'playlists' | 'community'>('videos')
const isSubscribed = ref(false)
const pollSelection = ref<number | null>(null)
const pollRawVotes = ref([640, 210, 150])

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

const channelVideos = [
  {
    id: 'vid_arch_2026',
    title: 'Building a 10M Concurrent Viewer Global Video Platform with Go 1.24, Nuxt 4 & ClickHouse',
    views: '412.9K views',
    duration: '30:45',
    thumbnailUrl: 'https://images.unsplash.com/photo-1518770660439-4636190af475?w=700&q=80'
  },
  {
    id: 'vid_ai_agents',
    title: 'Multimodal Video Indexing with Audio-Visual Fingerprints & Vector Embeddings',
    views: '189.4K views',
    duration: '18:40',
    thumbnailUrl: 'https://images.unsplash.com/photo-1620712943543-bcc4688e7485?w=700&q=80'
  },
  {
    id: 'vid_seoul_4k',
    title: 'Zero-Downtime PostgreSQL 18 Partitioning for Double-Entry Settlement Ledgers',
    views: '98.2K views',
    duration: '26:15',
    thumbnailUrl: 'https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=700&q=80'
  }
]

const channelPlaylists = [
  {
    id: 'pl_01',
    title: 'Distributed Video Infrastructure Masterclass (12 Videos)',
    videoCount: 12,
    thumbnailUrl: 'https://images.unsplash.com/photo-1518770660439-4636190af475?w=600&q=80'
  },
  {
    id: 'pl_02',
    title: 'Nuxt 4 & Vue 3 Performance Patterns (8 Videos)',
    videoCount: 8,
    thumbnailUrl: 'https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=600&q=80'
  }
]
</script>

<template>
  <div class="mx-auto max-w-[1600px] pb-12">
    <!-- Channel Banner -->
    <div class="h-44 w-full overflow-hidden bg-[#1A1A1A] sm:h-60">
      <img :src="channel.bannerUrl" :alt="channel.name" class="h-full w-full object-cover" />
    </div>

    <div class="px-4 sm:px-8">
      <!-- Channel Header -->
      <div class="mt-6 flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
        <div class="flex items-center gap-5">
          <img
            :src="channel.avatarUrl"
            :alt="channel.name"
            class="h-20 w-20 rounded-full border-4 border-[#0F0F0F] object-cover sm:h-28 sm:w-28"
          />
          <div>
            <div class="flex items-center gap-2">
              <h1 class="text-xl font-extrabold sm:text-2xl">{{ channel.name }}</h1>
              <span v-if="channel.isVerified" class="rounded-full bg-[#3EA6FF]/20 px-2 py-0.5 text-xs font-bold text-[#3EA6FF]">
                ✓ Verified
              </span>
            </div>
            <p class="mt-1 text-xs text-neutral-400">
              {{ channel.handle }} · {{ channel.subscriberCount.toLocaleString() }} subscribers · {{ channel.videoCount }} videos
            </p>
            <p class="mt-2 max-w-2xl text-xs text-neutral-300">
              {{ channel.description }}
            </p>
          </div>
        </div>

        <button
          type="button"
          class="rounded-full px-6 py-2.5 text-xs font-bold transition"
          :class="isSubscribed ? 'bg-[#272727] text-white' : 'bg-white text-black hover:bg-neutral-200'"
          @click="isSubscribed = !isSubscribed"
        >
          {{ isSubscribed ? (appStore.locale === 'ko' ? '구독중 ✓' : 'Subscribed ✓') : (appStore.locale === 'ko' ? '구독하기' : 'Subscribe') }}
        </button>
      </div>

      <!-- Navigation Tabs -->
      <div class="mt-8 flex gap-6 border-b border-white/10 text-sm font-semibold">
        <button
          v-for="tab in (['videos', 'shorts', 'live', 'playlists', 'community'] as const)"
          :key="tab"
          type="button"
          class="border-b-2 pb-3 capitalize transition"
          :class="activeTab === tab ? 'border-white text-white' : 'border-transparent text-neutral-400 hover:text-white'"
          @click="activeTab = tab"
        >
          {{ tab }}
        </button>
      </div>

      <!-- Tab Content: Videos -->
      <div v-if="activeTab === 'videos'" class="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        <NuxtLink
          v-for="vid in channelVideos"
          :key="vid.id"
          :to="`/watch/${vid.id}`"
          class="group rounded-2xl border border-white/10 bg-[#161616] p-3 transition hover:border-white/25"
        >
          <div class="relative aspect-video overflow-hidden rounded-xl">
            <img :src="vid.thumbnailUrl" :alt="vid.title" class="h-full w-full object-cover transition group-hover:scale-105" />
            <span class="absolute bottom-2 right-2 rounded bg-black/85 px-1.5 py-0.5 font-mono text-xs text-white">
              {{ vid.duration }}
            </span>
          </div>
          <h3 class="mt-3 font-bold leading-snug text-white group-hover:text-[#3EA6FF]">{{ vid.title }}</h3>
          <p class="mt-1 text-xs text-neutral-400">{{ vid.views }}</p>
        </NuxtLink>
      </div>

      <!-- Tab Content: Shorts -->
      <div v-else-if="activeTab === 'shorts'" class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
        <NuxtLink
          v-for="i in 4"
          :key="i"
          to="/shorts"
          class="relative aspect-[9/16] overflow-hidden rounded-2xl bg-[#1F1F1F]"
        >
          <img src="https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=500&q=80" alt="Short" class="h-full w-full object-cover" />
          <div class="absolute inset-x-3 bottom-3 text-xs font-bold text-white">
            HLS Segment Caching Explained #{{ i }}
          </div>
        </NuxtLink>
      </div>

      <!-- Tab Content: Live -->
      <div v-else-if="activeTab === 'live'" class="mt-6">
        <NuxtLink
          to="/live/live_main_01"
          class="flex flex-col gap-4 rounded-2xl border border-[#FF0033]/40 bg-[#181818] p-4 sm:flex-row"
        >
          <img src="https://images.unsplash.com/photo-1540575467063-178a50c2df87?w=600&q=80" alt="Live" class="aspect-video w-full rounded-xl object-cover sm:w-72" />
          <div>
            <span class="rounded bg-[#FF0033] px-2 py-0.5 text-xs font-bold text-white">LIVE NOW</span>
            <h3 class="mt-2 text-lg font-bold text-white">[LIVE] 2026 Global Architecture Summit — Keynote & Live Q&A</h3>
            <p class="mt-1 text-xs text-neutral-400">14,820 concurrent viewers · Ultra-low latency HLS + WebSocket Chat</p>
          </div>
        </NuxtLink>
      </div>

      <!-- Tab Content: Playlists -->
      <div v-else-if="activeTab === 'playlists'" class="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-2">
        <div
          v-for="pl in channelPlaylists"
          :key="pl.id"
          class="flex items-center gap-4 rounded-2xl border border-white/10 bg-[#161616] p-4"
        >
          <img :src="pl.thumbnailUrl" :alt="pl.title" class="h-24 w-40 rounded-xl object-cover" />
          <div>
            <h3 class="font-bold text-white">{{ pl.title }}</h3>
            <p class="mt-1 text-xs text-[#3EA6FF]">{{ pl.videoCount }} videos in playlist</p>
          </div>
        </div>
      </div>

      <!-- Tab Content: Community Posts & Polls -->
      <div v-else class="mt-6 max-w-2xl rounded-2xl border border-white/10 bg-[#181818] p-5">
        <div class="flex items-center gap-3">
          <img :src="channel.avatarUrl" :alt="channel.name" class="h-10 w-10 rounded-full object-cover" />
          <div>
            <p class="text-sm font-bold">{{ channel.name }}</p>
            <p class="text-xs text-neutral-400">2 hours ago · Community Poll</p>
          </div>
        </div>
        <p class="mt-3 text-sm text-neutral-200">
          다음 주 기술 딥다이브 영상 주제로 어떤 아키텍처를 가장 보고 싶으신가요? 투표해 주세요!
        </p>
        <div class="mt-4 space-y-2">
          <button
            v-for="(option, idx) in ['AV1 Hardware Encoding on GPU Fleet', 'ClickHouse Real-Time Creator Analytics', 'WebTransport Low-Latency Live Ingest']"
            :key="option"
            type="button"
            class="relative flex w-full items-center justify-between overflow-hidden rounded-xl border px-4 py-3 text-left text-xs font-semibold transition"
            :class="pollSelection === idx ? 'border-[#3EA6FF] text-white' : 'border-white/15 bg-white/5 hover:bg-white/10 text-neutral-200'"
            @click="votePoll(idx)"
          >
            <div
              v-if="pollSelection !== null"
              class="absolute inset-y-0 left-0 bg-[#3EA6FF]/20 transition-all duration-500 ease-out"
              :style="{ width: `${pollPercentages[idx]}%` }"
            />
            <span class="relative z-10">{{ option }}</span>
            <span v-if="pollSelection !== null" class="relative z-10 font-mono font-bold">{{ pollPercentages[idx] }}%</span>
          </button>
        </div>
        <p v-if="pollSelection !== null" class="mt-2 text-[11px] text-neutral-400">
          {{ totalPollVotes.toLocaleString() }} {{ appStore.locale === 'ko' ? '표' : 'votes' }}
        </p>
      </div>
    </div>
  </div>
</template>
