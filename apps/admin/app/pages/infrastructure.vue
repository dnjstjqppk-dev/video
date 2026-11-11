<script setup lang="ts">
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'

const api = useAdminApi()

useSeoMeta({
  title: 'Platform Health & Service Status — Super Admin'
})

const autoRefresh = ref(true)
let timer: ReturnType<typeof setInterval> | null = null

const { data: health, pending: healthPending, error: healthError, refresh: refreshHealth } = await useAsyncData(
  'admin-service-health',
  async () => await api.getServiceHealth()
)

const { data: dashboard, pending: dashPending, error: dashError, refresh: refreshDashboard } = await useAsyncData(
  'admin-infra-dashboard',
  async () => await api.getAdminDashboard()
)

function refreshAll() {
  refreshHealth()
  refreshDashboard()
}

onMounted(() => {
  timer = setInterval(() => {
    if (autoRefresh.value && !document.hidden) refreshAll()
  }, 30000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})

const apiAlive = computed(() => health.value?.alive?.status === 'UP')
const apiReady = computed(() => health.value?.ready?.ready === true)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-900">
          플랫폼 헬스 및 서비스 상태 (Service Health)
        </h1>
        <p class="mt-1 text-sm text-slate-600">
          Live API liveness/readiness probes · Platform totals · 30초 자동 새로고침
        </p>
      </div>
      <div class="flex items-center gap-3 text-sm">
        <label class="flex items-center gap-2 text-slate-700">
          <input v-model="autoRefresh" type="checkbox" class="rounded border-slate-300 accent-slate-900" />
          <span class="tabular-nums">자동 새로고침 (30s)</span>
        </label>
        <button
          type="button"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
          @click="refreshAll"
        >
          새로고침
        </button>
      </div>
    </div>

    <!-- Liveness / Readiness -->
    <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
      <div class="overflow-hidden rounded-lg border border-slate-200 bg-white">
        <div class="border-b border-slate-200 bg-slate-50 px-5 py-3.5">
          <h2 class="text-sm font-semibold text-slate-900">
            API Liveness (<span class="font-mono text-xs">/healthz</span>)
          </h2>
        </div>
        <div class="p-5">
          <div v-if="healthPending" class="h-16 animate-pulse rounded-lg bg-slate-100" role="status" aria-label="불러오는 중" />
          <GvpBanner v-else-if="healthError" tone="danger">
            헬스 정보를 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refreshAll()">다시 시도</button>
          </GvpBanner>
          <div v-else class="flex items-center gap-3">
            <span
              class="rounded-md px-2.5 py-1 font-mono text-xs font-semibold"
              :class="apiAlive ? 'border border-emerald-200 bg-emerald-50 text-emerald-700' : 'border border-rose-200 bg-rose-50 text-rose-700'"
              role="status"
            >
              {{ health?.alive?.status || 'UNKNOWN' }}
            </span>
            <span class="font-mono text-xs text-slate-600 tabular-nums">
              v{{ health?.alive?.version || '?' }} · {{ health?.alive?.timestamp || '-' }}
            </span>
          </div>
        </div>
      </div>

      <div class="overflow-hidden rounded-lg border border-slate-200 bg-white">
        <div class="border-b border-slate-200 bg-slate-50 px-5 py-3.5">
          <h2 class="text-sm font-semibold text-slate-900">
            API Readiness (<span class="font-mono text-xs">/ready</span>)
          </h2>
        </div>
        <div class="p-5">
          <div v-if="healthPending" class="h-16 animate-pulse rounded-lg bg-slate-100" role="status" aria-label="불러오는 중" />
          <div v-else-if="!healthError">
            <div class="flex items-center gap-3">
              <span
                class="rounded-md px-2.5 py-1 font-mono text-xs font-semibold"
                :class="apiReady ? 'border border-emerald-200 bg-emerald-50 text-emerald-700' : 'border border-rose-200 bg-rose-50 text-rose-700'"
                role="status"
              >
                {{ health?.ready?.status || 'NOT READY' }}
              </span>
              <span class="font-mono text-xs text-slate-600 tabular-nums">{{ health?.ready?.timestamp || '-' }}</span>
            </div>
            <div v-if="health?.ready?.subsystems" class="mt-3 flex flex-wrap gap-2 text-xs">
              <span
                v-for="(state, name) in health.ready.subsystems"
                :key="name"
                class="rounded-md border border-slate-200 bg-slate-50 px-2.5 py-1 font-mono text-slate-700"
              >
                {{ name }}: <strong :class="state === 'operational' ? 'text-emerald-700' : 'text-rose-700'">{{ state }}</strong>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Platform Totals -->
    <div class="overflow-hidden rounded-lg border border-slate-200 bg-white">
      <div class="border-b border-slate-200 bg-slate-50 px-5 py-3.5">
        <h2 class="text-sm font-semibold text-slate-900">플랫폼 현황 (Live Totals)</h2>
      </div>
      <div class="p-5">
        <div v-if="dashPending" class="grid grid-cols-2 gap-4 md:grid-cols-4" role="status" aria-label="불러오는 중">
          <div v-for="n in 4" :key="n" class="h-20 animate-pulse rounded-lg bg-slate-100" />
        </div>
        <GvpBanner v-else-if="dashError" tone="danger">
          현황을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refreshAll()">다시 시도</button>
        </GvpBanner>
        <div v-else-if="dashboard" class="grid grid-cols-2 gap-4 md:grid-cols-4" role="status">
          <div class="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <p class="text-xs font-medium text-slate-500">총 사용자</p>
            <p class="mt-1 font-mono text-xl font-bold text-slate-900 tabular-nums">{{ dashboard.metrics.totalUsers.toLocaleString() }}</p>
          </div>
          <div class="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <p class="text-xs font-medium text-slate-500">총 영상</p>
            <p class="mt-1 font-mono text-xl font-bold text-slate-900 tabular-nums">{{ dashboard.metrics.totalVideos.toLocaleString() }}</p>
          </div>
          <div class="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <p class="text-xs font-medium text-slate-500">진행 중 라이브</p>
            <p class="mt-1 font-mono text-xl font-bold text-rose-600 tabular-nums">{{ dashboard.metrics.liveStreamsActive.toLocaleString() }}</p>
          </div>
          <div class="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <p class="text-xs font-medium text-slate-500">시스템 상태</p>
            <p class="mt-1 font-mono text-xl font-bold text-emerald-700 tabular-nums">{{ dashboard.metrics.systemHealth }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Deep Observability & DR Simulation Reference -->
    <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <h2 class="text-sm font-semibold text-slate-900">심층 관측 (Deep Observability)</h2>
        <p class="mt-1.5 text-xs text-slate-600 leading-relaxed">
          CPU·메모리·파드·카나리·WAL 지표는 Prometheus/Grafana에서 확인하세요. Compose 스택은
          <span class="font-mono text-slate-900 tabular-nums">prometheus:9090</span>·<span class="font-mono text-slate-900 tabular-nums">grafana:3000</span>으로 노출됩니다.
        </p>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <div class="flex items-center justify-between gap-2">
          <h2 class="text-sm font-semibold text-slate-900">DR 페일오버 시뮬레이션 시연 (Simulation Only)</h2>
          <span class="rounded-md border border-amber-200 bg-amber-50 px-2 py-0.5 font-mono text-xs font-semibold text-amber-700">
            SIMULATION ONLY
          </span>
        </div>
        <p class="mt-1.5 text-xs text-slate-600 leading-relaxed">
          실제 데이터베이스 승격이나 쿠버네티스 클러스터 조작을 수행하지 않으며, 재해 복구 런북 절차를 검증하기 위한 시뮬레이션 안내입니다.
        </p>
      </div>
    </div>
  </div>
</template>
