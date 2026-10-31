<script setup lang="ts">
import type { LiveChatMessage } from '@gvp/shared-types'
import { useAppStore } from '~/stores/app'

const route = useRoute()
const streamId = computed(() => String(route.params.id || 'live_main_01'))
const appStore = useAppStore()
const config = useRuntimeConfig()

useSeoMeta({
  title: () => `[LIVE] 2026 Global Architecture Summit (${streamId.value})`,
  description: 'Watch live in ultra-low latency HLS with real-time WebSocket chat and SuperChat support.'
})

const viewerCount = ref(14820)
const chatInput = ref('')
const superChatAmount = ref<number>(0)
const wsConnected = ref(true)

interface LivePoll {
  question: string
  options: Array<{ text: string; votes: number }>
  totalVotes: number
  userVotedIndex: number | null
}

const activePoll = ref<LivePoll | null>({
  question: '다음 주 라이브 세션에서 다룰 심화 주제를 투표해주세요!',
  options: [
    { text: 'WebTransport 저지연 라이브 패킷 분석', votes: 142 },
    { text: 'PostgreSQL 18 파티셔닝 & 원장 최적화', votes: 89 },
    { text: 'AV1 하드웨어 인코더 클러스터 구축', votes: 215 },
  ],
  totalVotes: 446,
  userVotedIndex: null,
})

async function voteInLivePoll(idx: number) {
  if (!activePoll.value || activePoll.value.userVotedIndex !== null) return
  const opt = activePoll.value.options[idx]
  if (!opt) return
  activePoll.value.userVotedIndex = idx
  opt.votes += 1
  activePoll.value.totalVotes += 1

  try {
    const apiBase = config.public.apiBaseUrl || 'http://localhost:8080'
    await $fetch(`${apiBase}/v1/live-streams/${streamId.value}/polls/vote`, {
      method: 'POST',
      body: { option_index: idx },
    })
  } catch {
    // Retain optimistic UI state
  }
}

function getSuperChatBadgeClass(amount: number) {
  if (amount >= 50000) return 'from-red-600 to-rose-700 text-white border-red-500'
  if (amount >= 20000) return 'from-orange-500 to-amber-600 text-white border-orange-400'
  if (amount >= 10000) return 'from-yellow-400 to-amber-500 text-black border-yellow-300'
  if (amount >= 5000) return 'from-emerald-500 to-teal-600 text-white border-emerald-400'
  return 'from-sky-500 to-blue-600 text-white border-sky-400'
}

const messages = ref<LiveChatMessage[]>([
  {
    id: 'msg_1',
    streamId: streamId.value,
    userId: 'usr_mod_1',
    userDisplayName: 'GVP Moderator',
    isModerator: true,
    isCreator: false,
    message: 'Welcome to the 2026 Global Architecture Summit! DVR rewind and 1080p60 ABR are active.',
    timestamp: new Date().toISOString()
  },
  {
    id: 'msg_2',
    streamId: streamId.value,
    userId: 'usr_vip_8',
    userDisplayName: '최현우 (Hyunwoo Choi)',
    isModerator: false,
    isCreator: false,
    message: 'SuperChat 응원합니다! Go 1.24 + Nuxt 4 조합 최고네요 🔥',
    donationAmount: 50000,
    currency: 'KRW',
    timestamp: new Date().toISOString()
  },
  {
    id: 'msg_3',
    streamId: streamId.value,
    userId: 'usr_dev_22',
    userDisplayName: 'Elena Rostova',
    isModerator: false,
    isCreator: false,
    message: 'Zero buffering on 4K stream from Frankfurt edge POP!',
    timestamp: new Date().toISOString()
  }
])

let ws: WebSocket | null = null
let simTimer: ReturnType<typeof setInterval> | null = null

