<script setup lang="ts">
import { useLibraryStore } from '~/stores/library'
import type { Video } from '@gvp/shared-types'
const library = useLibraryStore()
const route = useRoute()
const api = useApi()
const section = computed(() => String(route.params.section))
const titles: Record<string, string> = {
  history: '시청 기록', saved: '나중에 볼 동영상', liked: '좋아요 표시한 동영상', subscriptions: '구독',
  trending: '인기 급상승', music: '음악', gaming: '게임', tech: '테크/개발',
}
if (!titles[section.value]) throw createError({ statusCode: 404, statusMessage: 'Page not found' })
const apiSections = new Set(['subscriptions', 'trending', 'music', 'gaming', 'tech'])
const isApiSection = computed(() => apiSections.has(section.value))
const query = ref('')
const confirmClear = ref(false)
const feed = ref<Video[]>([])
const loading = ref(false)
const failed = ref(false)
async function fetchFeed() {
  loading.value = true
  failed.value = false
  try { feed.value = (await api.getHomeFeed(1, 100, 'all')).items || [] }
  catch { failed.value = true }
  finally { loading.value = false }
}
onMounted(() => { library.hydrate(); if (isApiSection.value) fetchFeed() })
watch(section, () => { query.value = ''; confirmClear.value = false; if (isApiSection.value) fetchFeed() })
function matchesCategory(v: Video, cat: string) {
  const hay = `${v.category || ''} ${(v.tags || []).join(' ')}`.toLowerCase()
  if (cat === 'tech') return hay.includes('tech') || hay.includes('개발') || hay.includes('technology')
  return hay.includes(cat)
}
const videos = computed(() => {
  let list: Video[]
  if (section.value === 'subscriptions') list = feed.value.filter(v => library.subscriptions.includes(v.channelId))
  else if (section.value === 'trending') list = [...feed.value].sort((a, b) => b.viewCount - a.viewCount)
  else if (section.value === 'music' || section.value === 'gaming' || section.value === 'tech') list = feed.value.filter(v => matchesCategory(v, section.value))
  else if (section.value === 'history') list = library.history
  else if (section.value === 'saved') list = library.saved
  else list = library.liked
  return list.filter(v => `${v.title} ${v.channel?.name || ''}`.toLowerCase().includes(query.value.toLowerCase()))
})
function remove(id: string) {
  if (section.value === 'history' || section.value === 'saved' || section.value === 'liked') library.remove(section.value, id)
}
useSeoMeta({ title: () => titles[section.value] || '보관함' })
</script>

