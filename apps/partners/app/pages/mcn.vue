<script setup lang="ts">
import type { PartnerContract, UserProfile } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import GvpModal from '@gvp/ui/src/components/GvpModal.vue'
import GvpTextInput from '@gvp/ui/src/components/GvpTextInput.vue'

useSeoMeta({
  title: 'MCN Console — Managed Channels, Revenue Share & Access Revocation'
})

const api = usePartnerApi()
const sessionUser = useState<UserProfile | null>('partner-session', () => null)
const canLink = computed(() => {
  const role = sessionUser.value?.role || ''
  return ['partner_manager', 'mcn_manager', 'partner_admin', 'super_admin', 'platform_admin'].includes(role)
})

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

const linkChannelId = ref('')
const linkSplit = ref<number | null>(85)
const linkValidUntil = ref('')
const linkError = ref('')
const linking = ref(false)

async function submitLink() {
  linkError.value = ''
  actionError.value = null
  const channelId = linkChannelId.value.trim()
  if (!channelId) {
    linkError.value = '채널 ID를 입력하세요.'
    return
  }
  const split = linkSplit.value ?? 85
  if (!Number.isInteger(split) || split < 0 || split > 100) {
    linkError.value = '분배율은 0~100 정수여야 합니다.'
    return
  }
  linking.value = true
  try {
    await api.createPartnerContract({ channelId, creatorSplitPct: split, validUntil: linkValidUntil.value.trim() || undefined })
    linkChannelId.value = ''
    linkSplit.value = 85
    linkValidUntil.value = ''
    await refresh()
  } catch (e: unknown) {
    linkError.value = e instanceof Error ? e.message : '계약 연결에 실패했습니다.'
  } finally {
    linking.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-[var(--gvp-border)] pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">MCN 소속 크리에이터 및 계약 관리 (MCN Roster Console)</h1>
      <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
        Automatic revenue split routing in the double-entry settlement ledger · Immediate access revocation on contract termination
      </p>
    </div>

    <GvpBanner v-if="actionError" tone="danger" dismissible @dismiss="actionError = null">
      {{ actionError }}
    </GvpBanner>

    <div v-if="pending" class="space-y-2" role="status" aria-label="불러오는 중">
      <div v-for="n in 4" :key="n" class="h-14 animate-pulse rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
    </div>

    <GvpBanner v-else-if="error" tone="danger">
      계약 목록을 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <GvpEmptyState
      v-else-if="contracts.length === 0"
      title="등록된 MCN 계약이 없습니다"
      description="새 채널 계약이 체결되면 이 목록에 표시됩니다."
    />

    <div v-else class="overflow-x-auto rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
      <table class="w-full text-left">
        <thead>
          <tr class="bg-[var(--gvp-canvas)] border-y border-[var(--gvp-border)] text-xs font-semibold text-[var(--gvp-text-secondary)] uppercase tracking-wider">
            <th scope="col" class="py-2.5 px-4">Channel</th>
            <th scope="col" class="py-2.5 px-4">Contract Split (Creator : MCN)</th>
            <th scope="col" class="py-2.5 px-4">Valid Until</th>
            <th scope="col" class="py-2.5 px-4">Status / Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--gvp-border)]">
          <tr v-for="ch in contracts" :key="ch.id" class="hover:bg-[var(--gvp-canvas)]">
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)]">
              <p class="font-medium text-[var(--gvp-text)]">{{ ch.channelName }}</p>
              <p class="text-xs text-[var(--gvp-text-secondary)] tabular-nums">{{ ch.channelHandle }} · {{ ch.id }}</p>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <span class="rounded-md bg-[var(--gvp-elevated)] px-2.5 py-1 text-xs font-medium text-[var(--gvp-text)] tabular-nums">
                {{ ch.creatorSplitPct }}% Creator / {{ ch.partnerSplitPct }}% MCN
              </span>
            </td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">{{ ch.validUntil }}</td>
            <td class="py-3 px-4 text-sm text-[var(--gvp-text)] border-b border-[var(--gvp-border)] tabular-nums">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  class="rounded-md px-2 py-0.5 text-xs font-medium uppercase"
                  :class="ch.status === 'active' ? 'border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]' : 'border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] text-[var(--gvp-live)]'"
                >
                  {{ ch.status }}
                </span>
                <button
                  v-if="ch.status === 'active'"
                  type="button"
                  class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
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

    <div v-if="canLink" class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4">
      <h2 class="text-sm font-semibold text-[var(--gvp-text)]">새 채널 연결</h2>
      <div class="mt-3 flex flex-wrap items-end gap-3">
        <label class="flex min-w-64 flex-1 flex-col gap-1 text-xs font-medium text-[var(--gvp-text-secondary)]">
          채널 ID
          <input
            v-model="linkChannelId"
            type="text"
            placeholder="예: aaaaaaaa-…"
            class="h-9 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 font-mono text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
          />
        </label>
        <label class="flex w-32 flex-col gap-1 text-xs font-medium text-[var(--gvp-text-secondary)]">
          크리에이터 %
          <input
            v-model.number="linkSplit"
            type="number"
            min="0"
            max="100"
            class="h-9 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)] tabular-nums"
          />
        </label>
        <label class="flex w-40 flex-col gap-1 text-xs font-medium text-[var(--gvp-text-secondary)]">
          만료일 (선택)
          <input
            v-model="linkValidUntil"
            type="text"
            placeholder="2027-12-31"
            class="h-9 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-focus)]"
          />
        </label>
        <button
          type="button"
          :disabled="linking"
          class="inline-flex h-9 items-center justify-center rounded-md bg-[var(--gvp-brand)] px-4 text-sm font-medium text-white transition-colors hover:bg-[var(--gvp-brand-hover)] disabled:opacity-50"
          @click="submitLink"
        >
          {{ linking ? '연결 중…' : '채널 연결' }}
        </button>
      </div>
      <p v-if="linkError" role="alert" class="mt-2 text-xs font-medium text-[var(--gvp-live)]">{{ linkError }}</p>
    </div>

    <GvpModal :model-value="revokeTarget !== null" title="계약 즉시 해지" @update:model-value="(v: boolean) => { if (!v) revokeTarget = null }">
      <p class="text-sm text-[var(--gvp-text-secondary)]">
        <strong class="font-semibold text-[var(--gvp-text)]">{{ revokeTarget?.channelName }}</strong> ({{ revokeTarget?.id }}) 계약을 해지합니다.
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
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
          @click="revokeTarget = null"
        >
          취소
        </button>
        <button
          type="button"
          :disabled="revoking"
          class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors disabled:opacity-50"
          @click="confirmRevoke"
        >
          {{ revoking ? '해지 중…' : '계약 해지 확정' }}
        </button>
      </template>
    </GvpModal>
  </div>
</template>
