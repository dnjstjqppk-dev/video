<script setup lang="ts">
import { VideoProcessingStatus, VideoVisibility } from '@gvp/shared-types'

const api = useStudioApi()
const route = useRoute()

useSeoMeta({
  title: '동영상 업로드'
})

const steps = ['세부정보', '동영상 요소', '검토', '공개 상태'] as const
const step = ref(0)
const title = ref('')
const description = ref('')
const category = ref('일반')
const tagsText = ref('')
const drmProtected = ref(false)
const aiGenerated = ref(false)
const aiTool = ref('')

function parseTags(): string[] {
  return tagsText.value.split(',').map(t => t.trim()).filter(t => t !== '')
}
const videoType = ref<'standard' | 'shorts'>(route.query.type === 'shorts' ? 'shorts' : 'standard')
const visibility = ref<VideoVisibility>(VideoVisibility.PRIVATE)
const detailsNotice = ref<string | null>(null)
const detailsFailed = ref<string | null>(null)
const detailsBusy = ref(false)

const selectedFile = ref<File | null>(null)
const selectedFileName = ref('동영상 파일을 선택하거나 드래그하세요')
const selectedFileSizeBytes = ref(0)
const dragOver = ref(false)
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
const jobError = ref('')

let cancelled = false
let pollTimer: ReturnType<typeof setTimeout> | null = null

const progressPercent = computed(() =>
  totalParts.value === 0 ? 0 : Math.min(100, Math.round((uploadedParts.value / totalParts.value) * 100))
)
const phaseLabel = computed(() => {
  if (isUploading.value) return `업로드 ${progressPercent.value}%`
  switch (currentStatus.value) {
    case VideoProcessingStatus.READY: return '검사 완료'
    case VideoProcessingStatus.FAILED: return '처리 실패'
    case VideoProcessingStatus.PROCESSING:
    case VideoProcessingStatus.VALIDATING: return 'HD 화질 처리 중'
    case VideoProcessingStatus.QUEUED:
    case VideoProcessingStatus.UPLOADED: return '처리 대기 중'
    default: return '준비됨'
  }
})

function pickFile(file: File | undefined) {
  if (!file) return
  // The picker suggests video/*, but drag-and-drop and some browsers let
  // anything through: reject declared non-video types here (ffprobe is the
  // final gate for empty-type files).
  if (file.type && !file.type.startsWith('video/')) {
    uploadError.value = `동영상 파일만 업로드할 수 있습니다 (${file.type}).`
    return
  }
  selectedFile.value = file
  selectedFileName.value = file.name
  selectedFileSizeBytes.value = file.size
  totalParts.value = Math.max(1, Math.ceil(file.size / partSizeBytes.value))
  uploadedParts.value = 0
  completedEtags.value = {}
  uploadId.value = null
  videoId.value = null
  uploadError.value = null
  jobError.value = ''
  currentStatus.value = VideoProcessingStatus.CREATED
  if (!title.value) title.value = file.name.replace(/\.[^.]+$/, '')
}

function onFileSelected(event: Event) {
  pickFile((event.target as HTMLInputElement).files?.[0])
}

function onDrop(event: DragEvent) {
  dragOver.value = false
  pickFile(event.dataTransfer?.files?.[0])
}

async function saveDetails() {
  if (!videoId.value || detailsBusy.value) return
  detailsBusy.value = true
  detailsNotice.value = null
  detailsFailed.value = null
  try {
    await api.updateVideo(videoId.value, { title: title.value, description: description.value, category: category.value, tags: parseTags(), aiGenerated: aiGenerated.value, aiTool: aiTool.value.trim() } as never)
    detailsNotice.value = '세부정보가 저장되었습니다.'
  } catch { detailsFailed.value = '세부정보 저장에 실패했습니다.' } finally { detailsBusy.value = false }
}

