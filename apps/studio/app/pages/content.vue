<script setup lang="ts">
import { VideoVisibility, VideoProcessingStatus, type Video } from '@gvp/shared-types'
import { useStudioApi } from '~/composables/useStudioApi'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import { formatDurationBadge, resolveStatusBadgeTone } from '@gvp/ui'

useSeoMeta({
  title: 'Content Manager — Videos, Shorts, Live & Playlists'
})

const api = useStudioApi()
const route = useRoute()

interface StudioContentRow {
  id: string
  title: string
  type: string
  visibility: VideoVisibility
  status: VideoProcessingStatus
  views: number
  comments: number
  likes: number
  dislikes: number
  thumbnailUrl: string
  durationSeconds: number
  ageRestricted: boolean
  publishedAt: string
  aiGenerated: boolean
  aiTool: string
  aiReviewStatus: '' | 'pending' | 'approved' | 'rejected'
  submittingAi: boolean
  confirmAiOff: boolean
  savingAi: boolean
  selected: boolean
}

const aiReviewLabel: Record<StudioContentRow['aiReviewStatus'], string> = {
  '': '미제출',
  pending: '심사 중',
  approved: '선정됨',
  rejected: '반려됨'
}

const activeTab = ref<'all' | 'video' | 'shorts' | 'live'>('all')
const statusFilter = ref<'all' | VideoProcessingStatus>('all')
const visibilityFilter = ref<'all' | VideoVisibility>('all')
const viewsFilter = ref(0)
const searchQuery = ref(typeof route.query.q === 'string' ? route.query.q : '')
const config = useRuntimeConfig()
const page = ref(1)
const limit = 20
const savingIds = ref<Set<string>>(new Set())
const rowError = ref<string | null>(null)

watch(() => route.query.q, (q) => {
  searchQuery.value = typeof q === 'string' ? q : ''
})

function toRow(v: Video): StudioContentRow {
  return {
    id: v.id,
    title: v.title,
    type: String(v.videoType).toLowerCase(),
    visibility: v.visibility,
    status: v.status,
    views: v.viewCount,
    comments: v.commentCount,
    likes: v.likeCount,
    dislikes: v.dislikeCount,
    thumbnailUrl: v.thumbnailUrl,
    durationSeconds: v.durationSeconds,
    ageRestricted: v.isAgeRestricted,
    publishedAt: v.publishedAt || v.createdAt,
    aiGenerated: v.aiGenerated === true,
    aiTool: v.aiTool || '',
    aiReviewStatus: (v.aiReviewStatus || '') as StudioContentRow['aiReviewStatus'],
    submittingAi: false,
    confirmAiOff: false,
    savingAi: false,
    selected: false
  }
}

async function toggleAiFlag(row: StudioContentRow) {
  if (row.savingAi) return
  // Unflagging drops the video (and its review) from Originals: confirm first.
  if (row.aiGenerated && row.aiReviewStatus !== '' && !row.confirmAiOff) {
    row.confirmAiOff = true
    return
  }
  row.savingAi = true
  rowError.value = null
  try {
    const updated = await api.updateVideo(row.id, { aiGenerated: !row.aiGenerated } as Partial<Video>)
    row.aiGenerated = updated.aiGenerated === true
    row.aiReviewStatus = (updated.aiReviewStatus || '') as StudioContentRow['aiReviewStatus']
    row.confirmAiOff = false
  } catch (e: unknown) {
    rowError.value = e instanceof Error ? e.message : 'AI 표시 변경에 실패했습니다.'
  } finally {
    row.savingAi = false
  }
}

async function submitAiReview(row: StudioContentRow) {
  if (row.submittingAi) return
  row.submittingAi = true
  rowError.value = null
  try {
    const updated = await api.submitAIReview(row.id)
    row.aiReviewStatus = (updated.aiReviewStatus || 'pending') as StudioContentRow['aiReviewStatus']
  } catch (e: unknown) {
    rowError.value = e instanceof Error ? e.message : 'Originals 심사 제출에 실패했습니다.'
  } finally {
    row.submittingAi = false
  }
}

