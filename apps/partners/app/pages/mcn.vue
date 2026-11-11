<script setup lang="ts">
import type { PartnerContract } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import GvpModal from '@gvp/ui/src/components/GvpModal.vue'
import GvpTextInput from '@gvp/ui/src/components/GvpTextInput.vue'

useSeoMeta({
  title: 'MCN Console — Managed Channels, Revenue Share & Access Revocation'
})

const api = usePartnerApi()

const { data, pending, error, refresh } = await useAsyncData('partner-contracts', async () => {
  return await api.getPartnerContracts()
})

const contracts = computed(() => data.value ?? [])
const revokeTarget = ref<PartnerContract | null>(null)
const revokeReason = ref('')
const revokeReasonError = ref('')
const revoking = ref(false)
const actionError = ref<string | null>(null)

function openRevoke(contract: PartnerContract) {
  revokeTarget.value = contract
  revokeReason.value = ''
  revokeReasonError.value = ''
  actionError.value = null
}

async function confirmRevoke() {
  if (!revokeTarget.value) return
  if (!revokeReason.value.trim()) {
    revokeReasonError.value = '해지 사유를 입력해야 합니다.'
    return
  }
  revoking.value = true
  actionError.value = null
  try {
    const updated = await api.revokePartnerContract(revokeTarget.value.id, revokeReason.value.trim())
    revokeTarget.value.status = updated.status
    revokeTarget.value.revokedAt = updated.revokedAt
    revokeTarget.value = null
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '계약 해지에 실패했습니다.'
  } finally {
    revoking.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-slate-200 pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-slate-900">MCN 소속 크리에이터 및 계약 관리 (MCN Roster Console)</h1>
      <p class="mt-1 text-sm text-slate-600">
        Automatic revenue split routing in the double-entry settlement ledger · Immediate access revocation on contract termination
      </p>
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <div v-if="pending" class="space-y-2" role="status" aria-label="불러오는 중">
      <div v-for="n in 4" :key="n" class="h-14 animate-pulse rounded-lg border border-slate-200 bg-slate-100" />
    </div>

    <GvpBanner v-else-if="error" tone="danger">
      계약 목록을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <GvpEmptyState
      v-else-if="contracts.length === 0"
      title="등록된 MCN 계약이 없습니다"
      description="새 채널 계약이 체결되면 이 목록에 표시됩니다."
    />

    <div v-else class="overflow-x-auto rounded-lg border border-slate-200 bg-white">
      <table class="w-full text-left">
        <thead>
          <tr class="bg-slate-50 border-y border-slate-200 text-xs font-semibold text-slate-600 uppercase tracking-wider">
            <th scope="col" class="py-2.5 px-4">Channel</th>
            <th scope="col" class="py-2.5 px-4">Contract Split (Creator : MCN)</th>
            <th scope="col" class="py-2.5 px-4">Valid Until</th>
            <th scope="col" class="py-2.5 px-4">Status / Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <tr v-for="ch in contracts" :key="ch.id" class="hover:bg-slate-50">
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100">
              <p class="font-medium text-slate-900">{{ ch.channelName }}</p>
              <p class="text-xs text-slate-500 tabular-nums">{{ ch.channelHandle }} · {{ ch.id }}</p>
            </td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              <span class="rounded-md bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-800 tabular-nums">
                {{ ch.creatorSplitPct }}% Creator / {{ ch.partnerSplitPct }}% MCN
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">{{ ch.validUntil }}</td>
            <td class="py-3 px-4 text-sm text-slate-900 border-b border-slate-100 tabular-nums">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  class="rounded-md px-2 py-0.5 text-xs font-medium uppercase"
                  :class="ch.status === 'active' ? 'border border-emerald-200 bg-emerald-50 text-emerald-700' : 'border border-red-200 bg-red-50 text-red-700'"
                >
                  {{ ch.status }}
                </span>
                <button
                  v-if="ch.status === 'active'"
                  type="button"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
                  @click="openRevoke(ch)"
                >
                  Revoke Contract Immediately
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <GvpModal :model-value="revokeTarget !== null" title="계약 즉시 해지" @update:model-value="(v: boolean) => { if (!v) revokeTarget = null }">
      <p class="text-sm text-slate-600">
        <strong class="font-semibold text-slate-900">{{ revokeTarget?.channelName }}</strong> ({{ revokeTarget?.id }}) 계약을 해지합니다.
        해지 즉시 정산 분배가 중단되며, 이 작업은 감사 로그에 기록됩니다.
      </p>
      <div class="mt-4">
        <GvpTextInput
          v-model="revokeReason"
          input-id="mcn-revoke-reason"
          label="해지 사유 (필수)"
          placeholder="예: 크리에이터 요청에 의한 합의 해지"
          :error="revokeReasonError"
        />
      </div>
      <template #footer>
        <button
          type="button"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
          @click="revokeTarget = null"
        >
          취소
        </button>
        <button
          type="button"
          :disabled="revoking"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors disabled:opacity-50"
          @click="confirmRevoke"
        >
          {{ revoking ? '해지 중…' : '계약 해지 확정' }}
        </button>
      </template>
    </GvpModal>
  </div>
</template>
