/** First currently-live stream id for drawer / explore navigation. Null when none. */
export function useDrawerLive() {
  const api = useApi()
  const liveId = ref<string | null>(null)

  async function load() {
    try {
      const res = await api.listLiveStreams(1, 10)
      liveId.value = (res.items || []).find(s => s.status === 'live')?.id || null
    } catch { liveId.value = null }
  }

  onMounted(load)
  return { liveId, reload: load }
}
