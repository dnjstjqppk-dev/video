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

const selectedFile = ref<File | null>(null)
const selectedFileName = ref('선택된 파일 없음 (Select a video file)')
const selectedFileSizeBytes = ref(0)
const uploadId = ref<string | null>(null)
const videoId = ref<string | null>(null)
const partSizeBytes = ref(10 * 1024 * 1024)

const totalParts = ref(0)
const uploadedParts = ref(0)
const isUploading = ref(false)
const isPaused = ref(false)
const uploadError = ref<string | null>(null)
const completedEtags = ref<Record<number, string>>({})

const transcodeSteps: VideoProcessingStatus[] = [
  VideoProcessingStatus.UPLOADED,
  VideoProcessingStatus.VALIDATING,
  VideoProcessingStatus.QUEUED,
  VideoProcessingStatus.PROCESSING,
  VideoProcessingStatus.READY
]
const currentStatus = ref<VideoProcessingStatus>(VideoProcessingStatus.CREATED)

let cancelled = false

const progressPercent = computed(() =>
  totalParts.value === 0 ? 0 : Math.min(100, Math.round((uploadedParts.value / totalParts.value) * 100))
)

function onFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  selectedFile.value = file
  selectedFileName.value = file.name
  selectedFileSizeBytes.value = file.size
  totalParts.value = Math.max(1, Math.ceil(file.size / partSizeBytes.value))
  uploadedParts.value = 0
  completedEtags.value = {}
  uploadId.value = null
  videoId.value = null
  uploadError.value = null
  currentStatus.value = VideoProcessingStatus.CREATED
}

async function startMultipartUpload() {
  if (!selectedFile.value) {
    uploadError.value = '먼저 업로드할 동영상 파일을 선택하세요.'
    return
  }
  uploadError.value = null
  isUploading.value = true
  isPaused.value = false
  cancelled = false
  currentStatus.value = VideoProcessingStatus.UPLOADING

  try {
    const session = await api.createUploadSession({
      title: title.value,
      description: description.value,
      fileSizeBytes: selectedFile.value.size,
      mimeType: selectedFile.value.type || 'video/mp4',
      videoType: 'standard'
    })
    uploadId.value = session.uploadId
    videoId.value = session.videoId
    partSizeBytes.value = session.partSize || partSizeBytes.value
    totalParts.value = session.totalParts || 1
    await uploadMissingParts()
    if (!cancelled && !uploadError.value) {
      await finalizeMultipartUpload()
    }
  } catch (err: any) {
    isUploading.value = false
    uploadError.value = `Upload session initialization failed: ${err?.message || 'Server connection error'}`
    currentStatus.value = VideoProcessingStatus.FAILED
  }
}

const studioToken = useCookie<string | null>('gvp_studio_token')

