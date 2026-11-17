<script setup lang="ts">
import type { ModerationCase } from '@gvp/shared-types'
import { ModerationStatus } from '@gvp/shared-types'
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

const api = useAdminApi()

useSeoMeta({
  title: 'Trust & Safety Moderation Queue'
})

const { data: queueData, pending: queuePending, error: queueError } = await useAsyncData('admin-moderation-queue', async () => {
  return await api.getAdminModerationQueue()
})

const cases = computed(() => queueData.value || [])
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
    <div class="border-b border-slate-200 pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-slate-900">
        콘텐츠 신뢰 및 안전 심의 큐 (Trust &amp; Safety Moderation Queue)
      </h1>
      <p class="mt-1 text-sm text-slate-600">
        Connected to <span class="font-mono text-xs">/v1/admin/moderation</span> · Policy classification, age/geo restriction, and takedown with legal evidence preservation
      </p>
    </div>

    <div v-if="queuePending" class="space-y-4">
      <div v-for="i in 3" :key="i" class="h-32 animate-pulse rounded-lg border border-slate-200 bg-slate-100" />
    </div>
    <div v-else-if="queueError" class="rounded-lg border border-red-200 bg-red-50 p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-red-700">심의 큐를 불러오지 못했습니다.</p>
    </div>
    <GvpEmptyState
      v-else-if="cases.length === 0"
      title="심의 대기 건이 없습니다"
      description="새 신고가 접수되면 여기에 표시됩니다."
    />

    <div v-else class="space-y-4">
      <div
        v-for="item in cases"
        :key="item.id"
        class="rounded-lg border border-slate-200 bg-white p-5"
      >
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-slate-100 pb-3">
          <div class="flex flex-wrap items-center gap-2.5">
            <span class="rounded-md border border-rose-200 bg-rose-50 px-2.5 py-0.5 font-mono text-xs font-semibold text-rose-700 tabular-nums">
              {{ item.id }}
            </span>
            <span class="text-sm font-semibold uppercase text-slate-900 tabular-nums">
              [{{ item.targetType }}] {{ item.targetId }}
            </span>
            <span class="rounded-md border border-slate-200 bg-slate-100 px-2.5 py-0.5 font-mono text-xs font-medium text-slate-700">
              {{ item.policyCategory }}
            </span>
          </div>

          <div class="flex items-center gap-3">
            <span class="font-mono text-xs text-slate-400 tabular-nums">{{ item.createdAt }}</span>
            <span
              class="rounded-md px-2.5 py-0.5 text-xs font-semibold uppercase"
              :class="item.status === ModerationStatus.PENDING ? 'border border-amber-200 bg-amber-50 text-amber-700' : 'border border-emerald-200 bg-emerald-50 text-emerald-700'"
            >
              {{ item.status }}
            </span>
          </div>
        </div>

        <p class="mt-3 text-sm text-slate-600">
          <strong class="font-semibold text-slate-900">Report Reason:</strong> {{ item.reason }}
        </p>

        <p v-if="resolveError" class="mt-3 rounded-md border border-red-200 bg-red-50 p-2 text-xs font-semibold text-red-700" role="alert">
          {{ resolveError }}
        </p>

        <div class="mt-4 flex flex-wrap items-center gap-2.5">
          <input
            v-model="reviewerNotes"
            type="text"
            aria-label="Reviewer notes"
            placeholder="Reviewer notes (recorded in audit log)"
            class="h-9 min-w-[240px] flex-1 rounded-md border border-slate-300 bg-slate-50 px-3 text-sm text-slate-900 focus:border-slate-900 focus:bg-white outline-none transition-colors"
          />
          <input
            v-model="reviewerTotp"
            type="text"
            inputmode="numeric"
            maxlength="6"
            aria-label="Step-up TOTP code"
            placeholder="TOTP"
            class="h-9 w-24 rounded-md border border-slate-300 bg-slate-50 px-3 text-sm text-slate-900 focus:border-slate-900 focus:bg-white outline-none transition-colors"
          />

          <button
            type="button"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
            @click="resolveCase(item, 'approve')"
          >
            Dismiss Report (Approve Content)
          </button>
          <button
            type="button"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50 transition-colors"
            @click="resolveCase(item, 'reject')"
          >
            Apply 19+ Age / Geo Restriction
          </button>
          <button
            type="button"
            class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors"
            @click="resolveCase(item, 'remove')"
          >
            Takedown &amp; Preserve Evidence Snapshot
          </button>
        </div>

        <p v-if="item.actionTaken" class="mt-3 font-mono text-xs font-medium text-emerald-700">
          Action Recorded: {{ item.actionTaken }}
        </p>
      </div>
    </div>
  </div>
</template>
