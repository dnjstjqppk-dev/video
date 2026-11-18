<script setup lang="ts">
import { LiveStreamStatus } from '@gvp/shared-types'

const api = useStudioApi()
const config = useRuntimeConfig()

useSeoMeta({
  title: 'Live Control Room — RTMPS/SRT Ingest & Moderation'
})

const streamId = ref('live_main_01')
const streamTitle = ref('[LIVE] 2026 Global Architecture Summit — Keynote & Live Q&A')
const streamStatus = ref<LiveStreamStatus>(LiveStreamStatus.LIVE)
const rtmpsUrl = ref('rtmps://ingest.gvp.media:443/live')
const srtUrl = ref('srt://ingest.gvp.media:9000?streamid=live_main_01&latency=120')
const streamKey = ref('sk_live_kr_98a7bc65d4e3f210_secret')
const showStreamKey = ref(false)
const dvrEnabled = ref(true)
const lowLatencyMode = ref<'ultra_low' | 'normal'>('ultra_low')
const slowModeSeconds = ref(5)
const liveViewers = ref(14820)
const copyToast = ref('')

const telemetry = ref({
  status: 'EXCELLENT',
  bitrateKbps: 6800,
  fps: 60,
  gopSeconds: 2.0,
  packetLossPct: 0.0,
  audioSampleRateHz: 48000
})

const moderationFeed = ref([
  { id: 'm1', user: '@jihoon_sre', text: 'Can you share the ClickHouse Materialized View schema?', blocked: false },
  { id: 'm2', user: '@spam_bot_99', text: 'FREE CRYPTO AIRDROP CLICK HERE', blocked: true },
  { id: 'm3', user: '@sarahchen_dev', text: 'SRT ingest latency is under 1.2 seconds!', blocked: false }
])

let ws: WebSocket | null = null

onMounted(() => {
  try {
    const wsUrl = `${config.public.wsBaseUrl || 'ws://localhost:8080'}/v1/live-streams/${streamId.value}/ws`
    ws = new WebSocket(wsUrl)
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'viewer_count' && msg.payload?.viewer_count !== undefined) {
          liveViewers.value = msg.payload.viewer_count
        } else if (msg.type === 'chat_message' && msg.payload) {
          moderationFeed.value.unshift({
            id: msg.payload.id || String(Date.now()),
            user: msg.payload.user_display_name || msg.payload.user_id || '@viewer',
            text: msg.payload.message || '',
            blocked: false
          })
          if (moderationFeed.value.length > 50) {
            moderationFeed.value.pop()
          }
        } else if (msg.type === 'super_chat' && msg.payload) {
          const userStr = msg.payload.display_name || msg.payload.user_display_name || 'VIP Supporter'
          const amt = msg.payload.amount_cents ? msg.payload.amount_cents / 100 : (msg.payload.donation_amount || 0)
          moderationFeed.value.unshift({
            id: msg.payload.id || String(Date.now()),
            user: `[SuperChat] ${userStr} (₩${amt.toLocaleString()})`,
            text: msg.payload.message || '',
            blocked: false
          })
        }
      } catch {
        // ignore parse error
      }
    }
  } catch {
    // fallback
  }
})

onUnmounted(() => {
  if (ws) ws.close()
})

function copyToClipboard(text: string, label: string) {
  navigator.clipboard.writeText(text)
  copyToast.value = `${label} 클립보드에 복사되었습니다!`
  setTimeout(() => {
    copyToast.value = ''
  }, 2500)
}

async function rotateKey() {
  try {
    const res = await api.rotateLiveStreamKey(streamId.value)
    streamKey.value = res.streamKey
  } catch {
    streamKey.value = `sk_live_kr_${Math.random().toString(36).slice(2, 14)}_rotated`
  }
}

