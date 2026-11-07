<script setup lang="ts">
import type { Comment } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import { useStudioApi } from '~/composables/useStudioApi'

useSeoMeta({
  title: 'Channel Comments & Moderation — Creator Studio'
})

const api = useStudioApi()
const searchQuery = ref('')
const page = ref(1)
const limit = 20
const total = ref(0)
const actionError = ref<string | null>(null)
const actingIds = ref<Set<string>>(new Set())

const { data, pending, error, refresh } = await useAsyncData(
  () => `studio-comments-${page.value}`,
  async () => {
    const res = await api.getStudioComments(page.value, limit)
    total.value = res.total
    return res.items
  },
  { watch: [page] }
)

const comments = computed(() => data.value ?? [])
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit)))

const filteredComments = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return comments.value
  return comments.value.filter(c =>
    c.content.toLowerCase().includes(q) || c.author.displayName.toLowerCase().includes(q)
  )
})

function initials(name: string): string {
  return (name || '?').trim().slice(0, 1).toUpperCase()
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '-' : d.toLocaleDateString('ko-KR')
}

async function toggleHeart(cmt: Comment) {
  actionError.value = null
  actingIds.value.add(cmt.id)
  try {
    const res = await api.heartComment(cmt.id)
    cmt.isHeartedByCreator = res.isHearted
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '하트 처리에 실패했습니다.'
  } finally {
    actingIds.value.delete(cmt.id)
  }
}

async function togglePin(cmt: Comment) {
  actionError.value = null
  actingIds.value.add(cmt.id)
  try {
    const res = await api.pinComment(cmt.id)
    cmt.isPinned = res.isPinned
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '고정 처리에 실패했습니다.'
  } finally {
    actingIds.value.delete(cmt.id)
  }
}

async function removeComment(cmt: Comment) {
  actionError.value = null
  actingIds.value.add(cmt.id)
  try {
    await api.deleteComment(cmt.id)
    await refresh()
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '댓글 삭제에 실패했습니다.'
  } finally {
    actingIds.value.delete(cmt.id)
  }
}
</script>

<template>
  <div class="max-w-6xl">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 댓글 및 검토 관리 (Comments & Mentions)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          내 채널 영상에 달린 댓글을 하트·고정·삭제하고 시청자와 소통하세요
        </p>
      </div>
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible class="mt-4" @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <!-- Filter Bar -->
    <div class="mt-6 flex flex-wrap items-center justify-between gap-4 border-b border-white/10 pb-3">
      <p class="text-sm font-semibold text-white" role="status">
        전체 댓글 <span class="text-[#3EA6FF]">{{ total.toLocaleString() }}</span>개
      </p>
      <label class="flex items-center gap-2 text-xs text-neutral-400">
        검색
        <input
          v-model="searchQuery"
          type="search"
          placeholder="댓글 내용 또는 작성자..."
          class="w-64 rounded-xl border border-white/15 bg-[#151822] px-3 py-1.5 text-xs text-white outline-none focus:border-[#3EA6FF]"
        />
      </label>
    </div>

    <!-- Loading -->
    <div v-if="pending" class="mt-6 space-y-4" role="status" aria-label="불러오는 중">
      <div v-for="n in 4" :key="n" class="h-28 animate-pulse rounded-2xl bg-white/5" />
    </div>

    <!-- Error -->
    <GvpBanner v-else-if="error" tone="danger" class="mt-6">
      댓글을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <!-- Empty -->
    <GvpEmptyState
      v-else-if="filteredComments.length === 0"
      class="mt-6"
      title="표시할 댓글이 없습니다"
      :description="searchQuery ? '검색어를 바꾸거나 필터를 초기화하세요.' : '아직 채널 영상에 댓글이 없습니다.'"
    />

    <!-- Comments List -->
    <div v-else class="mt-6 space-y-4">
      <div
        v-for="cmt in filteredComments"
        :key="cmt.id"
        class="flex items-start gap-4 rounded-2xl border border-white/10 bg-[#151822] p-4 text-xs"
      >
        <img
          v-if="cmt.author.avatarUrl"
          :src="cmt.author.avatarUrl"
          :alt="cmt.author.displayName"
          class="h-10 w-10 shrink-0 rounded-full object-cover"
        />
        <span
          v-else
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[#3EA6FF]/20 text-sm font-black text-[#3EA6FF]"
          aria-hidden="true"
        >
          {{ initials(cmt.author.displayName) }}
        </span>
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-bold text-white">{{ cmt.author.displayName }}</span>
            <span class="text-neutral-400">· {{ formatDate(cmt.createdAt) }}</span>
            <span class="text-neutral-500">· ♥ {{ cmt.likeCount.toLocaleString() }}</span>
            <span v-if="cmt.isPinned" class="rounded bg-white/10 px-1.5 py-0.5 text-[10px] text-neutral-300">
              Pinned
            </span>
            <span v-if="cmt.isHeartedByCreator" class="text-rose-400" title="크리에이터 하트">♥</span>
          </div>

          <p class="mt-0.5 font-mono text-[11px] text-[#3EA6FF]">on video: {{ cmt.videoId }}</p>
          <p class="mt-2 text-sm text-neutral-200">{{ cmt.content }}</p>

          <!-- Actions -->
          <div class="mt-3 flex flex-wrap items-center gap-2">
            <button
              type="button"
              :disabled="actingIds.has(cmt.id)"
              class="rounded-lg bg-white/10 px-2.5 py-1 text-white hover:bg-white/20 disabled:opacity-40"
              @click="toggleHeart(cmt)"
            >
              {{ cmt.isHeartedByCreator ? '♥ 하트 취소' : '♡ 하트' }}
            </button>
            <button
              type="button"
              :disabled="actingIds.has(cmt.id)"
              class="rounded-lg bg-white/10 px-2.5 py-1 text-white hover:bg-white/20 disabled:opacity-40"
              @click="togglePin(cmt)"
            >
              {{ cmt.isPinned ? '고정 해제' : '상단 고정' }}
            </button>
            <button
              type="button"
              :disabled="actingIds.has(cmt.id)"
              class="rounded-lg bg-red-600/20 px-2.5 py-1 font-bold text-red-400 hover:bg-red-600/40 disabled:opacity-40"
              @click="removeComment(cmt)"
            >
              삭제
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="!pending && !error && totalPages > 1" class="mt-6 flex items-center justify-center gap-2 text-xs">
      <button
        type="button"
        :disabled="page <= 1"
        class="rounded-lg border border-white/15 px-3 py-1.5 text-neutral-200 disabled:opacity-40"
        @click="page -= 1"
      >
        이전
      </button>
      <span class="text-neutral-400" role="status">{{ page }} / {{ totalPages }} (총 {{ total }}개)</span>
      <button
        type="button"
        :disabled="page >= totalPages"
        class="rounded-lg border border-white/15 px-3 py-1.5 text-neutral-200 disabled:opacity-40"
        @click="page += 1"
      >
        다음
      </button>
    </div>
  </div>
</template>
