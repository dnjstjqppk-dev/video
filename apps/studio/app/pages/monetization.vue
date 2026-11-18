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
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-[var(--gvp-border)] pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">수익 창출 및 복식부기 정산 (Partner Program &amp; Monetization)</h1>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          Transparent double-entry ledger accounting · Tax withholding verification · Audited payout requests
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button
          type="button"
          :class="
            currency === 'KRW'
              ? 'inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors'
              : 'inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors'
          "
          @click="currency = 'KRW'"
        >
          KRW
        </button>
        <button
          type="button"
          :class="
            currency === 'USD'
              ? 'inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors'
              : 'inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors'
          "
          @click="currency = 'USD'"
        >
          USD
        </button>
      </div>
    </div>

    <!-- Earnings Overview -->
    <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-semibold text-[var(--gvp-text)]">이번 달 예상 수익 (Estimated Earnings)</h2>
        <button
          type="button"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
          @click="refreshEarnings()"
        >
          새로고침
        </button>
      </div>
      <div v-if="earningsPending" class="mt-4 h-10 w-48 animate-pulse rounded-md bg-[var(--gvp-elevated)]" role="status" aria-label="불러오는 중" />
      <GvpBanner v-else-if="earningsError" tone="danger" class="mt-4">
        수익 정보를 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refreshEarnings()">다시 시도</button>
      </GvpBanner>
      <div v-else-if="earnings">
        <p class="mt-2 text-3xl font-semibold text-[var(--gvp-text)] tabular-nums">{{ formatMoney(earnings.currentMonthEstimated) }}</p>
        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
          <div v-for="item in revenueStreams" :key="item.stream" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4">
            <div class="flex items-center justify-between gap-2 text-xs">
              <span class="font-medium text-[var(--gvp-text)]">{{ item.stream }}</span>
              <span class="font-semibold text-[var(--gvp-text)] tabular-nums">{{ formatMoney(item.amount) }}</span>
            </div>
            <div class="mt-2 h-2 w-full overflow-hidden rounded-md bg-[var(--gvp-elevated)]">
              <div class="h-full bg-[var(--gvp-brand)]" :style="{ width: `${item.pct}%` }" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Payout Ledger Table & Payout Request -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <div class="overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] lg:col-span-2">
        <div class="border-b border-[var(--gvp-border)] px-5 py-4">
          <h2 class="text-sm font-semibold text-[var(--gvp-text)]">월별 정산 원장 내역 (Monthly Payout Statements)</h2>
        </div>
        <div v-if="settlementsPending" class="p-5 space-y-2" role="status" aria-label="불러오는 중">
          <div v-for="n in 3" :key="n" class="h-12 animate-pulse rounded-md bg-[var(--gvp-elevated)]" />
        </div>
        <div v-else-if="settlementsError" class="p-5">
          <GvpBanner tone="danger">
            정산 내역을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refreshSettlements()">다시 시도</button>
          </GvpBanner>
        </div>
        <div v-else-if="!settlements || settlements.length === 0" class="p-5">
          <GvpEmptyState
            title="정산 내역이 없습니다"
            description="수익이 발생하면 월별 정산 명세가 여기에 표시됩니다."
          />
        </div>
        <table v-else class="w-full text-left">
          <thead class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider">
            <tr>
              <th scope="col" class="py-2.5 px-4">Statement ID</th>
              <th scope="col" class="py-2.5 px-4">Period</th>
              <th scope="col" class="py-2.5 px-4">Gross</th>
              <th scope="col" class="py-2.5 px-4">Tax Withheld</th>
              <th scope="col" class="py-2.5 px-4">Net Payout</th>
              <th scope="col" class="py-2.5 px-4">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--gvp-border)]">
            <tr v-for="p in settlements" :key="p.id" class="hover:bg-[var(--gvp-canvas)] transition-colors">
              <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums font-mono">{{ p.id }}</td>
              <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums">{{ p.periodStart }} ~ {{ p.periodEnd }}</td>
              <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">{{ formatMoney(p.amount) }}</td>
              <td class="py-3 px-4 text-sm text-[var(--gvp-live)] border-b border-[var(--gvp-border)] tabular-nums">-{{ formatMoney(p.taxWithheld) }}</td>
              <td class="py-3 px-4 text-sm font-semibold text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">{{ formatMoney(p.netAmount) }}</td>
              <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
                <span
                  class="inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium uppercase"
                  :class="String(p.status).toLowerCase() === 'paid' ? 'border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]' : 'border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] text-[var(--gvp-warning)]'"
                >
                  {{ p.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Payout Request -->
      <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5">
        <h2 class="text-sm font-semibold text-[var(--gvp-text)]">정산금 출금 요청 (Request Payout)</h2>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          요청은 감사 로그에 기록되며, 재무 운영자가 승인 후 등록 계좌로 송금합니다.
        </p>
        <GvpBanner v-if="payoutResult" tone="success" class="mt-3">{{ payoutResult }}</GvpBanner>
        <GvpBanner v-if="payoutError" tone="danger" class="mt-3">{{ payoutError }}</GvpBanner>
        <button
          type="button"
          :disabled="requesting"
          class="mt-4 w-full inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors disabled:opacity-50"
          @click="requestPayoutNow"
        >
          {{ requesting ? '요청 중…' : '출금 요청하기' }}
        </button>
        <NuxtLink to="/copyright" class="mt-3 block text-center text-xs font-medium text-[var(--gvp-text-secondary)] underline hover:text-[var(--gvp-text)]">
          정산·저작권 이의가 있으신가요? 권리 보호 센터로 이동
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
