<script setup lang="ts">
const api = useStudioApi()
const selectedRange = ref<'7d' | '28d' | '90d' | '365d' | 'all'>('28d')
const activeTab = ref<'overview' | 'content' | 'audience' | 'revenue'>('overview')
const selectedMetric = ref<'views' | 'watch' | 'subscribers' | 'revenue'>('views')
const chartCanvasRef = ref<HTMLElement | null>(null)
let chart: { setOption: (o: unknown, notMerge?: boolean) => void; resize: () => void; dispose: () => void } | null = null

useSeoMeta({
  title: 'Creator Analytics — Views, Retention, CTR & Revenue'
})

const tabs = [
  { id: 'overview', label: '개요' },
  { id: 'content', label: '콘텐츠' },
  { id: 'audience', label: '시청자층' },
  { id: 'revenue', label: '수익' }
] as const

const metricOfTab = { overview: null, content: 'views', audience: 'subscribers', revenue: 'revenue' } as const
const activeMetric = computed(() => metricOfTab[activeTab.value] ?? selectedMetric.value)

const metricMeta = {
  views: { label: '조회수', color: '#2563EB' },
  watch: { label: '시청 시간(시간)', color: '#FF2E4C' },
  subscribers: { label: '구독자', color: '#22C55E' },
  revenue: { label: '예상 수익(원)', color: '#F59E0B' }
} as const

const tabNote: Record<string, string> = {
  content: '동영상별 노출 클릭률·시청 지속률 등 세부 콘텐츠 분석은 준비 중입니다.',
  audience: '연령·지역·신규/재방문자 등 세부 시청자층 분석은 준비 중입니다.',
  revenue: '수익원별(광고·Premium·SuperChat) 세부 분석은 준비 중입니다.'
}

const { data: analyticsData, error: analyticsError } = await useAsyncData(
  () => `studio-analytics-${selectedRange.value}`,
  async () => {
    return await api.getStudioAnalytics(selectedRange.value)
  },
  { watch: [selectedRange] }
)

const seriesPoints = computed(() => {
  const d = analyticsData.value
  return {
    dates: (d?.viewsTimeSeries || []).map(p => p.date),
    views: (d?.viewsTimeSeries || []).map(p => p.value),
    watch: (d?.watchTimeTimeSeries || []).map(p => p.value),
    subscribers: (d?.subscriberTimeSeries || []).map(p => p.value),
    revenue: (d?.revenueTimeSeries || []).map(p => p.value)
  }
})
const totals = computed(() => ({
  views: seriesPoints.value.views.reduce((s, v) => s + v, 0),
  watch: seriesPoints.value.watch.reduce((s, v) => s + v, 0),
  subscribers: seriesPoints.value.subscribers.reduce((s, v) => s + v, 0),
  revenue: seriesPoints.value.revenue.reduce((s, v) => s + v, 0)
}))
const hasData = computed(() => seriesPoints.value.dates.length > 0)

function renderChart() {
  if (!chart) return
  const m = activeMetric.value
  chart.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: seriesPoints.value.dates, axisLabel: { color: '#94A3B8' } },
    yAxis: { type: 'value', axisLabel: { color: '#94A3B8' }, splitLine: { lineStyle: { color: 'rgba(148, 163, 184, 0.16)' } } },
    series: [{
      name: metricMeta[m].label,
      type: 'line',
      smooth: true,
      data: seriesPoints.value[m],
      itemStyle: { color: metricMeta[m].color },
      lineStyle: { color: metricMeta[m].color },
      areaStyle: { color: metricMeta[m].color, opacity: 0.12 }
    }]
  }, true)
}

watch([activeMetric, seriesPoints], async () => {
  await nextTick()
  renderChart()
  chart?.resize()
})

async function initChart() {
  if (chart || !chartCanvasRef.value) return
  try {
    const echarts = await import('echarts')
    chart = echarts.init(chartCanvasRef.value)
    renderChart()
  } catch {
    chart = null
  }
}

watch(chartCanvasRef, () => { void initChart() })

onMounted(() => {
  void initChart()
  window.addEventListener('resize', onResize)
})
function onResize() { chart?.resize() }
onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  chart?.dispose()
  chart = null
})

const { data: realtimeData, pending: realtimePending, error: realtimeError } = await useAsyncData('studio-realtime-48h', async () => {
  try {
    return await api.getStudioRealtime()
  } catch {
    return null
  }
})
const realtimeHours = computed(() => realtimeData.value?.hours || [])
const realtimeTotal = computed(() => realtimeData.value?.total48h || 0)
const realtimeMax = computed(() => Math.max(1, ...realtimeHours.value.map(h => h.views)))
const realtimeHasViews = computed(() => realtimeHours.value.some(h => h.views > 0))