async function startMultipartUpload() {
  if (!selectedFile.value) {
    uploadError.value = '먼저 업로드할 동영상 파일을 선택하세요.'
    return
  }
  if (!title.value.trim()) {
    uploadError.value = '동영상 제목을 입력하세요.'
    step.value = 0
    return
  }
  uploadError.value = null
  isUploading.value = true
  isPaused.value = false
  cancelled = false
  currentStatus.value = VideoProcessingStatus.UPLOADING

  try {
    const session = await api.createUploadSession({
      title: title.value.trim(),
      description: description.value,
      fileSizeBytes: selectedFile.value.size,
      mimeType: selectedFile.value.type || 'video/mp4',
      videoType: videoType.value,
      tags: parseTags(),
      drmProtected: drmProtected.value || undefined
    })
    uploadId.value = session.uploadId
    videoId.value = session.videoId
    partSizeBytes.value = session.partSize || partSizeBytes.value
    totalParts.value = session.totalParts || 1
    await uploadMissingParts()
    if (!cancelled && !uploadError.value) {
      await finalizeMultipartUpload()
    }
  } catch (err: unknown) {
    isUploading.value = false
    uploadError.value = `업로드 세션 시작 실패: ${err instanceof Error ? err.message : '서버 연결 오류'}`
    currentStatus.value = VideoProcessingStatus.FAILED
  }
}

const studioToken = useCookie<string | null>('gvp_studio_token')

async function putPartBytes(uploadUrl: string, blob: Blob): Promise<string> {
  const headers: Record<string, string> = {
    'Content-Type': selectedFile.value?.type || 'video/mp4'
  }
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
        throw new Error(`청크 PUT 실패 (HTTP ${res.status})`)
      }
      const etag = (res.headers.get('etag') || '').replace(/"/g, '')
      if (!etag) {
        throw new Error('스토리지에서 파트 ETag을 반환하지 않았습니다')
      }
      return etag
    } catch (err) {
      lastError = err
      await new Promise(resolve => setTimeout(resolve, attempt * 500))
    }
  }
  throw lastError instanceof Error ? lastError : new Error('청크 PUT 실패')
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
    } catch (err: unknown) {
      isUploading.value = false
      uploadError.value = `파트 ${part} 업로드 실패: ${err instanceof Error ? err.message : '네트워크 오류'}`
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

function stopPolling() {
  if (pollTimer) clearTimeout(pollTimer)
  pollTimer = null
}

async function pollStatus() {
  if (!uploadId.value || cancelled) return
  try {
    const st = await api.getUploadStatus(uploadId.value)
    currentStatus.value = (st.status as VideoProcessingStatus) || currentStatus.value
    jobError.value = st.error || ''
    if (st.status === 'READY' || st.status === 'FAILED') {
      stopPolling()
      return
    }
  } catch { /* keep polling */ }
  stopPolling()
  pollTimer = setTimeout(pollStatus, 3000)
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
      stopPolling()
      pollTimer = setTimeout(pollStatus, 3000)
    }
  } catch (err: unknown) {
    uploadError.value = `업로드 완료 처리 실패: ${err instanceof Error ? err.message : '미디어 큐 오류'}`
    currentStatus.value = VideoProcessingStatus.FAILED
  } finally {
    isUploading.value = false
  }
}

onBeforeUnmount(() => {
  cancelled = true
  stopPolling()
})
</script>

