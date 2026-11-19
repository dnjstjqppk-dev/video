<script setup lang="ts">
import type { SubChannel } from '~/composables/useSubscriptionChannels'

defineProps<{ sub: SubChannel }>()
</script>

<template>
  <NuxtLink :to="`/channels/${sub.id}`" class="flex items-center justify-between rounded-lg px-3 py-1.5 text-sm text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]">
    <span class="flex min-w-0 items-center gap-2.5">
      <img
        v-if="sub.avatarUrl" :src="sub.avatarUrl" :alt="sub.name" loading="lazy"
        class="h-6 w-6 shrink-0 rounded-full object-cover"
        :class="sub.isLive ? 'ring-2 ring-[var(--gvp-live)] ring-offset-1 ring-offset-[var(--gvp-surface)]' : ''"
      />
      <span v-else class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[var(--gvp-elevated)] text-xs font-semibold">{{ sub.name.slice(0, 1) }}</span>
      <span class="truncate">{{ sub.name }}</span>
    </span>
    <span v-if="sub.isLive" class="tabular-nums flex shrink-0 items-center gap-1 rounded bg-[var(--gvp-live)] px-1 py-0.5 text-[10px] font-bold text-white">
      <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-white" />LIVE
    </span>
    <span v-else class="h-2 w-2 shrink-0 rounded-full bg-[var(--gvp-brand)]" title="새 영상" />
  </NuxtLink>
</template>
