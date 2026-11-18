import type { Channel, LiveStream } from '@gvp/shared-types'

export interface SubChannel { id: string; name: string; handle: string; avatarUrl?: string; isLive: boolean }

/** Sidebar subscriptions derived from real API feed + live list. */
export function useSubscriptionChannels() {
  const api = useApi()
  const channels = ref<SubChannel[]>([])
  const pending = ref(true)

  async function load() {
    pending.value = true
    try {
      const [feed, live] = await Promise.all([
        api.getSubscriptionsFeed(1, 24).catch(() => ({ items: [] as never[] })),
        api.listLiveStreams(1, 20).catch(() => ({ items: [] as LiveStream[] })),
      ])
      const liveSet = new Set((live.items || []).filter(s => s.status === 'live').map(s => s.channelId))
      const seen = new Map<string, SubChannel>()
      for (const v of (feed.items || []) as Array<{ channelId: string; channel?: Channel }>) {
        if (!v.channel || seen.has(v.channelId)) continue
        seen.set(v.channelId, {
          id: v.channelId, name: v.channel.name, handle: v.channel.handle,
          avatarUrl: v.channel.avatarUrl, isLive: liveSet.has(v.channelId),
        })
      }
      channels.value = [...seen.values()].slice(0, 12)
    } catch { channels.value = [] } finally { pending.value = false }
  }

  onMounted(load)
  return { channels, pending, reload: load }
}