<template>
  <div class="mx-auto max-w-5xl space-y-6">
    <div class="border-b border-[var(--gvp-border)] pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">동영상 업로드</h1>
      <p class="tabular-nums mt-1 text-xs text-[var(--gvp-text-secondary)]">현재 단계: {{ phaseLabel }}</p>
    </div>

    <div
      v-if="uploadError"
      class="rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-4 py-3 text-xs font-medium text-[var(--gvp-live)]"
      role="alert"
    >
      {{ uploadError }}
    </div>

    <!-- File Dropzone & Chunk Controls -->
    <div
      class="rounded-lg border bg-[var(--gvp-surface)] p-6"
      :class="dragOver ? 'border-[var(--gvp-studio)]' : 'border-[var(--gvp-border)]'"
      @dragover.prevent="dragOver = true" @dragleave.prevent="dragOver = false" @drop.prevent="onDrop"
    >
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex min-w-0 items-start gap-3">
          <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text-secondary)]">
            <svg class="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.5 2H4.5a1 1 0 00-1 1v10a1 1 0 001 1h7a1 1 0 001-1V5L9.5 2z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.5 2v3h3" />
            </svg>
          </div>
          <div class="min-w-0">
            <p class="truncate text-sm font-medium text-[var(--gvp-text)]">{{ selectedFileName }}</p>
            <p class="tabular-nums mt-0.5 text-xs text-[var(--gvp-text-secondary)]">
              크기: {{ (selectedFileSizeBytes / (1024 * 1024)).toFixed(1) }} MB · {{ totalParts }}개 청크 · ETag 검증
            </p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <label class="inline-flex h-9 cursor-pointer items-center justify-center rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-sm font-medium text-[var(--gvp-text)] transition-colors hover:bg-[var(--gvp-canvas)]">
            파일 선택
            <input type="file" accept="video/*" class="hidden" @change="onFileSelected" />
          </label>

          <button
            v-if="!isUploading && uploadedParts === 0"
            type="button"
            class="inline-flex h-9 items-center justify-center rounded-full bg-[var(--gvp-studio)] px-4 text-sm font-medium text-white transition hover:opacity-90 active:scale-[0.97]"
            @click="startMultipartUpload"
          >
            업로드 시작
          </button>

          <button
            v-if="isUploading"
            type="button"
            class="inline-flex h-9 items-center justify-center rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-sm font-medium text-[var(--gvp-text)] transition-colors hover:bg-[var(--gvp-canvas)]"
            @click="togglePauseResume"
          >
            {{ isPaused ? '이어서 업로드' : '일시중지' }}
          </button>

          <button
            v-if="uploadError"
            type="button"
            class="inline-flex h-9 items-center justify-center rounded-full bg-[var(--gvp-studio)] px-4 text-sm font-medium text-white transition-colors hover:opacity-90"
            @click="retryUpload"
          >
            실패한 파트 재시도
          </button>
        </div>
      </div>

      <!-- Progress Bar -->
      <div class="mt-5">
        <div class="flex justify-between text-xs font-medium">
          <span class="text-[var(--gvp-text-secondary)] tabular-nums">
            파트 {{ uploadedParts }} / {{ totalParts }} ({{ progressPercent }}%)
          </span>
          <span class="font-mono text-[var(--gvp-text-secondary)] tabular-nums">
            {{ uploadId ? `업로드 ID: ${uploadId}` : phaseLabel }}
          </span>
        </div>
        <div class="mt-2 h-2 w-full overflow-hidden rounded-md bg-[var(--gvp-elevated)]" role="progressbar" :aria-valuenow="progressPercent" aria-valuemin="0" aria-valuemax="100">
          <div
            class="h-full bg-[var(--gvp-studio)] transition-all"
            :style="{ width: `${progressPercent}%` }"
          />
        </div>
      </div>

      <!-- Transcoding Pipeline Stepper -->
      <div class="mt-6 border-t border-[var(--gvp-border)] pt-5">
        <p class="tabular-nums mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-secondary)]">
          인코딩 파이프라인 (240p → 1080p → 4K HLS)
        </p>
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-5">
          <div
            v-for="s in transcodeSteps"
            :key="s"
            class="rounded-md border px-3 py-2 text-center text-xs font-medium transition-colors"
            :class="
              currentStatus === s
                ? 'border-[var(--gvp-studio)] bg-[var(--gvp-studio)] text-white'
                : transcodeSteps.indexOf(currentStatus) > transcodeSteps.indexOf(s)
                  ? 'border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]'
                  : 'border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text-secondary)]'
            "
          >
            {{ s }}
          </div>
        </div>
        <p v-if="jobError" role="alert" class="mt-2 text-xs font-medium text-[var(--gvp-live)]">{{ jobError }}</p>
      </div>
    </div>

    <!-- 4-step wizard -->
    <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <div class="flex gap-1 overflow-x-auto border-b border-[var(--gvp-border)] p-2" role="tablist" aria-label="업로드 단계">
        <button
          v-for="(label, i) in steps"
          :key="label"
          type="button"
          role="tab"
          :aria-selected="step === i"
          :disabled="i > 0 && !videoId"
          class="flex h-10 flex-1 items-center justify-center gap-2 whitespace-nowrap rounded-md px-4 text-xs font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-40"
          :class="step === i ? 'bg-[var(--gvp-text)] text-[var(--gvp-canvas)]' : 'text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]'"
          @click="step = i"
        >
          <span class="tabular-nums flex h-5 w-5 items-center justify-center rounded-full text-[10px]" :class="step === i ? 'bg-[var(--gvp-canvas)] text-[var(--gvp-text)]' : 'bg-[var(--gvp-elevated)]'">{{ i + 1 }}</span>
          {{ label }}
        </button>
      </div>

      <div class="p-6">
        <div v-show="step === 0" role="tabpanel" aria-label="세부정보">
          <UploadDetailsStep
            v-model:title="title"
            v-model:description="description"
            v-model:video-type="videoType"
            v-model:category="category"
            v-model:tags-text="tagsText"
            v-model:drm-protected="drmProtected"
            v-model:ai-generated="aiGenerated"
            v-model:ai-tool="aiTool"
            :type-locked="!!uploadId"
          />
          <section aria-label="재생목록" class="mt-6 border-t border-[var(--gvp-border)] pt-4">
            <h3 class="mb-2 text-sm font-semibold text-[var(--gvp-text)]">재생목록</h3>
            <UploadPlaylistAdd :video-id="videoId" />
          </section>
          <div class="mt-4 flex items-center gap-2">
            <button v-if="videoId" type="button" :disabled="detailsBusy" class="inline-flex h-9 items-center rounded-full bg-[var(--gvp-studio)] px-5 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50" @click="saveDetails">
              {{ detailsBusy ? '저장 중…' : '세부정보 저장' }}
            </button>
            <p v-if="detailsNotice" role="status" class="text-xs font-medium text-[var(--gvp-success)]">{{ detailsNotice }}</p>
            <p v-if="detailsFailed" role="alert" class="text-xs font-medium text-[var(--gvp-live)]">{{ detailsFailed }}</p>
          </div>
          <section aria-label="썸네일" class="mt-8 border-t border-[var(--gvp-border)] pt-6">
            <h3 class="mb-3 text-sm font-semibold text-[var(--gvp-text)]">썸네일</h3>
            <UploadThumbnailPicker :video-id="videoId" :status="currentStatus" />
          </section>
        </div>
        <div v-show="step === 1" role="tabpanel" aria-label="동영상 요소" class="space-y-8">
          <section aria-label="자막">
            <h3 class="mb-3 text-sm font-semibold text-[var(--gvp-text)]">자막</h3>
            <UploadCaptionsStep :video-id="videoId" />
          </section>
          <section aria-label="최종 화면 및 카드">
            <h3 class="mb-3 text-sm font-semibold text-[var(--gvp-text)]">최종 화면 및 카드</h3>
            <UploadElementsStep :video-id="videoId" />
          </section>
        </div>
        <div v-show="step === 2" role="tabpanel" aria-label="검토">
          <UploadReviewStep :video-id="videoId" :status="currentStatus" :job-error="jobError" />
        </div>
        <div v-show="step === 3" role="tabpanel" aria-label="공개 상태">
          <UploadVisibilityStep :video-id="videoId" :initial="visibility" />
        </div>

        <div class="mt-6 flex items-center justify-between border-t border-[var(--gvp-border)] pt-4">
          <button type="button" :disabled="step === 0" class="inline-flex h-9 items-center rounded-full border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-xs font-medium hover:bg-[var(--gvp-canvas)] disabled:opacity-40" @click="step -= 1">이전 단계</button>
          <button v-if="step < 3" type="button" :disabled="!videoId" class="inline-flex h-9 items-center rounded-full bg-[var(--gvp-studio)] px-5 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-40" @click="step += 1">다음 단계</button>
          <NuxtLink v-else-if="videoId" to="/content" class="inline-flex h-9 items-center rounded-full bg-[var(--gvp-studio)] px-5 text-xs font-semibold text-white hover:opacity-90">콘텐츠에서 확인</NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>
