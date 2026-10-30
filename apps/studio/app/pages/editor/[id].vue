<script setup lang="ts">
const route = useRoute()
const videoId = computed(() => String(route.params.id || 'vid_arch_2026'))

useSeoMeta({
  title: () => `Non-Destructive Timeline Video Editor (${videoId.value})`
})

const trimStart = ref(12)
const trimEnd = ref(1820)
const splitPoints = ref<number[]>([310, 890])
const blurRegions = ref<Array<{ id: string; label: string; x: number; y: number; startSec: number; endSec: number }>>([
  { id: 'blur_1', label: 'License Plate / Personal Info Blur', x: 42, y: 68, startSec: 120, endSec: 165 }
])
const mutedSegments = ref<Array<{ startSec: number; endSec: number; reason: string }>>([
  { startSec: 410, endSec: 445, reason: 'Copyrighted Background Music Segment' }
])
const replacementAudioTrack = ref('Royalty-Free Synthwave Ambient #04 (StreamVerse Audio Library)')
const jobStatus = ref<'IDLE' | 'DISPATCHED' | 'RENDERING' | 'COMPLETED'>('IDLE')

function addSplitAtPlayhead() {
  splitPoints.value.push(540)
}

function addBlurRegion() {
  blurRegions.value.push({
    id: `blur_${Date.now()}`,
    label: 'Face / Screen Blur Box',
    x: 50,
    y: 50,
    startSec: 200,
    endSec: 260
  })
}

function dispatchFfmpegEditJob() {
  jobStatus.value = 'DISPATCHED'
  setTimeout(() => {
    jobStatus.value = 'RENDERING'
  }, 600)
  setTimeout(() => {
    jobStatus.value = 'COMPLETED'
  }, 1800)
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">
          비파괴 타임라인 비디오 에디터 (Non-Destructive Video Editor)
        </h1>
        <p class="mt-1 text-xs text-neutral-400">
          Video ID: <code class="text-[#3EA6FF]">{{ videoId }}</code> · Preserves original URL, view count, and comments while re-rendering HLS segments via FFmpeg worker pool
        </p>
      </div>

      <button
        type="button"
        class="rounded-xl bg-[#FF0033] px-5 py-2.5 text-xs font-bold text-white shadow hover:bg-[#D9002B]"
        @click="dispatchFfmpegEditJob"
      >
        🎬 Save & Dispatch FFmpeg Render Job
      </button>
    </div>

    <!-- Job Status Banner -->
    <div
      v-if="jobStatus !== 'IDLE'"
      class="mt-4 rounded-xl border border-emerald-500/40 bg-emerald-500/10 px-4 py-3 text-xs font-bold text-emerald-300"
    >
      FFmpeg Worker Status: {{ jobStatus }} — Original R2 source preserved; non-destructive EDL (Edit Decision List) applied to HLS manifest.
    </div>

    <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Preview Canvas with Interactive Blur Overlay -->
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-4 lg:col-span-2">
        <div class="relative aspect-video w-full overflow-hidden rounded-xl bg-black">
          <img
            src="https://images.unsplash.com/photo-1518770660439-4636190af475?w=1200&q=80"
            alt="Editor frame"
            class="h-full w-full object-cover"
          />
          <!-- Blur Region Indicator Box -->
          <div
            v-for="box in blurRegions"
            :key="box.id"
            class="absolute flex h-16 w-36 items-center justify-center rounded-lg border-2 border-[#3EA6FF] bg-black/50 text-[10px] font-bold text-white backdrop-blur-md"
            :style="{ left: `${box.x}%`, top: `${box.y}%` }"
          >
            {{ box.label }}
          </div>
        </div>

        <!-- Multi-Track Non-Destructive Timeline -->
        <div class="mt-5 space-y-3 rounded-xl border border-white/10 bg-[#0D0F14] p-4 text-xs">
          <div class="flex items-center justify-between">
            <span class="font-bold text-white">Timeline Tracks (Trim: {{ trimStart }}s – {{ trimEnd }}s)</span>
            <div class="flex gap-2">
              <button
                type="button"
                class="rounded-lg bg-white/10 px-3 py-1 font-semibold text-white hover:bg-white/20"
                @click="addSplitAtPlayhead"
              >
                ✂️ Split Segment
              </button>
              <button
                type="button"
                class="rounded-lg bg-white/10 px-3 py-1 font-semibold text-white hover:bg-white/20"
                @click="addBlurRegion"
              >
                🔲 Add Blur Region
              </button>
            </div>
          </div>

          <!-- Video Track -->
          <div class="relative h-8 w-full overflow-hidden rounded-lg bg-[#1E293B] px-3 leading-8 text-white">
            <span>🎞️ Video Track (2160p60 Master)</span>
            <span
              v-for="sp in splitPoints"
              :key="sp"
              class="absolute top-0 h-full w-0.5 bg-[#FF0033]"
              :style="{ left: `${(sp / 1845) * 100}%` }"
            />
          </div>

          <!-- Blur Mask Track -->
          <div class="h-7 w-full rounded-lg bg-purple-950/60 px-3 leading-7 text-purple-200">
            🔲 Blur Regions: {{ blurRegions.length }} active mask(s)
          </div>

          <!-- Audio & Copyright Mute Track -->
          <div class="h-7 w-full rounded-lg bg-amber-950/60 px-3 leading-7 text-amber-200">
            🔇 Muted Claim Segment ({{ (mutedSegments[0]?.startSec ?? 0) }}s–{{ (mutedSegments[0]?.endSec ?? 0) }}s) · Replaced with: {{ replacementAudioTrack }}
          </div>
        </div>
      </div>

      <!-- Right Inspector Panel -->
      <div class="space-y-4 rounded-2xl border border-white/10 bg-[#151822] p-5 text-xs">
        <h2 class="text-sm font-bold text-white">Edit Decision List (EDL) Controls</h2>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="text-neutral-400">Trim In (sec)</label>
            <input
              v-model.number="trimStart"
              type="number"
              class="mt-1 w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-1.5 text-white"
            />
          </div>
          <div>
            <label class="text-neutral-400">Trim Out (sec)</label>
            <input
              v-model.number="trimEnd"
              type="number"
              class="mt-1 w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-1.5 text-white"
            />
          </div>
        </div>

        <div>
          <label class="font-bold text-neutral-300">Audio Track Replacement (Resolve Copyright Claims)</label>
          <select
            v-model="replacementAudioTrack"
            class="mt-1.5 w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-2 text-white"
          >
            <option>Royalty-Free Synthwave Ambient #04 (StreamVerse Audio Library)</option>
            <option>Lo-Fi Chill Coding Beats #12 (Cleared Global License)</option>
            <option>Mute Claimed Segment Only (Preserve Speech via AI Stem Split)</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>
