<script setup lang="ts">
const props = defineProps<{ channelId: string; initialCount: number }>()
const emit = defineEmits<{ (e: 'change', v: { subscribed: boolean; count: number }): void }>()
const api = useApi()
const ko = computed(() => useAppStore().locale === 'ko')
const subscribed = ref(false)
const count = ref(props.initialCount)
const open = ref(false)
const failed = ref(false)

watch(() => props.initialCount, v => { count.value = v })
watch(() => props.channelId, () => { subscribed.value = false; void sync() })

async function sync() {
  try {
    const feed = await api.getSubscriptionsFeed(1, 100)
    subscribed.value = (feed.items || []).some(v => v.channelId === props.channelId)
  } catch { /* stays unsubscribed */ }
}

async function subscribe() {
  failed.value = false
  try {
    const res = await api.subscribeChannel(props.channelId)
    subscribed.value = res.subscribed
    count.value = res.subscriberCount
    emit('change', { subscribed: true, count: count.value })
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
    <div v-if="open && subscribed" role="menu" class="gvp-dropdown gvp-elevated absolute left-0 top-full z-30 mt-2 w-48 rounded-xl p-1.5">
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
