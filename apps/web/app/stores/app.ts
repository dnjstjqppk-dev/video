import { defineStore } from 'pinia'
import type { UserProfile } from '@gvp/shared-types'
import type { ApiClient } from '@gvp/ts-api-client'

export const useAppStore = defineStore('gvp-web-app', () => {
  const sidebarCollapsed = ref(false)
  const locale = ref<'ko' | 'en'>('ko')
  const currency = ref<'KRW' | 'USD'>('KRW')
  const sessionId = ref(`sess_${Math.random().toString(36).slice(2, 11)}`)

  const currentUser = ref<UserProfile | null>(null)
  const authLoading = ref(false)
  const authError = ref<string | null>(null)

  async function fetchSession(api: ApiClient) {
    authLoading.value = true
    authError.value = null
    try {
      currentUser.value = await api.getMe()
    } catch (err: any) {
      currentUser.value = null
      authError.value = err?.message || 'Session lookup failed'
    } finally {
      authLoading.value = false
    }
  }

  function clearSession() {
    currentUser.value = null
    authError.value = null
  }

  const watchProgressMap = ref<Record<string, number>>({})
  const hiddenVideoIds = ref<Set<string>>(new Set())
  const blockedChannelIds = ref<Set<string>>(new Set())

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function toggleLocale() {
    locale.value = locale.value === 'ko' ? 'en' : 'ko'
  }

  function toggleCurrency() {
    currency.value = currency.value === 'KRW' ? 'USD' : 'KRW'
  }

  function saveWatchProgress(videoId: string, seconds: number) {
    watchProgressMap.value[videoId] = Math.floor(seconds)
    if (import.meta.client) {
      try {
        localStorage.setItem('gvp_watch_progress', JSON.stringify(watchProgressMap.value))
      } catch {
        // Ignore storage quota errors
      }
    }
  }

  function getWatchProgress(videoId: string): number {
    if (import.meta.client && Object.keys(watchProgressMap.value).length === 0) {
      try {
        const raw = localStorage.getItem('gvp_watch_progress')
        if (raw) {
          watchProgressMap.value = JSON.parse(raw)
        }
      } catch {
        // Ignore parse errors
      }
    }
    return watchProgressMap.value[videoId] || 0
  }

  function formatCurrency(krwAmount: number): string {
    if (currency.value === 'KRW') {
      return `₩${krwAmount.toLocaleString('ko-KR')}`
    }
    const usd = (krwAmount / 1350).toFixed(2)
    return `$${usd}`
  }

  return {
    sidebarCollapsed,
    locale,
    currency,
    sessionId,
    currentUser,
    authLoading,
    authError,
    fetchSession,
    clearSession,
    watchProgressMap,
    hiddenVideoIds,
    blockedChannelIds,
    toggleSidebar,
    toggleLocale,
    toggleCurrency,
    saveWatchProgress,
    getWatchProgress,
    formatCurrency
  }
})
