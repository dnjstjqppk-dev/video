<script setup lang="ts">
useSeoMeta({
  title: 'MCN Console — Managed Channels, Revenue Share & Access Revocation'
})

interface ManagedChannelContract {
  channelId: string
  channelName: string
  handle: string
  subscribers: number
  creatorSplitPct: number
  mcnSplitPct: number
  monthlyGrossKrw: number
  contractExpiresAt: string
  status: 'ACTIVE' | 'REVOKED'
}

const roster = ref<ManagedChannelContract[]>([
  {
    channelId: 'ch_tech_kr',
    channelName: 'GVP Tech Engineering',
    handle: '@gvptech',
    subscribers: 284000,
    creatorSplitPct: 85,
    mcnSplitPct: 15,
    monthlyGrossKrw: 18450000,
    contractExpiresAt: '2027-06-30',
    status: 'ACTIVE'
  },
  {
    channelId: 'ch_seoul_4k',
    channelName: 'Seoul 4K Cinema',
    handle: '@seoul4k',
    subscribers: 890000,
    creatorSplitPct: 80,
    mcnSplitPct: 20,
    monthlyGrossKrw: 42100000,
    contractExpiresAt: '2027-03-31',
    status: 'ACTIVE'
  },
  {
    channelId: 'ch_esports_global',
    channelName: 'Global Esports Arena',
    handle: '@esports',
    subscribers: 2450000,
    creatorSplitPct: 90,
    mcnSplitPct: 10,
    monthlyGrossKrw: 96400000,
    contractExpiresAt: '2026-12-31',
    status: 'ACTIVE'
  }
])

function revokeChannelContract(ch: ManagedChannelContract) {
  ch.status = 'REVOKED'
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-white">MCN 소속 크리에이터 및 계약 관리 (MCN Roster Console)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Automatic revenue split routing in the double-entry settlement ledger · Immediate access revocation on contract termination
    </p>

    <div class="mt-6 overflow-x-auto rounded-2xl border border-white/10 bg-[#141A2C]">
      <table class="w-full text-left text-xs">
        <thead class="border-b border-white/10 bg-white/5 uppercase text-neutral-400">
          <tr>
            <th class="p-3.5">Channel</th>
            <th class="p-3.5">Subscribers</th>
            <th class="p-3.5">Contract Split (Creator : MCN)</th>
            <th class="p-3.5">Monthly Gross</th>
            <th class="p-3.5">MCN Commission</th>
            <th class="p-3.5">Contract Expiration</th>
            <th class="p-3.5">Status / Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/10">
          <tr v-for="ch in roster" :key="ch.channelId" class="hover:bg-white/5">
            <td class="p-3.5">
              <p class="font-bold text-white">{{ ch.channelName }}</p>
              <p class="text-[11px] text-neutral-400">{{ ch.handle }}</p>
            </td>
            <td class="p-3.5 font-mono">{{ ch.subscribers.toLocaleString() }}</td>
            <td class="p-3.5">
              <span class="rounded bg-[#3EA6FF]/20 px-2 py-0.5 font-bold text-[#3EA6FF]">
                {{ ch.creatorSplitPct }}% Creator / {{ ch.mcnSplitPct }}% MCN
              </span>
            </td>
            <td class="p-3.5 font-mono">₩{{ ch.monthlyGrossKrw.toLocaleString() }}</td>
            <td class="p-3.5 font-mono font-bold text-emerald-400">
              ₩{{ Math.round((ch.monthlyGrossKrw * ch.mcnSplitPct) / 100).toLocaleString() }}
            </td>
            <td class="p-3.5">{{ ch.contractExpiresAt }}</td>
            <td class="p-3.5">
              <div class="flex items-center gap-2">
                <span
                  class="rounded px-2 py-0.5 text-[10px] font-bold"
                  :class="ch.status === 'ACTIVE' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-red-500/20 text-red-300'"
                >
                  {{ ch.status }}
                </span>
                <button
                  v-if="ch.status === 'ACTIVE'"
                  type="button"
                  class="rounded-lg bg-red-500/20 px-2.5 py-1 text-[11px] font-bold text-red-300 hover:bg-red-500/30"
                  @click="revokeChannelContract(ch)"
                >
                  Revoke Contract Immediately
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
