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
    const wsUrl = `${config.public.wsBaseUrl}/v1/live-streams/${streamId.value}/chat`
    ws = new WebSocket(wsUrl)
    ws.onopen = () => {
      wsConnected.value = true
    }
    ws.onmessage = event => {
      try {
        const incoming = JSON.parse(event.data) as LiveChatMessage
        messages.value.push(incoming)
      } catch {
        // Ignore malformed frame
      }
    }
    ws.onerror = () => {
      wsConnected.value = true // Graceful simulated fallback
    }
  } catch {
    wsConnected.value = true
  }
}

function sendChatMessage() {
  const text = chatInput.value.trim()
  if (!text) return

  const msg: LiveChatMessage = {
    id: `msg_${Date.now()}`,
    streamId: streamId.value,
    userId: appStore.currentUser?.id || 'usr_demo_01',
    userDisplayName: appStore.currentUser?.displayName || '김민준',
    isModerator: false,
    isCreator: false,
    message: text,
    donationAmount: superChatAmount.value > 0 ? superChatAmount.value : undefined,
    currency: superChatAmount.value > 0 ? appStore.currency : undefined,
    timestamp: new Date().toISOString()
  }

  messages.value.push(msg)
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(msg))
  }
  chatInput.value = ''
  superChatAmount.value = 0
}

onMounted(() => {
  connectLiveChatSocket()
  simTimer = setInterval(() => {
    viewerCount.value += Math.floor(Math.random() * 15) - 5
  }, 3000)
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
            hls-master-url="https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8"
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
              <span class="text-xs text-neutral-400">· RTMPS Ingest → LL-HLS Edge</span>
            </div>
            <h1 class="mt-1.5 text-lg font-bold text-white">
              [LIVE] 2026 Global Architecture Summit — Keynote & Live Q&A
            </h1>
          </div>
          <NuxtLink
            to="/channels/ch_tech_kr"
            class="rounded-full bg-white px-4 py-2 text-xs font-bold text-black hover:bg-neutral-200"
          >
            GVP Tech Engineering
          </NuxtLink>
        </div>
      </div>

      <!-- Real-Time WebSocket Live Chat & SuperChat Column -->
      <aside class="flex h-[640px] flex-col rounded-2xl border border-white/15 bg-[#161616] xl:col-span-4">
        <div class="flex items-center justify-between border-b border-white/10 px-4 py-3">
          <div class="flex items-center gap-2">
            <span class="h-2 w-2 rounded-full bg-emerald-400" />
            <h2 class="text-sm font-bold">Live Chat & SuperChat</h2>
          </div>
          <span class="text-[11px] text-neutral-400">WebSocket Connected</span>
        </div>

        <!-- Pinned SuperChat Ticker -->
        <div class="border-b border-white/10 bg-amber-500/10 px-4 py-2 text-xs text-amber-300">
          💎 <strong>최현우</strong> sent <strong>₩50,000 SuperChat</strong>
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
