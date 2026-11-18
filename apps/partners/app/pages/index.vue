<script setup lang="ts">
const api = usePartnerApi()
const activeOrg = useState<{ id: string; name: string; type: string }>('partner-org', () => ({
  id: '',
  name: '',
  type: ''
}))

useSeoMeta({
  title: 'Partner Organization Overview'
})

const { data: overview, pending: overviewPending, error: overviewError } = await useAsyncData('partner-overview', async () => {
  const res = await api.getPartnerOverview()
  activeOrg.value = { ...res.organization }
  return res
})
</script>

<template>
  <div class="space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-[var(--gvp-border)] pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">
          파트너 조직 종합 대시보드 ({{ overview?.organization?.name || 'Partner' }})
        </h1>
        <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
          Multi-tenant RBAC isolation · Real-time ad pacing, MCN channel roster, Content ID claims, and ledger settlements
        </p>
      </div>
    </div>

    <!-- KPI Grid -->
    <div v-if="overviewPending" class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div v-for="i in 3" :key="i" class="h-28 animate-pulse rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="overviewError || !overview" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">조직 개요를 불러오지 못했습니다.</p>
    </div>
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <p class="text-xs font-medium uppercase tracking-wider text-[var(--gvp-text-secondary)]">Active Ad Campaigns</p>
        <p class="mt-2 text-2xl font-semibold text-[var(--gvp-text)] tabular-nums">{{ overview?.campaignCount ?? 0 }}</p>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">Pre-roll, Mid-roll &amp; Shorts Feed</p>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <p class="text-xs font-medium uppercase tracking-wider text-[var(--gvp-text-secondary)]">Managed MCN Channels</p>
        <p class="mt-2 text-2xl font-semibold text-[var(--gvp-text)] tabular-nums">{{ overview?.activeChannelsCount ?? 0 }}</p>
      </div>

      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <p class="text-xs font-medium uppercase tracking-wider text-[var(--gvp-text-secondary)]">Pending Double-Entry Settlement</p>
        <p class="mt-2 text-2xl font-semibold text-[var(--gvp-text)] tabular-nums">
          ₩{{ (overview?.pendingSettlementAmount ?? 0).toLocaleString() }}
        </p>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)] tabular-nums">Next clearing: 15th of month</p>
      </div>
    </div>

    <!-- Quick Action Cards -->
    <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
      <NuxtLink
        to="/campaigns"
        class="group rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5 transition-colors hover:border-[var(--gvp-border-strong)] hover:bg-[var(--gvp-canvas)]"
      >
        <div class="flex h-8 w-8 items-center justify-center rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text)]">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
          </svg>
        </div>
        <h2 class="mt-3 text-sm font-semibold text-[var(--gvp-text)]">Advertiser Campaign Manager</h2>
        <p class="mt-1 text-xs leading-relaxed text-[var(--gvp-text-secondary)]">
          Launch pre-roll, mid-roll, and Shorts ad campaigns with CPM/CPC bidding and real-time conversion attribution.
        </p>
      </NuxtLink>

      <NuxtLink
        to="/mcn"
        class="group rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5 transition-colors hover:border-[var(--gvp-border-strong)] hover:bg-[var(--gvp-canvas)]"
      >
        <div class="flex h-8 w-8 items-center justify-center rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text)]">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </div>
        <h2 class="mt-3 text-sm font-semibold text-[var(--gvp-text)]">MCN Channel &amp; Contract Roster</h2>
        <p class="mt-1 text-xs leading-relaxed text-[var(--gvp-text-secondary)]">
          Configure creator-MCN split ratios, monitor contract expirations, and trigger immediate access revocation.
        </p>
      </NuxtLink>

      <NuxtLink
        to="/rights"
        class="group rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5 transition-colors hover:border-[var(--gvp-border-strong)] hover:bg-[var(--gvp-canvas)]"
      >
        <div class="flex h-8 w-8 items-center justify-center rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text)]">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </div>
        <h2 class="mt-3 text-sm font-semibold text-[var(--gvp-text)]">Rights Owner Content ID</h2>
        <p class="mt-1 text-xs leading-relaxed text-[var(--gvp-text-secondary)]">
          Register SHA-256 reference fingerprints, set territory policies (Monetize / Track / Block), and resolve disputes.
        </p>
      </NuxtLink>
    </div>
  </div>
</template>