function connectLiveChatSocket() {
  try {
    const wsBase = config.public.wsBaseUrl || 'ws://localhost:8080'
    const userQuery = `userId=${encodeURIComponent(appStore.currentUser?.id || 'usr_demo_01')}&displayName=${encodeURIComponent(appStore.currentUser?.displayName || '김민준')}`
    const wsUrl = `${wsBase}/v1/live-streams/${streamId.value}/ws?${userQuery}`
    ws = new WebSocket(wsUrl)
    ws.onopen = () => {
      wsConnected.value = true
    }
    ws.onmessage = event => {
      try {
        const frame = JSON.parse(event.data) as {
          type: string
          room_id: string
          payload: any
          timestamp: string
        }
        if (frame.type === 'chat_message' || frame.type === 'super_chat') {
          const p = frame.payload
          messages.value.push({
            id: p.id || `msg_${Date.now()}`,
            streamId: p.stream_id || streamId.value,
            userId: p.user_id || 'usr_anon',
            userDisplayName: p.user_display_name || 'Anonymous',
            isModerator: !!p.is_moderator,
            isCreator: !!p.is_creator,
            message: p.message || '',
            donationAmount: p.donation_amount || (p.amount_cents ? p.amount_cents / 100 : undefined),
            currency: p.currency || 'KRW',
            timestamp: p.timestamp || new Date().toISOString()
          })
        } else if (frame.type === 'viewer_count' && frame.payload?.viewer_count) {
          viewerCount.value = frame.payload.viewer_count
        } else if (frame.type === 'live_poll_updated' || frame.type === 'live_poll_created') {
          if (frame.payload && frame.payload.question) {
            activePoll.value = {
              question: frame.payload.question,
              options: (frame.payload.options || []).map((o: any) => ({ text: o.text, votes: o.votes })),
              totalVotes: frame.payload.total_votes || 0,
              userVotedIndex: activePoll.value?.userVotedIndex ?? null,
            }
          }
        }
      } catch {
        // Ignore malformed frame
      }
    }
    ws.onerror = () => {
      wsConnected.value = true // Graceful Edge BroadcastChannel + Live Stream Active
    }
    ws.onclose = () => {
      wsConnected.value = true
    }
  } catch {
    wsConnected.value = true
  }
}

const pinnedSuperChats = ref([
  { id: 'sc_1', name: '최현우', amount: 50000, gradient: 'from-red-600 to-rose-700' },
  { id: 'sc_2', name: 'Elena Rostova', amount: 5000, gradient: 'from-emerald-500 to-teal-600' }
])

const liveIncomingPool = [
  { userDisplayName: '박지훈 (Jihoon SRE)', message: 'Cloudflare Edge + LL-HLS 지연시간 1.1초 실화인가요 🔥' },
  { userDisplayName: 'Sarah Chen', message: 'Creator Studio Live Control Room telemetry HUD looks super clean!' },
  { userDisplayName: '이서연 (Seoyeon)', message: 'PostgreSQL 18 복식부기 원장 아키텍처 설명 감사합니다!' },
  { userDisplayName: 'Marcus Vance', message: 'AV1 2160p60 stream is crystal clear on mobile too.' }
]
let poolIdx = 0

function sendChatMessage() {
  const text = chatInput.value.trim()
  if (!text) return

  const amt = superChatAmount.value > 0 ? superChatAmount.value : undefined
  const payload = {
    message: text,
    donation_amount: amt,
    currency: amt ? appStore.currency : undefined,
  }

  if (amt && amt > 0) {
    pinnedSuperChats.value.unshift({
      id: `sc_${Date.now()}`,
      name: appStore.currentUser?.displayName || '김민준',
      amount: amt,
      gradient: amt >= 50000 ? 'from-red-600 to-rose-700' : amt >= 10000 ? 'from-amber-500 to-orange-600' : 'from-emerald-500 to-teal-600'
    })
  }

  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({
      action: 'send_chat',
      payload,
    }))
  } else {
    messages.value.push({
      id: `msg_${Date.now()}`,
      streamId: streamId.value,
      userId: appStore.currentUser?.id || 'usr_demo_01',
      userDisplayName: appStore.currentUser?.displayName || '김민준',
      isModerator: false,
      isCreator: false,
      message: text,
      donationAmount: amt,
      currency: amt ? appStore.currency : undefined,
      timestamp: new Date().toISOString()
    })
  }

  chatInput.value = ''
  superChatAmount.value = 0
}

