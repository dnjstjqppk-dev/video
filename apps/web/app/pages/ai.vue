<script setup lang="ts">
import { BRAND } from '@gvp/design-tokens/src/brand'
import { formatTimecode, formatViewCount } from '@gvp/ui'

const appStore = useAppStore()
const api = useApi()
const ko = computed(() => appStore.locale === 'ko')
const toast = ref<string | null>(null)
function onCardToast(msg: string) {
  toast.value = msg
  setTimeout(() => { toast.value = null }, 4000)
}

useSeoMeta({
  title: computed(() => ko.value ? `AI Originals — AI 영화·드라마의 본진 | ${BRAND.name}` : `AI Originals — Home of AI film & drama | ${BRAND.name}`),
  description: computed(() => ko.value ? '심사를 통과한 AI 영화·AI 드라마 선정작과 시리즈를 정주행하세요.' : 'Binge curated AI films and drama series.'),
})

const { data, pending, error, refresh } = await useAsyncData('ai-originals', async () => {
  return await api.getOriginals({ type: 'all', limit: 48 })
})
const films = computed(() => (data.value?.videos.items || []).filter(v => v.videoType !== 'shorts'))
const shorts = computed(() => (data.value?.videos.items || []).filter(v => v.videoType === 'shorts'))
const series = computed(() => data.value?.series.items || [])

// Billboard feature: the first series wins (drama-first), else the top film.
const featuredSeries = computed(() => series.value[0] || null)
const { data: featuredDetail } = await useAsyncData('ai-featured-detail', async () => {
  if (!featuredSeries.value) return null
  try {
    return await api.getSeries(featuredSeries.value.id)
  } catch {
    return null
  }
})
const featuredPlayable = computed(() =>
  (featuredDetail.value?.episodes || []).find(ep => ep.status === 'READY' && ep.visibility === 'public') || null
)
const featuredFilm = computed(() => (!featuredSeries.value ? films.value[0] || null : null))
const heroImage = computed(() =>
  featuredSeries.value?.coverUrl
  || featuredDetail.value?.episodes[0]?.thumbnailUrl
  || featuredFilm.value?.thumbnailUrl
  || 'https://images.unsplash.com/photo-1531297484001-80022131f5a1?w=1600'
)
const heroPlayTo = computed(() =>
  featuredPlayable.value ? `/watch/${featuredPlayable.value.videoId}`
  : featuredFilm.value ? `/watch/${featuredFilm.value.id}`
  : null
)
const heroInfoTo = computed(() =>
  featuredSeries.value ? `/series/${featuredSeries.value.id}`
  : featuredFilm.value ? `/watch/${featuredFilm.value.id}`
  : null
)

const seriesRail = ref<HTMLElement | null>(null)
const filmsRail = ref<HTMLElement | null>(null)
const shortsRail = ref<HTMLElement | null>(null)
function railScroll(el: HTMLElement | null, dir: 1 | -1) {
  if (!el) return
  el.scrollBy({ left: dir * Math.max(320, el.clientWidth * 0.8), behavior: 'smooth' })
}
</script>

