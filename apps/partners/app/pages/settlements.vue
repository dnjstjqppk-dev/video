<script setup lang="ts">
import type { Payout } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

const api = usePartnerApi()

useSeoMeta({
  title: 'Partner Double-Entry Settlement Statements & Invoices'
})

const { data: settlements, pending: settlementsPending, error: settlementsError } = await useAsyncData('partner-settlements', async () => {
  return await api.getPartnerSettlements()
})

function downloadInvoice(statement: Payout) {
  const content = [
    `STREAMVERSE DOUBLE-ENTRY SETTLEMENT TAX INVOICE`,
    `Statement ID: ${statement.id}`,
    `Period: ${statement.periodStart} ~ ${statement.periodEnd}`,
    `Gross Amount: ${statement.amount.toLocaleString()} ${statement.currency}`,
    `Tax / VAT Withheld: ${statement.taxWithheld.toLocaleString()} ${statement.currency}`,
    `Net Settlement Payable: ${statement.netAmount.toLocaleString()} ${statement.currency}`,
    `Status: ${statement.status.toUpperCase()}`
  ].join('\n')

  const blob = new Blob([content], { type: 'text/plain;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `invoice_${statement.id}.txt`
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-slate-200 pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-slate-900">복식부기 정산 명세서 및 세금계산서 (Partner Settlements)</h1>
      <p class="mt-1 text-sm text-slate-600">
        Immutable double-entry ledger reconciliation (<code class="font-mono text-xs text-slate-700">SUM(debit) == SUM(credit)</code>) · Downloadable VAT &amp; withholding invoices
      </p>
    </div>

    <div v-if="settlementsPending" class="rounded-lg border border-slate-200 bg-white p-8">
      <div class="h-6 animate-pulse rounded-md bg-slate-100" />
    </div>
    <div v-else-if="settlementsError" class="rounded-lg border border-red-200 bg-red-50 p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-red-700">정산 명세서를 불러오지 못했습니다.</p>
    </div>
    <GvpEmptyState
      v-else-if="(settlements || []).length === 0"
      title="정산 명세가 없습니다"
      description="정산 주기가 도래하면 여기에 표시됩니다."
    />

    <div v-else class="overflow-x-auto rounded-lg border border-slate-200 bg-white">
      <table class="w-full text-left">
        <thead>
          <tr class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider">
            <th class="py-2.5 px-4">Statement ID</th>
            <th class="py-2.5 px-4">Billing Period</th>
            <th class="py-2.5 px-4">Gross Ledger Credit</th>
            <th class="py-2.5 px-4">Tax / VAT Withheld</th>
            <th class="py-2.5 px-4">Net Settlement</th>
            <th class="py-2.5 px-4">Status</th>
            <th class="py-2.5 px-4">Invoice</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <tr v-for="stl in (settlements || [])" :key="stl.id" class="hover:bg-slate-50">
            <td class="py-3 px-4 font-mono text-sm font-medium text-slate-900 border-b border-slate-100 tabular-nums">{{ stl.id }}</td>
            <td class="py-3 px-4 text-sm text-slate-600 border-b border-slate-100 tabular-nums">{{ stl.periodStart }} ~ {{ stl.periodEnd }}</td>
            <td class="py-3 px-4 font-mono text-sm text-slate-900 border-b border-slate-100 tabular-nums">₩{{ stl.amount.toLocaleString() }}</td>
            <td class="py-3 px-4 font-mono text-sm text-red-700 border-b border-slate-100 tabular-nums">-₩{{ stl.taxWithheld.toLocaleString() }}</td>
            <td class="py-3 px-4 font-mono text-sm font-semibold text-slate-900 border-b border-slate-100 tabular-nums">₩{{ stl.netAmount.toLocaleString() }}</td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              <span class="rounded-md border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-xs font-medium uppercase text-emerald-700">
                {{ stl.status }}
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              <button
                type="button"
                class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
                @click="downloadInvoice(stl)"
              >
                Download Invoice
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
