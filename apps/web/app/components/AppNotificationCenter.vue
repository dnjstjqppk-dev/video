<script setup lang="ts">
import type { NotificationItem } from '@gvp/shared-types'

const api = useApi()
const router = useRouter()
const store = useAppStore()
const open = ref(false)
const tab = ref('all')
const items = ref<NotificationItem[]>([])
const unread = ref(0)
const tabs = [
  { id: 'all', label: '전체' },
  { id: 'subs', label: '구독 채널' },
  { id: 'comments', label: '댓글·멘션' },
  { id: 'studio', label: '스튜디오·정산' },
]
function inTab(n: NotificationItem) {
  const t = n.type.toLowerCase()
  if (tab.value === 'subs') return /sub|upload|live|video|premiere/.test(t)
  if (tab.value === 'comments') return /comment|mention|reply/.test(t)
  if (tab.value === 'studio') return /studio|payout|settle|monet|copyright|claim/.test(t)
  return true
}
const visible = computed(() => items.value.filter(inTab))
async function load() {
  if (!store.currentUser) return
  try { const r = await api.getNotifications(1, 30); items.value = r.items; unread.value = r.unread } catch { /* empty */ }
}
async function openItem(n: NotificationItem) {
  open.value = false
  if (!n.read) { try { await api.markNotificationRead(n.id); n.read = true; unread.value = Math.max(0, unread.value - 1) } catch { /* keep */ } }
  if (n.refType === 'video' && n.refId) router.push(`/watch/${n.refId}`)
  else if (n.refType === 'channel' && n.refId) router.push(`/channels/${n.refId}`)
}
async function markAll() {
  for (const n of items.value.filter(i => !i.read)) { try { await api.markNotificationRead(n.id); n.read = true } catch { /* keep */ } }
  unread.value = 0
}
onMounted(load)
watch(() => store.currentUser, load)
</script>

<template>
  <div v-click-outside="() => (open = false)" class="relative">
    <button type="button" :aria-expanded="open" aria-label="알림 센터" class="relative flex h-9 w-9 items-center justify-center rounded-full hover:bg-[var(--gvp-elevated)]" @click="open = !open">
      <svg class="h-5 w-5 text-[var(--gvp-text)]" viewBox="0 0 24 24" fill="currentColor"><path d="M12 22c1.1 0 2-.9 2-2h-4a2 2 0 0 0 2 2zm6-6v-5a6 6 0 0 0-4.5-5.8V4.5a1.5 1.5 0 0 0-3 0v.7A6 6 0 0 0 6 11v5l-2 2v1h16v-1l-2-2z" /></svg>
      <span v-if="unread > 0" class="tabular-nums absolute right-0.5 top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-[var(--gvp-live)] px-1 text-[10px] font-bold text-white">{{ unread > 9 ? '9+' : unread }}</span>
    </button>
    <div v-if="open" role="dialog" aria-label="알림" class="gvp-dropdown gvp-elevated absolute right-0 top-full z-50 mt-2 flex max-h-[560px] w-[420px] max-w-[calc(100vw-2rem)] flex-col rounded-xl">
      <div class="flex gap-1 border-b border-[var(--gvp-border)] p-2" role="tablist">
        <button v-for="t in tabs" :key="t.id" role="tab" :aria-selected="tab === t.id" @click="tab = t.id"
          :class="tab === t.id ? 'bg-[var(--gvp-text)] text-[var(--gvp-canvas)]' : 'text-[var(--gvp-text-secondary)] hover:text-[var(--gvp-text)]'"
          class="rounded-md px-2.5 py-1.5 text-xs font-medium">{{ t.label }}</button>
      </div>
      <div class="flex-1 overflow-y-auto p-1.5">
        <p v-if="!visible.length" class="p-6 text-center text-xs text-[var(--gvp-text-secondary)]">새 알림이 없습니다.</p>
        <button v-for="n in visible" :key="n.id" type="button" class="flex w-full items-start gap-2.5 rounded-lg px-2.5 py-2 text-left hover:bg-[var(--gvp-elevated)]" @click="openItem(n)">
          <span v-if="!n.read" class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-[var(--gvp-accent)]" aria-label="안읽음" />
          <span v-else class="w-2 shrink-0" />
          <span class="min-w-0 flex-1"><span class="block text-xs font-semibold text-[var(--gvp-text)]">{{ n.title }}</span>
          <span class="clamp-2 mt-0.5 block text-xs text-[var(--gvp-text-secondary)]">{{ n.body }}</span></span>
        </button>
      </div>
      <div class="flex items-center justify-between border-t border-[var(--gvp-border)] p-2 text-xs">
        <button type="button" class="rounded-md px-2.5 py-1.5 font-medium text-[var(--gvp-brand)] hover:bg-[var(--gvp-elevated)]" @click="markAll()">모두 읽음 처리</button>
        <NuxtLink to="/premium" class="rounded-md px-2.5 py-1.5 text-[var(--gvp-text-secondary)] hover:text-[var(--gvp-text)]" @click="open = false">알림 설정</NuxtLink>
      </div>
    </div>
  </div>
</template>
