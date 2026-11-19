<script setup lang="ts">
import type { VideoCaption } from '@gvp/shared-types'

const props = defineProps<{ videoId: string | null }>()
const api = useStudioApi()
const captions = ref<VideoCaption[]>([])
const loading = ref(false)
const failed = ref(false)
const lang = ref('ko')
const label = ref('')
const vttUrl = ref('')
const busy = ref(false)
const notice = ref<string | null>(null)

async function load() {
  if (!props.videoId) return
  loading.value = true
  failed.value = false
  try {
    captions.value = await api.listVideoCaptions(props.videoId)
  } catch { failed.value = true } finally { loading.value = false }
}

async function add() {
  if (!props.videoId || !vttUrl.value.trim() || busy.value) return
  busy.value = true
  notice.value = null
  try {
    const created = await api.addVideoCaption(props.videoId, { language: lang.value, label: label.value.trim() || undefined, vttUrl: vttUrl.value.trim() })
    captions.value.push(created)
    label.value = ''
    vttUrl.value = ''
  } catch { notice.value = '자막 등록에 실패했습니다.' } finally { busy.value = false }
}

async function remove(id: string) {
  if (!props.videoId) return
  try {
    await api.deleteVideoCaption(props.videoId, id)
    captions.value = captions.value.filter(c => c.id !== id)
  } catch { notice.value = '자막 삭제에 실패했습니다.' }
}

watch(() => props.videoId, load, { immediate: true })
</script>

<template>
  <div class="space-y-4">
    <p v-if="!videoId" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">
      먼저 파일을 업로드하면 자막을 추가할 수 있습니다.
    </p>
    <template v-else>
      <p v-if="loading" role="status" class="text-xs text-[var(--gvp-text-secondary)]">자막 목록을 불러오는 중…</p>
      <p v-else-if="failed" role="alert" class="text-xs font-medium text-[var(--gvp-live)]">자막 목록을 불러오지 못했습니다.</p>
      <ul v-else-if="captions.length" class="divide-y divide-[var(--gvp-border)] rounded-lg border border-[var(--gvp-border)]">
        <li v-for="c in captions" :key="c.id" class="flex items-center justify-between gap-3 px-3 py-2 text-xs">
          <span class="min-w-0 truncate text-[var(--gvp-text)]"><strong>{{ c.language }}</strong> · {{ c.label || c.vttUrl }}</span>
          <button type="button" class="shrink-0 rounded-md border border-[var(--gvp-border-strong)] px-2 py-1 font-medium hover:bg-[var(--gvp-canvas)]" @click="remove(c.id)">삭제</button>
        </li>
      </ul>
      <p v-else class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">등록된 자막이 없습니다.</p>
      <div class="grid grid-cols-1 gap-2 rounded-lg border border-[var(--gvp-border)] p-3 sm:grid-cols-[88px_1fr_2fr_auto]">
        <input v-model="lang" type="text" aria-label="자막 언어 코드" placeholder="ko" class="h-9 rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 font-mono text-xs outline-none focus:border-[var(--gvp-focus)]" />
        <input v-model="label" type="text" aria-label="자막 라벨" placeholder="라벨 (선택)" class="h-9 rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 text-xs outline-none focus:border-[var(--gvp-focus)]" />
        <input v-model="vttUrl" type="url" aria-label="VTT 파일 URL" placeholder="https://…/captions.vtt" class="h-9 rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 font-mono text-xs outline-none focus:border-[var(--gvp-focus)]" />
        <button type="button" :disabled="busy || !vttUrl.trim()" class="h-9 rounded-md bg-[var(--gvp-studio)] px-4 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50" @click="add">추가</button>
      </div>
      <p v-if="notice" role="status" class="text-xs text-[var(--gvp-text-secondary)]">{{ notice }}</p>
    </template>
  </div>
</template>
