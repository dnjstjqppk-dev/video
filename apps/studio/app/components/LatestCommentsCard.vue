<script setup lang="ts">
import type { Comment } from '@gvp/shared-types'
import { formatRelativeTime } from '@gvp/ui'
import GvpChannelAvatar from '@gvp/ui/src/components/GvpChannelAvatar.vue'

const props = defineProps<{ comments: Comment[] }>()
const shown = computed(() => props.comments.slice(0, 4))
</script>

<template>
  <section class="flex flex-col rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-5" aria-label="최신 댓글">
    <h2 class="text-sm font-semibold text-[var(--gvp-text)]">최신 댓글</h2>
    <p v-if="shown.length === 0" class="py-6 text-center text-xs text-[var(--gvp-text-secondary)]">아직 달린 댓글이 없습니다.</p>
    <ul v-else class="mt-3 flex-1 space-y-3">
      <li v-for="c in shown" :key="c.id" class="flex gap-2.5">
        <GvpChannelAvatar :src="c.author.avatarUrl" :name="c.author.displayName" :size-px="30" />
        <div class="min-w-0">
          <p class="text-xs text-[var(--gvp-text-secondary)]"><span class="font-semibold text-[var(--gvp-text)]">{{ c.author.displayName }}</span> · {{ formatRelativeTime(c.createdAt) }}</p>
          <p class="mt-0.5 line-clamp-2 text-xs leading-relaxed text-[var(--gvp-text)]">{{ c.content }}</p>
        </div>
      </li>
    </ul>
    <NuxtLink to="/comments" class="mt-3 rounded-md px-2 py-1 text-xs font-medium text-[var(--gvp-brand-hover)] hover:bg-[var(--gvp-elevated)]">댓글 모두 보기</NuxtLink>
  </section>
</template>
