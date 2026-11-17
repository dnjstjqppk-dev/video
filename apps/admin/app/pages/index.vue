<script setup lang="ts">
const api = useAdminApi()

useSeoMeta({
  title: 'Platform Operations Dashboard — Infrastructure & Global Metrics'
})

const { data: adminDash, pending: dashPending, error: dashError } = await useAsyncData('admin-dashboard', async () => {
  return await api.getAdminDashboard()
})

const { data: modQueue } = await useAsyncData('admin-index-modcount', async () => {
  try {
    return await api.getAdminModerationQueue()
  } catch {
    return []
  }
})
const openModCount = computed(() => (modQueue.value || []).length)
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-slate-200 pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-slate-900">
        플랫폼 통합 운영 관제 (Platform Operations Dashboard)
      </h1>
      <p class="mt-1 text-sm text-slate-600">
        Real-time global platform KPIs, transcoding queue depth, gross revenue, and 7-tier infrastructure health
      </p>
    </div>

    <!-- Top KPI Grid -->
    <div v-if="dashPending" class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
      <div v-for="i in 5" :key="i" class="h-24 animate-pulse rounded-lg border border-slate-200 bg-slate-100" />
    </div>
    <div v-else-if="dashError || !adminDash" class="rounded-lg border border-red-200 bg-red-50 p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-red-700">운영 지표를 불러오지 못했습니다.</p>
    </div>
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <p class="text-xs font-medium text-slate-500">Total Registered Users</p>
        <p class="mt-1.5 text-xl font-bold text-slate-900 tabular-nums">
          {{ adminDash?.metrics.totalUsers.toLocaleString() }}
        </p>
      </div>
      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <p class="text-xs font-medium text-slate-500">Active Viewers Today (DAU)</p>
        <p class="mt-1.5 text-xl font-bold text-slate-900 tabular-nums">
          {{ adminDash?.metrics.activeUsersToday.toLocaleString() }}
        </p>
      </div>
      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <p class="text-xs font-medium text-slate-500">Active Live Streams</p>
        <p class="mt-1.5 text-xl font-bold text-rose-600 tabular-nums">
          {{ adminDash?.metrics.liveStreamsActive.toLocaleString() }}
        </p>
      </div>
      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <p class="text-xs font-medium text-slate-500">Gross Ad + Sub Revenue</p>
        <p class="mt-1.5 text-xl font-bold text-emerald-700 tabular-nums">
          ₩{{ ((adminDash?.metrics.totalAdRevenueKrw ?? 0) / 100000000).toFixed(1) }}억
        </p>
      </div>
      <div class="rounded-lg border border-slate-200 bg-white p-4">
        <p class="text-xs font-medium text-slate-500">Open Moderation Reports</p>
        <p class="mt-1.5 text-xl font-bold text-amber-700 tabular-nums">{{ openModCount }} cases</p>
      </div>
    </div>

    <!-- Infrastructure Health Matrix & Audit Logs -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <div class="rounded-lg border border-dashed border-slate-300 bg-slate-50 p-8 text-center lg:col-span-2">
        <h2 class="text-sm font-semibold text-slate-600">
          핵심 인프라 클러스터 상태 (Core Infrastructure Health)
        </h2>
        <p class="mt-1 text-xs text-slate-500">관측 스택(Prometheus/Grafana) 연동 준비 중입니다.</p>
      </div>

      <!-- Recent Immutable Audit Logs -->
      <div class="overflow-hidden rounded-lg border border-slate-200 bg-white">
        <div class="border-b border-slate-200 bg-slate-50 px-5 py-3.5">
          <h2 class="text-sm font-semibold text-slate-900">
            실시간 관리자 감사 로그 (Immutable Audit Trail)
          </h2>
        </div>
        <p v-if="(adminDash?.recentAuditLogs || []).length === 0" class="p-5 text-center text-xs text-slate-500">
          기록된 감사 로그가 없습니다.
        </p>
        <div v-else class="divide-y divide-slate-100">
          <div
            v-for="log in adminDash?.recentAuditLogs"
            :key="log.id"
            class="p-4 hover:bg-slate-50 transition-colors"
          >
            <p class="font-mono text-xs font-semibold text-slate-900">{{ log.action }}</p>
            <p class="mt-1 text-xs text-slate-600">Actor: <span class="font-mono text-slate-900">{{ log.actor }}</span></p>
            <p class="mt-0.5 font-mono text-xs text-slate-400 tabular-nums">{{ log.timestamp }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