const { data, pending, error, refresh } = await useAsyncData(
  () => `studio-content-${activeTab.value}-${page.value}`,
  async () => {
    const res = await api.getStudioVideos(page.value, limit)
    const rows = res.items.map(toRow)
    return {
      total: res.total,
      rows: activeTab.value === 'all' ? rows : rows.filter(r => r.type === activeTab.value)
    }
  },
  { watch: [activeTab, page] }
)

const total = computed(() => data.value?.total ?? 0)

const items = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return (data.value?.rows ?? []).filter((r) => {
    if (statusFilter.value !== 'all' && r.status !== statusFilter.value) return false
    if (visibilityFilter.value !== 'all' && r.visibility !== visibilityFilter.value) return false
    if (r.views < viewsFilter.value) return false
    if (q && !r.title.toLowerCase().includes(q)) return false
    return true
  })
})
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit)))
const selectedCount = computed(() => items.value.filter(i => i.selected).length)

async function setRowVisibility(row: StudioContentRow, vis: VideoVisibility) {
  const prev = row.visibility
  if (vis === VideoVisibility.SCHEDULED && row.visibility !== VideoVisibility.SCHEDULED) {
    rowError.value = '예약 공개는 업로드 위자드에서 시각과 함께 설정하세요.'
    return
  }
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

const bulkBusy = ref(false)
const bulkPlaylistTitle = ref('')

async function applyBulkDelete() {
  rowError.value = null
  bulkBusy.value = true
  try {
    const targets = items.value.filter(i => i.selected)
    for (const row of targets) {
      await api.deleteVideo(row.id)
    }
    await refresh()
  } catch (e: unknown) {
    rowError.value = e instanceof Error ? e.message : '일괄 삭제에 실패했습니다.'
  } finally {
    bulkBusy.value = false
  }
}

async function applyBulkPlaylist() {
  const title = bulkPlaylistTitle.value.trim()
  if (!title || bulkBusy.value) return
  rowError.value = null
  bulkBusy.value = true
  try {
    const pl = await api.createPlaylist({ title, visibility: 'private' })
    const targets = items.value.filter(i => i.selected)
    for (const row of targets) {
      await api.addPlaylistItem(pl.id, row.id)
    }
    bulkPlaylistTitle.value = ''
  } catch (e: unknown) {
    rowError.value = e instanceof Error ? e.message : '재생목록 추가에 실패했습니다.'
  } finally {
    bulkBusy.value = false
  }
}

function likePct(row: StudioContentRow): string {
  const total = row.likes + row.dislikes
  if (total <= 0) return '-'
  return `${Math.round((row.likes / total) * 100)}%`
}

function restrictionOf(row: StudioContentRow): string | null {
  if (row.status === VideoProcessingStatus.FAILED) return '처리 실패'
  if (row.status === VideoProcessingStatus.PROCESSING) return '처리 중'
  if (row.status === VideoProcessingStatus.QUEUED) return '대기 중'
  if (row.ageRestricted) return '연령 제한'
  return null
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
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">채널 콘텐츠</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          동영상 · Shorts · 라이브 VOD의 공개 상태와 실적을 관리하세요
        </p>
      </div>
      <div v-if="selectedCount > 0" class="flex flex-wrap items-center gap-2 rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-1.5 text-xs">
        <span class="font-semibold text-[var(--gvp-text)] tabular-nums">{{ selectedCount }}개 선택됨</span>
        <button type="button" class="inline-flex h-8 items-center rounded-full border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-3 text-xs font-medium hover:bg-[var(--gvp-canvas)]" @click="applyBulkVisibility(VideoVisibility.PUBLIC)">공개로 설정</button>
        <button type="button" class="inline-flex h-8 items-center rounded-full border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-3 text-xs font-medium hover:bg-[var(--gvp-canvas)]" @click="applyBulkVisibility(VideoVisibility.PRIVATE)">비공개로 설정</button>
        <button type="button" :disabled="bulkBusy" class="inline-flex h-8 items-center rounded-full border border-[var(--gvp-live)] bg-[var(--gvp-surface)] px-3 text-xs font-medium text-[var(--gvp-live)] hover:bg-[var(--gvp-live-soft)] disabled:opacity-50" @click="applyBulkDelete">삭제</button>
        <span class="flex items-center gap-1.5">
          <input v-model="bulkPlaylistTitle" type="text" placeholder="새 재생목록 이름" aria-label="새 재생목록 이름" class="h-8 w-36 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 text-xs outline-none placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)]" @keydown.enter="applyBulkPlaylist" />
          <button type="button" :disabled="bulkBusy || !bulkPlaylistTitle.trim()" class="inline-flex h-8 items-center rounded-full bg-[var(--gvp-studio)] px-3 text-xs font-medium text-white hover:opacity-90 disabled:opacity-50" @click="applyBulkPlaylist">재생목록에 추가</button>
        </span>
      </div>
    </div>

    <GvpBanner v-if="rowError" tone="danger" dismissible @dismiss="rowError = null">
      {{ rowError }}
    </GvpBanner>

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
      <div class="flex flex-wrap items-center gap-2 pb-2">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="제목으로 필터"
          aria-label="제목으로 필터"
          class="h-8 w-44 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 text-xs text-[var(--gvp-text)] outline-none placeholder:text-[var(--gvp-text-muted)] focus:border-[var(--gvp-focus)]"
        />
        <label class="flex items-center gap-2 text-xs font-medium text-[var(--gvp-text-secondary)]">
          처리 상태
          <select v-model="statusFilter" class="h-8 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs text-[var(--gvp-text)]" aria-label="처리 상태 필터">
            <option value="all">전체</option>
            <option :value="VideoProcessingStatus.READY">READY</option>
            <option :value="VideoProcessingStatus.PROCESSING">PROCESSING</option>
            <option :value="VideoProcessingStatus.QUEUED">QUEUED</option>
            <option :value="VideoProcessingStatus.FAILED">FAILED</option>
          </select>
        </label>
        <label class="flex items-center gap-2 text-xs font-medium text-[var(--gvp-text-secondary)]">
          공개 상태
          <select v-model="visibilityFilter" class="h-8 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs text-[var(--gvp-text)]" aria-label="공개 상태 필터">
            <option value="all">전체</option>
            <option :value="VideoVisibility.PUBLIC">공개</option>
            <option :value="VideoVisibility.UNLISTED">일부 공개</option>
            <option :value="VideoVisibility.PRIVATE">비공개</option>
            <option :value="VideoVisibility.MEMBERS_ONLY">멤버십 전용</option>
            <option :value="VideoVisibility.SCHEDULED">예약</option>
          </select>
        </label>
        <label class="flex items-center gap-2 text-xs font-medium text-[var(--gvp-text-secondary)]">
          조회수
          <select v-model.number="viewsFilter" class="h-8 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs text-[var(--gvp-text)]" aria-label="조회수 조건">
            <option :value="0">전체</option>
            <option :value="1000">1천 이상</option>
            <option :value="10000">1만 이상</option>
            <option :value="100000">10만 이상</option>
          </select>
        </label>
      </div>
    </div>

    <div v-if="pending" class="space-y-2" role="status" aria-label="불러오는 중">
      <div v-for="n in 5" :key="n" class="h-12 animate-pulse rounded-lg bg-[var(--gvp-elevated)]" />
    </div>

    <GvpBanner v-else-if="error" tone="danger">
      콘텐츠를 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <GvpEmptyState
      v-else-if="items.length === 0"
      title="아직 업로드한 영상이 없습니다"
      description="첫 영상을 업로드하면 처리 상태와 댓글·조회수가 여기에 표시됩니다."
      action-label="업로드하러 가기"
      @action="navigateTo('/upload')"
    />

    <div v-else class="overflow-x-auto rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <table class="w-full text-left">
        <thead class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-secondary)]">
          <tr>
            <th scope="col" class="px-4 py-2.5"><span class="sr-only">선택</span></th>
            <th scope="col" class="px-4 py-2.5">동영상</th>
            <th scope="col" class="px-4 py-2.5">공개 상태</th>
            <th scope="col" class="px-4 py-2.5">제한사항</th>
            <th scope="col" class="px-4 py-2.5">AI Originals</th>
            <th scope="col" class="px-4 py-2.5">날짜</th>
            <th scope="col" class="px-4 py-2.5 text-right">조회수</th>
            <th scope="col" class="px-4 py-2.5 text-right">댓글</th>
            <th scope="col" class="px-4 py-2.5 text-right">좋아요 비율</th>
            <th scope="col" class="px-4 py-2.5 text-right">관리</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--gvp-border)]">
          <tr v-for="row in items" :key="row.id" class="transition-colors hover:bg-[var(--gvp-canvas)]">
            <td class="px-4 py-3">
              <input v-model="row.selected" type="checkbox" class="rounded-md border-[var(--gvp-border-strong)] accent-[var(--gvp-brand)]" :aria-label="`${row.title} 선택`" />
            </td>
            <td class="px-4 py-3">
              <div class="flex min-w-0 items-center gap-3">
                <span class="relative shrink-0">
                  <img :src="row.thumbnailUrl" :alt="row.title" class="h-12 w-[86px] rounded-md object-cover" loading="lazy" />
                  <span class="absolute bottom-1 right-1 rounded bg-black/80 px-1 text-[10px] font-medium text-white tabular-nums">{{ formatDurationBadge(row.durationSeconds) }}</span>
                </span>
                <span class="min-w-0">
                  <span class="block max-w-64 truncate text-sm font-medium text-[var(--gvp-text)]">{{ row.title }}</span>
                  <span class="mt-1 inline-block rounded-md bg-[var(--gvp-elevated)] px-2 py-0.5 text-[11px] font-medium uppercase text-[var(--gvp-text-secondary)]">{{ row.type }}</span>
                </span>
              </div>
            </td>
            <td class="px-4 py-3">
              <select
                :value="row.visibility"
                :disabled="savingIds.has(row.id)"
                :aria-label="`${row.title} 공개 범위`"
                class="rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 py-1.5 text-xs font-medium text-[var(--gvp-text)] disabled:opacity-50"
                @change="setRowVisibility(row, ($event.target as HTMLSelectElement).value as VideoVisibility)"
              >
                <option :value="VideoVisibility.PUBLIC">공개</option>
                <option :value="VideoVisibility.UNLISTED">일부 공개</option>
                <option :value="VideoVisibility.PRIVATE">비공개</option>
                <option :value="VideoVisibility.MEMBERS_ONLY">멤버십 전용</option>
                <option :value="VideoVisibility.SCHEDULED">예약</option>
              </select>
            </td>
            <td class="px-4 py-3 text-xs">
              <span v-if="restrictionOf(row)" class="inline-flex items-center rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-2 py-0.5 font-medium text-[var(--gvp-warning)]">{{ restrictionOf(row) }}</span>
              <span v-else class="text-[var(--gvp-text-muted)]">없음</span>
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-xs">
              <span v-if="!row.aiGenerated" class="inline-flex items-center gap-1.5">
                <span class="text-[var(--gvp-text-muted)]">일반</span>
                <button
                  type="button"
                  :disabled="row.savingAi"
                  class="inline-flex h-7 items-center rounded-full border border-[var(--gvp-border-strong)] px-2.5 font-medium hover:bg-[var(--gvp-canvas)] disabled:opacity-50"
                  @click="toggleAiFlag(row)"
                >{{ row.savingAi ? '저장 중…' : 'AI로 표시' }}</button>
              </span>
              <span v-else class="inline-flex items-center gap-1.5">
                <span
                  class="inline-flex items-center rounded-md border px-2 py-0.5 font-medium"
                  :class="row.aiReviewStatus === 'approved'
                    ? 'border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]'
                    : row.aiReviewStatus === 'pending'
                      ? 'border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] text-[var(--gvp-warning)]'
                      : row.aiReviewStatus === 'rejected'
                        ? 'border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] text-[var(--gvp-live)]'
                        : 'border-[var(--gvp-border-strong)] bg-[var(--gvp-elevated)] text-[var(--gvp-text-secondary)]'"
                  :title="row.aiTool || undefined"
                >AI · {{ aiReviewLabel[row.aiReviewStatus] }}</span>
                <button
                  v-if="(row.aiReviewStatus === '' || row.aiReviewStatus === 'rejected') && row.status === VideoProcessingStatus.READY"
                  type="button"
                  :disabled="row.submittingAi"
                  class="inline-flex h-7 items-center rounded-full bg-[var(--gvp-studio)] px-2.5 font-medium text-white hover:opacity-90 disabled:opacity-50"
                  @click="submitAiReview(row)"
                >{{ row.submittingAi ? '제출 중…' : '심사 제출' }}</button>
                <button
                  type="button"
                  :disabled="row.savingAi"
                  class="inline-flex h-7 items-center rounded-full border px-2.5 font-medium disabled:opacity-50"
                  :class="row.confirmAiOff ? 'border-[var(--gvp-live)] text-[var(--gvp-live)]' : 'border-[var(--gvp-border-strong)] hover:bg-[var(--gvp-canvas)]'"
                  @click="toggleAiFlag(row)"
                >{{ row.savingAi ? '저장 중…' : row.confirmAiOff ? '정말 해제?' : 'AI 해제' }}</button>
              </span>
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-xs text-[var(--gvp-text-secondary)] tabular-nums">{{ formatDate(row.publishedAt) }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-sm text-[var(--gvp-text)] tabular-nums">{{ row.views.toLocaleString() }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-sm text-[var(--gvp-text)] tabular-nums">{{ row.comments.toLocaleString() }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-sm text-[var(--gvp-text)] tabular-nums" :title="`좋아요 ${row.likes.toLocaleString()} / 싫어요 ${row.dislikes.toLocaleString()}`">{{ likePct(row) }}</td>
            <td class="px-4 py-3 text-right">
              <div class="inline-flex items-center gap-1.5">
                <NuxtLink :to="`/editor/${row.id}`" title="세부정보 수정" class="inline-flex h-8 items-center rounded-full border border-[var(--gvp-border-strong)] px-2.5 text-xs font-medium hover:bg-[var(--gvp-canvas)]">편집</NuxtLink>
                <NuxtLink to="/analytics" title="분석 보기" class="inline-flex h-8 items-center rounded-full border border-[var(--gvp-border-strong)] px-2.5 text-xs font-medium hover:bg-[var(--gvp-canvas)]">분석</NuxtLink>
                <NuxtLink to="/comments" title="댓글 보기" class="inline-flex h-8 items-center rounded-full border border-[var(--gvp-border-strong)] px-2.5 text-xs font-medium hover:bg-[var(--gvp-canvas)]">댓글</NuxtLink>
                <a :href="`${config.public.webUrl}/watch/${row.id}`" target="_blank" rel="noopener noreferrer" title="vravra에서 보기" class="inline-flex h-8 items-center rounded-full bg-[var(--gvp-studio)] px-2.5 text-xs font-medium text-white hover:opacity-90">보기↗</a>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="!pending && !error && totalPages > 1" class="flex items-center justify-center gap-3 text-xs">
      <button type="button" :disabled="page <= 1" class="inline-flex h-9 items-center rounded-full border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-sm font-medium hover:bg-[var(--gvp-canvas)] disabled:opacity-40" @click="page -= 1">이전</button>
      <span class="text-[var(--gvp-text-secondary)] tabular-nums" role="status">{{ page }} / {{ totalPages }} (총 {{ total }}개)</span>
      <button type="button" :disabled="page >= totalPages" class="inline-flex h-9 items-center rounded-full border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-sm font-medium hover:bg-[var(--gvp-canvas)] disabled:opacity-40" @click="page += 1">다음</button>
    </div>
  </div>
</template>
