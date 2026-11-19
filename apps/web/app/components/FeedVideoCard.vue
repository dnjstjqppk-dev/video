<script setup lang="ts">
import type { Video } from '@gvp/shared-types'
import { formatRelativeTime, formatTimecode, formatViewCount } from '@gvp/ui'

const props = defineProps<{ video: Video }>()
const emit = defineEmits<{ (e: 'toast', msg: string): void }>()
const appStore = useAppStore()
const library = useLibraryStore()
const api = useApi()
const ko = computed(() => appStore.locale === 'ko'); const menuOpen = ref(false); const hovering = ref(false)
let hoverTimer: ReturnType<typeof setTimeout> | null = null
const progress = computed(() => {
  const p = appStore.getWatchProgress(props.video.id); const d = props.video.durationSeconds
  return d > 0 ? Math.min(100, (p / d) * 100) : 0
})
function onEnter() { if (hoverTimer) clearTimeout(hoverTimer); hoverTimer = setTimeout(() => { hovering.value = true }, 400) }
function onLeave() { if (hoverTimer) clearTimeout(hoverTimer); hovering.value = false }
async function feedback(action: 'not_interested' | 'dont_recommend_channel') {
  menuOpen.value = false
  if (action === 'not_interested') appStore.hiddenVideoIds.add(props.video.id)
  else appStore.blockedChannelIds.add(props.video.channelId)
  emit('toast', ko.value ? '피드백이 반영되었습니다.' : 'Feedback applied.')
  try { await api.sendFeedFeedback(props.video.id, action) } catch { /* non-blocking */ }
}
function save() {
  menuOpen.value = false
  library.toggleSaved(props.video)
  emit('toast', ko.value ? '나중에 볼 동영상에 저장했습니다.' : 'Saved to Watch Later.')
}
async function share() {
  menuOpen.value = false
  const url = `${location.origin}/watch/${props.video.id}`
  try {
    if (navigator.share) await navigator.share({ title: props.video.title, url })
    else { await navigator.clipboard.writeText(url); emit('toast', ko.value ? '링크가 복사되었습니다.' : 'Link copied.') }
  } catch { /* dismissed */ }
}
async function report() {
  menuOpen.value = false
  try {
    await api.reportContent({ targetType: 'video', targetId: props.video.id, reason: 'user_report', policyCategory: 'user-report' })
    emit('toast', ko.value ? '신고가 접수되었습니다.' : 'Report submitted.')
  } catch { emit('toast', ko.value ? '신고 접수에 실패했습니다.' : 'Report failed.') }
}
</script>

<template>
  <article class="group relative flex flex-col">
    <NuxtLink :to="`/watch/${video.id}`" class="relative aspect-video w-full overflow-hidden rounded-xl bg-[var(--gvp-elevated)]" @mouseenter="onEnter" @mouseleave="onLeave" @focus="onEnter" @blur="onLeave">
      <video v-if="hovering && video.previewUrl" :src="video.previewUrl" muted loop autoplay playsinline class="h-full w-full object-cover" />
      <img v-else :src="video.thumbnailUrl" :alt="video.title" class="h-full w-full object-cover" loading="lazy" />
      <span class="absolute bottom-2 right-2 rounded bg-black/80 px-1 py-0.5 text-xs font-medium text-white tabular-nums">{{ formatTimecode(video.durationSeconds) }}</span>
      <div v-if="progress > 0" class="absolute inset-x-0 bottom-0 h-[3px] bg-white/25"><div class="h-full bg-[var(--gvp-brand)]" :style="{ width: `${progress}%` }" /></div>
    </NuxtLink>
    <div class="mt-3 flex items-start justify-between gap-2">
      <NuxtLink :to="`/channels/${video.channelId}`" class="shrink-0" :aria-label="video.channel?.name">
        <img :src="video.channel?.avatarUrl || `https://api.dicebear.com/9.x/initials/svg?seed=${video.channelId}`" :alt="video.channel?.name" class="h-9 w-9 rounded-full object-cover" loading="lazy" />
      </NuxtLink>
      <div class="min-w-0 flex-1">
        <NuxtLink :to="`/watch/${video.id}`" class="line-clamp-2 text-[15px] font-semibold leading-[1.4] text-[var(--gvp-text)]">{{ video.title }}</NuxtLink>
        <NuxtLink :to="`/channels/${video.channelId}`" class="mt-1 flex items-center gap-1 text-xs text-[var(--gvp-text-secondary)] hover:text-[var(--gvp-text)]">
          <span class="truncate">{{ video.channel?.name || 'vravra 크리에이터' }}</span>
          <svg v-if="video.channel?.isVerified" class="h-3.5 w-3.5 shrink-0 text-[var(--gvp-brand)]" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10 10-4.5 10-10S17.5 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" /></svg>
        </NuxtLink>
        <p class="tabular-nums text-xs text-[var(--gvp-text-secondary)]">{{ formatViewCount(video.viewCount, appStore.locale) }} · {{ formatRelativeTime(video.publishedAt || video.createdAt, appStore.locale) }}</p>
      </div>
      <div v-click-outside="() => (menuOpen = false)" class="relative shrink-0">
        <button type="button" aria-label="동영상 옵션" :aria-expanded="menuOpen" class="rounded-full p-1.5 text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]" @click="menuOpen = !menuOpen">⋮</button>
        <div v-if="menuOpen" role="menu" class="gvp-dropdown gvp-elevated absolute right-0 top-8 z-30 w-52 rounded-xl p-1.5 text-xs">
          <button type="button" role="menuitem" class="w-full rounded-lg px-3 py-2 text-left text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="save()">{{ ko ? '나중에 볼 동영상에 저장' : 'Save to Watch Later' }}</button>
          <button type="button" role="menuitem" class="w-full rounded-lg px-3 py-2 text-left text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="share()">{{ ko ? '공유' : 'Share' }}</button>
          <button type="button" role="menuitem" class="w-full rounded-lg px-3 py-2 text-left text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="feedback('not_interested')">{{ ko ? '관심 없음' : 'Not interested' }}</button>
          <button type="button" role="menuitem" class="w-full rounded-lg px-3 py-2 text-left text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="feedback('dont_recommend_channel')">{{ ko ? '이 채널 추천 안함' : "Don't recommend channel" }}</button>
          <button type="button" role="menuitem" class="w-full rounded-lg px-3 py-2 text-left text-[var(--gvp-live)] hover:bg-[var(--gvp-elevated)]" @click="report()">{{ ko ? '신고' : 'Report' }}</button>
        </div>
      </div>
    </div>
  </article>
</template>
