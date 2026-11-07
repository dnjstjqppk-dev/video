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

const items = computed(() => data.value ?? [])
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
  <div>
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 콘텐츠 관리 (Content Manager)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Manage visibility, processing status, and comments across Videos, Shorts, and Live VODs
        </p>
      </div>

      <!-- Bulk Actions Toolbar -->
      <div v-if="selectedCount > 0" class="flex items-center gap-2 rounded-xl border border-[#3EA6FF]/40 bg-[#182330] px-4 py-2 text-xs">
        <span class="font-bold text-white">{{ selectedCount }} selected</span>
        <button
          type="button"
          class="rounded-lg bg-white/10 px-2.5 py-1 text-white hover:bg-white/20"
          @click="applyBulkVisibility(VideoVisibility.PUBLIC)"
        >
          Set Public
        </button>
        <button
          type="button"
          class="rounded-lg bg-white/10 px-2.5 py-1 text-white hover:bg-white/20"
          @click="applyBulkVisibility(VideoVisibility.PRIVATE)"
        >
          Set Private
        </button>
      </div>
    </div>

    <GvpBanner v-if="rowError" tone="danger" dismissible class="mt-4" @dismiss="rowError = null">
      {{ rowError }}
    </GvpBanner>

    <!-- Filter Tabs -->
    <div class="mt-6 flex gap-4 border-b border-white/10 text-xs font-bold" role="tablist" aria-label="콘텐츠 유형">
      <button
        v-for="tab in (['all', 'video', 'shorts', 'live'] as const)"
        :key="tab"
        type="button"
        role="tab"
        :aria-selected="activeTab === tab"
        class="border-b-2 pb-3 uppercase transition"
        :class="activeTab === tab ? 'border-[#FF0033] text-white' : 'border-transparent text-neutral-400 hover:text-white'"
        @click="activeTab = tab; page = 1"
      >
        {{ tab }}
      </button>
    </div>

    <!-- Loading -->
    <div v-if="pending" class="mt-4 space-y-2" role="status" aria-label="불러오는 중">
      <div v-for="n in 5" :key="n" class="h-14 animate-pulse rounded-xl bg-white/5" />
    </div>

    <!-- Error -->
    <GvpBanner v-else-if="error" tone="danger" class="mt-4">
      콘텐츠를 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <!-- Empty -->
    <GvpEmptyState
      v-else-if="items.length === 0"
      class="mt-4"
      title="아직 업로드한 영상이 없습니다"
      description="첫 영상을 업로드하면 처리 상태와 댓글·조회수가 여기에 표시됩니다."
      action-label="업로드하러 가기"
      @action="navigateTo('/upload')"
    />

    <!-- Content Table -->
    <div v-else class="mt-4 overflow-x-auto rounded-2xl border border-white/10 bg-[#151822]">
      <table class="w-full text-left text-xs">
        <thead class="border-b border-white/10 bg-white/5 uppercase text-neutral-400">
          <tr>
            <th scope="col" class="p-3.5">Select</th>
            <th scope="col" class="p-3.5">Video / Content</th>
            <th scope="col" class="p-3.5">Visibility</th>
            <th scope="col" class="p-3.5">Status</th>
            <th scope="col" class="p-3.5">Views</th>
            <th scope="col" class="p-3.5">Comments</th>
            <th scope="col" class="p-3.5">Published</th>
            <th scope="col" class="p-3.5">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/10">
          <tr v-for="row in items" :key="row.id" class="hover:bg-white/5">
            <td class="p-3.5">
              <input v-model="row.selected" type="checkbox" class="rounded accent-[#FF0033]" :aria-label="`${row.title} 선택`" />
            </td>
            <td class="p-3.5">
              <p class="font-bold text-white">{{ row.title }}</p>
              <span class="mt-0.5 inline-block rounded bg-white/10 px-1.5 py-0.5 text-[10px] uppercase text-neutral-300">
                {{ row.type }}
              </span>
            </td>
            <td class="p-3.5">
              <select
                :value="row.visibility"
                :disabled="savingIds.has(row.id)"
                :aria-label="`${row.title} 공개 범위`"
                class="rounded-lg border border-white/15 bg-[#0D0F14] px-2.5 py-1 text-xs text-white disabled:opacity-50"
                @change="setRowVisibility(row, ($event.target as HTMLSelectElement).value as VideoVisibility)"
              >
                <option :value="VideoVisibility.PUBLIC">Public</option>
                <option :value="VideoVisibility.UNLISTED">Unlisted</option>
                <option :value="VideoVisibility.PRIVATE">Private</option>
                <option :value="VideoVisibility.MEMBERS_ONLY">Members Only</option>
              </select>
            </td>
            <td class="p-3.5">
              <span
                class="rounded-full px-2.5 py-0.5 text-[11px] font-bold"
                :class="{
                  'bg-emerald-500/20 text-emerald-400': resolveStatusBadgeTone(row.status) === 'success',
                  'bg-amber-500/20 text-amber-300': resolveStatusBadgeTone(row.status) === 'warning',
                  'bg-red-500/20 text-red-400': resolveStatusBadgeTone(row.status) === 'danger',
                  'bg-white/10 text-neutral-300': resolveStatusBadgeTone(row.status) === 'neutral',
                  'bg-[#3EA6FF]/20 text-[#3EA6FF]': resolveStatusBadgeTone(row.status) === 'info'
                }"
              >
                {{ row.status }}
              </span>
            </td>
            <td class="p-3.5 font-mono text-white">{{ row.views.toLocaleString() }}</td>
            <td class="p-3.5 font-mono text-white">{{ row.comments.toLocaleString() }}</td>
            <td class="p-3.5 text-neutral-300">{{ formatDate(row.publishedAt) }}</td>
            <td class="p-3.5">
              <NuxtLink
                :to="`/editor/${row.id}`"
                class="rounded-lg bg-[#3EA6FF]/20 px-3 py-1.5 font-bold text-[#3EA6FF] hover:bg-[#3EA6FF]/30"
              >
                Timeline Editor
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="!pending && !error && totalPages > 1" class="mt-4 flex items-center justify-center gap-2 text-xs">
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
