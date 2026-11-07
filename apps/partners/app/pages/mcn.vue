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
  <div>
    <h1 class="text-2xl font-extrabold text-white">MCN 소속 크리에이터 및 계약 관리 (MCN Roster Console)</h1>
    <p class="mt-1 text-xs text-neutral-400">
      Automatic revenue split routing in the double-entry settlement ledger · Immediate access revocation on contract termination
    </p>

    <GvpBanner v-if="actionError" tone="danger" dismissible class="mt-4" @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <div v-if="pending" class="mt-6 space-y-2" role="status" aria-label="불러오는 중">
      <div v-for="n in 4" :key="n" class="h-14 animate-pulse rounded-xl bg-white/5" />
    </div>

    <GvpBanner v-else-if="error" tone="danger" class="mt-6">
      계약 목록을 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <GvpEmptyState
      v-else-if="contracts.length === 0"
      class="mt-6"
      title="등록된 MCN 계약이 없습니다"
      description="새 채널 계약이 체결되면 이 목록에 표시됩니다."
    />

    <div v-else class="mt-6 overflow-x-auto rounded-2xl border border-white/10 bg-[#141A2C]">
      <table class="w-full text-left text-xs">
        <thead class="border-b border-white/10 bg-white/5 uppercase text-neutral-400">
          <tr>
            <th scope="col" class="p-3.5">Channel</th>
            <th scope="col" class="p-3.5">Contract Split (Creator : MCN)</th>
            <th scope="col" class="p-3.5">Valid Until</th>
            <th scope="col" class="p-3.5">Status / Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/10">
          <tr v-for="ch in contracts" :key="ch.id" class="hover:bg-white/5">
            <td class="p-3.5">
              <p class="font-bold text-white">{{ ch.channelName }}</p>
              <p class="text-[11px] text-neutral-400">{{ ch.channelHandle }} · {{ ch.id }}</p>
            </td>
            <td class="p-3.5">
              <span class="rounded bg-[#3EA6FF]/20 px-2 py-0.5 font-bold text-[#3EA6FF]">
                {{ ch.creatorSplitPct }}% Creator / {{ ch.partnerSplitPct }}% MCN
              </span>
            </td>
            <td class="p-3.5">{{ ch.validUntil }}</td>
            <td class="p-3.5">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  class="rounded px-2 py-0.5 text-[10px] font-bold uppercase"
                  :class="ch.status === 'active' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-red-500/20 text-red-300'"
                >
                  {{ ch.status }}
                </span>
                <button
                  v-if="ch.status === 'active'"
                  type="button"
                  class="rounded-lg bg-red-500/20 px-2.5 py-1 text-[11px] font-bold text-red-300 hover:bg-red-500/30"
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
      <p class="text-sm text-neutral-300">
        <strong class="text-white">{{ revokeTarget?.channelName }}</strong> ({{ revokeTarget?.id }}) 계약을 해지합니다.
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
          class="rounded-full bg-white/10 px-5 py-2 text-xs font-bold text-neutral-200 hover:bg-white/20"
          @click="revokeTarget = null"
        >
          취소
        </button>
        <button
          type="button"
          :disabled="revoking"
          class="rounded-full bg-[#FF0033] px-5 py-2 text-xs font-bold text-white hover:bg-[#D9002B] disabled:opacity-50"
          @click="confirmRevoke"
        >
          {{ revoking ? '해지 중…' : '계약 해지 확정' }}
        </button>
      </template>
    </GvpModal>
  </div>
</template>
