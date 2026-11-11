<script setup lang="ts">
import type { StudioEditDecisionList, StudioEditJob, StudioEditJobStatus, Video } from '@gvp/shared-types'
import { useStudioApi } from '~/composables/useStudioApi'
import GvpButton from '@gvp/ui/src/components/GvpButton.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpSkeleton from '@gvp/ui/src/components/GvpSkeleton.vue'

const route = useRoute()
const api = useStudioApi()
const videoId = computed(() => String(route.params.id ?? ''))

useSeoMeta({
  title: () => `Video Editor (${videoId.value}) — StreamVerse Studio`
})

const { data: video, pending: videoPending, error: videoError, refresh: refreshVideo } = await useAsyncData<Video>(
  () => `studio-editor-video-${videoId.value}`,
  () => api.getVideo(videoId.value),
  { watch: [videoId] }
)

const duration = computed(() => Math.max(0, Math.floor(video.value?.durationSeconds ?? 0)))
const trimStart = ref(0)
const trimEnd = ref(0)
const splitPoints = ref<number[]>([])
const blurRegions = ref<Array<{ id: string; label: string; x: number; y: number; startSec: number; endSec: number }>>([])
const mutedSegments = ref<Array<{ startSec: number; endSec: number; reason: string }>>([])
const replacementAudio = ref('Mute Claimed Segment Only (Preserve Speech via AI Stem Split)')
const blurCounter = ref(0)

watch(duration, (d) => {
  if (d > 0 && trimEnd.value === 0) {
    trimEnd.value = d
  }
}, { immediate: true })

const job = ref<StudioEditJob | null>(null)
const jobError = ref<string | null>(null)
const dispatching = ref(false)
const polling = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const jobTone = computed(() => {
  const status: StudioEditJobStatus | undefined = job.value?.status
  if (status === 'READY') return 'success'
  if (status === 'FAILED') return 'danger'
  return 'info'
})

const jobMessage = computed(() => {
  const status = job.value?.status
  if (status === 'READY') return `Render complete. Edited rendition published to ${job.value?.outputPrefix} — the source video, URL, views, and comments are unchanged.`
  if (status === 'FAILED') return `Render failed: ${job.value?.error || 'the media worker rejected this edit'}. Adjust the EDL and dispatch again.`
  if (status === 'PROCESSING') return 'The media worker is rendering your edit (trim, splits, blurs, audio replacement).'
  return 'Edit queued on the media worker. This usually finishes in seconds.'
})

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  polling.value = false
}

onUnmounted(stopPolling)

async function pollJob(jobId: string) {
  polling.value = true
  const startedAt = Date.now()
  pollTimer = setInterval(async () => {
    try {
      const current = await api.getStudioEditJob(videoId.value, jobId)
      job.value = current
      if (current.status === 'READY' || current.status === 'FAILED' || Date.now() - startedAt > 90000) {
        stopPolling()
      }
    } catch (err) {
      jobError.value = err instanceof Error ? err.message : 'Could not reach the render job status endpoint.'
      stopPolling()
    }
  }, 1500)
}

function buildEdl(): StudioEditDecisionList {
  return {
    trimStartSec: Math.max(0, Math.floor(trimStart.value)),
    trimEndSec: Math.max(0, Math.floor(trimEnd.value)),
    splitPoints: [...splitPoints.value].sort((a, b) => a - b),
    blurRegions: blurRegions.value.map(b => ({ ...b })),
    mutedSegments: mutedSegments.value.map(m => ({ ...m })),
    replacementAudio: replacementAudio.value
  }
}

async function dispatchRenderJob() {
  if (dispatching.value || polling.value) return
  dispatching.value = true
  jobError.value = null
  try {
    const created = await api.createStudioEditJob(videoId.value, buildEdl())
    job.value = created
    stopPolling()
    await pollJob(created.jobId)
  } catch (err) {
    jobError.value = err instanceof Error ? err.message : 'Dispatch failed. Check the EDL values and try again.'
  } finally {
    dispatching.value = false
  }
}

