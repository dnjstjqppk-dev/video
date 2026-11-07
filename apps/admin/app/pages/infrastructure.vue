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
  <div>
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">플랫폼 헬스 및 서비스 상태 (Service Health)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Live API liveness/readiness probes · Platform totals · 30초 자동 새로고침
        </p>
      </div>
      <div class="flex items-center gap-3 text-xs">
        <label class="flex items-center gap-2 text-neutral-300">
          <input v-model="autoRefresh" type="checkbox" class="rounded accent-red-600" />
          자동 새로고침 (30s)
        </label>
        <button
          type="button"
          class="rounded-xl border border-white/15 px-4 py-2 font-bold text-neutral-200 hover:bg-white/10"
          @click="refreshAll"
        >
          새로고침
        </button>
      </div>
    </div>

    <!-- Liveness / Readiness -->
    <div class="mt-6 grid grid-cols-1 gap-6 md:grid-cols-2">
      <div class="rounded-2xl border border-white/10 bg-[#131722] p-5">
        <h2 class="text-sm font-bold text-white">API Liveness (`/healthz`)</h2>
        <div v-if="healthPending" class="mt-3 h-16 animate-pulse rounded-xl bg-white/5" role="status" aria-label="불러오는 중" />
        <GvpBanner v-else-if="healthError" tone="danger" class="mt-3">
          헬스 정보를 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refreshAll()">다시 시도</button>
        </GvpBanner>
        <div v-else class="mt-3 flex items-center gap-3">
          <span
            class="rounded px-2.5 py-1 font-mono text-xs font-extrabold"
            :class="apiAlive ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-300'"
            role="status"
          >
            {{ health?.alive?.status || 'UNKNOWN' }}
          </span>
          <span class="font-mono text-xs text-neutral-400">v{{ health?.alive?.version || '?' }} · {{ health?.alive?.timestamp || '-' }}</span>
        </div>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#131722] p-5">
        <h2 class="text-sm font-bold text-white">API Readiness (`/ready`)</h2>
        <div v-if="healthPending" class="mt-3 h-16 animate-pulse rounded-xl bg-white/5" role="status" aria-label="불러오는 중" />
        <div v-else-if="!healthError" class="mt-3">
          <div class="flex items-center gap-3">
            <span
              class="rounded px-2.5 py-1 font-mono text-xs font-extrabold"
              :class="apiReady ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-300'"
              role="status"
            >
              {{ health?.ready?.status || 'NOT READY' }}
            </span>
            <span class="font-mono text-xs text-neutral-400">{{ health?.ready?.timestamp || '-' }}</span>
          </div>
          <div v-if="health?.ready?.subsystems" class="mt-3 flex flex-wrap gap-2 text-xs">
            <span
              v-for="(state, name) in health.ready.subsystems"
              :key="name"
              class="rounded-lg bg-white/5 px-2.5 py-1 font-mono text-neutral-300"
            >
              {{ name }}: <strong :class="state === 'operational' ? 'text-emerald-400' : 'text-red-300'">{{ state }}</strong>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Platform Totals -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-5">
      <h2 class="text-sm font-bold text-white">플랫폼 현황 (Live Totals)</h2>
      <div v-if="dashPending" class="mt-4 grid grid-cols-2 gap-4 md:grid-cols-4" role="status" aria-label="불러오는 중">
        <div v-for="n in 4" :key="n" class="h-20 animate-pulse rounded-xl bg-white/5" />
      </div>
      <GvpBanner v-else-if="dashError" tone="danger" class="mt-4">
        현황을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refreshAll()">다시 시도</button>
      </GvpBanner>
      <div v-else-if="dashboard" class="mt-4 grid grid-cols-2 gap-4 md:grid-cols-4" role="status">
        <div class="rounded-xl bg-[#090B10] p-4">
          <p class="text-xs text-neutral-400">총 사용자</p>
          <p class="mt-1 font-mono text-xl font-extrabold text-white">{{ dashboard.metrics.totalUsers.toLocaleString() }}</p>
        </div>
        <div class="rounded-xl bg-[#090B10] p-4">
          <p class="text-xs text-neutral-400">총 영상</p>
          <p class="mt-1 font-mono text-xl font-extrabold text-white">{{ dashboard.metrics.totalVideos.toLocaleString() }}</p>
        </div>
        <div class="rounded-xl bg-[#090B10] p-4">
          <p class="text-xs text-neutral-400">진행 중 라이브</p>
          <p class="mt-1 font-mono text-xl font-extrabold text-[#FF0033]">{{ dashboard.metrics.liveStreamsActive.toLocaleString() }}</p>
        </div>
        <div class="rounded-xl bg-[#090B10] p-4">
          <p class="text-xs text-neutral-400">시스템 상태</p>
          <p class="mt-1 font-mono text-xl font-extrabold text-emerald-400">{{ dashboard.metrics.systemHealth }}</p>
        </div>
      </div>
    </div>

    <!-- Deep Observability -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#131722] p-5 text-xs text-neutral-300">
      <h2 class="text-sm font-bold text-white">심층 관측 (Deep Observability)</h2>
      <p class="mt-1 text-neutral-400">
        CPU·메모리·파드·카나리·WAL 지표는 Prometheus/Grafana에서 확인하세요. Compose 스택은
        <span class="font-mono">prometheus:9090</span>·<span class="font-mono">grafana:3000</span>으로 노출됩니다.
      </p>
    </div>
  </div>
</template>
