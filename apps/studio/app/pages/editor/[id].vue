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
  <div class="editor">
    <div class="editor__header">
      <div>
        <h1 class="editor__title">Video Editor</h1>
        <p class="editor__subtitle">
          Video ID: <code class="editor__code">{{ videoId }}</code>
          <span v-if="video"> · {{ video.title }} · {{ duration }}s · Non-destructive: the source, URL, views, and comments are preserved.</span>
        </p>
      </div>
      <GvpButton
        variant="primary"
        :loading="dispatching || polling"
        :disabled="videoPending || !!videoError || !video"
        @click="dispatchRenderJob"
      >
        Save &amp; Dispatch Render Job
      </GvpButton>
    </div>

    <div v-if="videoPending" class="editor__loading">
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
      <GvpBanner v-if="job" :tone="jobTone" class="editor__banner">
        <strong>Render job {{ job.jobId }} — {{ job.status }}</strong><br />
        {{ jobMessage }}
      </GvpBanner>
      <GvpBanner v-if="jobError" tone="danger" class="editor__banner">
        {{ jobError }}
        <button type="button" class="editor__link" @click="refreshVideo()">Retry loading video</button>
      </GvpBanner>

      <div class="editor__grid">
        <section class="editor__panel editor__preview" aria-label="Preview and timeline">
          <div class="editor__frame">
            <img
              v-if="video.thumbnailUrl"
              :src="video.thumbnailUrl"
              :alt="`Preview frame of ${video.title}`"
              class="editor__frame-img"
            />
            <div v-else class="editor__frame-fallback" role="img" :aria-label="`No thumbnail for ${video.title}`">
              {{ video.title }}
            </div>
            <div
              v-for="box in blurRegions"
              :key="box.id"
              class="editor__blur-box"
              :style="{ left: `${box.x}%`, top: `${box.y}%` }"
            >
              {{ box.label }}
            </div>
          </div>

          <div class="editor__timeline">
            <div class="editor__timeline-head">
              <span class="editor__timeline-title">Timeline (Trim: {{ trimStart }}s – {{ trimEnd }}s of {{ duration }}s)</span>
              <div class="editor__timeline-actions">
                <GvpButton variant="secondary" size="sm" @click="addSplitAtPlayhead">Split at midpoint</GvpButton>
                <GvpButton variant="secondary" size="sm" @click="addBlurRegion">Add blur region</GvpButton>
              </div>
            </div>

            <div class="editor__track editor__track--video" aria-label="Video track with split points">
              <span>Video track</span>
              <button
                v-for="sp in splitPoints"
                :key="sp"
                type="button"
                class="editor__split"
                :style="{ left: splitLeft(sp) }"
                :title="`Split at ${sp}s — activate to remove`"
                :aria-label="`Remove split at ${sp} seconds`"
                @click="removeSplit(sp)"
              />
              <span v-if="splitPoints.length === 0" class="editor__track-hint">No splits — the trim plays straight through.</span>
            </div>

            <div class="editor__track editor__track--blur">
              Blur regions: {{ blurRegions.length }} active mask(s)
              <button
                v-for="box in blurRegions"
                :key="box.id"
                type="button"
                class="editor__chip"
                :title="`Remove ${box.label}`"
                @click="removeBlurRegion(box.id)"
              >
                {{ box.label }} ✕
              </button>
            </div>

            <div class="editor__track editor__track--audio">
              <template v-if="mutedSegments.length > 0">
                Muted {{ mutedSegments[0]?.startSec }}s–{{ mutedSegments[0]?.endSec }}s · {{ replacementAudio }}
              </template>
              <template v-else>
                Full original audio · {{ replacementAudio }}
              </template>
            </div>
          </div>
        </section>

        <aside class="editor__panel editor__inspector" aria-label="Edit decision list controls">
          <h2 class="editor__inspector-title">Edit Decision List</h2>

          <div class="editor__fields">
            <div class="editor__field">
              <label for="editor-trim-in">Trim in (sec)</label>
              <input
                id="editor-trim-in"
                v-model.number="trimStart"
                type="number"
                min="0"
                :max="duration"
              />
            </div>
            <div class="editor__field">
              <label for="editor-trim-out">Trim out (sec)</label>
              <input
                id="editor-trim-out"
                v-model.number="trimEnd"
                type="number"
                min="0"
                :max="duration"
              />
            </div>
          </div>
          <p v-if="trimEnd <= trimStart" class="editor__field-error" role="alert">
            Trim out must be greater than trim in (0–{{ duration }}s).
          </p>

          <div class="editor__field">
            <label for="editor-audio">Audio replacement (resolves copyright claims)</label>
            <select id="editor-audio" v-model="replacementAudio">
              <option>Royalty-Free Synthwave Ambient #04 (StreamVerse Audio Library)</option>
              <option>Lo-Fi Chill Coding Beats #12 (Cleared Global License)</option>
              <option>Mute Claimed Segment Only (Preserve Speech via AI Stem Split)</option>
            </select>
          </div>

          <p class="editor__note">
            Dispatch sends this EDL to the media worker render queue. The original file keeps serving until the edited rendition is ready.
          </p>
        </aside>
      </div>
    </template>
  </div>
