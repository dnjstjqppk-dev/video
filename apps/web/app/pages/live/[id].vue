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

const viewerCount = ref(0)
const viewerCountKnown = ref(false)
const chatInput = ref('')
const superChatAmount = ref<number>(0)
const wsConnected = ref(false)
const chatStatus = ref<'connecting' | 'live' | 'reconnecting' | 'offline'>('connecting')
const reconnectAttempts = ref(0)
const queuedOutbox = ref(0)

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
  if (amount >= 50000) return 'bg-red-50 text-red-700 border-red-200'
  if (amount >= 20000) return 'bg-amber-50 text-amber-700 border-amber-200'
  if (amount >= 10000) return 'bg-amber-50 text-amber-700 border-amber-200'
  if (amount >= 5000) return 'bg-green-50 text-green-700 border-green-200'
  return 'bg-blue-50 text-blue-700 border-blue-200'
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
    message: 'SuperChat 응원합니다! Go 1.24 + Nuxt 4 조합 최고네요.',
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
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let intentionalClose = false
const outbox: Array<{ action: string; payload: Record<string, unknown> }> = []

const MAX_RECONNECT_ATTEMPTS = 10

function scheduleReconnect() {
  if (intentionalClose) return
  if (reconnectAttempts.value >= MAX_RECONNECT_ATTEMPTS) {
    chatStatus.value = 'offline'
    return
  }
  chatStatus.value = 'reconnecting'
  const delayMs = Math.min(1000 * 2 ** reconnectAttempts.value, 30000)
  reconnectAttempts.value += 1
  if (reconnectTimer) clearTimeout(reconnectTimer)
  reconnectTimer = setTimeout(() => connectLiveChatSocket(), delayMs)
}

function connectLiveChatSocket() {
  try {
    const wsBase = config.public.wsBaseUrl || 'ws://localhost:8080'
    const userQuery = `userId=${encodeURIComponent(appStore.currentUser?.id || 'usr_demo_01')}&displayName=${encodeURIComponent(appStore.currentUser?.displayName || '김민준')}`
    const wsUrl = `${wsBase}/v1/live-streams/${streamId.value}/ws?${userQuery}`
    ws = new WebSocket(wsUrl)
    ws.onopen = () => {
      wsConnected.value = true
      chatStatus.value = 'live'
      reconnectAttempts.value = 0
      while (outbox.length > 0 && ws && ws.readyState === WebSocket.OPEN) {
        const frame = outbox.shift()
        if (frame) ws.send(JSON.stringify(frame))
      }
      queuedOutbox.value = outbox.length
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
        } else if (frame.type === 'viewer_count' && typeof frame.payload?.viewer_count === 'number') {
          viewerCount.value = frame.payload.viewer_count
          viewerCountKnown.value = true
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
      wsConnected.value = false
      ws?.close()
    }
    ws.onclose = () => {
      wsConnected.value = false
      scheduleReconnect()
    }
  } catch {
    wsConnected.value = false
    scheduleReconnect()
  }
}

function retryChatConnection() {
  reconnectAttempts.value = 0
  chatStatus.value = 'connecting'
  connectLiveChatSocket()
}

const pinnedSuperChats = ref([
  { id: 'sc_1', name: '최현우', amount: 50000, gradient: 'border-red-200 bg-red-50 text-red-700' },
  { id: 'sc_2', name: 'Elena Rostova', amount: 5000, gradient: 'border-green-200 bg-green-50 text-green-700' }
])

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
      gradient: amt >= 50000 ? 'border-red-200 bg-red-50 text-red-700' : amt >= 10000 ? 'border-amber-200 bg-amber-50 text-amber-700' : 'border-green-200 bg-green-50 text-green-700'
    })
  }

  const frame = { action: 'send_chat', payload }
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(frame))
  } else {
    outbox.push(frame)
    queuedOutbox.value = outbox.length
    if (chatStatus.value === 'offline') retryChatConnection()
  }

  chatInput.value = ''
  superChatAmount.value = 0
}

onMounted(() => {
  connectLiveChatSocket()
})

onBeforeUnmount(() => {
  intentionalClose = true
  if (reconnectTimer) clearTimeout(reconnectTimer)
  ws?.close()
})
</script>

