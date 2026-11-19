<script setup lang="ts">
import type { CopyrightClaim } from '@gvp/shared-types'

const props = defineProps<{ videoId: string | null; status: string; jobError: string }>()
const api = useStudioApi()
const claims = ref<CopyrightClaim[]>([])
const loading = ref(false)
const failed = ref(false)

async function load() {
  if (!props.videoId) return
  loading.value = true
  failed.value = false
  try {
    const all = await api.getStudioClaims()
    claims.value = (all || []).filter(c => c.claimedVideoId === props.videoId)
  } catch { failed.value = true } finally { loading.value = false }
}

watch(() => props.videoId, load, { immediate: true })
</script>

<template>
  <div class="space-y-4">
    <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4 text-xs">
      <p class="font-semibold text-[var(--gvp-text)]">인코딩 검사 상태</p>
      <p class="tabular-nums mt-1 font-mono text-sm font-bold" :class="status === 'READY' ? 'text-[var(--gvp-success)]' : status === 'FAILED' ? 'text-[var(--gvp-live)]' : 'text-[var(--gvp-studio)]'">
        {{ status }}
      </p>
      <p v-if="jobError" role="alert" class="mt-1 text-[var(--gvp-live)]">{{ jobError }}</p>
      <p v-else-if="status !== 'READY'" class="mt-1 text-[var(--gvp-text-secondary)]">처리가 끝나면 자동으로 갱신됩니다.</p>
    </div>
    <div class="rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4 text-xs">
      <p class="font-semibold text-[var(--gvp-text)]">저작권 클레임</p>
      <p v-if="!videoId" class="mt-1 text-[var(--gvp-text-secondary)]">업로드 시작 후 확인할 수 있습니다.</p>
      <p v-else-if="loading" role="status" class="mt-1 text-[var(--gvp-text-secondary)]">확인 중…</p>
      <p v-else-if="failed" role="alert" class="mt-1 font-medium text-[var(--gvp-live)]">클레임을 불러오지 못했습니다.</p>
      <p v-else-if="!claims.length" class="mt-1 text-[var(--gvp-success)]">이 영상에 대한 클레임이 없습니다.</p>
      <ul v-else class="mt-2 divide-y divide-[var(--gvp-border)] rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)]">
        <li v-for="c in claims" :key="c.id" class="flex flex-wrap items-center justify-between gap-2 px-3 py-2">
          <span class="text-[var(--gvp-text)]">{{ c.ownerName }} · {{ c.claimAction }}</span>
          <span class="rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-2 py-0.5 font-semibold text-[var(--gvp-warning)]">{{ c.status }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>
