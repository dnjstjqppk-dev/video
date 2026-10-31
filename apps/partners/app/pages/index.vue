<script setup lang="ts">
const api = usePartnerApi()
const activeTenant = useState<'ADVERTISER' | 'MCN' | 'RIGHTS_OWNER'>('partner-tenant', () => 'ADVERTISER')

useSeoMeta({
  title: 'Partner Organization Overview'
})

const { data: overview } = await useAsyncData('partner-overview', async () => {
  try {
    return await api.getPartnerOverview()
  } catch {
    return {
      organization: {
        id: 'org_partner_01',
        name: 'Apex Global Media / Horizon MCN / Starlight Rights',
        type: activeTenant.value
      },
      campaignCount: 14,
      activeChannelsCount: 42,
      pendingSettlementAmount: 148500000
    }
  }
})
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">
          파트너 조직 종합 대시보드 ({{ activeTenant }} Console)
        </h1>
        <p class="mt-1 text-xs text-neutral-400">
          Multi-tenant RBAC isolation · Real-time ad pacing, MCN channel roster, Content ID claims, and ledger settlements
        </p>
      </div>
    </div>

    <!-- KPI Grid -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-5">
        <p class="text-xs text-neutral-400">Active Ad Campaigns</p>
        <p class="mt-2 text-2xl font-extrabold text-white">{{ overview?.campaignCount || 14 }}</p>
        <p class="mt-1 text-xs text-[#3EA6FF]">Pre-roll, Mid-roll & Shorts Feed</p>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-5">
        <p class="text-xs text-neutral-400">Managed MCN Channels</p>
        <p class="mt-2 text-2xl font-extrabold text-white">{{ overview?.activeChannelsCount || 42 }}</p>
        <p class="mt-1 text-xs text-emerald-400">12.4M combined subscribers</p>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-5">
        <p class="text-xs text-neutral-400">Active Reference Fingerprints</p>
        <p class="mt-2 text-2xl font-extrabold text-white">1,284 assets</p>
        <p class="mt-1 text-xs text-amber-300">19 active Content ID matches</p>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#141A2C] p-5">
        <p class="text-xs text-neutral-400">Pending Double-Entry Settlement</p>
        <p class="mt-2 text-2xl font-extrabold text-emerald-400">
          ₩{{ (overview?.pendingSettlementAmount || 148500000).toLocaleString() }}
        </p>
        <p class="mt-1 text-xs text-neutral-400">Next clearing: 15th of month</p>
      </div>
    </div>

    <!-- Quick Action Cards -->
    <div class="mt-8 grid grid-cols-1 gap-6 md:grid-cols-3">
      <NuxtLink to="/campaigns" class="rounded-2xl border border-white/10 bg-[#141A2C] p-5 transition hover:border-[#3EA6FF]">
        <span class="text-2xl">📣</span>
        <h2 class="mt-3 text-base font-bold text-white">Advertiser Campaign Manager</h2>
        <p class="mt-1 text-xs text-neutral-400">
          Launch pre-roll, mid-roll, and Shorts ad campaigns with CPM/CPC bidding and real-time conversion attribution.
        </p>
      </NuxtLink>

      <NuxtLink to="/mcn" class="rounded-2xl border border-white/10 bg-[#141A2C] p-5 transition hover:border-[#3EA6FF]">
        <span class="text-2xl">🤝</span>
        <h2 class="mt-3 text-base font-bold text-white">MCN Channel & Contract Roster</h2>
        <p class="mt-1 text-xs text-neutral-400">
          Configure creator-MCN split ratios, monitor contract expirations, and trigger immediate access revocation.
        </p>
      </NuxtLink>

      <NuxtLink to="/rights" class="rounded-2xl border border-white/10 bg-[#141A2C] p-5 transition hover:border-[#3EA6FF]">
        <span class="text-2xl">🛡️</span>
        <h2 class="mt-3 text-base font-bold text-white">Rights Owner Content ID</h2>
        <p class="mt-1 text-xs text-neutral-400">
          Register SHA-256 reference fingerprints, set territory policies (Monetize / Track / Block), and resolve disputes.
        </p>
      </NuxtLink>
    </div>
  </div>
</template>
