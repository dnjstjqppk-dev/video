import { defineStore } from 'pinia'
import type { Video } from '@gvp/shared-types'

export const useLibraryStore = defineStore('viewer-library', () => {
  const history = ref<Video[]>([])
  const saved = ref<Video[]>([])
  const liked = ref<Video[]>([])
  const queue = ref<Video[]>([])
  const subscriptions = ref<string[]>([])
  const paused = ref(false)
  const ready = ref(false)
  const storageError = ref(false)
  function hydrate() {
    if (!import.meta.client || ready.value) return
    try {
      const data = JSON.parse(localStorage.getItem('gvp_library_v1') || '{}')
      const videos = (value: unknown): Video[] => Array.isArray(value)
        ? value.filter(v => v && typeof v.id === 'string' && typeof v.title === 'string' && typeof v.channelId === 'string').slice(0, 200) : []
      history.value = videos(data.history)
      saved.value = videos(data.saved)
      liked.value = videos(data.liked)
      queue.value = videos(data.queue)
      subscriptions.value = Array.isArray(data.subscriptions) ? data.subscriptions.filter((v: unknown) => typeof v === 'string') : []
      paused.value = data.paused === true
    } catch { storageError.value = true }
    ready.value = true
  }
  function persist() {
    if (!import.meta.client) return
    try {
      localStorage.setItem('gvp_library_v1', JSON.stringify({ history: history.value, saved: saved.value, liked: liked.value, queue: queue.value, subscriptions: subscriptions.value, paused: paused.value }))
      storageError.value = false
    } catch { storageError.value = true }
  }
  function record(video: Video) {
    hydrate()
    if (paused.value) return
    history.value = [video, ...history.value.filter(v => v.id !== video.id)].slice(0, 200)
    persist()
  }
  function toggleSaved(video: Video) {
    hydrate()
    saved.value = saved.value.some(v => v.id === video.id) ? saved.value.filter(v => v.id !== video.id) : [video, ...saved.value].slice(0, 200)
    persist()
  }
  function enqueue(video: Video) {
    hydrate()
    if (queue.value.some(v => v.id === video.id)) return false
    queue.value = [...queue.value, video].slice(0, 200)
    persist()
    return true
  }
  function dequeue(id: string) {
    hydrate()
    queue.value = queue.value.filter(v => v.id !== id)
    persist()
  }
  function shiftQueue(): Video | undefined {
    hydrate()
    const next = queue.value[0]
    queue.value = queue.value.slice(1)
    persist()
    return next
  }
  function setLiked(video: Video, value: boolean) {
    hydrate()
    liked.value = liked.value.filter(v => v.id !== video.id)
    if (value) liked.value.unshift(video)
    liked.value = liked.value.slice(0, 200)
    persist()
  }
  function setSubscribed(id: string, value: boolean) {
    hydrate()
    subscriptions.value = subscriptions.value.filter(item => item !== id)
    if (value) subscriptions.value.push(id)
    persist()
  }
  function remove(kind: 'history' | 'saved' | 'liked', id: string) {
    hydrate()
    const target = { history, saved, liked }[kind]
    target.value = target.value.filter(v => v.id !== id)
    persist()
  }
  function clearHistory() { hydrate(); history.value = []; persist() }
  function togglePaused() { hydrate(); paused.value = !paused.value; persist() }
  return { history, saved, liked, queue, subscriptions, paused, ready, storageError, hydrate, record, toggleSaved, enqueue, dequeue, shiftQueue, setLiked, setSubscribed, remove, clearHistory, togglePaused }
})
