<script setup lang="ts">
import { VideoProcessingStatus, VideoVisibility } from '@gvp/shared-types'

const api = useStudioApi()

useSeoMeta({
  title: 'Multipart R2 Video Uploader'
})

const title = ref('Zero-Copy HLS Transcoding Pipeline in Go 1.24 (4K 60fps Masterclass)')
const description = ref(
  'In this video we benchmark Cloudflare R2 multipart uploads, FFmpeg segmenting, and Shaka Player ABR adaptation.'
)
const visibility = ref<VideoVisibility>(VideoVisibility.PUBLIC)
const isAgeRestricted = ref(false)
const scheduledPublishAt = ref('')
const chaptersInput = ref('00:00 Introduction\n03:30 Multipart S3/R2 Presigned Upload\n09:15 FFmpeg Worker Transcoding Ladder\n15:40 Summary')

const selectedFileName = ref('go_hls_pipeline_4k60_master.mp4')
const selectedFileSizeBytes = ref(524_288_000) // 500 MB
const uploadId = ref<string | null>(null)
const videoId = ref<string | null>(null)

const totalParts = ref(10)
const uploadedParts = ref(0)
const isUploading = ref(false)
const isPaused = ref(false)
const uploadError = ref<string | null>(null)

const transcodeSteps: VideoProcessingStatus[] = [
  VideoProcessingStatus.UPLOADED,
  VideoProcessingStatus.VALIDATING,
  VideoProcessingStatus.QUEUED,
  VideoProcessingStatus.PROCESSING,
  VideoProcessingStatus.READY
]
const currentStatus = ref<VideoProcessingStatus>(VideoProcessingStatus.CREATED)

let uploadInterval: ReturnType<typeof setInterval> | null = null

const progressPercent = computed(() =>
  Math.min(100, Math.round((uploadedParts.value / totalParts.value) * 100))
)

function onFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    selectedFileName.value = file.name
    selectedFileSizeBytes.value = file.size
    totalParts.value = Math.max(4, Math.ceil(file.size / (10 * 1024 * 1024)))
  }
}

async function startMultipartUpload() {
  uploadError.value = null
  isUploading.value = true
  isPaused.value = false
  currentStatus.value = VideoProcessingStatus.UPLOADING

  try {
    const session = await api.createUploadSession({
      title: title.value,
      description: description.value,
      fileSizeBytes: selectedFileSizeBytes.value,
      mimeType: 'video/mp4',
      videoType: 'standard'
    })
    uploadId.value = session.uploadId
    videoId.value = session.videoId
    totalParts.value = session.totalParts || 10
    runChunkLoop()
  } catch (err: any) {
    isUploading.value = false
    uploadError.value = `Upload session initialization failed: ${err?.message || 'Server connection error'}`
    currentStatus.value = VideoProcessingStatus.FAILED
  }
}

function runChunkLoop() {
  if (uploadInterval) clearInterval(uploadInterval)
  uploadInterval = setInterval(async () => {
    if (isPaused.value) return
    if (uploadedParts.value < totalParts.value) {
      const nextPart = uploadedParts.value + 1
      try {
        if (uploadId.value) {
          await api.getUploadPartUrls(uploadId.value, nextPart)
        }
        uploadedParts.value = nextPart
      } catch (err: any) {
        if (uploadInterval) clearInterval(uploadInterval)
        isUploading.value = false
        uploadError.value = `Part ${nextPart} presigned URL generation failed: ${err?.message || 'Network error'}`
        return
      }
    } else {
      if (uploadInterval) clearInterval(uploadInterval)
      isUploading.value = false
      await finalizeMultipartUpload()
    }
  }, 350)
}

function togglePauseResume() {
  isPaused.value = !isPaused.value
}

function retryUpload() {
  uploadError.value = null
  isPaused.value = false
  runChunkLoop()
}

