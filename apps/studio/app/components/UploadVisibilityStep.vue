<script setup lang="ts">
import { VideoVisibility } from '@gvp/shared-types'

const props = defineProps<{ videoId: string | null; initial: VideoVisibility }>()
const api = useStudioApi()
const visibility = ref<VideoVisibility>(props.initial)
const busy = ref(false)
const notice = ref<string | null>(null)
const failed = ref<string | null>(null)

watch(() => props.initial, v => { visibility.value = v })

const options = [
  { id: VideoVisibility.PUBLIC, label: '공개', hint: '누구나 검색하고 시청할 수 있습니다' },
  { id: VideoVisibility.UNLISTED, label: '일부 공개', hint: '링크를 아는 사람만 시청할 수 있습니다' },
  { id: VideoVisibility.PRIVATE, label: '비공개', hint: '나만 볼 수 있습니다' },
  { id: VideoVisibility.MEMBERS_ONLY, label: '멤버십 전용', hint: '멤버십 구독자만 시청할 수 있습니다' },
] as const

async function save() {
  if (!props.videoId || busy.value) return
  busy.value = true
  notice.value = null
  failed.value = null
  try {
    await api.updateVideo(props.videoId, { visibility: visibility.value } as never)
    notice.value = '공개 상태가 저장되었습니다.'
  } catch { failed.value = '공개 상태 저장에 실패했습니다.' } finally { busy.value = false }
}
</script>

<template>
  <div class="space-y-3">
    <p v-if="!videoId" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">
      먼저 파일을 업로드하면 공개 상태를 설정할 수 있습니다.
    </p>
    <template v-else>
      <div
        v-for="o in options" :key="o.id"
        role="radio" :aria-checked="visibility === o.id" tabindex="0"
        class="flex cursor-pointer items-start gap-3 rounded-lg border p-3.5"
        :class="visibility === o.id ? 'border-[var(--gvp-studio)] bg-[var(--gvp-studio-soft)]' : 'border-[var(--gvp-border)] hover:bg-[var(--gvp-canvas)]'"
        @click="visibility = o.id"
        @keydown.enter="visibility = o.id" @keydown.space.prevent="visibility = o.id"
      >
        <span class="mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border" :class="visibility === o.id ? 'border-[var(--gvp-studio)]' : 'border-[var(--gvp-border-strong)]'">
          <span v-if="visibility === o.id" class="h-2 w-2 rounded-full bg-[var(--gvp-studio)]" />
        </span>
        <span>
          <span class="block text-sm font-semibold text-[var(--gvp-text)]">{{ o.label }}</span>
          <span class="block text-xs text-[var(--gvp-text-secondary)]">{{ o.hint }}</span>
        </span>
      </div>
      <div class="flex items-center gap-2">
        <button type="button" :disabled="busy" class="inline-flex h-9 items-center rounded-md bg-[var(--gvp-studio)] px-5 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50" @click="save">
          {{ busy ? '저장 중…' : '공개 상태 저장' }}
        </button>
        <p v-if="notice" role="status" class="text-xs font-medium text-[var(--gvp-success)]">{{ notice }}</p>
        <p v-if="failed" role="alert" class="text-xs font-medium text-[var(--gvp-live)]">{{ failed }}</p>
      </div>
    </template>
  </div>
</template>
