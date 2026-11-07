<script setup lang="ts">
import type { Video, VideoCaption } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import GvpModal from '@gvp/ui/src/components/GvpModal.vue'
import GvpTextInput from '@gvp/ui/src/components/GvpTextInput.vue'
import { useStudioApi } from '~/composables/useStudioApi'

useSeoMeta({
  title: 'Video Subtitles & Captions — Creator Studio'
})

const api = useStudioApi()

const tracksByVideo = ref<Record<string, VideoCaption[]>>({})
const loadingTracks = ref<Set<string>>(new Set())
const trackError = ref<Record<string, string>>({})
const actionError = ref<string | null>(null)

const addTarget = ref<Video | null>(null)
const newLanguage = ref('ko')
const newLabel = ref('')
const newVttUrl = ref('')
const adding = ref(false)
const deletingIds = ref<Set<string>>(new Set())

const { data, pending, error, refresh } = await useAsyncData('studio-subtitle-videos', async () => {
  const res = await api.getStudioVideos(1, 20)
  return res.items
})

const videos = computed(() => data.value ?? [])

async function loadTracks(video: Video) {
  if (tracksByVideo.value[video.id] || loadingTracks.value.has(video.id)) return
  loadingTracks.value.add(video.id)
  try {
    tracksByVideo.value[video.id] = await api.listVideoCaptions(video.id)
  } catch (e: unknown) {
    trackError.value[video.id] = e instanceof Error ? e.message : '자막 목록을 불러오지 못했습니다.'
  } finally {
    loadingTracks.value.delete(video.id)
  }
}

watch(videos, (list) => {
  for (const v of list) void loadTracks(v)
}, { immediate: true })

function openAddModal(video: Video) {
  addTarget.value = video
  newLanguage.value = 'ko'
  newLabel.value = ''
  newVttUrl.value = ''
  actionError.value = null
}

async function submitAddTrack() {
  if (!addTarget.value || !newLanguage.value.trim() || !newVttUrl.value.trim()) return
  adding.value = true
  actionError.value = null
  try {
    const created = await api.addVideoCaption(addTarget.value.id, {
      language: newLanguage.value.trim(),
      label: newLabel.value.trim() || undefined,
      vttUrl: newVttUrl.value.trim(),
    })
    const list = tracksByVideo.value[addTarget.value.id] || []
    tracksByVideo.value[addTarget.value.id] = [...list, created]
    addTarget.value = null
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '자막 등록에 실패했습니다.'
  } finally {
    adding.value = false
  }
}

