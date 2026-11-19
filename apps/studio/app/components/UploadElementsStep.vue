<script setup lang="ts">
import type { VideoElement } from '@gvp/shared-types'

const props = defineProps<{ videoId: string | null }>()
const durationSeconds = ref(0)
const api = useStudioApi()
const elements = ref<VideoElement[]>([])
const loading = ref(false)
const failed = ref<string | null>(null)
const busy = ref(false)
const kind = ref<'end_screen' | 'card'>('card')
const title = ref('')
const targetType = ref<'video' | 'playlist' | 'channel' | 'url'>('video')
const target = ref('')
const startSec = ref(0)
const endSec = ref(10)

const kindLabel = (k: string) => k === 'end_screen' ? '최종 화면' : '카드'

async function load() {
  if (!props.videoId) return
  loading.value = true
  failed.value = null
  try {
    elements.value = await api.getVideoElements(props.videoId)
  } catch {
    failed.value = '요소 목록을 불러오지 못했습니다.'
  } finally {
    loading.value = false
  }
  try {
    const v = await api.getVideo(props.videoId)
    durationSeconds.value = v.durationSeconds || 0
  } catch { /* duration hint stays hidden */ }
}

async function add() {
  if (!props.videoId || busy.value) return
  busy.value = true
  failed.value = null
  try {
    const created = await api.addVideoElement(props.videoId, {
      kind: kind.value,
      title: title.value.trim(),
      targetType: targetType.value,
      target: target.value.trim(),
      startSec: Number(startSec.value),
      endSec: Number(endSec.value),
    })
    elements.value.push(created)
    title.value = ''
    target.value = ''
  } catch {
    failed.value = '요소 추가에 실패했습니다. (제목·대상·0초 이상 구간, 영상 길이 이내)'
  } finally {
    busy.value = false
  }
}

async function remove(id: string) {
  if (!props.videoId) return
  try {
    await api.deleteVideoElement(props.videoId, id)
    elements.value = elements.value.filter(e => e.id !== id)
  } catch {
    failed.value = '요소 삭제에 실패했습니다.'
  }
}

watch(() => props.videoId, load, { immediate: true })
</script>

<template>
  <div class="space-y-3">
    <p v-if="!videoId" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">
      먼저 파일을 업로드하면 최종 화면과 카드를 추가할 수 있습니다.
    </p>
    <template v-else>
      <p v-if="loading" role="status" class="text-xs text-[var(--gvp-text-secondary)]">요소 목록을 불러오는 중…</p>
      <ul v-else-if="elements.length" class="divide-y divide-[var(--gvp-border)] rounded-lg border border-[var(--gvp-border)]">
        <li v-for="el in elements" :key="el.id" class="flex flex-wrap items-center justify-between gap-2 px-3 py-2 text-xs">
          <span class="min-w-0 text-[var(--gvp-text)]">
            <strong class="mr-2 rounded-md bg-[var(--gvp-elevated)] px-1.5 py-0.5 text-[10px] font-bold uppercase">{{ kindLabel(el.kind) }}</strong>
            {{ el.title }}
            <span class="tabular-nums text-[var(--gvp-text-secondary)]">· {{ el.startSec }}초–{{ el.endSec }}초 → {{ el.target }}</span>
          </span>
          <button type="button" class="shrink-0 rounded-md border border-[var(--gvp-border-strong)] px-2 py-1 font-medium hover:bg-[var(--gvp-canvas)]" @click="remove(el.id)">삭제</button>
        </li>
      </ul>
      <p v-else class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">등록된 최종 화면·카드가 없습니다.</p>
      <div class="grid grid-cols-2 gap-2 rounded-lg border border-[var(--gvp-border)] p-3 sm:grid-cols-3">
        <label class="text-xs font-medium text-[var(--gvp-text-secondary)]">종류
          <select v-model="kind" class="mt-1 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs text-[var(--gvp-text)]">
            <option value="card">카드</option>
            <option value="end_screen">최종 화면</option>
          </select>
        </label>
        <label class="col-span-2 text-xs font-medium text-[var(--gvp-text-secondary)]">제목
          <input v-model="title" type="text" placeholder="예: 다음 영상 보기" class="mt-1 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 text-xs outline-none focus:border-[var(--gvp-focus)]" />
        </label>
        <label class="text-xs font-medium text-[var(--gvp-text-secondary)]">대상 종류
          <select v-model="targetType" class="mt-1 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs text-[var(--gvp-text)]">
            <option value="video">동영상</option>
            <option value="playlist">재생목록</option>
            <option value="channel">채널</option>
            <option value="url">외부 링크</option>
          </select>
        </label>
        <label class="col-span-2 text-xs font-medium text-[var(--gvp-text-secondary)]">대상 (ID 또는 URL)
          <input v-model="target" type="text" placeholder="동영상 ID 또는 https://…" class="mt-1 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 font-mono text-xs outline-none focus:border-[var(--gvp-focus)]" />
        </label>
        <label class="text-xs font-medium text-[var(--gvp-text-secondary)]">시작 (초)
          <input v-model.number="startSec" type="number" min="0" class="tabular-nums mt-1 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 text-xs outline-none focus:border-[var(--gvp-focus)]" />
        </label>
        <label class="text-xs font-medium text-[var(--gvp-text-secondary)]">종료 (초)
          <input v-model.number="endSec" type="number" min="1" class="tabular-nums mt-1 h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 text-xs outline-none focus:border-[var(--gvp-focus)]" />
        </label>
        <span class="col-span-2 flex items-end sm:col-span-1">
          <button type="button" :disabled="busy || !title.trim() || !target.trim()" class="h-9 w-full rounded-md bg-[var(--gvp-studio)] px-4 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50" @click="add">요소 추가</button>
        </span>
      </div>
      <p v-if="durationSeconds > 0" class="tabular-nums text-[11px] text-[var(--gvp-text-secondary)]">영상 길이: {{ durationSeconds }}초 (종료 시각은 길이 이내여야 합니다)</p>
      <p v-if="failed" role="alert" class="text-xs font-medium text-[var(--gvp-live)]">{{ failed }}</p>
    </template>
  </div>
</template>
