<script setup lang="ts">
import type { Comment } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import { useStudioApi } from '~/composables/useStudioApi'

useSeoMeta({
  title: 'Channel Comments & Moderation — vravra studio'
})

const api = useStudioApi()
const searchQuery = ref('')
const page = ref(1)
const limit = 20
const actionError = ref<string | null>(null)
const actingIds = ref<Set<string>>(new Set())

const { data, pending, error, refresh } = await useAsyncData(
  () => `studio-comments-${page.value}`,
  async () => {
    return await api.getStudioComments(page.value, limit)
  },
  { watch: [page] }
)

const comments = computed(() => data.value?.items ?? [])
const total = computed(() => data.value?.total ?? 0)
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
  <div class="max-w-6xl space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-[var(--gvp-border)] pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">채널 댓글 및 검토 관리 (Comments &amp; Mentions)</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          내 채널 영상에 달린 댓글을 하트·고정·삭제하고 시청자와 소통하세요
        </p>
      </div>
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <!-- Filter Bar -->
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-4 py-3">
      <p class="text-sm font-medium text-[var(--gvp-text)] tabular-nums" role="status">
        전체 댓글 <span class="font-semibold text-[var(--gvp-text)]">{{ total.toLocaleString() }}</span>개
      </p>
      <label class="flex items-center gap-2 text-xs font-medium text-[var(--gvp-text-secondary)]">
        <span>검색</span>
        <input
          v-model="searchQuery"
          type="search"
          placeholder="댓글 내용 또는 작성자..."
          class="h-9 w-64 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)]"
        />
      </label>
    </div>

    <!-- Loading -->
    <div v-if="pending" class="space-y-3" role="status" aria-label="불러오는 중">
      <div v-for="n in 4" :key="n" class="h-24 animate-pulse rounded-lg bg-[var(--gvp-elevated)]" />
    </div>

    <!-- Error -->
    <GvpBanner v-else-if="error" tone="danger">
      댓글을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <!-- Empty -->
    <GvpEmptyState
      v-else-if="filteredComments.length === 0"
      title="표시할 댓글이 없습니다"
      :description="searchQuery ? '검색어를 바꾸거나 필터를 초기화하세요.' : '아직 채널 영상에 댓글이 없습니다.'"
    />

    <!-- Comments List -->
    <div v-else class="divide-y divide-[var(--gvp-border)] overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <div
        v-for="cmt in filteredComments"
        :key="cmt.id"
        class="flex items-start gap-4 p-4 text-xs hover:bg-[var(--gvp-canvas)] transition-colors"
      >
        <img
          v-if="cmt.author.avatarUrl"
          :src="cmt.author.avatarUrl"
          :alt="cmt.author.displayName"
          class="h-10 w-10 shrink-0 rounded-full object-cover"
        />
        <span
          v-else
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[var(--gvp-elevated)] text-sm font-semibold text-[var(--gvp-text)]"
          aria-hidden="true"
        >
          {{ initials(cmt.author.displayName) }}
        </span>
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2 tabular-nums">
            <span class="font-semibold text-[var(--gvp-text)]">{{ cmt.author.displayName }}</span>
            <span class="text-[var(--gvp-text-muted)]">· {{ formatDate(cmt.createdAt) }}</span>
            <span class="text-[var(--gvp-text-secondary)]">· Likes {{ cmt.likeCount.toLocaleString() }}</span>
            <span v-if="cmt.isPinned" class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-elevated)] px-2 py-0.5 text-[11px] font-medium text-[var(--gvp-text)]">
              Pinned
            </span>
            <span v-if="cmt.isHeartedByCreator" class="rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-2 py-0.5 text-[11px] font-medium text-[var(--gvp-live)]" title="크리에이터 하트">
              Creator Hearted
            </span>
          </div>

          <p class="mt-0.5 font-mono text-[11px] text-[var(--gvp-text-secondary)]">on video: {{ cmt.videoId }}</p>
          <p class="mt-2 text-sm text-[var(--gvp-text)]">{{ cmt.content }}</p>

          <!-- Actions -->
          <div class="mt-3 flex flex-wrap items-center gap-2">
            <button
              type="button"
              :disabled="actingIds.has(cmt.id)"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-40"
              @click="toggleHeart(cmt)"
            >
              {{ cmt.isHeartedByCreator ? '하트 취소' : '하트' }}
            </button>
            <button
              type="button"
              :disabled="actingIds.has(cmt.id)"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-40"
              @click="togglePin(cmt)"
            >
              {{ cmt.isPinned ? '고정 해제' : '상단 고정' }}
            </button>
            <button
              type="button"
              :disabled="actingIds.has(cmt.id)"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-live)] bg-[var(--gvp-live-soft)] border border-[var(--gvp-live)] rounded-md hover:bg-[var(--gvp-live-soft)] transition-colors disabled:opacity-40"
              @click="removeComment(cmt)"
            >
              삭제
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="!pending && !error && totalPages > 1" class="flex items-center justify-center gap-3 text-xs">
      <button
        type="button"
        :disabled="page <= 1"
        class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-40"
        @click="page -= 1"
      >
        이전
      </button>
      <span class="text-[var(--gvp-text-secondary)] tabular-nums" role="status">{{ page }} / {{ totalPages }} (총 {{ total }}개)</span>
      <button
        type="button"
        :disabled="page >= totalPages"
        class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors disabled:opacity-40"
        @click="page += 1"
      >
        다음
      </button>
    </div>
  </div>
</template>