async function putPartBytes(uploadUrl: string, blob: Blob): Promise<string> {
  const headers: Record<string, string> = {
    'Content-Type': selectedFile.value?.type || 'video/mp4'
  }
  // Local relay chunks need the studio bearer token; S3 presigned URLs carry
  // their own SigV4 authorization and must not receive it.
  if (!uploadUrl.includes('X-Amz-Signature') && studioToken.value) {
    headers['Authorization'] = `Bearer ${studioToken.value}`
  }
  let lastError: unknown = null
  for (let attempt = 1; attempt <= 3; attempt++) {
    try {
      const res = await fetch(uploadUrl, {
        method: 'PUT',
        body: blob,
        headers,
        credentials: 'omit'
      })
      if (!res.ok) {
        throw new Error(`Chunk PUT failed with HTTP ${res.status}`)
      }
      const etag = (res.headers.get('etag') || '').replace(/"/g, '')
      if (!etag) {
        throw new Error('Storage did not return a part ETag')
      }
      return etag
    } catch (err) {
      lastError = err
      await new Promise(resolve => setTimeout(resolve, attempt * 500))
    }
  }
  throw lastError instanceof Error ? lastError : new Error('Chunk PUT failed')
}

async function uploadMissingParts() {
  const file = selectedFile.value
  if (!file || !uploadId.value) return
  for (let part = 1; part <= totalParts.value; part++) {
    if (cancelled) return
    while (isPaused.value && !cancelled) {
      await new Promise(resolve => setTimeout(resolve, 250))
    }
    if (completedEtags.value[part]) {
      uploadedParts.value = Math.max(uploadedParts.value, part)
      continue
    }
    try {
      const { uploadUrl } = await api.getUploadPartUrls(uploadId.value, part)
      const start = (part - 1) * partSizeBytes.value
      const chunk = file.slice(start, start + partSizeBytes.value)
      const etag = await putPartBytes(uploadUrl, chunk)
      completedEtags.value[part] = etag
      uploadedParts.value = part
    } catch (err: any) {
      isUploading.value = false
      uploadError.value = `Part ${part} upload failed: ${err?.message || 'Network error'}`
      currentStatus.value = VideoProcessingStatus.FAILED
      return
    }
  }
  currentStatus.value = VideoProcessingStatus.UPLOADED
}

function togglePauseResume() {
  isPaused.value = !isPaused.value
}

function retryUpload() {
  if (!selectedFile.value || !uploadId.value) return
  uploadError.value = null
  isPaused.value = false
  cancelled = false
  isUploading.value = true
  currentStatus.value = VideoProcessingStatus.UPLOADING
  void (async () => {
    await uploadMissingParts()
    if (!cancelled && !uploadError.value) {
      await finalizeMultipartUpload()
    } else {
      isUploading.value = false
    }
  })()
}

async function finalizeMultipartUpload() {
  try {
    if (uploadId.value) {
      const parts = Object.entries(completedEtags.value)
        .map(([partNumber, etag]) => ({ partNumber: Number(partNumber), etag }))
        .sort((a, b) => a.partNumber - b.partNumber)
      const res = await api.completeUpload(uploadId.value, parts)
      currentStatus.value = (res.status as VideoProcessingStatus) || VideoProcessingStatus.QUEUED
      if (res.videoId) {
        videoId.value = res.videoId
      }
    }
  } catch (err: any) {
    uploadError.value = `Multipart upload finalization failed: ${err?.message || 'Media queue error'}`
    currentStatus.value = VideoProcessingStatus.FAILED
  } finally {
    isUploading.value = false
  }
}

onBeforeUnmount(() => {
  cancelled = true
})
</script>

<template>
  <div class="mx-auto max-w-5xl space-y-6">
    <div class="border-b border-slate-200 pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-slate-900">
        클라우드플레어 R2 멀티파트 업로더 (Resumable R2 Multipart Upload)
      </h1>
      <p class="mt-1 text-xs text-slate-500">
        Presigned S3/R2 chunked upload with pause/resume/retry and real-time FFmpeg HLS ABR transcoding state machine
      </p>
    </div>

    <div
      v-if="uploadError"
      class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-xs font-medium text-red-700"
      role="alert"
    >
      {{ uploadError }}
    </div>

    <!-- File Dropzone & Chunk Controls -->
    <div class="rounded-lg border border-slate-200 bg-white p-6">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-start gap-3">
          <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-slate-200 bg-slate-50 text-slate-600">
            <svg class="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.5 2H4.5a1 1 0 00-1 1v10a1 1 0 001 1h7a1 1 0 001-1V5L9.5 2z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.5 2v3h3" />
            </svg>
          </div>
          <div>
            <p class="text-sm font-medium text-slate-900">{{ selectedFileName }}</p>
            <p class="mt-0.5 text-xs text-slate-500 tabular-nums">
              Size: {{ (selectedFileSizeBytes / (1024 * 1024)).toFixed(1) }} MB · {{ totalParts }} Multipart Chunks · ETag Verified
            </p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <label class="cursor-pointer inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors">
            Select Video File
            <input type="file" accept="video/*" class="hidden" @change="onFileSelected" />
          </label>

          <button
            v-if="!isUploading && uploadedParts === 0"
            type="button"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
            @click="startMultipartUpload"
          >
            Start Multipart Upload (POST /v1/uploads)
          </button>

          <button
            v-if="isUploading"
            type="button"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
            @click="togglePauseResume"
          >
            {{ isPaused ? 'Resume Chunks' : 'Pause Upload' }}
          </button>

          <button
            v-if="uploadError"
            type="button"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
            @click="retryUpload"
          >
            Retry Failed Part
          </button>
        </div>
      </div>

      <!-- Progress Bar -->
      <div class="mt-5">
        <div class="flex justify-between text-xs font-medium">
          <span class="text-slate-600 tabular-nums">
            Part {{ uploadedParts }} / {{ totalParts }} ({{ progressPercent }}%)
          </span>
          <span class="font-mono text-slate-500 tabular-nums">
            {{ uploadId ? `Upload ID: ${uploadId}` : 'Ready to initialize session' }}
          </span>
        </div>
        <div class="mt-2 h-2 w-full overflow-hidden rounded-md bg-slate-100">
          <div
            class="h-full bg-slate-900 transition-all"
            :style="{ width: `${progressPercent}%` }"
          />
        </div>
      </div>

      <!-- Transcoding Pipeline Stepper -->
      <div class="mt-6 border-t border-slate-200 pt-5">
        <p class="mb-3 text-xs font-semibold uppercase tracking-wider text-slate-500 tabular-nums">
          Live Transcoding Pipeline Status (240p → 360p → 480p → 720p → 1080p → 1440p → 2160p HLS)
        </p>
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-5">
          <div
            v-for="step in transcodeSteps"
            :key="step"
            class="rounded-md border px-3 py-2 text-center text-xs font-medium transition-colors"
            :class="
              currentStatus === step
                ? 'border-slate-900 bg-slate-900 text-white'
                : transcodeSteps.indexOf(currentStatus) > transcodeSteps.indexOf(step)
                  ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                  : 'border-slate-200 bg-slate-50 text-slate-500'
            "
          >
            {{ step }}
          </div>
        </div>
      </div>
    </div>

    <!-- Video Metadata Editor -->
    <div class="grid grid-cols-1 gap-6 rounded-lg border border-slate-200 bg-white p-6 lg:grid-cols-3">
      <div class="space-y-4 lg:col-span-2">
        <div>
          <label class="block text-xs font-medium text-slate-700">동영상 제목 (Title)</label>
          <input
            v-model="title"
            type="text"
            class="mt-1.5 w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 outline-none focus:border-slate-900"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-slate-700">설명 (Description)</label>
          <textarea
            v-model="description"
            rows="4"
            class="mt-1.5 w-full rounded-md border border-slate-300 bg-white p-3 text-sm text-slate-900 outline-none focus:border-slate-900"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-slate-700">타임라인 챕터 마커 (Chapters)</label>
          <textarea
            v-model="chaptersInput"
            rows="4"
            class="mt-1.5 w-full rounded-md border border-slate-300 bg-slate-50 p-3 font-mono text-xs text-slate-800 tabular-nums outline-none focus:border-slate-900"
          />
        </div>
      </div>

      <div class="space-y-4">
        <div>
          <label class="block text-xs font-medium text-slate-700">공개 상태 (Visibility)</label>
          <select
            v-model="visibility"
            class="mt-1.5 w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-xs font-medium text-slate-900 outline-none focus:border-slate-900"
          >
            <option :value="VideoVisibility.PUBLIC">Public (전체 공개)</option>
            <option :value="VideoVisibility.UNLISTED">Unlisted (일부 공개)</option>
            <option :value="VideoVisibility.PRIVATE">Private (비공개)</option>
            <option :value="VideoVisibility.MEMBERS_ONLY">Members Only (멤버십 전용)</option>
            <option :value="VideoVisibility.SCHEDULED">Scheduled (예약 공개)</option>
          </select>
        </div>

        <div>
          <label class="block text-xs font-medium text-slate-700">예약 게시 일시 (Scheduled Publish)</label>
          <input
            v-model="scheduledPublishAt"
            type="datetime-local"
            class="mt-1.5 w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-xs text-slate-900 tabular-nums outline-none focus:border-slate-900"
          />
        </div>

        <label class="flex items-center gap-2 pt-2 text-xs text-slate-700">
          <input v-model="isAgeRestricted" type="checkbox" class="rounded-md border-slate-300 accent-slate-900" />
          <span class="tabular-nums">연령 제한 콘텐츠 (19+ Age Restriction)</span>
        </label>

        <div class="rounded-lg border border-slate-200 bg-slate-50 p-3 text-xs">
          <p class="font-medium text-slate-900">Auto-Generated Thumbnails</p>
          <img
            src="https://images.unsplash.com/photo-1518770660439-4636190af475?w=400&q=80"
            alt="Thumbnail preview"
            class="mt-2 aspect-video w-full rounded-md border border-slate-200 object-cover"
          />
        </div>
      </div>
    </div>
  </div>
</template>
