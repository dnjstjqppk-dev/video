<script setup lang="ts">
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import { useStudioApi } from '~/composables/useStudioApi'

const api = useStudioApi()
const currency = useState<'KRW' | 'USD'>('studio-currency', () => 'KRW')

useSeoMeta({
  title: 'Earn & Monetization — Double-Entry Ledger Settlements'
})

const { data: earnings, pending: earningsPending, error: earningsError, refresh: refreshEarnings } = await useAsyncData(
  'studio-earnings',
  async () => await api.getStudioEarnings()
)

const { data: settlements, pending: settlementsPending, error: settlementsError, refresh: refreshSettlements } = await useAsyncData(
  'studio-settlements',
  async () => await api.getStudioSettlements()
)

const requesting = ref(false)
const payoutResult = ref<string | null>(null)
const payoutError = ref<string | null>(null)

const revenueStreams = computed(() => {
  if (!earnings.value) return []
  const total = earnings.value.currentMonthEstimated || 1
  return [
    { stream: 'Watch Page Ads (55% Creator Share)', amount: earnings.value.adRevenue, pct: Math.round((earnings.value.adRevenue / total) * 100) },
    { stream: 'Premium Watch-Time Pool (55% Share)', amount: earnings.value.membershipRevenue, pct: Math.round((earnings.value.membershipRevenue / total) * 100) },
    { stream: 'Live SuperChat & Donations (70% Creator Share)', amount: earnings.value.superChatRevenue, pct: Math.round((earnings.value.superChatRevenue / total) * 100) },
  ]
})

async function requestPayoutNow() {
  requesting.value = true
  payoutResult.value = null
  payoutError.value = null
  try {
    const res = await api.requestPayout()
    payoutResult.value = `${res.message} (ID: ${res.payout_id}, 상태: ${res.status})`
    await refreshSettlements()
  } catch (e: unknown) {
    payoutError.value = e instanceof Error ? e.message : '출금 요청에 실패했습니다.'
  } finally {
    requesting.value = false
  }
}

function formatMoney(cents: number): string {
  const krw = cents / 100
  if (currency.value === 'KRW') return `₩${Math.round(krw).toLocaleString('ko-KR')}`
  return `$${(krw / 1350).toFixed(2)}`
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">수익 창출 및 복식부기 정산 (Partner Program & Monetization)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Transparent double-entry ledger accounting · Tax withholding verification · Audited payout requests
        </p>
      </div>
      <div class="flex items-center gap-2 text-xs">
        <button
          type="button"
          class="rounded-lg border px-3 py-1.5 font-bold"
          :class="currency === 'KRW' ? 'border-[#3EA6FF] text-[#3EA6FF]' : 'border-white/15 text-neutral-300'"
          @click="currency = 'KRW'"
        >
          KRW
        </button>
        <button
          type="button"
          class="rounded-lg border px-3 py-1.5 font-bold"
          :class="currency === 'USD' ? 'border-[#3EA6FF] text-[#3EA6FF]' : 'border-white/15 text-neutral-300'"
          @click="currency = 'USD'"
        >
          USD
        </button>
      </div>
    </div>

    <!-- Earnings Overview -->
    <div class="mt-6 rounded-2xl border border-white/10 bg-[#151822] p-5">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-bold text-white">이번 달 예상 수익 (Estimated Earnings)</h2>
        <button type="button" class="text-xs font-bold text-[#3EA6FF] hover:underline" @click="refreshEarnings()">새로고침</button>
      </div>
      <div v-if="earningsPending" class="mt-4 h-10 w-48 animate-pulse rounded-lg bg-white/5" role="status" aria-label="불러오는 중" />
      <GvpBanner v-else-if="earningsError" tone="danger" class="mt-4">
        수익 정보를 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refreshEarnings()">다시 시도</button>
      </GvpBanner>
      <div v-else-if="earnings">
        <p class="mt-2 text-3xl font-extrabold text-emerald-400">{{ formatMoney(earnings.currentMonthEstimated) }}</p>
        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
          <div v-for="item in revenueStreams" :key="item.stream" class="rounded-xl bg-white/5 p-4">
            <div class="flex items-center justify-between text-xs">
              <span class="font-semibold text-neutral-200">{{ item.stream }}</span>
              <span class="font-extrabold text-emerald-400">{{ formatMoney(item.amount) }}</span>
            </div>
            <div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-white/10">
              <div class="h-full bg-[#3EA6FF]" :style="{ width: `${item.pct}%` }" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Payout Ledger Table & Payout Request -->
    <div class="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5 lg:col-span-2">
        <h2 class="text-sm font-bold text-white">월별 정산 원장 내역 (Monthly Payout Statements)</h2>
        <div v-if="settlementsPending" class="mt-4 space-y-2" role="status" aria-label="불러오는 중">
          <div v-for="n in 3" :key="n" class="h-12 animate-pulse rounded-xl bg-white/5" />
        </div>
        <GvpBanner v-else-if="settlementsError" tone="danger" class="mt-4">
          정산 내역을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refreshSettlements()">다시 시도</button>
        </GvpBanner>
        <GvpEmptyState
          v-else-if="!settlements || settlements.length === 0"
          class="mt-4"
          title="정산 내역이 없습니다"
          description="수익이 발생하면 월별 정산 명세가 여기에 표시됩니다."
        />
        <table v-else class="mt-4 w-full text-left text-xs">
          <thead class="border-b border-white/10 text-neutral-400">
            <tr>
              <th scope="col" class="py-2">Statement ID</th>
              <th scope="col" class="py-2">Period</th>
              <th scope="col" class="py-2">Gross</th>
              <th scope="col" class="py-2">Tax Withheld</th>
              <th scope="col" class="py-2">Net Payout</th>
              <th scope="col" class="py-2">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/10">
            <tr v-for="p in settlements" :key="p.id">
              <td class="py-3 font-mono text-[#3EA6FF]">{{ p.id }}</td>
              <td class="py-3">{{ p.periodStart }} ~ {{ p.periodEnd }}</td>
              <td class="py-3">{{ formatMoney(p.amount) }}</td>
              <td class="py-3 text-red-300">-{{ formatMoney(p.taxWithheld) }}</td>
              <td class="py-3 font-bold text-emerald-400">{{ formatMoney(p.netAmount) }}</td>
              <td class="py-3">
                <span
                  class="rounded px-2 py-0.5 text-[10px] font-bold uppercase"
                  :class="String(p.status).toLowerCase() === 'paid' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-amber-500/20 text-amber-300'"
                >
                  {{ p.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Payout Request -->
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="text-sm font-bold text-white">정산금 출금 요청 (Request Payout)</h2>
        <p class="mt-1 text-xs text-neutral-400">
          요청은 감사 로그에 기록되며, 재무 운영자가 승인 후 등록 계좌로 송금합니다.
        </p>
        <GvpBanner v-if="payoutResult" tone="success" class="mt-3">{{ payoutResult }}</GvpBanner>
        <GvpBanner v-if="payoutError" tone="danger" class="mt-3">{{ payoutError }}</GvpBanner>
        <button
          type="button"
          :disabled="requesting"
          class="mt-3 w-full rounded-xl bg-[#FF0033] py-2 text-xs font-bold text-white hover:bg-[#D9002B] disabled:opacity-50"
          @click="requestPayoutNow"
        >
          {{ requesting ? '요청 중…' : '출금 요청하기' }}
        </button>
        <NuxtLink to="/copyright" class="mt-3 block text-center text-xs font-bold text-[#3EA6FF] hover:underline">
          정산·저작권 이의가 있으신가요? 권리 보호 센터로 이동
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