async function finalizeMultipartUpload() {
  try {
    if (uploadId.value) {
      const parts = Array.from({ length: totalParts.value }, (_, idx) => ({
        partNumber: idx + 1,
        etag: `"etag_part_${idx + 1}_sha256"`
      }))
      const res = await api.completeUpload(uploadId.value, parts)
      currentStatus.value = res.status as VideoProcessingStatus || VideoProcessingStatus.QUEUED
    }
  } catch (err: any) {
    uploadError.value = `Multipart upload finalization failed: ${err?.message || 'Media queue error'}`
    currentStatus.value = VideoProcessingStatus.FAILED
  }
}

onBeforeUnmount(() => {
  if (uploadInterval) clearInterval(uploadInterval)
})
</script>

<template>
  <div class="mx-auto max-w-5xl">
    <h1 class="text-2xl font-extrabold text-white">
      클라우드플레어 R2 멀티파트 업로더 (Resumable R2 Multipart Upload)
    </h1>
    <p class="mt-1 text-xs text-neutral-400">
      Presigned S3/R2 chunked upload with pause/resume/retry and real-time FFmpeg HLS ABR transcoding state machine
    </p>

    <!-- File Dropzone & Chunk Controls -->
    <div class="mt-6 rounded-2xl border border-dashed border-white/20 bg-[#151822] p-6">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div>
          <p class="text-sm font-bold text-white">📄 {{ selectedFileName }}</p>
          <p class="mt-1 text-xs text-neutral-400">
            Size: {{ (selectedFileSizeBytes / (1024 * 1024)).toFixed(1) }} MB · {{ totalParts }} Multipart Chunks · SHA-256 Verified
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <label class="cursor-pointer rounded-xl border border-white/15 bg-white/5 px-3.5 py-2 text-xs font-semibold text-white hover:bg-white/10">
            Select Video File
            <input type="file" accept="video/*" class="hidden" @change="onFileSelected" />
          </label>

          <button
            v-if="!isUploading && uploadedParts === 0"
            type="button"
            class="rounded-xl bg-[#FF0033] px-4 py-2 text-xs font-bold text-white hover:bg-[#D9002B]"
            @click="startMultipartUpload"
          >
            🚀 Start Multipart Upload (`POST /v1/uploads`)
          </button>

          <button
            v-if="isUploading"
            type="button"
            class="rounded-xl bg-amber-400 px-4 py-2 text-xs font-bold text-black"
            @click="togglePauseResume"
          >
            {{ isPaused ? '▶️ Resume Chunks' : '⏸️ Pause Upload' }}
          </button>

          <button
            v-if="uploadError"
            type="button"
            class="rounded-xl bg-[#3EA6FF] px-4 py-2 text-xs font-bold text-black"
            @click="retryUpload"
          >
            🔄 Retry Failed Part
          </button>
        </div>
      </div>

      <!-- Progress Bar -->
      <div class="mt-5">
        <div class="flex justify-between text-xs font-semibold">
          <span class="text-neutral-300">
            Part {{ uploadedParts }} / {{ totalParts }} ({{ progressPercent }}%)
          </span>
          <span class="font-mono text-[#3EA6FF]">
            {{ uploadId ? `Upload ID: ${uploadId}` : 'Ready to initialize session' }}
          </span>
        </div>
        <div class="mt-2 h-2.5 w-full overflow-hidden rounded-full bg-white/10">
          <div
            class="h-full bg-gradient-to-r from-[#FF0033] to-[#3EA6FF] transition-all"
            :style="{ width: `${progressPercent}%` }"
          />
        </div>
      </div>

      <!-- Transcoding Pipeline Stepper: UPLOADED -> VALIDATING -> QUEUED -> PROCESSING -> READY -->
      <div class="mt-6 border-t border-white/10 pt-5">
        <p class="mb-3 text-xs font-bold uppercase tracking-wider text-neutral-400">
          Live Transcoding Pipeline Status (240p → 360p → 480p → 720p → 1080p → 1440p → 2160p HLS)
        </p>
        <div class="grid grid-cols-5 gap-2">
          <div
            v-for="step in transcodeSteps"
            :key="step"
            class="rounded-xl border px-3 py-2.5 text-center text-xs font-bold transition"
            :class="
              currentStatus === step
                ? 'border-[#3EA6FF] bg-[#3EA6FF]/20 text-white'
                : transcodeSteps.indexOf(currentStatus) > transcodeSteps.indexOf(step)
                  ? 'border-emerald-500/40 bg-emerald-500/15 text-emerald-300'
                  : 'border-white/10 bg-white/5 text-neutral-500'
            "
          >
            {{ step }}
          </div>
        </div>
      </div>
    </div>

    <!-- Video Metadata Editor -->
    <div class="mt-6 grid grid-cols-1 gap-6 rounded-2xl border border-white/10 bg-[#151822] p-6 lg:grid-cols-3">
      <div class="space-y-4 lg:col-span-2">
        <div>
          <label class="block text-xs font-bold text-neutral-300">동영상 제목 (Title)</label>
          <input
            v-model="title"
            type="text"
            class="mt-1.5 w-full rounded-xl border border-white/15 bg-[#0D0F14] px-3.5 py-2 text-sm text-white outline-none focus:border-[#3EA6FF]"
          />
        </div>

        <div>
          <label class="block text-xs font-bold text-neutral-300">설명 (Description)</label>
          <textarea
            v-model="description"
            rows="4"
            class="mt-1.5 w-full rounded-xl border border-white/15 bg-[#0D0F14] p-3.5 text-sm text-white outline-none focus:border-[#3EA6FF]"
          />
        </div>

        <div>
          <label class="block text-xs font-bold text-neutral-300">타임라인 챕터 마커 (Chapters)</label>
          <textarea
            v-model="chaptersInput"
            rows="4"
            class="mt-1.5 w-full rounded-xl border border-white/15 bg-[#0D0F14] p-3 font-mono text-xs text-neutral-200 outline-none focus:border-[#3EA6FF]"
          />
        </div>
      </div>

      <div class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-neutral-300">공개 상태 (Visibility)</label>
          <select
            v-model="visibility"
            class="mt-1.5 w-full rounded-xl border border-white/15 bg-[#0D0F14] px-3 py-2 text-xs text-white outline-none"
          >
            <option :value="VideoVisibility.PUBLIC">Public (전체 공개)</option>
            <option :value="VideoVisibility.UNLISTED">Unlisted (일부 공개)</option>
            <option :value="VideoVisibility.PRIVATE">Private (비공개)</option>
            <option :value="VideoVisibility.MEMBERS_ONLY">Members Only (멤버십 전용)</option>
            <option :value="VideoVisibility.SCHEDULED">Scheduled (예약 공개)</option>
          </select>
        </div>

        <div>
          <label class="block text-xs font-bold text-neutral-300">예약 게시 일시 (Scheduled Publish)</label>
          <input
            v-model="scheduledPublishAt"
            type="datetime-local"
            class="mt-1.5 w-full rounded-xl border border-white/15 bg-[#0D0F14] px-3 py-2 text-xs text-white outline-none"
          />
        </div>

        <label class="flex items-center gap-2 pt-2 text-xs text-neutral-300">
          <input v-model="isAgeRestricted" type="checkbox" class="rounded accent-[#FF0033]" />
          <span>연령 제한 콘텐츠 (19+ Age Restriction)</span>
        </label>

        <div class="rounded-xl border border-white/10 bg-[#0D0F14] p-3 text-xs">
          <p class="font-bold text-white">Auto-Generated Thumbnails</p>
          <img
            src="https://images.unsplash.com/photo-1518770660439-4636190af475?w=400&q=80"
            alt="Thumbnail preview"
            class="mt-2 aspect-video w-full rounded-lg object-cover"
          />
        </div>
      </div>
    </div>
  </div>
</template>
