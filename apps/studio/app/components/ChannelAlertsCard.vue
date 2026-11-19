<script setup lang="ts">
import type { CopyrightClaim } from '@gvp/shared-types'

const props = defineProps<{ claims: CopyrightClaim[] }>()
const active = computed(() => props.claims.filter(c => c.status !== 'resolved'))
</script>

<template>
  <section class="rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5" aria-label="저작권 및 정책 알림">
    <h2 class="border-b border-[var(--gvp-border)] pb-3 text-sm font-semibold text-[var(--gvp-text)]">저작권 및 정책 알림</h2>
    <div class="mt-4 space-y-3 text-xs">
      <div v-if="active.length === 0" class="rounded-lg border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] p-3.5 text-[var(--gvp-success)]">
        <p class="font-semibold">채널 상태 양호</p>
        <p class="mt-1 tabular-nums">진행 중인 저작권 클레임이 없습니다.</p>
      </div>
      <div v-for="claim in active" :key="claim.id" class="rounded-lg border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] p-3.5 text-[var(--gvp-warning)]">
        <p class="font-semibold">Content ID 클레임 · {{ claim.id }}</p>
        <p class="mt-1">{{ claim.ownerName || claim.rightsOwnerId }}의 {{ claim.claimAction }} 클레임이 있습니다.</p>
        <NuxtLink to="/copyright" class="mt-2 inline-block font-semibold text-[var(--gvp-text)] underline">검토 또는 이의 제기</NuxtLink>
      </div>
    </div>
  </section>
</template>
