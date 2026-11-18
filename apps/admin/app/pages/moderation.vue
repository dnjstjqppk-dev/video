<script setup lang="ts">
import type { ModerationCase } from '@gvp/shared-types'
import { ModerationStatus } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

const api = useAdminApi()
const config = useRuntimeConfig()

useSeoMeta({
  title: 'Trust & Safety Moderation Queue'
})

const { data: queueData, pending: queuePending, error: queueError } = await useAsyncData('admin-moderation-queue', async () => {
  return await api.getAdminModerationQueue()
})

const cases = computed(() => queueData.value || [])
const statusFilter = ref<'all' | 'pending' | 'resolved'>('all')
const visibleCases = computed(() => {
  if (statusFilter.value === 'pending') return cases.value.filter(c => c.status === ModerationStatus.PENDING)
  if (statusFilter.value === 'resolved') return cases.value.filter(c => c.status !== ModerationStatus.PENDING)
  return cases.value
})
const reviewerNotes = ref('')
const reviewerTotp = ref('')
const resolveError = ref<string | null>(null)

async function resolveCase(item: ModerationCase, action: 'approve' | 'reject' | 'remove') {
  resolveError.value = null
  if (reviewerTotp.value.trim().length !== 6) {
    resolveError.value = '6자리 TOTP 코드를 입력하세요.'
    return
  }
  try {
    const updated = await api.resolveModerationCase(item.id, action, reviewerNotes.value, reviewerTotp.value.trim())
    item.status = updated.status as ModerationStatus
    item.actionTaken = updated.actionTaken
    reviewerTotp.value = ''
  } catch (e: unknown) {
    resolveError.value = e instanceof Error ? e.message : '조치에 실패했습니다.'
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-[var(--gvp-border)] pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-[var(--gvp-text)]">
        콘텐츠 신뢰 및 안전 심의 큐 (Trust &amp; Safety Moderation Queue)
      </h1>
      <p class="mt-1 text-sm text-[var(--gvp-text-secondary)]">
        Connected to <span class="font-mono text-xs">/v1/admin/moderation</span> · Policy classification, age/geo restriction, and takedown with legal evidence preservation
      </p>
    </div>

    <div v-if="queuePending" class="space-y-4">
      <div v-for="i in 3" :key="i" class="h-32 animate-pulse rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="queueError" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">심의 큐를 불러오지 못했습니다.</p>
    </div>
    <GvpEmptyState
      v-else-if="cases.length === 0"
      title="심의 대기 건이 없습니다"
      description="새 신고가 접수되면 여기에 표시됩니다."
    />

    <div v-else class="space-y-4">
      <div class="flex flex-wrap items-center gap-3">
        <label class="flex items-center gap-2 text-xs font-medium text-[var(--gvp-text-secondary)]">
          우선순위 · 상태 필터
          <select
            v-model="statusFilter"
            class="h-8 rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2 text-xs text-[var(--gvp-text)]"
            aria-label="심의 상태 필터"
          >
            <option value="all">전체 ({{ cases.length }})</option>
            <option value="pending">처리 대기</option>
            <option value="resolved">처리 완료</option>
          </select>
        </label>
        <span class="tabular-nums text-xs text-[var(--gvp-text-secondary)]">{{ visibleCases.length }}건 표시</span>
      </div>
      <div
        v-for="item in visibleCases"
        :key="item.id"
        class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5"
      >
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-[var(--gvp-border)] pb-3">
          <div class="flex flex-wrap items-center gap-2.5">
            <span class="rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-2.5 py-0.5 font-mono text-xs font-semibold text-[var(--gvp-live)] tabular-nums">
              {{ item.id }}
            </span>
            <span class="text-sm font-semibold uppercase text-[var(--gvp-text)] tabular-nums">
              [{{ item.targetType }}] {{ item.targetId }}
            </span>
            <span class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-elevated)] px-2.5 py-0.5 font-mono text-xs font-medium text-[var(--gvp-text)]">
              {{ item.policyCategory }}
            </span>
            <a
              v-if="item.targetType === 'video'"
              :href="`${config.public.webUrl}/watch/${item.targetId}`"
              target="_blank"
              rel="noopener noreferrer"
              class="rounded-md bg-[var(--gvp-brand)] px-2.5 py-1 text-xs font-medium text-white hover:bg-[var(--gvp-brand-hover)]"
            >
              원본 보기 ↗
            </a>
          </div>

          <div class="flex items-center gap-3">
            <span class="font-mono text-xs text-[var(--gvp-text-muted)] tabular-nums">{{ item.createdAt }}</span>
            <span
              class="rounded-md px-2.5 py-0.5 text-xs font-semibold uppercase"
              :class="item.status === ModerationStatus.PENDING ? 'border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] text-[var(--gvp-warning)]' : 'border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] text-[var(--gvp-success)]'"
            >
              {{ item.status }}
            </span>
          </div>
        </div>

        <div class="mt-4 grid gap-4 lg:grid-cols-[1fr_340px]">
          <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4">
            <p class="text-xs font-semibold uppercase tracking-wider text-[var(--gvp-text-muted)]">신고 내용 · 검수 메모</p>
            <p class="mt-2 text-sm text-[var(--gvp-text-secondary)]">
              <strong class="font-semibold text-[var(--gvp-text)]">Report Reason:</strong> {{ item.reason }}
            </p>
            <p v-if="item.actionTaken" class="mt-2 font-mono text-xs font-medium text-[var(--gvp-success)]">
              Action Recorded: {{ item.actionTaken }}
            </p>
            <p v-if="resolveError" class="mt-2 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-2 text-xs font-semibold text-[var(--gvp-live)]" role="alert">
              {{ resolveError }}
            </p>
          </div>

          <div class="flex flex-col gap-2">
            <input
              v-model="reviewerNotes"
              type="text"
              aria-label="Reviewer notes"
              placeholder="Reviewer notes (audit log 기록)"
              class="h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 text-sm text-[var(--gvp-text)] focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
            />
            <input
              v-model="reviewerTotp"
              type="text"
              inputmode="numeric"
              maxlength="6"
              aria-label="Step-up TOTP code"
              placeholder="TOTP 6자리"
              class="tabular-nums h-9 w-full rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-canvas)] px-3 text-sm text-[var(--gvp-text)] focus:border-[var(--gvp-focus)] focus:bg-[var(--gvp-surface)] outline-none transition-colors"
            />
            <button
              type="button"
              class="inline-flex h-9 items-center justify-center rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-sm font-medium text-[var(--gvp-text)] transition-colors hover:bg-[var(--gvp-canvas)]"
              @click="resolveCase(item, 'approve')"
            >
              정상 승인
            </button>
            <button
              type="button"
              class="inline-flex h-9 items-center justify-center rounded-md border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-4 text-sm font-medium text-[var(--gvp-text)] transition-colors hover:bg-[var(--gvp-canvas)]"
              @click="resolveCase(item, 'reject')"
            >
              19+ 연령 제한
            </button>
            <button
              type="button"
              class="inline-flex h-9 items-center justify-center rounded-md bg-[var(--gvp-live)] px-4 text-sm font-medium text-white transition-colors hover:opacity-90"
              @click="resolveCase(item, 'remove')"
            >
              차단 · 삭제 (증거 보존)
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
