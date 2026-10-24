<script setup lang="ts">
import { PayoutStatus } from '@gvp/shared-types'

const currency = useState<'KRW' | 'USD'>('studio-currency', () => 'KRW')

useSeoMeta({
  title: 'Earn & Monetization — Double-Entry Ledger Settlements'
})

const disputeReason = ref('')
const disputeSubmitted = ref(false)

const revenueBreakdown = [
  { stream: 'Watch Page Pre-Roll & Mid-Roll Ads (55% Creator Share)', krw: 11250000, pct: 61 },
  { stream: 'StreamVerse Premium Watch-Time Pool (55% Share)', krw: 3840000, pct: 21 },
  { stream: 'Channel Memberships & Perks (70% Creator Share)', krw: 2160000, pct: 12 },
  { stream: 'Live SuperChat & Donations (70% Creator Share)', krw: 1200000, pct: 6 }
]

const payoutHistory = [
  {
    id: 'pay_2026_09',
    period: '2026-09-01 ~ 2026-09-30',
    grossKrw: 16800000,
    taxWithheldKrw: 554400,
    netKrw: 16245600,
    status: PayoutStatus.PAID,
    paidAt: '2026-10-15'
  },
  {
    id: 'pay_2026_10',
    period: '2026-10-01 ~ 2026-10-31',
    grossKrw: 18450000,
    taxWithheldKrw: 608850,
    netKrw: 17841150,
    status: PayoutStatus.APPROVED,
    paidAt: 'Scheduled 2026-11-15'
  }
]

function formatMoney(krw: number): string {
  if (currency.value === 'KRW') return `₩${krw.toLocaleString('ko-KR')}`
  return `$${(krw / 1350).toFixed(2)}`
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-extrabold text-white">수익 창출 및 복식부기 정산 (Partner Program & Monetization)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Transparent double-entry ledger accounting · Tax withholding verification · Settlement dispute workflow
    </p>

    <!-- Eligibility & Tax Status Cards -->
    <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
      <div class="rounded-2xl border border-emerald-500/30 bg-[#151822] p-5">
        <span class="rounded-full bg-emerald-500/20 px-3 py-1 text-xs font-bold text-emerald-400">
          ✓ StreamVerse Partner Program Active
        </span>
        <h2 class="mt-3 text-base font-bold text-white">Eligibility Requirements Met</h2>
        <div class="mt-3 space-y-2 text-xs text-neutral-300">
          <p>✓ Subscribers: <strong>284,000</strong> / 1,000 required (100%)</p>
          <p>✓ Valid Public Watch Hours (12m): <strong>148,520 hrs</strong> / 4,000 required (100%)</p>
          <p>✓ 2-Step Verification (TOTP + WebAuthn) & Clean Policy Record</p>
        </div>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="text-base font-bold text-white">세금 및 지급 계좌 정보 (Tax & Payout Profile)</h2>
        <div class="mt-3 space-y-1.5 text-xs text-neutral-300">
          <p>Tax Residency: <strong class="text-white">Republic of Korea (KR-NTS Verified)</strong></p>
          <p>US W-8BEN Treaty Status: <strong class="text-emerald-400">Approved (10% Treaty Rate)</strong></p>
          <p>Settlement Bank: <strong class="text-white">Shinhan Bank KRW / Wise Multi-Currency USD</strong></p>
        </div>
      </div>
    </div>

    <!-- Revenue Split Breakdown -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#151822] p-5">
      <h2 class="text-sm font-bold text-white">수익원별 정산 내역 (Revenue Split by Stream)</h2>
      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
        <div
          v-for="item in revenueBreakdown"
          :key="item.stream"
          class="rounded-xl bg-white/5 p-4"
        >
          <div class="flex items-center justify-between text-xs">
            <span class="font-semibold text-neutral-200">{{ item.stream }}</span>
            <span class="font-extrabold text-emerald-400">{{ formatMoney(item.krw) }}</span>
          </div>
          <div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-white/10">
            <div class="h-full bg-[#3EA6FF]" :style="{ width: `${item.pct}%` }" />
          </div>
        </div>
      </div>
    </div>

    <!-- Payout Ledger Table & Dispute Filing -->
    <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5 lg:col-span-2">
        <h2 class="text-sm font-bold text-white">월별 정산 원장 내역 (Monthly Payout Statements)</h2>
        <table class="mt-4 w-full text-left text-xs">
          <thead class="border-b border-white/10 text-neutral-400">
            <tr>
              <th class="py-2">Statement ID</th>
              <th class="py-2">Period</th>
              <th class="py-2">Gross</th>
              <th class="py-2">Tax Withheld</th>
              <th class="py-2">Net Payout</th>
              <th class="py-2">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/10">
            <tr v-for="p in payoutHistory" :key="p.id">
              <td class="py-3 font-mono text-[#3EA6FF]">{{ p.id }}</td>
              <td class="py-3">{{ p.period }}</td>
              <td class="py-3">{{ formatMoney(p.grossKrw) }}</td>
              <td class="py-3 text-red-300">-{{ formatMoney(p.taxWithheldKrw) }}</td>
              <td class="py-3 font-bold text-emerald-400">{{ formatMoney(p.netKrw) }}</td>
              <td class="py-3">
                <span class="rounded bg-emerald-500/20 px-2 py-0.5 text-[10px] font-bold uppercase text-emerald-300">
                  {{ p.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Settlement / Copyright Dispute Form -->
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="text-sm font-bold text-white">정산 및 저작권 이의 신청 (File Settlement Dispute)</h2>
        <p class="mt-1 text-xs text-neutral-400">
          Escalate ledger discrepancy or Content ID revenue-share hold to Finance & Rights Admin.
        </p>
        <textarea
          v-model="disputeReason"
          rows="4"
          placeholder="Enter statement ID or claim ID and supporting license details..."
          class="mt-3 w-full rounded-xl border border-white/15 bg-[#0D0F14] p-3 text-xs text-white outline-none"
        />
        <p v-if="disputeSubmitted" class="mt-2 text-xs font-bold text-emerald-400">
          ✓ Dispute ticket #DSP-2026-884 logged on immutable ledger audit trail.
        </p>
        <button
          type="button"
          class="mt-3 w-full rounded-xl bg-[#FF0033] py-2 text-xs font-bold text-white hover:bg-[#D9002B]"
          @click="disputeSubmitted = true"
        >
          Submit Formal Dispute
        </button>
      </div>
    </div>
  </div>
</template>
