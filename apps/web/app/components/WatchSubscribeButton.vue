<script setup lang="ts">
const props = defineProps<{ channelId: string; initialCount: number }>()
const emit = defineEmits<{ (e: 'change', v: { subscribed: boolean; count: number }): void }>()
const api = useApi()
const ko = computed(() => useAppStore().locale === 'ko')
const subscribed = ref(false)
const notifyLevel = ref<'all' | 'custom' | 'none'>('all')
const count = ref(props.initialCount)
const open = ref(false)
const failed = ref(false)

const notifyOpts = [
  { id: 'all', labelKo: '전체', labelEn: 'All', hintKo: '모든 새 동영상 알림', hintEn: 'Notify on every upload' },
  { id: 'custom', labelKo: '맞춤설정', labelEn: 'Personalized', hintKo: '관심 있을 만한 소식만', hintEn: 'Highlights only' },
  { id: 'none', labelKo: '없음', labelEn: 'None', hintKo: '알림 끄기', hintEn: 'Muted' },
] as const

watch(() => props.initialCount, v => { count.value = v })
watch(() => props.channelId, () => { subscribed.value = false; notifyLevel.value = 'all'; void sync() })

async function sync() {
  try {
    const sub = await api.getSubscription(props.channelId)
    subscribed.value = sub.subscribed
    count.value = sub.subscriberCount
    if (sub.subscribed && (sub.notifyLevel === 'all' || sub.notifyLevel === 'custom' || sub.notifyLevel === 'none')) {
      notifyLevel.value = sub.notifyLevel
    }
  } catch { /* stays unsubscribed */ }
}

async function subscribe() {
  failed.value = false
  try {
    const res = await api.subscribeChannel(props.channelId)
    subscribed.value = res.subscribed
    count.value = res.subscriberCount
    if (res.notifyLevel === 'all' || res.notifyLevel === 'custom' || res.notifyLevel === 'none') {
      notifyLevel.value = res.notifyLevel
    }
    emit('change', { subscribed: true, count: count.value })
  } catch { failed.value = true }
}

async function setNotify(level: 'all' | 'custom' | 'none') {
  failed.value = false
  try {
    const res = await api.setSubscriptionNotify(props.channelId, level)
    notifyLevel.value = res.notifyLevel as 'all' | 'custom' | 'none'
    count.value = res.subscriberCount
  } catch { failed.value = true }
}

async function unsubscribe() {
  open.value = false
  failed.value = false
  try {
    const res = await api.unsubscribeChannel(props.channelId)
    subscribed.value = res.subscribed
    count.value = res.subscriberCount
    emit('change', { subscribed: false, count: count.value })
  } catch { failed.value = true }
}

onMounted(sync)
</script>

<template>
  <div v-click-outside="() => (open = false)" class="relative ml-3">
    <button
      v-if="!subscribed"
      type="button"
      class="rounded-full bg-[var(--gvp-brand)] px-4 py-2 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
      @click="subscribe"
    >
      {{ ko ? '구독' : 'Subscribe' }}
    </button>
    <button
      v-else
      type="button"
      :aria-expanded="open"
      class="flex items-center gap-1 rounded-full bg-[var(--gvp-elevated)] px-4 py-2 text-xs font-semibold text-[var(--gvp-text)] transition hover:bg-[var(--gvp-canvas)]"
      @click="open = !open"
    >
      {{ ko ? '구독중 ▾' : 'Subscribed ▾' }}
    </button>
    <div v-if="open && subscribed" role="menu" aria-label="구독 알림 설정" class="gvp-dropdown gvp-elevated absolute left-0 top-full z-30 mt-2 w-60 rounded-xl p-1.5">
      <p class="px-3 pb-1 pt-1.5 text-[11px] font-semibold text-[var(--gvp-text-muted)]">{{ ko ? '알림 설정' : 'Notifications' }}</p>
      <button
        v-for="o in notifyOpts" :key="o.id" type="button" role="menuitemradio" :aria-checked="notifyLevel === o.id"
        class="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-left hover:bg-[var(--gvp-elevated)]"
        @click="setNotify(o.id)"
      >
        <span class="flex h-4 w-4 shrink-0 items-center justify-center rounded-full border" :class="notifyLevel === o.id ? 'border-[var(--gvp-brand)]' : 'border-[var(--gvp-border-strong)]'">
          <span v-if="notifyLevel === o.id" class="h-2 w-2 rounded-full bg-[var(--gvp-brand)]" />
        </span>
        <span>
          <span class="block text-xs font-medium text-[var(--gvp-text)]">{{ ko ? o.labelKo : o.labelEn }}</span>
          <span class="block text-[11px] text-[var(--gvp-text-secondary)]">{{ ko ? o.hintKo : o.hintEn }}</span>
        </span>
      </button>
      <div class="my-1 border-t border-[var(--gvp-border)]" />
      <NuxtLink :to="`/channels/${channelId}`" role="menuitem" class="block rounded-lg px-3 py-2 text-xs text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]" @click="open = false">
        {{ ko ? '채널로 이동' : 'Visit channel' }}
      </NuxtLink>
      <button type="button" role="menuitem" class="w-full rounded-lg px-3 py-2 text-left text-xs text-[var(--gvp-live)] hover:bg-[var(--gvp-elevated)]" @click="unsubscribe">
        {{ ko ? '구독 취소' : 'Unsubscribe' }}
      </button>
    </div>
    <p v-if="failed" role="alert" class="mt-1 text-[11px] font-medium text-[var(--gvp-live)]">{{ ko ? '구독 변경에 실패했습니다.' : 'Subscription update failed.' }}</p>
  </div>
</template>
