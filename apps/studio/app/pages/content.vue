<script setup lang="ts">
import { VideoVisibility, VideoProcessingStatus, type Video } from '@gvp/shared-types'
import { useStudioApi } from '~/composables/useStudioApi'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import { resolveStatusBadgeTone } from '@gvp/ui'

useSeoMeta({
  title: 'Content Manager — Videos, Shorts, Live & Playlists'
})

const api = useStudioApi()

interface StudioContentRow {
  id: string
  title: string
  type: string
  visibility: VideoVisibility
  status: VideoProcessingStatus
  views: number
  comments: number
  publishedAt: string
  selected: boolean
}

const activeTab = ref<'all' | 'video' | 'shorts' | 'live'>('all')
const statusFilter = ref<'all' | VideoProcessingStatus>('all')
const config = useRuntimeConfig()
const page = ref(1)
const limit = 20
const total = ref(0)
const savingIds = ref<Set<string>>(new Set())
const rowError = ref<string | null>(null)

function toRow(v: Video): StudioContentRow {
  return {
    id: v.id,
    title: v.title,
    type: String(v.videoType).toLowerCase(),
    visibility: v.visibility,
    status: v.status,
    views: v.viewCount,
    comments: v.commentCount,
    publishedAt: v.publishedAt || v.createdAt,
    selected: false
  }
}

const { data, pending, error, refresh } = await useAsyncData(
  () => `studio-content-${activeTab.value}-${page.value}`,
  async () => {
    const res = await api.getStudioVideos(page.value, limit)
    total.value = res.total
    const rows = res.items.map(toRow)
    return activeTab.value === 'all' ? rows : rows.filter(r => r.type === activeTab.value)
  },
  { watch: [activeTab, page] }
)

const items = computed(() => {
  const rows = data.value ?? []
  return statusFilter.value === 'all' ? rows : rows.filter(r => r.status === statusFilter.value)
})
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit)))
const selectedCount = computed(() => items.value.filter(i => i.selected).length)

async function setRowVisibility(row: StudioContentRow, vis: VideoVisibility) {
  const prev = row.visibility
  row.visibility = vis
  rowError.value = null
  savingIds.value.add(row.id)
  try {
    await api.updateVideo(row.id, { visibility: vis } as Partial<Video>)
  } catch (e: unknown) {
    row.visibility = prev
    rowError.value = e instanceof Error ? e.message : '공개 범위 변경에 실패했습니다.'
  } finally {
    savingIds.value.delete(row.id)
  }
}

async function applyBulkVisibility(vis: VideoVisibility) {
  rowError.value = null
  const targets = items.value.filter(i => i.selected)
  for (const row of targets) {
    await setRowVisibility(row, vis)
  }
}

