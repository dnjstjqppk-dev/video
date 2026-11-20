<script setup lang="ts">
import type { CopyrightClaim, UserProfile } from '@gvp/shared-types'
import { CopyrightClaimAction, CopyrightClaimStatus } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

const api = usePartnerApi()
const sessionUser = useState<UserProfile | null>('partner-session', () => null)

// Policy changes are reporter-or-operator only (mirrors the API guard);
// everyone else keeps read-only visibility into their own disputes.
function canManageClaim(claim: CopyrightClaim): boolean {
  const user = sessionUser.value
  if (!user) return false
  if (claim.rightsOwnerId && claim.rightsOwnerId === user.id) return true
  return ['super_admin', 'platform_admin', 'moderator'].includes(user.role)
}

useSeoMeta({
  title: 'Rights Owner Console — SHA-256 Fingerprints & Content ID Claims'
})

const claimError = ref<string | null>(null)

const { data: claimsData, pending: claimsPending, error: claimsError, refresh: refreshClaims } = await useAsyncData('partner-claims', async () => {
  return await api.getPartnerClaims()
})

const claims = computed(() => claimsData.value || [])

async function updateClaimAction(claim: CopyrightClaim, action: CopyrightClaimAction) {
  claimError.value = null
  try {
    await api.updateCopyrightClaim(claim.id, { action })
    await refreshClaims()
  } catch (e: unknown) {
    claimError.value = e instanceof Error ? e.message : '클레임 처리에 실패했습니다.'
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-[var(--gvp-border)] pb-5">
      <h1 class="text-xl font-semibold tracking-tight text-[var(--gvp-text)]">저작권 권리자 및 Content ID 콘솔 (Rights Owner Console)</h1>
      <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
        Register SHA-256 &amp; perceptual audio/video fingerprints · Enforce territory rules (Monetize / Track / Block) · Resolve creator counter-disputes
      </p>
    </div>

    <!-- Register Reference Fingerprint Asset -->
    <div class="rounded-lg border border-dashed border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] p-6 text-center">
      <h2 class="text-sm font-semibold text-[var(--gvp-text-secondary)]">원본 레퍼런스 자산 등록 연동 준비 중</h2>
      <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">핑거프린트 등록 API 연결 후 제공됩니다.</p>
    </div>

    <!-- Suspected Infringement Matches & Creator Counter-Disputes -->
    <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-6">
      <h2 class="text-sm font-semibold text-[var(--gvp-text)]">Content ID 매칭 목록 및 크리에이터 이의 신청 처리</h2>

      <p v-if="claimError" class="mt-3 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-2 text-xs font-semibold text-[var(--gvp-live)]" role="alert">
        {{ claimError }}
      </p>
      <div v-if="claimsPending" class="mt-4 space-y-3">
        <div v-for="i in 2" :key="i" class="h-20 animate-pulse rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
      </div>
      <div v-else-if="claimsError" class="mt-4 rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-6 text-center" role="alert">
        <p class="text-xs font-semibold text-[var(--gvp-live)]">클레임 목록을 불러오지 못했습니다.</p>
      </div>
      <GvpEmptyState
        v-else-if="claims.length === 0"
        title="진행 중인 클레임이 없습니다"
        description="새 매칭이 접수되면 여기에 표시됩니다."
      />

      <div v-else class="mt-4 space-y-4">
        <div
          v-for="claim in claims"
          :key="claim.id"
          class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4 text-sm"
        >
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-md bg-[var(--gvp-elevated)] px-2 py-0.5 font-mono text-xs font-semibold text-[var(--gvp-text)] tabular-nums">{{ claim.id }}</span>
              <span class="font-semibold text-[var(--gvp-text)]">Target Video: {{ claim.claimedVideoId }}</span>
              <span class="text-xs text-[var(--gvp-text-secondary)] tabular-nums">
                Segment: {{ claim.matchStartSeconds }}s – {{ claim.matchEndSeconds }}s (Confidence: {{ claim.matchConfidence }}%)
              </span>
            </div>
            <span
              class="rounded-md px-2.5 py-0.5 text-xs font-medium uppercase"
              :class="claim.status === CopyrightClaimStatus.DISPUTED ? 'border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] text-[var(--gvp-warning)]' : 'border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]'"
            >
              {{ claim.status }} · Policy: {{ claim.claimAction.toUpperCase() }}
            </span>
          </div>

          <div v-if="claim.disputeReason" class="mt-3 rounded-lg border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] p-3 text-xs text-[var(--gvp-warning)]">
            <p class="font-semibold">Creator Counter-Dispute Filed</p>
            <p class="mt-1 text-[var(--gvp-text)]">"{{ claim.disputeReason }}"</p>
            <p class="mt-1 text-[var(--gvp-text-secondary)]">
              Note: Ad revenue is held in the Double-Entry Escrow Account until resolved.
            </p>
          </div>

          <div v-if="canManageClaim(claim)" class="mt-4 flex flex-wrap gap-2">
            <button
              type="button"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-[var(--gvp-brand)] rounded-md hover:bg-[var(--gvp-brand-hover)] transition-colors"
              @click="updateClaimAction(claim, CopyrightClaimAction.MONETIZE)"
            >
              Monetize (Claim Revenue)
            </button>
            <button
              type="button"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
              @click="updateClaimAction(claim, CopyrightClaimAction.TRACK)"
            >
              Track Only (Release Escrow to Creator)
            </button>
            <button
              type="button"
              class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-[var(--gvp-text)] bg-[var(--gvp-surface)] border border-[var(--gvp-border-strong)] rounded-md hover:bg-[var(--gvp-canvas)] transition-colors"
              @click="updateClaimAction(claim, CopyrightClaimAction.BLOCK)"
            >
              Enforce Block Worldwide
            </button>
          </div>
          <p v-else class="mt-4 text-xs text-[var(--gvp-text-secondary)]">
            정책 변경은 신고한 권리자 또는 운영자만 가능합니다. (읽기 전용으로 보고 있습니다)
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
