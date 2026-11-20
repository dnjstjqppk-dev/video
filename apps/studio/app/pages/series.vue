<script setup lang="ts">
import type { Series, SeriesDetail, Video } from '@gvp/shared-types'
import { useStudioApi } from '~/composables/useStudioApi'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'

useSeoMeta({ title: 'Series Manager — Seasons, Episodes & AI Originals' })

const api = useStudioApi()
const loadError = ref<string | null>(null)
const actionError = ref<string | null>(null)
const notice = ref<string | null>(null)
const busy = ref(false)

const channelId = ref<string | null>(null)
const seriesList = ref<Series[]>([])
const videos = ref<Video[]>([])
const selectedId = ref<string | null>(null)
const detail = ref<SeriesDetail | null>(null)

const formTitle = ref('')
const formDescription = ref('')
const formStatus = ref<'ongoing' | 'completed'>('ongoing')

const attachVideoId = ref('')
const attachSeason = ref(1)
const attachEpisode = ref(1)

const attachable = computed(() => videos.value.filter(v => v.status === 'READY'))
const nextSlot = computed(() => {
  const eps = detail.value?.episodes || []
  const last = eps[eps.length - 1]
  if (!last) return { season: 1, episode: 1 }
  return { season: last.seasonNo, episode: last.episodeNo + 1 }
})

async function loadAll() {
  loadError.value = null
  try {
    const dashboard = await api.getStudioDashboard()
    channelId.value = dashboard.channel.id
    const [seriesRes, videosRes] = await Promise.all([
      api.listSeries(dashboard.channel.id, 1, 50),
      api.getStudioVideos(1, 100)
    ])
    seriesList.value = seriesRes.items
    videos.value = videosRes.items
    if (selectedId.value) {
      detail.value = await api.getSeries(selectedId.value)
    } else if (seriesRes.items.length) {
      const first = seriesRes.items[0]
      if (first) await selectSeries(first.id)
    }
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : '시리즈를 불러오지 못했습니다.'
  }
}

await useAsyncData('studio-series', loadAll)

async function selectSeries(id: string) {
  selectedId.value = id
  actionError.value = null
  try {
    detail.value = await api.getSeries(id)
    attachSeason.value = nextSlot.value.season
    attachEpisode.value = nextSlot.value.episode
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '시리즈 상세를 불러오지 못했습니다.'
  }
}

async function createSeries() {
  const title = formTitle.value.trim()
  if (!title || !channelId.value || busy.value) return
  busy.value = true
  actionError.value = null
  notice.value = null
  try {
    const created = await api.createSeries({
      channelId: channelId.value,
      title,
      description: formDescription.value.trim() || undefined,
      status: formStatus.value
    })
    seriesList.value.unshift(created.series)
    formTitle.value = ''
    formDescription.value = ''
    await selectSeries(created.series.id)
    notice.value = '시리즈를 만들었습니다. 회차를 붙여보세요.'
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '시리즈 생성에 실패했습니다.'
  } finally {
    busy.value = false
  }
}

async function attach() {
  if (!selectedId.value || !attachVideoId.value || busy.value) return
  busy.value = true
  actionError.value = null
  notice.value = null
  try {
    detail.value = await api.attachEpisode(selectedId.value, {
      videoId: attachVideoId.value,
      seasonNo: attachSeason.value,
      episodeNo: attachEpisode.value
    })
    attachVideoId.value = ''
    attachSeason.value = nextSlot.value.season
    attachEpisode.value = nextSlot.value.episode
    notice.value = '회차를 붙였습니다.'
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '회차 등록에 실패했습니다.'
  } finally {
    busy.value = false
  }
}

async function detach(videoId: string) {
  if (!selectedId.value || busy.value) return
  busy.value = true
  actionError.value = null
  try {
    detail.value = await api.detachEpisode(selectedId.value, videoId)
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '회차 해제에 실패했습니다.'
  } finally {
    busy.value = false
  }
}