<template>
  <section class="mx-auto max-w-6xl bg-[var(--gvp-canvas)] px-4 py-8 text-[var(--gvp-text)] sm:px-8">
    <div class="flex flex-wrap items-start justify-between gap-5 border-b border-[var(--gvp-border)] pb-6">
      <div>
        <h1 class="text-2xl font-bold text-[var(--gvp-text)] sm:text-3xl">{{ titles[section] }}</h1>
        <p class="mt-2 text-sm text-[var(--gvp-text-secondary)] tabular-nums">
          {{ videos.length }}개의 동영상<span v-if="!isApiSection"> · 이 브라우저에 저장됨</span>
        </p>
      </div>
      <input
        v-model="query"
        type="search"
        :aria-label="`${titles[section]} 검색`"
        placeholder="동영상 검색"
        class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-4 py-2 text-sm text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-brand)] focus:bg-[var(--gvp-surface)]"
      />
    </div>

    <p v-if="library.storageError" role="alert" class="mt-4 rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-4 py-2.5 text-xs font-medium text-[var(--gvp-warning)]">
      브라우저 저장소에 접근할 수 없습니다. 변경 사항은 이 화면에서만 유지됩니다.
    </p>

    <div v-if="section === 'history'" class="my-6 flex flex-wrap gap-3">
      <button
        type="button"
        class="rounded-md bg-[var(--gvp-brand)] px-4 py-2 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
        @click="library.togglePaused()"
      >
        {{ library.paused ? '시청 기록 다시 저장' : '시청 기록 일시중지' }}
      </button>
      <button
        v-if="library.history.length"
        type="button"
        class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-4 py-2 text-xs font-semibold text-[var(--gvp-text)] transition hover:bg-[var(--gvp-canvas)]"
        @click="confirmClear = true"
      >
        시청 기록 모두 지우기
      </button>
      <div
        v-if="confirmClear"
        role="alert"
        class="flex w-full flex-wrap items-center justify-between gap-4 rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-4 text-xs text-[var(--gvp-live)]"
      >
        <span class="font-medium">이 브라우저의 모든 시청 기록을 삭제할까요?</span>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-surface)] px-3 py-1.5 font-semibold text-[var(--gvp-live)] hover:bg-[var(--gvp-live-soft)]"
            @click="library.clearHistory(); confirmClear = false"
          >
            삭제
          </button>
          <button
            type="button"
            class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-3 py-1.5 font-semibold text-[var(--gvp-text)] hover:bg-[var(--gvp-canvas)]"
            @click="confirmClear = false"
          >
            취소
          </button>
        </div>
      </div>
    </div>

    <p v-if="loading" role="status" class="py-16 text-center text-sm text-[var(--gvp-text-secondary)]">
      {{ titles[section] }} 불러오는 중…
    </p>

    <div v-else-if="failed && isApiSection" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] py-16 text-center">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">{{ titles[section] }} 불러오지 못했습니다.</p>
      <button
        type="button"
        class="mt-4 rounded-md bg-[var(--gvp-brand)] px-4 py-2 text-xs font-semibold text-white hover:bg-[var(--gvp-brand-hover)]"
        @click="fetchFeed"
      >
        다시 시도
      </button>
    </div>

    <div v-else-if="!videos.length" class="mt-6 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] py-20 text-center">
      <h2 class="text-lg font-bold text-[var(--gvp-text)]">{{ query ? '검색 결과가 없습니다' : '아직 동영상이 없습니다' }}</h2>
      <p class="mt-2 text-sm text-[var(--gvp-text-secondary)]">
        {{ section === 'saved' ? '동영상 메뉴에서 나중에 보기를 선택해 보세요.' : section === 'subscriptions' ? '좋아하는 채널을 구독하면 새로운 동영상을 모아 볼 수 있습니다.' : section === 'liked' ? '마음에 드는 동영상에 좋아요를 눌러 보세요.' : section === 'trending' ? '아직 집계된 인기 동영상이 없습니다.' : '이 카테고리에 표시할 동영상이 아직 없습니다.' }}
      </p>
      <NuxtLink to="/" class="mt-6 inline-block rounded-md bg-[var(--gvp-brand)] px-5 py-2.5 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]">
        동영상 둘러보기
      </NuxtLink>
    </div>

    <div v-else class="mt-6 space-y-4">
      <article
        v-for="video in videos"
        :key="video.id"
        class="flex items-start gap-4 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-3 transition hover:bg-[var(--gvp-canvas)]"
      >
        <NuxtLink :to="`/watch/${video.id}`" class="relative aspect-video w-40 shrink-0 overflow-hidden rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-elevated)] sm:w-72">
          <img :src="video.thumbnailUrl" :alt="video.title" loading="lazy" class="h-full w-full object-cover" />
          <span class="absolute bottom-2 right-2 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)]/95 px-1.5 py-0.5 font-mono text-xs font-semibold text-[var(--gvp-text)] tabular-nums">
            {{ Math.floor(video.durationSeconds / 60) }}:{{ String(video.durationSeconds % 60).padStart(2, '0') }}
          </span>
        </NuxtLink>
        <div class="min-w-0 flex-1">
          <NuxtLink :to="`/watch/${video.id}`" class="line-clamp-2 font-bold text-[var(--gvp-text)] hover:text-[var(--gvp-brand)] sm:text-base">
            {{ video.title }}
          </NuxtLink>
          <NuxtLink :to="`/channels/${video.channelId}`" class="mt-1.5 block text-xs font-medium text-[var(--gvp-text-secondary)] hover:text-[var(--gvp-text)]">
            {{ video.channel?.name }}
          </NuxtLink>
          <p class="mt-1 text-xs text-[var(--gvp-text-secondary)] tabular-nums">
            조회수 {{ video.viewCount.toLocaleString() }}회
          </p>
        </div>
        <button
          v-if="!isApiSection"
          type="button"
          :aria-label="`${video.title} 목록에서 삭제`"
          class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-2.5 py-1 text-xs font-semibold text-[var(--gvp-text-secondary)] transition hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]"
          @click="remove(video.id)"
        >
          ×
        </button>
      </article>
    </div>
  </section>
</template>