<template>
  <div class="mx-auto max-w-[1720px] bg-white px-4 py-5 text-slate-900 lg:px-8">
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

        <div class="mt-4 flex flex-wrap items-center justify-between gap-4 rounded-lg border border-slate-200 bg-white p-4">
          <div>
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-md border border-red-200 bg-red-50 px-2 py-0.5 text-xs font-bold uppercase text-red-700">
                LIVE
              </span>
              <span v-if="viewerCountKnown" class="text-xs font-semibold text-green-700 tabular-nums">
                ● {{ viewerCount.toLocaleString() }} watching now
              </span>
              <span v-else class="text-xs font-semibold text-slate-500">
                시청자 수 집계 중
              </span>
              <span class="rounded-md border border-green-200 bg-green-50 px-2 py-0.5 text-[11px] font-bold text-green-700 tabular-nums">
                6.8 Mbps · 60 fps · 2.0s GOP · 0% Loss
              </span>
              <span class="text-xs text-slate-500">· RTMPS Ingest → LL-HLS Edge</span>
            </div>
            <h1 class="mt-2 text-lg font-bold text-slate-900">
              [LIVE] 2026 Global Architecture Summit — Keynote & Live Q&A
            </h1>
          </div>
          <div class="flex items-center gap-2">
            <a
              :href="`${config.public.studioUrl}/live`"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 rounded-md bg-slate-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-slate-800"
            >
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5.636 18.364a9 9 0 010-12.728m12.728 0a9 9 0 010 12.728m-9.9-2.829a5 5 0 010-7.07m7.072 0a5 5 0 010 7.07M13 12a1 1 0 11-2 0 1 1 0 012 0z" />
              </svg>
              <span>라이브 관제실 (Studio Live Room)</span>
            </a>
            <NuxtLink
              to="/channels/ch_tech_kr"
              class="rounded-md border border-slate-200 bg-slate-50 px-4 py-2 text-xs font-semibold text-slate-700 transition hover:bg-slate-100"
            >
              GVP Tech Engineering
            </NuxtLink>
          </div>
        </div>
      </div>

      <!-- Real-Time WebSocket Live Chat & SuperChat Column -->
      <aside class="flex h-[640px] flex-col rounded-lg border border-slate-200 bg-white xl:col-span-4">
        <div class="flex items-center justify-between border-b border-slate-200 bg-slate-50 px-4 py-3">
          <div class="flex items-center gap-2">
            <span class="h-2 w-2 rounded-full" :class="wsConnected ? 'bg-green-600 animate-pulse' : 'bg-slate-400'" />
            <h2 class="text-sm font-bold text-slate-900">Live Chat & SuperChat</h2>
          </div>
          <span class="flex items-center gap-2">
            <span
              class="text-[11px] font-semibold tabular-nums"
              :class="chatStatus === 'live' ? 'text-green-700' : chatStatus === 'offline' ? 'text-red-700' : 'text-amber-700'"
              role="status"
            >
              {{ chatStatus === 'live' ? '● WebSocket Live' : chatStatus === 'connecting' ? '연결 중…' : chatStatus === 'reconnecting' ? `재연결 중… (${reconnectAttempts})` : '연결 끊김' }}
            </span>
            <button
              v-if="chatStatus === 'offline' || chatStatus === 'reconnecting'"
              type="button"
              class="rounded-md border border-slate-200 bg-white px-2.5 py-1 text-[11px] font-semibold text-slate-700 hover:bg-slate-50"
              @click="retryChatConnection"
            >
              다시 연결
            </button>
          </span>
        </div>
        <p v-if="queuedOutbox > 0" class="border-b border-amber-200 bg-amber-50 px-4 py-1.5 text-[11px] font-medium text-amber-700 tabular-nums" role="status">
          연결이 복구되면 대기 중인 메시지 {{ queuedOutbox }}건이 전송됩니다.
        </p>

        <!-- Pinned SuperChat Horizontal Ticker -->
        <div class="flex items-center gap-2 overflow-x-auto border-b border-slate-200 bg-slate-50 px-3 py-2 text-xs">
          <div
            v-for="sc in pinnedSuperChats"
            :key="sc.id"
            class="flex shrink-0 items-center gap-1.5 rounded-md border px-2.5 py-1 font-semibold tabular-nums"
            :class="sc.gradient"
          >
            <span class="text-[10px] font-bold uppercase">SuperChat</span>
            <span>{{ sc.name }} ₩{{ sc.amount.toLocaleString() }}</span>
          </div>
        </div>

        <!-- Real-Time Live Poll Widget -->
        <div v-if="activePoll" class="border-b border-slate-200 bg-slate-50 p-3 text-xs">
          <div class="flex items-center justify-between text-slate-600">
            <span class="flex items-center gap-1.5 font-bold text-slate-900">
              <svg class="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              <span>{{ appStore.locale === 'ko' ? '라이브 실시간 투표' : 'Live Poll' }}</span>
            </span>
            <span class="text-[11px] font-medium text-slate-500 tabular-nums">{{ activePoll.totalVotes }} {{ appStore.locale === 'ko' ? '표' : 'votes' }}</span>
          </div>
          <p class="mt-1 font-medium text-slate-700">{{ activePoll.question }}</p>
          <div class="mt-2 space-y-1.5">
            <button
              v-for="(opt, idx) in activePoll.options"
              :key="idx"
              type="button"
              class="relative flex w-full items-center justify-between overflow-hidden rounded-md border px-2.5 py-1.5 text-left text-xs transition"
              :class="activePoll.userVotedIndex === idx ? 'border-blue-200 bg-blue-50 text-blue-700 font-semibold' : 'border-slate-200 bg-white hover:bg-slate-50 text-slate-700'"
              @click="voteInLivePoll(idx)"
            >
              <div
                v-if="activePoll.userVotedIndex !== null"
                class="absolute inset-y-0 left-0 bg-blue-100/70"
                :style="{ width: `${Math.round((opt.votes / (activePoll.totalVotes || 1)) * 100)}%` }"
              />
              <span class="relative z-10">{{ opt.text }}</span>
              <span v-if="activePoll.userVotedIndex !== null" class="relative z-10 font-mono font-bold tabular-nums">
                {{ Math.round((opt.votes / (activePoll.totalVotes || 1)) * 100) }}%
              </span>
            </button>
          </div>
        </div>

        <!-- Chat Messages Scroll Area -->
        <div class="flex-1 space-y-2.5 overflow-y-auto p-4 text-xs">
          <div
            v-for="msg in messages"
            :key="msg.id"
            class="rounded-lg border p-2.5"
            :class="
              msg.donationAmount
                ? getSuperChatBadgeClass(msg.donationAmount)
                : 'border-slate-200 bg-slate-50 text-slate-700'
            "
          >
            <div class="flex items-center justify-between">
              <span class="font-bold" :class="msg.isModerator ? 'text-green-700' : 'text-slate-900'">
                {{ msg.userDisplayName }}
                <span v-if="msg.isModerator" class="ml-1 rounded-md border border-green-200 bg-green-50 px-1.5 py-0.5 text-[10px] text-green-700">MOD</span>
              </span>
              <span v-if="msg.donationAmount" class="rounded-md border border-amber-200 bg-amber-50 px-2 py-0.5 font-bold text-amber-700 tabular-nums">
                SuperChat {{ msg.currency === 'USD' ? `$${msg.donationAmount}` : `₩${msg.donationAmount.toLocaleString()}` }}
              </span>
            </div>
            <p class="mt-1">{{ msg.message }}</p>
          </div>
        </div>

        <!-- Chat Input & SuperChat Selector -->
        <form class="border-t border-slate-200 bg-slate-50 p-3" @submit.prevent="sendChatMessage">
          <div class="mb-2 flex items-center gap-1.5">
            <span class="text-[11px] font-medium text-slate-500">SuperChat:</span>
            <button
              v-for="amt in [0, 5000, 10000, 50000]"
              :key="amt"
              type="button"
              class="rounded-md border px-2 py-0.5 text-[11px] font-semibold transition tabular-nums"
              :class="superChatAmount === amt ? 'border-amber-200 bg-amber-50 text-amber-700' : 'border-slate-200 bg-white text-slate-600 hover:bg-slate-50'"
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
              class="flex-1 rounded-md border border-slate-200 bg-white px-3 py-2 text-xs text-slate-900 outline-none focus:border-blue-600"
            />
            <button
              type="submit"
              class="rounded-md bg-slate-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-slate-800"
            >
              Send
            </button>
          </div>
        </form>
      </aside>
    </div>
  </div>
</template>
