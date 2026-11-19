<script setup lang="ts">
const props = defineProps<{ videoId: string | null; status: string }>()
const api = useStudioApi()
const frames = ref<Array<{ source: string; url: string }>>([])
const current = ref('')
const loading = ref(false)
const failed = ref<string | null>(null)
const busy = ref(false)
const notice = ref<string | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const sourceLabel: Record<string, string> = {
  poster: '기본 포스터',
  auto_1: '자동 프레임 1',
  auto_2: '자동 프레임 2',
  auto_3: '자동 프레임 3',
  custom: '커스텀 썸네일',
}

async function load() {
  if (!props.videoId) return
  loading.value = true
  failed.value = null
  try {
    const res = await api.getVideoFrames(props.videoId)
    frames.value = res.frames || []
    current.value = res.current || ''
  } catch {
    failed.value = '썸네일 목록을 불러오지 못했습니다.'
  } finally {
    loading.value = false
  }
}

async function select(source: string) {
  if (!props.videoId || busy.value) return
  busy.value = true
  notice.value = null
  failed.value = null
  try {
    const res = await api.selectVideoThumbnail(props.videoId, source)
    current.value = res.thumbnailUrl
    notice.value = `'${sourceLabel[source] || source}'(으)로 변경했습니다.`
  } catch {
    failed.value = '썸네일 변경에 실패했습니다.'
  } finally {
    busy.value = false
  }
}

async function upload(file: File | undefined) {
  if (!props.videoId || !file || busy.value) return
  busy.value = true
  notice.value = null
  failed.value = null
  try {
    const res = await api.uploadVideoThumbnail(props.videoId, file)
    current.value = res.thumbnailUrl
    notice.value = '커스텀 썸네일을 업로드했습니다.'
    await load()
  } catch {
    failed.value = '썸네일 업로드에 실패했습니다. (JPEG/PNG/WebP, 5MB 이하)'
  } finally {
    busy.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

watch(() => [props.videoId, props.status], load, { immediate: true })
</script>

<template>
  <div class="space-y-3">
    <p v-if="!videoId" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">
      먼저 파일을 업로드하면 썸네일을 설정할 수 있습니다.
    </p>
    <template v-else>
      <div v-if="current" class="flex items-center gap-3 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-3">
        <img :src="current" alt="현재 썸네일" class="aspect-video w-32 shrink-0 rounded-md border border-[var(--gvp-border)] object-cover" />
        <p class="text-xs text-[var(--gvp-text-secondary)]">현재 썸네일</p>
      </div>
      <p v-if="loading" role="status" class="text-xs text-[var(--gvp-text-secondary)]">프레임 목록을 불러오는 중…</p>
      <div v-else-if="frames.length" class="grid grid-cols-2 gap-2 sm:grid-cols-3">
        <button
          v-for="f in frames" :key="f.source" type="button" :disabled="busy"
          class="group overflow-hidden rounded-lg border text-left transition-colors disabled:opacity-60"
          :class="current === f.url ? 'border-[var(--gvp-studio)]' : 'border-[var(--gvp-border)] hover:border-[var(--gvp-border-strong)]'"
          @click="select(f.source)"
        >
          <img :src="f.url" :alt="sourceLabel[f.source] || f.source" loading="lazy" class="aspect-video w-full object-cover" />
          <span class="tabular-nums block px-2 py-1.5 text-[11px] font-medium" :class="current === f.url ? 'bg-[var(--gvp-studio-soft)] text-[var(--gvp-studio)]' : 'text-[var(--gvp-text-secondary)]'">
            {{ sourceLabel[f.source] || f.source }}{{ current === f.url ? ' · 사용 중' : '' }}
          </span>
        </button>
      </div>
      <p v-else-if="!failed" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">
        자동 추출 프레임은 인코딩이 끝나면 표시됩니다. 커스텀 썸네일은 지금 바로 올릴 수 있습니다.
      </p>
      <div class="flex flex-wrap items-center gap-2">
        <label class="inline-flex h-9 cursor-pointer items-center rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-xs font-medium hover:bg-[var(--gvp-canvas)]">
          커스텀 썸네일 업로드
          <input ref="fileInput" type="file" accept="image/jpeg,image/png,image/webp" class="hidden" @change="upload(($event.target as HTMLInputElement).files?.[0])" />
        </label>
        <p v-if="notice" role="status" class="text-xs font-medium text-[var(--gvp-success)]">{{ notice }}</p>
        <p v-if="failed" role="alert" class="text-xs font-medium text-[var(--gvp-live)]">{{ failed }}</p>
      </div>
    </template>
  </div>
</template>