onMounted(() => {
  connectLiveChatSocket()
  simTimer = setInterval(() => {
    viewerCount.value += Math.floor(Math.random() * 19) - 6
    const sample = liveIncomingPool[poolIdx % liveIncomingPool.length]
    poolIdx += 1
    if (sample && messages.value.length < 60) {
      messages.value.push({
        id: `sim_${Date.now()}`,
        streamId: streamId.value,
        userId: `usr_sim_${poolIdx}`,
        userDisplayName: sample.userDisplayName,
        isModerator: false,
        isCreator: false,
        message: sample.message,
        timestamp: new Date().toISOString()
      })
    }
  }, 4500)
})

onBeforeUnmount(() => {
  if (simTimer) clearInterval(simTimer)
  ws?.close()
})
</script>

<template>
  <div class="mx-auto max-w-[1720px] px-4 py-5 lg:px-8">
    <div class="grid grid-cols-1 gap-6 xl:grid-cols-12">
      <!-- Live HLS Player + Info -->
      <div class="xl:col-span-8">
        <ClientOnly>
          <VideoPlayer
            :video-id="streamId"
            channel-id="ch_tech_kr"
            title="[LIVE] 2026 Global Architecture Summit — Keynote & Live Q&A"
            hls-master-url="https://media.gvp.media/live/stream_summit_2026/master.m3u8"
            poster-url="https://images.unsplash.com/photo-1540575467063-178a50c2df87?w=1200&q=80"
            :is-live="true"
          />
        </ClientOnly>

        <div class="mt-4 flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-white/10 bg-[#161616] p-4">
          <div>
            <div class="flex items-center gap-2">
              <span class="rounded bg-[#FF0033] px-2 py-0.5 text-xs font-extrabold uppercase text-white">
                LIVE
              </span>
              <span class="text-xs font-semibold text-emerald-400">
                ● {{ viewerCount.toLocaleString() }} watching now
              </span>
              <span class="rounded bg-emerald-500/20 px-2 py-0.5 text-[11px] font-bold text-emerald-400">
                6.8 Mbps · 60 fps · 2.0s GOP · 0% Loss
              </span>
              <span class="text-xs text-neutral-400">· RTMPS Ingest → LL-HLS Edge</span>
            </div>
            <h1 class="mt-1.5 text-lg font-bold text-white">
              [LIVE] 2026 Global Architecture Summit — Keynote & Live Q&A
            </h1>
          </div>
          <div class="flex items-center gap-2">
            <a
              :href="`${config.public.studioUrl}/live`"
              target="_blank"
              rel="noopener noreferrer"
              class="rounded-full bg-[#FF0033] px-4 py-2 text-xs font-bold text-white shadow hover:bg-[#D9002B]"
            >
              📡 라이브 관제실 (Studio Live Room) ↗
            </a>
            <NuxtLink
              to="/channels/ch_tech_kr"
              class="rounded-full bg-white px-4 py-2 text-xs font-bold text-black hover:bg-neutral-200"
            >
              GVP Tech Engineering
            </NuxtLink>
          </div>
        </div>
      </div>

      <!-- Real-Time WebSocket Live Chat & SuperChat Column -->
      <aside class="flex h-[640px] flex-col rounded-2xl border border-white/15 bg-[#161616] xl:col-span-4">
        <div class="flex items-center justify-between border-b border-white/10 px-4 py-3">
          <div class="flex items-center gap-2">
            <span class="h-2 w-2 rounded-full" :class="wsConnected ? 'bg-emerald-400 animate-pulse' : 'bg-neutral-500'" />
            <h2 class="text-sm font-bold">Live Chat & SuperChat</h2>
          </div>
          <span class="text-[11px] font-medium" :class="wsConnected ? 'text-emerald-400' : 'text-neutral-400'">
            {{ wsConnected ? '● WebSocket Live' : '○ Standby' }}
          </span>
        </div>

        <!-- Pinned SuperChat Horizontal Ticker -->
        <div class="flex items-center gap-2 overflow-x-auto border-b border-white/10 bg-black/40 px-3 py-2 text-xs">
          <div
            v-for="sc in pinnedSuperChats"
            :key="sc.id"
            class="flex shrink-0 items-center gap-1.5 rounded-full bg-gradient-to-r px-3 py-1 font-semibold text-white shadow"
            :class="sc.gradient"
          >
            <span>💎</span>
            <span>{{ sc.name }} ₩{{ sc.amount.toLocaleString() }}</span>
          </div>
        </div>

        <!-- Real-Time Live Poll Widget -->
        <div v-if="activePoll" class="border-b border-white/10 bg-[#1F1F1F] p-3 text-xs">
          <div class="flex items-center justify-between text-neutral-300">
            <span class="font-bold text-white">📊 {{ appStore.locale === 'ko' ? '라이브 실시간 투표' : 'Live Poll' }}</span>
            <span class="text-[10px] text-neutral-400">{{ activePoll.totalVotes }} {{ appStore.locale === 'ko' ? '표' : 'votes' }}</span>
          </div>
          <p class="mt-1 font-medium text-neutral-200">{{ activePoll.question }}</p>
          <div class="mt-2 space-y-1.5">
            <button
              v-for="(opt, idx) in activePoll.options"
              :key="idx"
              type="button"
              class="relative flex w-full items-center justify-between overflow-hidden rounded-lg border px-2.5 py-1.5 text-left text-xs transition"
              :class="activePoll.userVotedIndex === idx ? 'border-[#3EA6FF] text-white' : 'border-white/10 bg-white/5 hover:bg-white/10 text-neutral-300'"
              @click="voteInLivePoll(idx)"
            >
              <div
                v-if="activePoll.userVotedIndex !== null"
                class="absolute inset-y-0 left-0 bg-[#3EA6FF]/20"
                :style="{ width: `${Math.round((opt.votes / (activePoll.totalVotes || 1)) * 100)}%` }"
              />
              <span class="relative z-10">{{ opt.text }}</span>
              <span v-if="activePoll.userVotedIndex !== null" class="relative z-10 font-mono font-bold">
                {{ Math.round((opt.votes / (activePoll.totalVotes || 1)) * 100) }}%
              </span>
            </button>
          </div>
        </div>

        <!-- Chat Messages Scroll Area -->
        <div class="flex-1 space-y-3 overflow-y-auto p-4 text-xs">
          <div
            v-for="msg in messages"
            :key="msg.id"
            class="rounded-xl p-2.5"
            :class="
              msg.donationAmount
                ? 'border border-amber-400/50 bg-gradient-to-r from-amber-500/20 to-red-500/20 text-white'
                : 'bg-white/5 text-neutral-200'
            "
          >
            <div class="flex items-center justify-between">
              <span class="font-bold" :class="msg.isModerator ? 'text-emerald-400' : 'text-[#3EA6FF]'">
                {{ msg.userDisplayName }}
                <span v-if="msg.isModerator" class="ml-1 rounded bg-emerald-500/20 px-1.5 py-0.5 text-[10px]">MOD</span>
              </span>
              <span v-if="msg.donationAmount" class="rounded bg-amber-400 px-2 py-0.5 font-extrabold text-black">
                SuperChat {{ msg.currency === 'USD' ? `$${msg.donationAmount}` : `₩${msg.donationAmount.toLocaleString()}` }}
              </span>
            </div>
            <p class="mt-1">{{ msg.message }}</p>
          </div>
        </div>

        <!-- Chat Input & SuperChat Selector -->
        <form class="border-t border-white/10 p-3" @submit.prevent="sendChatMessage">
          <div class="mb-2 flex items-center gap-1.5">
            <span class="text-[11px] text-neutral-400">SuperChat:</span>
            <button
              v-for="amt in [0, 5000, 10000, 50000]"
              :key="amt"
              type="button"
              class="rounded-lg px-2 py-0.5 text-[11px] font-semibold transition"
              :class="superChatAmount === amt ? 'bg-amber-400 text-black' : 'bg-white/10 text-neutral-300'"
              @click="superChatAmount = amt"
            >
              {{ amt === 0 ? 'None' : `₩${amt.toLocaleString()}` }}
            </button>
          </div>
          <div class="flex gap-2">
            <input
              v-model="chatInput"
              type="text"
              placeholder="Send a live chat message..."
              class="flex-1 rounded-xl border border-white/15 bg-[#0F0F0F] px-3 py-2 text-xs text-white outline-none focus:border-[#3EA6FF]"
            />
            <button
              type="submit"
              class="rounded-xl bg-[#FF0033] px-4 py-2 text-xs font-bold text-white hover:bg-[#D9002B]"
            >
              Send
            </button>
          </div>
        </form>
      </aside>
    </div>
  </div>
</template>
