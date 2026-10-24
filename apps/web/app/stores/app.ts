import { defineStore } from 'pinia'
import type { UserProfile } from '@gvp/shared-types'
import { SubscriptionTier, UserRole } from '@gvp/shared-types'

export const useAppStore = defineStore('gvp-web-app', () => {
  const sidebarCollapsed = ref(false)
  const locale = ref<'ko' | 'en'>('ko')
  const currency = ref<'KRW' | 'USD'>('KRW')
  const theme = ref<'dark' | 'light'>('dark')
  const sessionId = ref(`sess_${Math.random().toString(36).slice(2, 11)}`)

  const currentUser = ref<UserProfile | null>({
    id: 'usr_demo_01',
    email: 'viewer@gvp.media',
    username: 'kim_viewer',
    displayName: '김민준 (Minjun Kim)',
    avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=120&q=80',
    role: UserRole.REGISTERED_USER,
    subscriptionTier: SubscriptionTier.FREE,
    isVerified: true,
    twoFactorEnabled: true,
    subscriberCount: 412,
    totalViews: 15820,
    createdAt: '2025-01-10T00:00:00Z',
    updatedAt: '2025-05-01T00:00:00Z'
  })

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

  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
    if (import.meta.client) {
      document.documentElement.classList.toggle('dark', theme.value === 'dark')
    }
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
    theme,
    sessionId,
    currentUser,
    watchProgressMap,
    hiddenVideoIds,
    blockedChannelIds,
    toggleSidebar,
    toggleLocale,
    toggleCurrency,
    toggleTheme,
    saveWatchProgress,
    getWatchProgress,
    formatCurrency
  }
})