function formatDate(iso: string) {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '-' : d.toLocaleDateString('ko-KR')
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">채널 콘텐츠 관리 (Content Manager)</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          Manage visibility, processing status, and comments across Videos, Shorts, and Live VODs
        </p>
      </div>

      <!-- Bulk Actions Toolbar -->
      <div v-if="selectedCount > 0" class="flex items-center gap-2 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-1.5 text-xs">
        <span class="font-semibold text-[var(--gvp-text)] tabular-nums">{{ selectedCount }} selected</span>
        <button
          type="button"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
          @click="applyBulkVisibility(VideoVisibility.PUBLIC)"
        >
          Set Public
        </button>
        <button
          type="button"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
          @click="applyBulkVisibility(VideoVisibility.PRIVATE)"
        >
          Set Private
        </button>
      </div>
    </div>

    <GvpBanner v-if="rowError" tone="danger" dismissible @dismiss="rowError = null">
      {{ rowError }}
    </GvpBanner>

    <!-- Filter Tabs + Status Filter -->
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--gvp-border)]">
      <div class="flex gap-6 text-xs font-semibold" role="tablist" aria-label="콘텐츠 유형">
        <button
          v-for="tab in (['all', 'video', 'shorts', 'live'] as const)"
          :key="tab"
          type="button"
          role="tab"
          :aria-selected="activeTab === tab"
          class="border-b-2 pb-3 uppercase tracking-wider transition-colors"
          :class="activeTab === tab ? 'border-[var(--gvp-focus)] text-[var(--gvp-text)]' : 'border-transparent text-[var(--gvp-text-secondary)] hover:text-[var(--gvp-text)]'"
          @click="activeTab = tab; page = 1"
        >
          {{ tab === 'all' ? '동영상' : tab === 'video' ? '일반' : tab === 'shorts' ? '숏폼' : '라이브' }}
        </button>
      </div>
      <label class="flex items-center gap-2 pb-2 text-xs font-medium text-[var(--gvp-text-secondary)]">
        처리 상태
        <select
          v-model="statusFilter"
          class="h-8 rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs text-[var(--gvp-text)]"
          aria-label="처리 상태 필터"
        >
          <option value="all">전체</option>
          <option :value="VideoProcessingStatus.READY">READY</option>
          <option :value="VideoProcessingStatus.PROCESSING">PROCESSING</option>
          <option :value="VideoProcessingStatus.QUEUED">QUEUED</option>
          <option :value="VideoProcessingStatus.FAILED">FAILED</option>
        </select>
      </label>
    </div>

    <!-- Loading -->
    <div v-if="pending" class="space-y-2" role="status" aria-label="불러오는 중">
      <div v-for="n in 5" :key="n" class="h-12 animate-pulse rounded-lg bg-[var(--gvp-elevated)]" />
    </div>

    <!-- Error -->
    <GvpBanner v-else-if="error" tone="danger">
      콘텐츠를 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <!-- Empty -->
    <GvpEmptyState
      v-else-if="items.length === 0"
      title="아직 업로드한 영상이 없습니다"
      description="첫 영상을 업로드하면 처리 상태와 댓글·조회수가 여기에 표시됩니다."
      action-label="업로드하러 가기"
      @action="navigateTo('/upload')"
    />

    <!-- Content Table -->
    <div v-else class="overflow-x-auto rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <table class="w-full text-left">
        <thead class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider">
          <tr>
            <th scope="col" class="py-2.5 px-4">Select</th>
            <th scope="col" class="py-2.5 px-4">Video / Content</th>
            <th scope="col" class="py-2.5 px-4">Visibility</th>
            <th scope="col" class="py-2.5 px-4">Status</th>
            <th scope="col" class="py-2.5 px-4">Views</th>
            <th scope="col" class="py-2.5 px-4">Comments</th>
            <th scope="col" class="py-2.5 px-4">Published</th>
            <th scope="col" class="py-2.5 px-4 text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--gvp-border)]">
          <tr v-for="row in items" :key="row.id" class="hover:bg-[var(--gvp-canvas)] transition-colors">
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <input v-model="row.selected" type="checkbox" class="rounded-md border-[var(--gvp-border-strong)] accent-[var(--gvp-brand)]" :aria-label="`${row.title} 선택`" />
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)]">
              <p class="font-medium text-[var(--gvp-text)]">{{ row.title }}</p>
              <span class="mt-1 inline-block rounded-md bg-[var(--gvp-elevated)] px-2 py-0.5 text-[11px] font-medium uppercase text-[var(--gvp-text-secondary)]">
                {{ row.type }}
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <select
                :value="row.visibility"
                :disabled="savingIds.has(row.id)"
                :aria-label="`${row.title} 공개 범위`"
                class="rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 py-1.5 text-xs font-medium text-[var(--gvp-text)] disabled:opacity-50"
                @change="setRowVisibility(row, ($event.target as HTMLSelectElement).value as VideoVisibility)"
              >
                <option :value="VideoVisibility.PUBLIC">Public</option>
                <option :value="VideoVisibility.UNLISTED">Unlisted</option>
                <option :value="VideoVisibility.PRIVATE">Private</option>
                <option :value="VideoVisibility.MEMBERS_ONLY">Members Only</option>
              </select>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <span
                class="inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium"
                :class="{
                  'border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]': resolveStatusBadgeTone(row.status) === 'success',
                  'border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] text-[var(--gvp-warning)]': resolveStatusBadgeTone(row.status) === 'warning',
                  'border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] text-[var(--gvp-live)]': resolveStatusBadgeTone(row.status) === 'danger',
                  'border-[var(--gvp-border)] bg-[var(--gvp-elevated)] text-[var(--gvp-text-secondary)]': resolveStatusBadgeTone(row.status) === 'neutral',
                  'border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] text-[var(--gvp-brand-hover)]': resolveStatusBadgeTone(row.status) === 'info'
                }"
              >
                {{ row.status }}
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums font-mono">{{ row.views.toLocaleString() }}</td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums font-mono">{{ row.comments.toLocaleString() }}</td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums">{{ formatDate(row.publishedAt) }}</td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums text-right">
              <div class="inline-flex items-center gap-1.5">
                <NuxtLink :to="`/editor/${row.id}`" title="세부정보 수정" class="inline-flex h-8 items-center rounded-md border border-[var(--gvp-border-strong)] px-2.5 text-xs font-medium hover:bg-[var(--gvp-canvas)]">편집</NuxtLink>
                <NuxtLink to="/analytics" title="분석 보기" class="inline-flex h-8 items-center rounded-md border border-[var(--gvp-border-strong)] px-2.5 text-xs font-medium hover:bg-[var(--gvp-canvas)]">분석</NuxtLink>
                <NuxtLink to="/comments" title="댓글 보기" class="inline-flex h-8 items-center rounded-md border border-[var(--gvp-border-strong)] px-2.5 text-xs font-medium hover:bg-[var(--gvp-canvas)]">댓글</NuxtLink>
                <a :href="`${config.public.webUrl}/watch/${row.id}`" target="_blank" rel="noopener noreferrer" title="실제 영상 보기" class="inline-flex h-8 items-center rounded-md bg-[var(--gvp-brand)] px-2.5 text-xs font-medium text-white hover:bg-[var(--gvp-brand-hover)]">보기↗</a>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
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