async function toggleStatus() {
  if (!detail.value || busy.value) return
  busy.value = true
  actionError.value = null
  try {
    const next = detail.value.series.status === 'completed' ? 'ongoing' : 'completed'
    detail.value = await api.updateSeries(detail.value.series.id, { status: next })
    const row = seriesList.value.find(s => s.id === detail.value!.series.id)
    if (row) row.status = next
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '상태 변경에 실패했습니다.'
  } finally {
    busy.value = false
  }
}

const confirmDelete = ref(false)
async function removeSeries() {
  if (!selectedId.value || busy.value) return
  if (!confirmDelete.value) {
    confirmDelete.value = true
    return
  }
  busy.value = true
  actionError.value = null
  try {
    await api.deleteSeries(selectedId.value)
    seriesList.value = seriesList.value.filter(s => s.id !== selectedId.value)
    selectedId.value = null
    detail.value = null
    confirmDelete.value = false
    notice.value = '시리즈를 삭제했습니다. 영상 자체는 유지됩니다.'
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '시리즈 삭제에 실패했습니다.'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">시리즈</h1>
      <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">시즌·회차로 묶으면 시청 페이지에서 정주행 패널과 다음화 자동재생이 붙습니다</p>
    </div>

    <GvpBanner v-if="loadError" tone="danger">{{ loadError }}</GvpBanner>
    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">{{ actionError }}</GvpBanner>
    <GvpBanner v-if="notice" tone="success" dismissible @dismiss="notice = null">{{ notice }}</GvpBanner>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-[320px_minmax(0,1fr)]">
      <section aria-label="시리즈 목록" class="space-y-4">
        <form class="space-y-2.5 rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4" @submit.prevent="createSeries">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">새 시리즈</h2>
          <input v-model="formTitle" type="text" required maxlength="255" placeholder="시리즈 제목 (예: 서울 2099)" aria-label="시리즈 제목" class="h-9 w-full rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-2.5 text-xs outline-none focus:border-[var(--gvp-focus)]" />
          <textarea v-model="formDescription" rows="2" maxlength="4000" placeholder="소개 (선택)" aria-label="시리즈 소개" class="w-full rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-2.5 py-2 text-xs outline-none focus:border-[var(--gvp-focus)]" />
          <div class="flex items-center gap-2">
            <select v-model="formStatus" aria-label="연재 상태" class="h-9 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-2 text-xs">
              <option value="ongoing">연재 중</option>
              <option value="completed">완결</option>
            </select>
            <button type="submit" :disabled="busy || !formTitle.trim()" class="inline-flex h-9 flex-1 items-center justify-center rounded-full bg-[var(--gvp-studio)] px-3 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50">만들기</button>
          </div>
        </form>

        <ul class="space-y-1.5">
          <li v-for="s in seriesList" :key="s.id">
            <button
              type="button"
              class="w-full rounded-xl border p-3 text-left transition"
              :class="selectedId === s.id ? 'border-[var(--gvp-studio)] bg-[var(--gvp-canvas)]' : 'border-[var(--gvp-border)] bg-[var(--gvp-surface)] hover:bg-[var(--gvp-canvas)]'"
              @click="selectSeries(s.id)"
            >
              <span class="flex items-center gap-2">
                <span class="truncate text-sm font-semibold text-[var(--gvp-text)]">{{ s.title }}</span>
                <span v-if="s.aiOriginal" class="shrink-0 rounded bg-[var(--gvp-brand)] px-1.5 py-0.5 text-[10px] font-bold uppercase text-white">AI</span>
              </span>
              <span class="mt-1 block text-[11px] text-[var(--gvp-text-secondary)]">{{ s.status === 'completed' ? '완결' : '연재 중' }}</span>
            </button>
          </li>
        </ul>
        <GvpEmptyState v-if="!seriesList.length && !loadError" title="아직 시리즈가 없습니다" description="위 폼에서 첫 시리즈를 만들고 회차를 붙여보세요." />
      </section>

      <section aria-label="회차 관리" class="min-w-0">
        <GvpEmptyState v-if="!detail" title="시리즈를 선택하세요" description="왼쪽 목록에서 관리할 시리즈를 고르면 회차 목록이 표시됩니다." />
        <div v-else class="space-y-4 rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4 sm:p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0">
              <h2 class="truncate text-base font-bold text-[var(--gvp-text)]">{{ detail.series.title }}</h2>
              <p class="mt-1 text-xs text-[var(--gvp-text-secondary)] tabular-nums">{{ detail.episodeCount }}화 · {{ detail.series.status === 'completed' ? '완결' : '연재 중' }}{{ detail.series.aiOriginal ? ' · AI Originals' : '' }}</p>
            </div>
            <div class="flex gap-2">
              <button type="button" :disabled="busy" class="inline-flex h-8 items-center rounded-full border border-[var(--gvp-border-strong)] px-3 text-xs font-medium hover:bg-[var(--gvp-canvas)] disabled:opacity-50" @click="toggleStatus">
                {{ detail.series.status === 'completed' ? '연재 중으로 전환' : '완결로 전환' }}
              </button>
              <button type="button" :disabled="busy" class="inline-flex h-8 items-center rounded-full border border-[var(--gvp-live)] px-3 text-xs font-medium text-[var(--gvp-live)] hover:bg-[var(--gvp-live-soft)] disabled:opacity-50" @click="removeSeries">
                {{ confirmDelete ? '정말 삭제할까요?' : '삭제' }}
              </button>
            </div>
          </div>

          <form class="flex flex-wrap items-end gap-2 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-3" @submit.prevent="attach">
            <label class="min-w-44 flex-1 text-[11px] font-medium text-[var(--gvp-text-secondary)]">영상
              <select v-model="attachVideoId" required aria-label="붙일 영상" class="mt-1 h-9 w-full rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs text-[var(--gvp-text)]">
                <option value="" disabled>READY 영상 선택</option>
                <option v-for="v in attachable" :key="v.id" :value="v.id">{{ v.title }}</option>
              </select>
            </label>
            <label class="text-[11px] font-medium text-[var(--gvp-text-secondary)]">시즌
              <input v-model.number="attachSeason" type="number" min="1" max="99" required aria-label="시즌 번호" class="mt-1 h-9 w-20 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs tabular-nums" />
            </label>
            <label class="text-[11px] font-medium text-[var(--gvp-text-secondary)]">회차
              <input v-model.number="attachEpisode" type="number" min="1" max="999" required aria-label="회차 번호" class="mt-1 h-9 w-20 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs tabular-nums" />
            </label>
            <button type="submit" :disabled="busy || !attachVideoId" class="inline-flex h-9 items-center rounded-full bg-[var(--gvp-studio)] px-4 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50">회차 붙이기</button>
          </form>

          <ol v-if="detail.episodes.length" class="divide-y divide-[var(--gvp-border)] overflow-hidden rounded-lg border border-[var(--gvp-border)]">
            <li v-for="ep in detail.episodes" :key="ep.videoId" class="flex items-center gap-3 bg-[var(--gvp-canvas)] p-2.5">
              <span class="w-14 shrink-0 text-center text-xs font-bold tabular-nums text-[var(--gvp-text-secondary)]">S{{ ep.seasonNo }}E{{ ep.episodeNo }}</span>
              <img :src="ep.thumbnailUrl" :alt="ep.title" class="h-11 w-[78px] shrink-0 rounded-md object-cover" loading="lazy" />
              <span class="min-w-0 flex-1">
                <span class="block truncate text-xs font-medium text-[var(--gvp-text)]">{{ ep.title }}</span>
                <span class="mt-0.5 block text-[11px] text-[var(--gvp-text-secondary)]">{{ ep.visibility }} · {{ ep.status }}{{ ep.aiGenerated ? ' · AI' : '' }}</span>
              </span>
              <button type="button" :disabled="busy" class="inline-flex h-7 shrink-0 items-center rounded-full border border-[var(--gvp-border-strong)] px-2.5 text-[11px] font-medium hover:bg-[var(--gvp-surface)] disabled:opacity-50" @click="detach(ep.videoId)">해제</button>
            </li>
          </ol>
          <p v-else class="rounded-lg border border-dashed border-[var(--gvp-border-strong)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">아직 붙은 회차가 없습니다. 위 폼에서 영상을 고르고 시즌·회차 번호를 정하세요.</p>
        </div>
      </section>
    </div>
  </div>
</template>