function addSplitAtPlayhead() {
  const mid = Math.floor((trimStart.value + trimEnd.value) / 2)
  if (mid > trimStart.value && mid < trimEnd.value && !splitPoints.value.includes(mid)) {
    splitPoints.value.push(mid)
  }
}

function removeSplit(point: number) {
  splitPoints.value = splitPoints.value.filter(p => p !== point)
}

function addBlurRegion() {
  blurCounter.value += 1
  blurRegions.value.push({
    id: `blur_${Date.now()}_${blurCounter.value}`,
    label: 'Face / Screen Blur Box',
    x: 50,
    y: 50,
    startSec: trimStart.value,
    endSec: Math.min(trimEnd.value, trimStart.value + 60)
  })
}

function removeBlurRegion(id: string) {
  blurRegions.value = blurRegions.value.filter(b => b.id !== id)
}

function splitLeft(point: number): string {
  const span = Math.max(1, trimEnd.value - trimStart.value)
  return `${Math.min(100, Math.max(0, ((point - trimStart.value) / span) * 100))}%`
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-slate-900">Video Editor</h1>
        <p class="mt-1 text-xs text-slate-500 tabular-nums">
          Video ID: <code class="rounded-md border border-slate-200 bg-slate-100 px-1.5 py-0.5 font-mono text-slate-900">{{ videoId }}</code>
          <span v-if="video"> · {{ video.title }} · {{ duration }}s · Non-destructive: the source, URL, views, and comments are preserved.</span>
        </p>
      </div>
      <GvpButton
        variant="primary"
        :loading="dispatching || polling"
        :disabled="videoPending || !!videoError || !video"
        class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
        @click="dispatchRenderJob"
      >
        Save &amp; Dispatch Render Job
      </GvpButton>
    </div>

    <div v-if="videoPending" class="grid gap-3">
      <GvpSkeleton />
      <GvpSkeleton />
    </div>

    <GvpEmptyState
      v-else-if="videoError || !video"
      title="Video not found"
      description="This video is missing or belongs to another channel. Open it from Content to edit."
      action-label="Back to Content"
      @action="navigateTo('/content')"
    />

    <template v-else>
      <GvpBanner v-if="job" :tone="jobTone">
        <strong class="tabular-nums">Render job {{ job.jobId }} — {{ job.status }}</strong><br />
        {{ jobMessage }}
      </GvpBanner>
      <GvpBanner v-if="jobError" tone="danger">
        {{ jobError }}
        <button type="button" class="ml-2 font-semibold underline" @click="refreshVideo()">Retry loading video</button>
      </GvpBanner>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <section class="rounded-lg border border-slate-200 bg-white p-5 lg:col-span-2" aria-label="Preview and timeline">
          <div class="relative aspect-video w-full overflow-hidden rounded-lg border border-slate-200 bg-slate-100">
            <img
              v-if="video.thumbnailUrl"
              :src="video.thumbnailUrl"
              :alt="`Preview frame of ${video.title}`"
              class="h-full w-full object-cover"
            />
            <div v-else class="flex h-full items-center justify-center p-8 text-center text-sm font-medium text-slate-600 bg-slate-50" role="img" :aria-label="`No thumbnail for ${video.title}`">
              {{ video.title }}
            </div>
            <div
              v-for="box in blurRegions"
              :key="box.id"
              class="absolute flex h-16 w-36 items-center justify-center rounded-md border-2 border-slate-900 bg-white/90 text-[11px] font-semibold text-slate-900"
              :style="{ left: `${box.x}%`, top: `${box.y}%` }"
            >
              {{ box.label }}
            </div>
          </div>

          <div class="mt-5 flex flex-col gap-3 rounded-lg border border-slate-200 bg-slate-50 p-4 text-xs">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <span class="font-semibold text-slate-900 tabular-nums">Timeline (Trim: {{ trimStart }}s – {{ trimEnd }}s of {{ duration }}s)</span>
              <div class="flex gap-2">
                <button
                  type="button"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
                  @click="addSplitAtPlayhead"
                >
                  Split at midpoint
                </button>
                <button
                  type="button"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
                  @click="addBlurRegion"
                >
                  Add blur region
                </button>
              </div>
            </div>

            <div class="relative w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-slate-900" aria-label="Video track with split points">
              <span class="font-medium">Video track</span>
              <button
                v-for="sp in splitPoints"
                :key="sp"
                type="button"
                class="absolute top-0 h-full w-1 -ml-0.5 cursor-pointer rounded-md bg-slate-900 hover:bg-slate-700"
                :style="{ left: splitLeft(sp) }"
                :title="`Split at ${sp}s — activate to remove`"
                :aria-label="`Remove split at ${sp} seconds`"
                @click="removeSplit(sp)"
              />
              <span v-if="splitPoints.length === 0" class="ml-2 text-slate-500">No splits — the trim plays straight through.</span>
            </div>

            <div class="flex flex-wrap items-center gap-2 rounded-md border border-slate-200 bg-white px-3 py-2 text-slate-700 tabular-nums">
              <span>Blur regions: {{ blurRegions.length }} active mask(s)</span>
              <button
                v-for="box in blurRegions"
                :key="box.id"
                type="button"
                class="inline-flex items-center gap-1 rounded-md border border-slate-300 bg-slate-100 px-2.5 py-0.5 text-[11px] font-medium text-slate-700 hover:bg-slate-200"
                :title="`Remove ${box.label}`"
                @click="removeBlurRegion(box.id)"
              >
                <span>{{ box.label }}</span>
                <span class="text-slate-500">(Remove)</span>
              </button>
            </div>

            <div class="rounded-md border border-slate-200 bg-white px-3 py-2 text-slate-700 tabular-nums">
              <template v-if="mutedSegments.length > 0">
                Muted {{ mutedSegments[0]?.startSec }}s–{{ mutedSegments[0]?.endSec }}s · {{ replacementAudio }}
              </template>
              <template v-else>
                Full original audio · {{ replacementAudio }}
              </template>
            </div>
          </div>
        </section>

        <aside class="flex flex-col gap-4 rounded-lg border border-slate-200 bg-white p-5 text-xs" aria-label="Edit decision list controls">
          <h2 class="text-sm font-semibold text-slate-900">Edit Decision List</h2>

          <div class="grid grid-cols-2 gap-3">
            <div class="flex flex-col gap-1.5">
              <label for="editor-trim-in" class="font-medium text-slate-600">Trim in (sec)</label>
              <input
                id="editor-trim-in"
                v-model.number="trimStart"
                type="number"
                min="0"
                :max="duration"
                class="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 tabular-nums outline-none focus:border-slate-900"
              />
            </div>
            <div class="flex flex-col gap-1.5">
              <label for="editor-trim-out" class="font-medium text-slate-600">Trim out (sec)</label>
              <input
                id="editor-trim-out"
                v-model.number="trimEnd"
                type="number"
                min="0"
                :max="duration"
                class="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 tabular-nums outline-none focus:border-slate-900"
              />
            </div>
          </div>
          <p v-if="trimEnd <= trimStart" class="font-semibold text-red-600 tabular-nums" role="alert">
            Trim out must be greater than trim in (0–{{ duration }}s).
          </p>

          <div class="flex flex-col gap-1.5">
            <label for="editor-audio" class="font-medium text-slate-600">Audio replacement (resolves copyright claims)</label>
            <select
              id="editor-audio"
              v-model="replacementAudio"
              class="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-xs text-slate-900 outline-none focus:border-slate-900"
            >
              <option>Royalty-Free Synthwave Ambient #04 (StreamVerse Audio Library)</option>
              <option>Lo-Fi Chill Coding Beats #12 (Cleared Global License)</option>
              <option>Mute Claimed Segment Only (Preserve Speech via AI Stem Split)</option>
            </select>
          </div>

          <p class="leading-relaxed text-slate-500">
            Dispatch sends this EDL to the media worker render queue. The original file keeps serving until the edited rendition is ready.
          </p>
        </aside>
      </div>
    </template>
  </div>
</template>
