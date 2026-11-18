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
    <div class="border-b border-[var(--gvp-border)] pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-[var(--gvp-text)]">
        플랫폼 통합 운영 관제 (Platform Operations Dashboard)
      </h1>
      <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
        Real-time global platform KPIs, transcoding queue depth, gross revenue, and 7-tier infrastructure health
      </p>
    </div>

    <!-- Top KPI Grid -->
    <div v-if="dashPending" class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
      <div v-for="i in 5" :key="i" class="h-24 animate-pulse rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="dashError || !adminDash" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">운영 지표를 불러오지 못했습니다.</p>
    </div>
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <p class="text-xs font-medium text-[var(--gvp-text-secondary)]">Total Registered Users</p>
        <p class="mt-1.5 text-xl font-bold text-[var(--gvp-text)] tabular-nums">
          {{ adminDash?.metrics.totalUsers.toLocaleString() }}
        </p>
      </div>
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <p class="text-xs font-medium text-[var(--gvp-text-secondary)]">Active Viewers Today (DAU)</p>
        <p class="mt-1.5 text-xl font-bold text-[var(--gvp-text)] tabular-nums">
          {{ adminDash?.metrics.activeUsersToday.toLocaleString() }}
        </p>
      </div>
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <p class="text-xs font-medium text-[var(--gvp-text-secondary)]">Active Live Streams</p>
        <p class="mt-1.5 text-xl font-bold text-[var(--gvp-live)] tabular-nums">
          {{ adminDash?.metrics.liveStreamsActive.toLocaleString() }}
        </p>
      </div>
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <p class="text-xs font-medium text-[var(--gvp-text-secondary)]">Gross Ad + Sub Revenue</p>
        <p class="mt-1.5 text-xl font-bold text-[var(--gvp-success)] tabular-nums">
          ₩{{ ((adminDash?.metrics.totalAdRevenueKrw ?? 0) / 100000000).toFixed(1) }}억
        </p>
      </div>
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
        <p class="text-xs font-medium text-[var(--gvp-text-secondary)]">Open Moderation Reports</p>
        <p class="mt-1.5 text-xl font-bold text-[var(--gvp-warning)] tabular-nums">{{ openModCount }} cases</p>
      </div>
    </div>

    <!-- Infrastructure Health Matrix & Audit Logs -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <div class="rounded-lg border border-dashed border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] p-8 text-center lg:col-span-2">
        <h2 class="text-sm font-semibold text-[var(--gvp-text-secondary)]">
          핵심 인프라 클러스터 상태 (Core Infrastructure Health)
        </h2>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">관측 스택(Prometheus/Grafana) 연동 준비 중입니다.</p>
      </div>

      <!-- Recent Immutable Audit Logs -->
      <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
        <div class="border-b border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-5 py-3.5">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">
            실시간 관리자 감사 로그 (Immutable Audit Trail)
          </h2>
        </div>
        <p v-if="(adminDash?.recentAuditLogs || []).length === 0" class="p-5 text-center text-xs text-[var(--gvp-text-secondary)]">
          기록된 감사 로그가 없습니다.
        </p>
        <div v-else class="divide-y divide-[var(--gvp-border)]">
          <div
            v-for="log in adminDash?.recentAuditLogs"
            :key="log.id"
            class="p-4 hover:bg-[var(--gvp-canvas)] transition-colors"
          >
            <p class="font-mono text-xs font-semibold text-[var(--gvp-text)]">{{ log.action }}</p>
            <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">Actor: <span class="font-mono text-[var(--gvp-text)]">{{ log.actor }}</span></p>
            <p class="mt-0.5 font-mono text-xs text-[var(--gvp-text-muted)] tabular-nums">{{ log.timestamp }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