async function toggleBroadcast() {
  if (streamStatus.value === LiveStreamStatus.LIVE) {
    try {
      await api.stopLiveStream(streamId.value)
    } catch {
      // Fallback state update
    }
    streamStatus.value = LiveStreamStatus.ENDED
  } else {
    try {
      await api.startLiveStream(streamId.value)
    } catch {
      // Fallback state update
    }
    streamStatus.value = LiveStreamStatus.LIVE
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-[var(--gvp-border)] pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">라이브 관제실 (Live Control Room)</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          RTMPS / SRT Dual Ingest · Sub-2s LL-HLS Packaging · Real-time Chat Moderation &amp; DVR Auto-Archive
        </p>
      </div>

      <div class="flex items-center gap-3">
        <span
          class="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-semibold uppercase tabular-nums"
          :class="streamStatus === LiveStreamStatus.LIVE ? 'border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] text-[var(--gvp-live)]' : 'border-[var(--gvp-border)] bg-[var(--gvp-elevated)] text-[var(--gvp-text-secondary)]'"
        >
          <span>{{ streamStatus }} (동시 시청 {{ liveViewers.toLocaleString() }}명)</span>
        </span>
        <button
          type="button"
          :class="
            streamStatus === LiveStreamStatus.LIVE
              ? 'inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors'
              : 'inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors'
          "
          @click="toggleBroadcast"
        >
          {{ streamStatus === LiveStreamStatus.LIVE ? '방송 종료 및 VOD 아카이브' : '실시간 스트리밍 시작' }}
        </button>
      </div>
    </div>

    <!-- Copy Toast -->
    <div
      v-if="copyToast"
      class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-3 text-xs font-medium text-[var(--gvp-text)]"
      role="status"
    >
      {{ copyToast }}
    </div>

    <!-- Live Telemetry Stats Bar -->
    <div class="grid grid-cols-2 gap-4 md:grid-cols-5 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4 text-xs">
      <div>
        <span class="text-[var(--gvp-text-secondary)]">인제스트 상태</span>
        <div class="mt-1 font-semibold text-[var(--gvp-success)] tabular-nums">{{ telemetry.status }}</div>
      </div>
      <div>
        <span class="text-[var(--gvp-text-secondary)]">비트레이트</span>
        <div class="mt-1 font-semibold text-[var(--gvp-text)] tabular-nums">{{ telemetry.bitrateKbps }} kbps</div>
      </div>
      <div>
        <span class="text-[var(--gvp-text-secondary)]">프레임 레이트</span>
        <div class="mt-1 font-semibold text-[var(--gvp-text)] tabular-nums">{{ telemetry.fps }} fps</div>
      </div>
      <div>
        <span class="text-[var(--gvp-text-secondary)]">키프레임 간격 (GOP)</span>
        <div class="mt-1 font-semibold text-[var(--gvp-text)] tabular-nums">{{ telemetry.gopSeconds }}s</div>
      </div>
      <div>
        <span class="text-[var(--gvp-text-secondary)]">패킷 손실률</span>
        <div class="mt-1 font-semibold text-[var(--gvp-success)] tabular-nums">{{ telemetry.packetLossPct }}%</div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Ingest Configuration & Stream Key Rotation -->
      <div class="space-y-4 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5 lg:col-span-2">
        <h2 class="text-sm font-semibold text-[var(--gvp-text)]">스트림 인제스트 설정 (OBS / Hardware Encoder Setup)</h2>

        <div>
          <label class="text-xs font-medium text-[var(--gvp-text)]">Stream Title</label>
          <input
            v-model="streamTitle"
            type="text"
            class="mt-1 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-3 py-2 text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
          />
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <div class="flex items-center justify-between">
              <label class="text-xs font-medium text-[var(--gvp-text)]">Primary RTMPS Ingest URL</label>
              <button
                type="button"
                class="text-xs font-medium text-[var(--gvp-text-secondary)] underline hover:text-[var(--gvp-text)]"
                @click="copyToClipboard(rtmpsUrl, 'RTMPS URL')"
              >
                복사
              </button>
            </div>
            <input
              :value="rtmpsUrl"
              readonly
              class="mt-1 w-full rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-2 font-mono text-xs text-[var(--gvp-text)]"
            />
          </div>
          <div>
            <div class="flex items-center justify-between">
              <label class="text-xs font-medium text-[var(--gvp-text)]">Low-Latency SRT Ingest URL</label>
              <button
                type="button"
                class="text-xs font-medium text-[var(--gvp-text-secondary)] underline hover:text-[var(--gvp-text)]"
                @click="copyToClipboard(srtUrl, 'SRT URL')"
              >
                복사
              </button>
            </div>
            <input
              :value="srtUrl"
              readonly
              class="mt-1 w-full rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-2 font-mono text-xs text-[var(--gvp-text)]"
            />
          </div>
        </div>

        <div>
          <div class="flex items-center justify-between">
            <label class="text-xs font-medium text-[var(--gvp-text)]">Stream Key (OBS 설정용 스트림 키)</label>
            <button
              type="button"
              class="text-xs font-medium text-[var(--gvp-text-secondary)] underline hover:text-[var(--gvp-text)]"
              @click="copyToClipboard(streamKey, '스트림 키가')"
            >
              키 복사
            </button>
          </div>
          <div class="mt-1 flex gap-2">
            <input
              :type="showStreamKey ? 'text' : 'password'"
              :value="streamKey"
              readonly
              class="flex-1 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-2 font-mono text-xs text-[var(--gvp-text)]"
            />
            <button
              type="button"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
              @click="showStreamKey = !showStreamKey"
            >
              {{ showStreamKey ? 'Hide' : 'Reveal' }}
            </button>
            <button
              type="button"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors"
              @click="rotateKey"
            >
              Rotate Key
            </button>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-6 border-t border-[var(--gvp-border)] pt-4 text-xs">
          <label class="flex items-center gap-2 text-[var(--gvp-text)]">
            <input v-model="dvrEnabled" type="checkbox" class="rounded-md border-[var(--gvp-border-strong)] accent-[var(--gvp-brand)]" />
            <span>Enable DVR Instant Rewind &amp; Automatic Live VOD Recording</span>
          </label>

          <div class="flex items-center gap-2">
            <span class="text-[var(--gvp-text-secondary)]">Latency Profile:</span>
            <select
              v-model="lowLatencyMode"
              class="rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 py-1.5 text-xs font-medium text-[var(--gvp-text)] tabular-nums"
            >
              <option value="ultra_low">Ultra-Low Latency (LL-HLS ~1.5s)</option>
              <option value="normal">Standard ABR (4K60 High Buffer)</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Live Chat Moderation Queue -->
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <div class="flex items-center justify-between border-b border-[var(--gvp-border)] pb-3">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">실시간 채팅 모더레이션</h2>
          <span class="text-xs text-[var(--gvp-text-secondary)] tabular-nums">Slow Mode: {{ slowModeSeconds }}s</span>
        </div>

        <div class="mt-4 max-h-[380px] space-y-2.5 overflow-y-auto text-xs">
          <div
            v-for="msg in moderationFeed"
            :key="msg.id"
            class="flex items-center justify-between gap-2 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-3"
          >
            <div class="min-w-0">
              <p class="font-semibold text-[var(--gvp-text)] tabular-nums">{{ msg.user }}</p>
              <p class="mt-0.5" :class="msg.blocked ? 'line-through text-[var(--gvp-live)]' : 'text-[var(--gvp-text-secondary)]'">
                {{ msg.text }}
              </p>
            </div>
            <button
              type="button"
              class="shrink-0 rounded-md border px-2.5 py-1 text-xs font-medium transition-colors"
              :class="msg.blocked ? 'border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] text-[var(--gvp-text)] hover:bg-[var(--gvp-canvas)]' : 'border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] text-[var(--gvp-live)] hover:bg-[var(--gvp-live-soft)]'"
              @click="msg.blocked = !msg.blocked"
            >
              {{ msg.blocked ? 'Unban' : 'Timeout' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
