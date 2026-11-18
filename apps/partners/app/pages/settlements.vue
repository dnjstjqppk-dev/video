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
    <div class="border-b border-[var(--gvp-border)] pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">복식부기 정산 명세서 및 세금계산서 (Partner Settlements)</h1>
      <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
        Immutable double-entry ledger reconciliation (<code class="font-mono text-xs text-[var(--gvp-text)]">SUM(debit) == SUM(credit)</code>) · Downloadable VAT &amp; withholding invoices
      </p>
    </div>

    <div v-if="settlementsPending" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-8">
      <div class="h-6 animate-pulse rounded-md bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="settlementsError" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">정산 명세서를 불러오지 못했습니다.</p>
    </div>
    <GvpEmptyState
      v-else-if="(settlements || []).length === 0"
      title="정산 명세가 없습니다"
      description="정산 주기가 도래하면 여기에 표시됩니다."
    />

    <div v-else class="overflow-x-auto rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <table class="w-full text-left">
        <thead>
          <tr class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider">
            <th class="py-2.5 px-4">Statement ID</th>
            <th class="py-2.5 px-4">Billing Period</th>
            <th class="py-2.5 px-4">Gross Ledger Credit</th>
            <th class="py-2.5 px-4">Tax / VAT Withheld</th>
            <th class="py-2.5 px-4">Net Settlement</th>
            <th class="py-2.5 px-4">Status</th>
            <th class="py-2.5 px-4">Invoice</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--gvp-border)]">
          <tr v-for="stl in (settlements || [])" :key="stl.id" class="hover:bg-[var(--gvp-canvas)]">
            <td class="py-3 px-4 font-mono text-sm font-medium text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">{{ stl.id }}</td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text-secondary)] border-b border-[var(--gvp-border)] tabular-nums">{{ stl.periodStart }} ~ {{ stl.periodEnd }}</td>
            <td class="py-3 px-4 font-mono text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">₩{{ stl.amount.toLocaleString() }}</td>
            <td class="py-3 px-4 font-mono text-sm text-[var(--gvp-live)] border-b border-[var(--gvp-border)] tabular-nums">-₩{{ stl.taxWithheld.toLocaleString() }}</td>
            <td class="py-3 px-4 font-mono text-sm font-semibold text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">₩{{ stl.netAmount.toLocaleString() }}</td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <span class="rounded-md border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] px-2 py-0.5 text-xs font-medium uppercase text-[var(--gvp-success)]">
                {{ stl.status }}
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <button
                type="button"
                class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
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
