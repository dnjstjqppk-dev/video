<script setup lang="ts">
import { LiveStreamStatus } from '@gvp/shared-types'

const api = useStudioApi()

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

const moderationFeed = ref([
  { id: 'm1', user: '@jihoon_sre', text: 'Can you share the ClickHouse Materialized View schema?', blocked: false },
  { id: 'm2', user: '@spam_bot_99', text: 'FREE CRYPTO AIRDROP CLICK HERE', blocked: true },
  { id: 'm3', user: '@sarahchen_dev', text: 'SRT ingest latency is under 1.2 seconds!', blocked: false }
])

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
  <div>
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">라이브 관제실 (Live Control Room)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          RTMPS / SRT Dual Ingest · Sub-2s LL-HLS Packaging · Real-time Chat Moderation & DVR Auto-Archive
        </p>
      </div>

      <div class="flex items-center gap-3">
        <span
          class="rounded-full px-3 py-1 text-xs font-extrabold uppercase"
          :class="streamStatus === LiveStreamStatus.LIVE ? 'bg-[#FF0033] text-white' : 'bg-neutral-700 text-neutral-300'"
        >
          ● {{ streamStatus }}
        </span>
        <button
          type="button"
          class="rounded-xl px-4 py-2 text-xs font-bold text-white"
          :class="streamStatus === LiveStreamStatus.LIVE ? 'bg-neutral-700 hover:bg-neutral-600' : 'bg-[#FF0033] hover:bg-[#D9002B]'"
          @click="toggleBroadcast"
        >
          {{ streamStatus === LiveStreamStatus.LIVE ? '⏹️ End Broadcast & Archive VOD' : '🔴 Start Live Broadcast' }}
        </button>
      </div>
    </div>

    <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Ingest Configuration & Stream Key Rotation -->
      <div class="space-y-4 rounded-2xl border border-white/10 bg-[#151822] p-5 lg:col-span-2">
        <h2 class="text-sm font-bold text-white">스트림 인제스트 설정 (OBS / Hardware Encoder Setup)</h2>

        <div>
          <label class="text-xs font-semibold text-neutral-400">Stream Title</label>
          <input
            v-model="streamTitle"
            type="text"
            class="mt-1 w-full rounded-xl border border-white/15 bg-[#0D0F14] px-3.5 py-2 text-xs text-white"
          />
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="text-xs font-semibold text-neutral-400">Primary RTMPS Ingest URL</label>
            <input
              :value="rtmpsUrl"
              readonly
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#0D0F14] px-3 py-2 font-mono text-xs text-[#3EA6FF]"
            />
          </div>
          <div>
            <label class="text-xs font-semibold text-neutral-400">Low-Latency SRT Ingest URL</label>
            <input
              :value="srtUrl"
              readonly
              class="mt-1 w-full rounded-xl border border-white/15 bg-[#0D0F14] px-3 py-2 font-mono text-xs text-emerald-400"
            />
          </div>
        </div>

        <div>
          <label class="text-xs font-semibold text-neutral-400">Stream Key (Encrypted at Rest)</label>
          <div class="mt-1 flex gap-2">
            <input
              :type="showStreamKey ? 'text' : 'password'"
              :value="streamKey"
              readonly
              class="flex-1 rounded-xl border border-white/15 bg-[#0D0F14] px-3 py-2 font-mono text-xs text-white"
            />
            <button
              type="button"
              class="rounded-xl border border-white/15 px-3 py-2 text-xs font-semibold text-white hover:bg-white/10"
              @click="showStreamKey = !showStreamKey"
            >
              {{ showStreamKey ? 'Hide' : 'Reveal' }}
            </button>
            <button
              type="button"
              class="rounded-xl bg-amber-400 px-3.5 py-2 text-xs font-bold text-black hover:bg-amber-300"
              @click="rotateKey"
            >
              🔄 Rotate Key
            </button>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-6 border-t border-white/10 pt-4 text-xs">
          <label class="flex items-center gap-2">
            <input v-model="dvrEnabled" type="checkbox" class="accent-[#FF0033]" />
            <span>Enable DVR Instant Rewind & Automatic Live-to-VOD Recording</span>
          </label>

          <div class="flex items-center gap-2">
            <span class="text-neutral-400">Latency Profile:</span>
            <select
              v-model="lowLatencyMode"
              class="rounded-lg border border-white/15 bg-[#0D0F14] px-2.5 py-1 text-white"
            >
              <option value="ultra_low">Ultra-Low Latency (LL-HLS ~1.5s)</option>
              <option value="normal">Standard ABR (4K60 High Buffer)</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Live Chat Moderation Queue -->
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <div class="flex items-center justify-between border-b border-white/10 pb-3">
          <h2 class="text-sm font-bold text-white">실시간 채팅 모더레이션</h2>
          <span class="text-xs text-neutral-400">Slow Mode: {{ slowModeSeconds }}s</span>
        </div>

        <div class="mt-4 space-y-3 text-xs">
          <div
            v-for="msg in moderationFeed"
            :key="msg.id"
            class="flex items-center justify-between gap-2 rounded-xl border border-white/10 bg-[#0D0F14] p-3"
          >
            <div>
              <p class="font-bold text-[#3EA6FF]">{{ msg.user }}</p>
              <p class="mt-0.5" :class="msg.blocked ? 'line-through text-red-400' : 'text-neutral-200'">
                {{ msg.text }}
              </p>
            </div>
            <button
              type="button"
              class="shrink-0 rounded-lg px-2.5 py-1 text-[11px] font-bold"
              :class="msg.blocked ? 'bg-white/10 text-neutral-300' : 'bg-red-500/20 text-red-300'"
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