function realtimeLabel(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return `${d.getUTCMonth() + 1}/${d.getUTCDate()} ${String(d.getUTCHours()).padStart(2, '0')}시(UTC)`
}

const { data: topData } = await useAsyncData('studio-top-content', async () => {
  try {
    const res = await api.getStudioVideos(1, 100)
    return [...(res.items || [])].sort((a, b) => b.viewCount - a.viewCount).slice(0, 10)
  } catch {
    return []
  }
})
const topContent = computed(() => topData.value || [])

function exportCsvReport() {
  const header = 'Date,Views,WatchHours,EstimatedRevenue_KRW\n'
  const rows = seriesPoints.value.dates
    .map((d, i) => `${d},${seriesPoints.value.views[i] ?? 0},${seriesPoints.value.watch[i] ?? 0},${seriesPoints.value.revenue[i] ?? 0}`)
    .join('\n')
  if (!rows) return
  const blob = new Blob([header + rows], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `vravra_studio_analytics_${selectedRange.value}.csv`
  link.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">채널 분석</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">일별 조회수 · 시청 시간 · 구독자 · 수익 추이 및 CSV 내보내기</p>
      </div>
      <div class="flex items-center gap-2">
        <select v-model="selectedRange" class="h-9 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-3 text-xs font-medium tabular-nums outline-none focus:border-[var(--gvp-focus)]">
          <option value="7d">지난 7일</option>
          <option value="28d">지난 28일</option>
          <option value="90d">지난 90일</option>
          <option value="365d">지난 365일</option>
          <option value="all">전체 기간</option>
        </select>
        <button type="button" class="inline-flex h-9 items-center gap-2 rounded-full bg-[var(--gvp-studio)] px-4 text-sm font-medium text-white transition-colors hover:opacity-90" @click="exportCsvReport">
          <span>CSV 내보내기</span>
        </button>
      </div>
    </div>

    <div class="flex gap-6 border-b border-[var(--gvp-border)] text-xs font-semibold" role="tablist" aria-label="분석 섹션">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        role="tab"
        :aria-selected="activeTab === tab.id"
        class="border-b-2 pb-3 tracking-wider transition-colors"
        :class="activeTab === tab.id ? 'border-[var(--gvp-focus)] text-[var(--gvp-text)]' : 'border-transparent text-[var(--gvp-text-secondary)] hover:text-[var(--gvp-text)]'"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
      </button>
    </div>

    <div v-if="activeTab === 'overview'" class="grid grid-cols-2 gap-3 xl:grid-cols-4" role="group" aria-label="지표 선택">
      <button
        v-for="m in (['views', 'watch', 'subscribers', 'revenue'] as const)"
        :key="m"
        type="button"
        :aria-pressed="selectedMetric === m"
        class="rounded-xl border p-4 text-left transition-colors"
        :class="selectedMetric === m ? 'border-[var(--gvp-focus)] bg-[var(--gvp-elevated)]' : 'border-[var(--gvp-border)] bg-[var(--gvp-surface)] hover:bg-[var(--gvp-canvas)]'"
        @click="selectedMetric = m"
      >
        <span class="text-xs font-medium text-[var(--gvp-text-secondary)]">{{ metricMeta[m].label }}</span>
        <span class="mt-1 block text-xl font-semibold text-[var(--gvp-text)] tabular-nums">{{ totals[m].toLocaleString() }}</span>
      </button>
    </div>

    <div class="grid grid-cols-1 gap-4 xl:grid-cols-3">
      <div class="rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5 xl:col-span-2">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-[var(--gvp-border)] pb-3">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">{{ metricMeta[activeMetric].label }} 추이</h2>
          <span class="text-xs font-medium text-[var(--gvp-text-secondary)] tabular-nums">{{ selectedRange }} · 일별 집계</span>
        </div>
        <div v-if="analyticsError" class="mt-4 rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
          <p class="text-sm font-semibold text-[var(--gvp-live)]">분석 데이터를 불러오지 못했습니다.</p>
        </div>
        <div v-else-if="!hasData" class="mt-4 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-12 text-center">
          <p class="text-sm font-semibold text-[var(--gvp-text-secondary)]">아직 집계된 시청 기록이 없습니다.</p>
        </div>
        <ClientOnly>
        <div v-show="hasData && !analyticsError" ref="chartCanvasRef" class="mt-4 h-72 w-full" />
        <template #fallback>
          <div class="mt-4 h-72 w-full animate-pulse rounded-lg bg-[var(--gvp-elevated)]" aria-hidden="true" />
        </template>
      </ClientOnly>
      </div>

      <div class="rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5" aria-label="실시간 48시간 조회수">
        <div class="border-b border-[var(--gvp-border)] pb-3">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">실시간 48시간</h2>
          <p class="tabular-nums mt-1 text-2xl font-semibold text-[var(--gvp-text)]">{{ realtimeTotal.toLocaleString() }}</p>
          <p class="text-xs text-[var(--gvp-text-secondary)]">시간당 조회수 합계</p>
        </div>
        <div v-if="realtimePending" class="mt-4 flex h-72 items-end gap-[3px]" role="status" aria-label="불러오는 중">
          <div v-for="i in 48" :key="i" class="flex-1 animate-pulse rounded-sm bg-[var(--gvp-elevated)]" :style="{ height: `${20 + ((i * 37) % 60)}%` }" />
        </div>
        <p v-else-if="realtimeError || !realtimeData" role="alert" class="mt-4 rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-6 text-center text-xs font-semibold text-[var(--gvp-live)]">
          실시간 데이터를 불러오지 못했습니다.
        </p>
        <p v-else-if="!realtimeHasViews" class="mt-4 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-6 text-center text-xs text-[var(--gvp-text-secondary)]">
          지난 48시간 동안 집계된 조회수가 없습니다.
        </p>
        <div v-else class="tabular-nums mt-4 flex h-72 items-end gap-[3px]" role="img" aria-label="시간당 조회수 막대 차트">
          <div
            v-for="h in realtimeHours" :key="h.hour"
            class="min-h-[2px] flex-1 rounded-sm bg-[var(--gvp-studio)]"
            :style="{ height: `${Math.max(2, Math.round((h.views / realtimeMax) * 100))}%` }"
            :title="`${realtimeLabel(h.hour)} · ${h.views.toLocaleString()}회`"
          />
        </div>
      </div>
    </div>

    <div v-if="activeTab !== 'overview'" class="rounded-xl border border-dashed border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] p-6 text-center">
      <p class="text-sm font-semibold text-[var(--gvp-text-secondary)]">{{ tabNote[activeTab] }}</p>
      <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">일별 추이는 위 차트에서 확인할 수 있습니다.</p>
    </div>

    <div class="overflow-hidden rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <div class="border-b border-[var(--gvp-border)] px-5 py-4">
        <h2 class="text-sm font-semibold text-[var(--gvp-text)]">인기 콘텐츠 순위</h2>
        <p class="mt-0.5 text-xs text-[var(--gvp-text-secondary)]">조회수 기준 상위 10개</p>
      </div>
      <p v-if="!topContent.length" class="p-8 text-center text-sm text-[var(--gvp-text-secondary)]">아직 집계할 동영상이 없습니다.</p>
      <table v-else class="w-full text-left">
        <thead class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-secondary)]">
          <tr>
            <th scope="col" class="w-12 px-4 py-2.5 text-right">순위</th>
            <th scope="col" class="px-4 py-2.5">동영상</th>
            <th scope="col" class="px-4 py-2.5 text-right">조회수</th>
            <th scope="col" class="px-4 py-2.5 text-right">좋아요</th>
            <th scope="col" class="px-4 py-2.5 text-right">댓글</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--gvp-border)]">
          <tr v-for="(v, i) in topContent" :key="v.id" class="transition-colors hover:bg-[var(--gvp-canvas)]">
            <td class="px-4 py-3 text-right font-mono text-sm font-bold text-[var(--gvp-text)] tabular-nums">{{ i + 1 }}</td>
            <td class="px-4 py-3">
              <span class="flex min-w-0 items-center gap-3">
                <img :src="v.thumbnailUrl" :alt="v.title" class="h-10 w-[72px] shrink-0 rounded-md object-cover" loading="lazy" />
                <span class="block max-w-64 truncate text-sm font-medium text-[var(--gvp-text)]">{{ v.title }}</span>
              </span>
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-sm text-[var(--gvp-text)] tabular-nums">{{ v.viewCount.toLocaleString() }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-sm text-[var(--gvp-text)] tabular-nums">{{ v.likeCount.toLocaleString() }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right font-mono text-sm text-[var(--gvp-text)] tabular-nums">{{ v.commentCount.toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