</template>

<style scoped>
.editor { display: flex; flex-direction: column; gap: 1rem; }
.editor__header { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 1rem; }
.editor__title { font-size: 1.5rem; font-weight: 800; color: var(--gvp-text); }
.editor__subtitle { margin-top: 0.25rem; font-size: 0.75rem; color: var(--gvp-text-muted); }
.editor__code { color: var(--gvp-link); font-family: var(--gvp-font-mono); }
.editor__loading { display: grid; gap: 0.75rem; }
.editor__banner { margin-top: 0.25rem; }
.editor__link { background: none; border: none; padding: 0; color: inherit; font-weight: 700; text-decoration: underline; cursor: pointer; }
.editor__grid { display: grid; grid-template-columns: 1fr; gap: 1.5rem; }
@media (min-width: 1024px) { .editor__grid { grid-template-columns: 2fr 1fr; } }
.editor__panel { border: 1px solid var(--gvp-border); background: var(--gvp-surface-1); border-radius: var(--gvp-radius-xl); padding: 1rem; }
.editor__frame { position: relative; aspect-ratio: 16 / 9; width: 100%; overflow: hidden; border-radius: var(--gvp-radius-md); background: #000; }
.editor__frame-img { height: 100%; width: 100%; object-fit: cover; }
.editor__frame-fallback { display: flex; height: 100%; align-items: center; justify-content: center; padding: 2rem; text-align: center; color: var(--gvp-text-secondary); background: linear-gradient(135deg, var(--gvp-surface-2), var(--gvp-surface-3)); }
.editor__blur-box { position: absolute; display: flex; height: 4rem; width: 9rem; align-items: center; justify-content: center; border: 2px solid var(--gvp-link); border-radius: var(--gvp-radius-md); background: rgba(0, 0, 0, 0.5); color: var(--gvp-text); font-size: 0.625rem; font-weight: 700; backdrop-filter: blur(6px); }
.editor__timeline { margin-top: 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; border: 1px solid var(--gvp-border); border-radius: var(--gvp-radius-md); background: var(--gvp-canvas); padding: 1rem; font-size: 0.75rem; }
.editor__timeline-head { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.75rem; }
.editor__timeline-title { font-weight: 700; color: var(--gvp-text); }
.editor__timeline-actions { display: flex; gap: 0.5rem; }
.editor__track { position: relative; width: 100%; border-radius: var(--gvp-radius-md); padding: 0.375rem 0.75rem; line-height: 1.75rem; color: var(--gvp-text); }
.editor__track--video { background: var(--gvp-surface-3); }
.editor__track--blur { background: var(--gvp-info-soft); color: var(--gvp-text-secondary); }
.editor__track--audio { background: var(--gvp-warning-soft); color: var(--gvp-text-secondary); }
.editor__track-hint { margin-left: 0.5rem; color: var(--gvp-text-faint); }
.editor__split { position: absolute; top: 0; height: 100%; width: 4px; margin-left: -2px; background: var(--gvp-brand); border: none; border-radius: 2px; cursor: pointer; }
.editor__split:hover { background: var(--gvp-brand-hover); }
.editor__chip { margin-left: 0.5rem; border: 1px solid var(--gvp-border-strong); background: var(--gvp-surface-2); color: var(--gvp-text); border-radius: var(--gvp-radius-pill); padding: 0.125rem 0.625rem; font-size: 0.6875rem; cursor: pointer; }
.editor__inspector { display: flex; flex-direction: column; gap: 1rem; font-size: 0.75rem; }
.editor__inspector-title { font-size: 0.875rem; font-weight: 700; color: var(--gvp-text); }
.editor__fields { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.editor__field { display: flex; flex-direction: column; gap: 0.375rem; color: var(--gvp-text-muted); }
.editor__field input, .editor__field select { width: 100%; border: 1px solid var(--gvp-border-strong); border-radius: var(--gvp-radius-md); background: var(--gvp-canvas); color: var(--gvp-text); padding: 0.5rem 0.75rem; font-size: 0.8125rem; }
.editor__field input:focus-visible, .editor__field select:focus-visible, .editor__split:focus-visible, .editor__chip:focus-visible { outline: none; box-shadow: var(--gvp-focus-ring); }
.editor__field-error { color: var(--gvp-danger); font-weight: 600; }
.editor__note { color: var(--gvp-text-faint); line-height: 1.5; }
</style>
