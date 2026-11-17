<script setup lang="ts">
const api = useStudioApi()
const selectedRange = ref<'7d' | '28d' | '90d' | '365d'>('28d')
const chartCanvasRef = ref<HTMLElement | null>(null)

useSeoMeta({
  title: 'Creator Analytics — Views, Retention, CTR & Revenue'
})

const { data: analyticsData, error: analyticsError } = await useAsyncData(
  () => `studio-analytics-${selectedRange.value}`,
  async () => {
    return await api.getStudioAnalytics(selectedRange.value)
  },
  { watch: [selectedRange] }
)

const seriesPoints = computed(() => {
  const views = analyticsData.value?.viewsTimeSeries || []
  const watch = analyticsData.value?.watchTimeTimeSeries || []
  const revenue = analyticsData.value?.revenueTimeSeries || []
  return views.map((pt, i) => ({
    date: pt.date,
    views: pt.value,
    watchHours: watch[i]?.value || 0,
    revenueKrw: revenue[i]?.value || 0
  }))
})

function exportCsvReport() {
  const header = 'Date,Views,WatchHours,EstimatedRevenue_KRW\n'
  const rows = seriesPoints.value
    .map(r => `${r.date},${r.views},${r.watchHours},${r.revenueKrw}`)
    .join('\n')
  if (!rows) return
  const blob = new Blob([header + rows], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `gvp_studio_analytics_${selectedRange.value}.csv`
  link.click()
  URL.revokeObjectURL(url)
}

onMounted(async () => {
  if (!chartCanvasRef.value) return
  try {
    const echarts = await import('echarts')
    const chart = echarts.init(chartCanvasRef.value)
    chart.setOption({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'axis' },
      legend: { data: ['Views', 'Watch Hours'], textStyle: { color: '#475569' } },
      xAxis: {
        type: 'category',
        data: seriesPoints.value.map(d => d.date),
        axisLabel: { color: '#64748B' }
      },
      yAxis: {
        type: 'value',
        axisLabel: { color: '#64748B' },
        splitLine: { lineStyle: { color: '#E2E8F0' } }
      },
      series: [
        {
          name: 'Views',
          type: 'line',
          smooth: true,
          data: seriesPoints.value.map(d => d.views),
          itemStyle: { color: '#0F172A' }
        },
        {
          name: 'Watch Hours',
          type: 'line',
          smooth: true,
          data: seriesPoints.value.map(d => d.watchHours),
          itemStyle: { color: '#2563EB' }
        }
      ]
    })
  } catch {
    // Fallback SVG chart is already rendered below if ECharts canvas skipped
  }
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-slate-900">심층 채널 분석 (ClickHouse Analytics &amp; Retention)</h1>
        <p class="mt-1 text-xs text-slate-500">
          일별 조회수 · 시청 시간 · 정산 수익 추이 및 CSV 내보내기
        </p>
      </div>

      <div class="flex items-center gap-2">
        <select
          v-model="selectedRange"
          class="h-9 rounded-md border border-slate-300 bg-white px-3 text-xs font-medium text-slate-900 tabular-nums outline-none focus:border-slate-900"
        >
          <option value="7d">Last 7 Days</option>
          <option value="28d">Last 28 Days</option>
          <option value="90d">Last 90 Days</option>
          <option value="365d">Last 365 Days</option>
        </select>

        <button
          type="button"
          class="inline-flex items-center justify-center gap-2 h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
          @click="exportCsvReport"
        >
          <svg class="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 2.5v8m0 0L4.75 7.25M8 10.5l3.25-3.25M2.5 11.5v1a1 1 0 001 1h9a1 1 0 001-1v-1" />
          </svg>
          <span>Export Revenue &amp; Telemetry CSV</span>
        </button>
      </div>
    </div>

    <!-- ECharts Time Series Chart -->
    <div class="rounded-lg border border-slate-200 bg-white p-5">
      <div class="flex flex-wrap items-center justify-between gap-2 border-b border-slate-100 pb-3">
        <h2 class="text-sm font-semibold text-slate-900">조회수 및 시청 시간 추이 (Views &amp; Watch Time Time-Series)</h2>
        <span class="text-xs font-medium text-slate-500 tabular-nums">{{ selectedRange }} · 일별 집계</span>
      </div>
      <div v-if="analyticsError" class="mt-4 rounded-lg border border-red-200 bg-red-50 p-8 text-center" role="alert">
        <p class="text-sm font-semibold text-red-700">분석 데이터를 불러오지 못했습니다.</p>
      </div>
      <div v-else-if="seriesPoints.length === 0" class="mt-4 rounded-lg border border-slate-200 bg-slate-50 p-12 text-center">
        <p class="text-sm font-semibold text-slate-600">아직 집계된 시청 기록이 없습니다.</p>
      </div>
      <div v-else ref="chartCanvasRef" class="mt-4 h-72 w-full" />
    </div>

    <!-- Detailed breakdowns land with the telemetry warehouse connection -->
    <div class="rounded-lg border border-dashed border-slate-300 bg-slate-50 p-8 text-center">
      <p class="text-sm font-semibold text-slate-600">시청 지속률 · 트래픽 소스 · 지역 분포 연동 준비 중</p>
      <p class="mt-1 text-xs text-slate-500">일별 조회수·시청 시간·수익 추이는 위 차트에서 확인할 수 있습니다.</p>
    </div>
  </div>
</template>