<template>
  <!-- Cinema mode: the AI hall stays dark in both themes, Netflix-style. -->
  <div class="min-h-screen bg-neutral-950 text-white">
    <!-- Billboard hero -->
    <section v-if="!pending && !error && (featuredSeries || featuredFilm)" class="relative h-[440px] w-full overflow-hidden sm:h-[560px]" aria-label="AI Originals 대표작">
      <img :src="heroImage" :alt="featuredSeries?.title || featuredFilm?.title || ''" class="absolute inset-0 h-full w-full object-cover" />
      <div class="absolute inset-0 bg-gradient-to-t from-neutral-950 via-neutral-950/40 to-transparent" />
      <div class="absolute inset-0 bg-gradient-to-r from-neutral-950/80 via-transparent to-transparent" />
      <div class="absolute inset-x-0 bottom-0 mx-auto max-w-[1720px] px-4 pb-10 sm:px-6">
        <p class="flex items-center gap-2 text-xs font-bold uppercase tracking-[0.2em] text-white/80">
          <AiBadge size="sm" approved /> Originals {{ ko ? '시리즈' : 'Series' }}
        </p>
        <h1 class="mt-3 max-w-3xl text-4xl font-black drop-shadow-lg sm:text-6xl">
          {{ featuredSeries?.title || featuredFilm?.title }}
        </h1>
        <p v-if="featuredSeries?.description || featuredFilm?.description" class="mt-3 max-w-xl text-sm text-white/80 sm:text-base">
          {{ (featuredSeries?.description || featuredFilm?.description || '').slice(0, 180) }}
        </p>
        <div class="mt-5 flex flex-wrap items-center gap-3">
          <NuxtLink
            v-if="heroPlayTo"
            :to="heroPlayTo"
            class="inline-flex items-center gap-2 rounded-md bg-white px-6 py-2.5 text-sm font-bold text-black transition hover:bg-white/80"
          >
            <span aria-hidden="true">▶</span> {{ ko ? '재생' : 'Play' }}
          </NuxtLink>
          <NuxtLink
            v-if="heroInfoTo"
            :to="heroInfoTo"
            class="inline-flex items-center gap-2 rounded-md bg-white/20 px-6 py-2.5 text-sm font-bold text-white backdrop-blur transition hover:bg-white/30"
          >
            {{ ko ? '상세 정보' : 'More Info' }}
          </NuxtLink>
        </div>
      </div>
    </section>

    <div class="mx-auto max-w-[1720px] px-4 pb-10 sm:px-6">
      <header class="border-b border-white/10 py-6">
        <h2 class="text-2xl font-bold sm:text-3xl">{{ ko ? 'AI 영화·드라마의 본진' : 'Home of AI film & drama' }}</h2>
        <p class="mt-2 max-w-2xl text-sm text-white/60">
          {{ ko
            ? '심사를 통과한 AI 창작물만 모았습니다. AI 라벨은 숨기지 않고 자랑합니다 — 메인 노출은 선정작의 특권입니다.'
            : 'Only reviewed AI works make this shelf. The AI label is a badge of honor — main exposure is earned.' }}
        </p>
      </header>

      <p v-if="toast" class="mt-4 rounded-lg border border-white/20 bg-white/10 px-4 py-3 text-sm text-white">{{ toast }}</p>

      <div v-if="pending" class="mt-8 grid grid-cols-1 gap-x-4 gap-y-8 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" aria-label="로딩 중">
        <div v-for="n in 8" :key="n" class="animate-pulse">
          <div class="aspect-video rounded-xl bg-white/10" />
          <div class="mt-3 h-4 w-3/4 rounded bg-white/10" />
          <div class="mt-2 h-3 w-1/2 rounded bg-white/10" />
        </div>
      </div>

      <div v-else-if="error" class="mt-8 rounded-xl border border-white/10 bg-white/5 p-10 text-center">
        <p class="text-sm text-white/60">{{ ko ? '선반을 불러오지 못했습니다.' : 'Could not load the shelf.' }}</p>
        <button type="button" class="mt-4 rounded-full bg-[var(--gvp-brand)] px-5 py-2 text-xs font-semibold text-white" @click="refresh()">{{ ko ? '다시 시도' : 'Retry' }}</button>
      </div>

      <template v-else>
        <!-- Series rail -->
        <section v-if="series.length" class="group/rail relative mt-8" aria-label="AI Originals 시리즈">
          <h3 class="mb-3 text-lg font-bold">{{ ko ? '시리즈 정주행' : 'Binge the series' }}</h3>
          <button type="button" :aria-label="ko ? '왼쪽으로 스크롤' : 'Scroll left'" class="absolute -left-2 top-1/2 z-10 hidden -translate-y-1/2 rounded-full bg-black/60 p-2 text-white opacity-0 transition hover:bg-black/80 group-hover/rail:opacity-100 md:block" @click="railScroll(seriesRail, -1)">‹</button>
          <div ref="seriesRail" class="flex gap-3 overflow-x-auto scroll-smooth pb-4 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
            <NuxtLink
              v-for="s in series"
              :key="s.id"
              :to="`/series/${s.id}`"
              class="group w-[280px] shrink-0 snap-start flex-col sm:w-[320px]"
            >
              <span class="relative block aspect-video w-full overflow-hidden rounded-lg bg-white/10 transition duration-300 group-hover:scale-[1.03]">
                <img :src="s.coverUrl || 'https://images.unsplash.com/photo-1531297484001-80022131f5a1?w=800'" :alt="s.title" class="h-full w-full object-cover" loading="lazy" />
                <AiBadge approved class="absolute left-2 top-2" />
              </span>
              <span class="mt-2 line-clamp-1 block text-sm font-semibold">{{ s.title }}</span>
              <span class="mt-0.5 block text-xs text-white/50">
                {{ s.status === 'completed' ? (ko ? '완결' : 'Completed') : (ko ? '연재 중' : 'Ongoing') }}
              </span>
            </NuxtLink>
          </div>
          <button type="button" :aria-label="ko ? '오른쪽으로 스크롤' : 'Scroll right'" class="absolute -right-2 top-1/2 z-10 hidden -translate-y-1/2 rounded-full bg-black/60 p-2 text-white opacity-0 transition hover:bg-black/80 group-hover/rail:opacity-100 md:block" @click="railScroll(seriesRail, 1)">›</button>
        </section>

        <!-- Films rail -->
        <section v-if="films.length" class="group/rail relative mt-8" aria-label="AI Originals 선정작">
          <h3 class="mb-3 text-lg font-bold">{{ ko ? '선정작' : 'Originals picks' }}</h3>
          <button type="button" :aria-label="ko ? '왼쪽으로 스크롤' : 'Scroll left'" class="absolute -left-2 top-1/2 z-10 hidden -translate-y-1/2 rounded-full bg-black/60 p-2 text-white opacity-0 transition hover:bg-black/80 group-hover/rail:opacity-100 md:block" @click="railScroll(filmsRail, -1)">‹</button>
          <div ref="filmsRail" class="flex gap-3 overflow-x-auto scroll-smooth pb-4 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
            <NuxtLink
              v-for="v in films"
              :key="v.id"
              :to="`/watch/${v.id}`"
              class="group w-[280px] shrink-0 snap-start flex-col sm:w-[320px]"
            >
              <span class="relative block aspect-video w-full overflow-hidden rounded-lg bg-white/10 transition duration-300 group-hover:scale-[1.03]">
                <img :src="v.thumbnailUrl" :alt="v.title" class="h-full w-full object-cover" loading="lazy" />
                <AiBadge :tool="v.aiTool" approved class="absolute left-2 top-2" />
                <span class="absolute bottom-2 right-2 rounded bg-black/80 px-1 py-0.5 text-xs font-medium text-white tabular-nums">{{ formatTimecode(v.durationSeconds) }}</span>
              </span>
              <span class="mt-2 line-clamp-1 block text-sm font-semibold">{{ v.title }}</span>
              <span class="mt-0.5 block text-xs tabular-nums text-white/50">{{ formatViewCount(v.viewCount, appStore.locale) }}</span>
            </NuxtLink>
          </div>
          <button type="button" :aria-label="ko ? '오른쪽으로 스크롤' : 'Scroll right'" class="absolute -right-2 top-1/2 z-10 hidden -translate-y-1/2 rounded-full bg-black/60 p-2 text-white opacity-0 transition hover:bg-black/80 group-hover/rail:opacity-100 md:block" @click="railScroll(filmsRail, 1)">›</button>
        </section>
        <p v-if="!films.length && !shorts.length" class="mt-8 rounded-xl border border-white/10 bg-white/5 p-10 text-center text-sm text-white/60">
          {{ ko ? '아직 선정작이 없습니다. 스튜디오에서 AI 작품을 제출해 보세요.' : 'No picks yet. Submit your AI work from the studio.' }}
        </p>

        <!-- Shorts rail -->
        <section v-if="shorts.length" class="group/rail relative mt-8" aria-label="AI 쇼츠">
          <h3 class="mb-3 text-lg font-bold">{{ ko ? 'AI 쇼츠' : 'AI Shorts' }}</h3>
          <button type="button" :aria-label="ko ? '왼쪽으로 스크롤' : 'Scroll left'" class="absolute -left-2 top-1/2 z-10 hidden -translate-y-1/2 rounded-full bg-black/60 p-2 text-white opacity-0 transition hover:bg-black/80 group-hover/rail:opacity-100 md:block" @click="railScroll(shortsRail, -1)">‹</button>
          <div ref="shortsRail" class="flex gap-3 overflow-x-auto scroll-smooth pb-4 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
            <NuxtLink v-for="v in shorts" :key="v.id" :to="`/watch/${v.id}`" class="group w-36 shrink-0 snap-start flex-col sm:w-44">
              <span class="relative block aspect-[9/16] w-full overflow-hidden rounded-lg bg-white/10 transition duration-300 group-hover:scale-[1.03]">
                <img :src="v.thumbnailUrl" :alt="v.title" class="h-full w-full object-cover" loading="lazy" />
                <AiBadge :tool="v.aiTool" approved class="absolute left-2 top-2" />
              </span>
              <span class="mt-2 line-clamp-2 block text-xs font-semibold">{{ v.title }}</span>
            </NuxtLink>
          </div>
          <button type="button" :aria-label="ko ? '오른쪽으로 스크롤' : 'Scroll right'" class="absolute -right-2 top-1/2 z-10 hidden -translate-y-1/2 rounded-full bg-black/60 p-2 text-white opacity-0 transition hover:bg-black/80 group-hover/rail:opacity-100 md:block" @click="railScroll(shortsRail, 1)">›</button>
        </section>
      </template>
    </div>
  </div>
</template>