async function removeTrack(video: Video, captionId: string) {
  deletingIds.value.add(captionId)
  actionError.value = null
  try {
    await api.deleteVideoCaption(video.id, captionId)
    tracksByVideo.value[video.id] = (tracksByVideo.value[video.id] || []).filter(t => t.id !== captionId)
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '자막 삭제에 실패했습니다.'
  } finally {
    deletingIds.value.delete(captionId)
  }
}

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${String(s).padStart(2, '0')}`
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-extrabold text-white">동영상 자막 관리 (Subtitles & Captions)</h1>
      <p class="mt-1 text-xs text-neutral-400">
        영상별 WebVTT 자막 트랙을 등록·삭제하세요. 등록된 자막은 플레이어에서 선택할 수 있습니다.
      </p>
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <div v-if="pending" class="space-y-4" role="status" aria-label="불러오는 중">
      <div v-for="n in 3" :key="n" class="h-36 animate-pulse rounded-2xl bg-white/5" />
    </div>

    <GvpBanner v-else-if="error" tone="danger">
      영상을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <GvpEmptyState
      v-else-if="videos.length === 0"
      title="자막을 관리할 영상이 없습니다"
      description="영상을 먼저 업로드하면 자막 트랙을 등록할 수 있습니다."
      action-label="업로드하러 가기"
      @action="navigateTo('/upload')"
    />

    <div v-else class="space-y-4">
      <div
        v-for="video in videos"
        :key="video.id"
        class="rounded-2xl border border-white/10 bg-[#151822] p-5"
      >
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="min-w-0">
            <p class="truncate text-sm font-bold text-white">{{ video.title }}</p>
            <p class="mt-0.5 font-mono text-[11px] text-neutral-400">{{ formatDuration(video.durationSeconds) }} · {{ video.id }}</p>
          </div>
          <button
            type="button"
            class="rounded-xl bg-[#3EA6FF] px-4 py-2 text-xs font-bold text-black hover:bg-[#65B8FF]"
            @click="openAddModal(video)"
          >
            + 자막 추가
          </button>
        </div>

        <div v-if="loadingTracks.has(video.id)" class="mt-3 h-10 animate-pulse rounded-xl bg-white/5" role="status" aria-label="자막 불러오는 중" />
        <GvpBanner v-else-if="trackError[video.id]" tone="danger" class="mt-3">
          {{ trackError[video.id] }}
          <button type="button" class="font-bold underline" @click="delete trackError[video.id]; loadTracks(video)">다시 시도</button>
        </GvpBanner>
        <p v-else-if="(tracksByVideo[video.id] || []).length === 0" class="mt-3 text-xs text-neutral-400">
          등록된 자막이 없습니다.
        </p>
        <ul v-else class="mt-3 space-y-2">
          <li
            v-for="track in tracksByVideo[video.id]"
            :key="track.id"
            class="flex flex-wrap items-center justify-between gap-2 rounded-xl bg-white/5 px-4 py-2.5 text-xs"
          >
            <div class="flex items-center gap-3">
              <span class="rounded bg-[#3EA6FF]/20 px-2 py-0.5 font-mono font-bold text-[#3EA6FF]">{{ track.language }}</span>
              <span class="text-neutral-200">{{ track.label }}</span>
              <a :href="track.vttUrl" target="_blank" rel="noopener noreferrer" class="font-mono text-[11px] text-neutral-400 underline hover:text-neutral-200">
                VTT 열기
              </a>
            </div>
            <button
              type="button"
              :disabled="deletingIds.has(track.id)"
              class="rounded-lg bg-red-600/20 px-2.5 py-1 font-bold text-red-300 hover:bg-red-600/40 disabled:opacity-40"
              @click="removeTrack(video, track.id)"
            >
              삭제
            </button>
          </li>
        </ul>
      </div>
    </div>

    <GvpModal :model-value="addTarget !== null" title="자막 트랙 추가" @update:model-value="(v: boolean) => { if (!v) addTarget = null }">
      <p class="text-xs text-neutral-400">
        <strong class="text-white">{{ addTarget?.title }}</strong>에 WebVTT(.vtt) 자막을 등록합니다.
      </p>
      <div class="mt-4 space-y-3">
        <GvpTextInput v-model="newLanguage" input-id="cap-lang" label="언어 코드 (ISO 639-1, 예: ko, en, ja)" placeholder="ko" />
        <GvpTextInput v-model="newLabel" input-id="cap-label" label="표시 이름 (선택사항)" placeholder="한국어" />
        <GvpTextInput v-model="newVttUrl" input-id="cap-url" label="WebVTT 파일 URL" placeholder="https://…" />
      </div>
      <template #footer>
        <button
          type="button"
          class="rounded-full bg-white/10 px-5 py-2 text-xs font-bold text-neutral-200 hover:bg-white/20"
          @click="addTarget = null"
        >
          취소
        </button>
        <button
          type="button"
          :disabled="adding || !newLanguage.trim() || !newVttUrl.trim()"
          class="rounded-full bg-[#3EA6FF] px-5 py-2 text-xs font-bold text-black hover:bg-[#65B8FF] disabled:opacity-50"
          @click="submitAddTrack"
        >
          {{ adding ? '등록 중…' : '자막 등록' }}
        </button>
      </template>
    </GvpModal>
  </div>
</template>
