<script setup lang="ts">
import GvpEmptyState from '@gvp/ui/src/components/GvpEmptyState.vue'

const api = useAdminApi()
const config = useRuntimeConfig()

useSeoMeta({
  title: 'AI Originals Review Queue'
})

const { data: queueData, pending, error, refresh } = await useAsyncData('admin-ai-reviews', async () => {
  return await api.listAIReviews(1, 50)
})

const items = computed(() => queueData.value?.items || [])
const total = computed(() => queueData.value?.total || 0)
const notes = ref<Record<string, string>>({})
const busyId = ref<string | null>(null)
const actionError = ref<string | null>(null)
const doneIds = ref<Set<string>>(new Set())

async function decide(videoId: string, decision: 'approved' | 'rejected') {
  if (busyId.value) return
  busyId.value = videoId
  actionError.value = null
  try {
    await api.reviewAI(videoId, decision, notes.value[videoId]?.trim() || undefined)
    doneIds.value.add(videoId)
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : '심사에 실패했습니다.'
  } finally {
    busyId.value = null
  }
}

const remaining = computed(() => items.value.filter(i => !doneIds.value.has(i.video.id)))
</script>

<template>
  <div class="space-y-6">
    <div class="border-b border-[var(--gvp-border)] pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-[var(--gvp-text)]">
        AI Originals 심사 큐 (AI Reviews)
      </h1>
      <p class="mt-1 text-sm text-[var(--gvp-text-secondary)] tabular-nums">
        Connected to <span class="font-mono text-xs">/v1/admin/ai-reviews</span> · {{ total }}건 대기 · 오래된 순
      </p>
    </div>

    <div v-if="pending" class="space-y-4">
      <div v-for="i in 3" :key="i" class="h-32 animate-pulse rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="error" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-8 text-center" role="alert">
      <p class="text-sm font-semibold text-[var(--gvp-live)]">심사 큐를 불러오지 못했습니다.</p>
      <button type="button" class="mt-3 text-xs font-semibold underline" @click="refresh()">다시 시도</button>
    </div>
    <GvpEmptyState
      v-else-if="!remaining.length"
      title="심사 대기 건이 없습니다"
      description="크리에이터가 콘텐츠 탭에서 Originals 심사를 신청하면 여기에 표시됩니다."
    />

    <div v-else class="space-y-4">
      <p v-if="actionError" role="alert" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-4 py-2.5 text-xs font-medium text-[var(--gvp-live)]">{{ actionError }}</p>
      <article
        v-for="item in remaining"
        :key="item.video.id"
        class="rounded-xl border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-4 sm:p-5"
      >
        <div class="flex flex-col gap-4 sm:flex-row">
          <a :href="`${config.public.webUrl}/watch/${item.video.id}`" target="_blank" rel="noopener noreferrer" class="relative block h-28 w-48 shrink-0 overflow-hidden rounded-lg bg-[var(--gvp-elevated)]">
            <img :src="item.video.thumbnailUrl" :alt="item.video.title" class="h-full w-full object-cover" loading="lazy" />
          </a>
          <div class="min-w-0 flex-1">
            <a :href="`${config.public.webUrl}/watch/${item.video.id}`" target="_blank" rel="noopener noreferrer" class="line-clamp-2 text-sm font-bold text-[var(--gvp-text)] hover:underline">
              {{ item.video.title }}
            </a>
            <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
              {{ item.channelName }} ({{ item.channelHandle }})
              <span v-if="item.seriesTitle"> · 시리즈: {{ item.seriesTitle }}</span>
            </p>
            <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
              AI 도구: {{ item.video.aiTool || '미표기' }} · {{ item.video.visibility }} · {{ item.video.status }}
            </p>
            <p class="mt-1 line-clamp-2 text-xs text-[var(--gvp-text-muted)]">{{ item.video.description }}</p>
          </div>
        </div>
        <div class="mt-3 flex flex-col gap-2 sm:flex-row">
          <input
            v-model="notes[item.video.id]"
            type="text"
            maxlength="2000"
            placeholder="심사 메모 (반려 사유 등, 창작자에게 전달됩니다)"
            :aria-label="`${item.video.title} 심사 메모`"
            class="h-9 flex-1 rounded-lg border border-[var(--gvp-border-strong)] bg-[var(--gvp-surface)] px-2.5 text-xs outline-none focus:border-[var(--gvp-focus)]"
          />
          <div class="flex shrink-0 gap-2">
            <button
              type="button"
              :disabled="busyId === item.video.id"
              class="inline-flex h-9 items-center rounded-full bg-[var(--gvp-brand)] px-4 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50"
              @click="decide(item.video.id, 'approved')"
            >{{ busyId === item.video.id ? '처리 중…' : '선정' }}</button>
            <button
              type="button"
              :disabled="busyId === item.video.id"
              class="inline-flex h-9 items-center rounded-full border border-[var(--gvp-live)] px-4 text-xs font-semibold text-[var(--gvp-live)] hover:bg-[var(--gvp-live-soft)] disabled:opacity-50"
              @click="decide(item.video.id, 'rejected')"
            >반려</button>
          </div>
        </div>
      </article>
    </div>
  </div>
</template>
