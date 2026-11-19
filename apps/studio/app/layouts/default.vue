<script setup lang="ts">
import type { UserProfile } from '@gvp/shared-types'

const route = useRoute()
const api = useStudioApi()
const sessionUser = useState<UserProfile | null>('studio-session', () => null)
const mobileNavOpen = ref(false)
const collapsed = useState('studio-sidebar-collapsed', () => false)

watch(() => route.path, () => {
  mobileNavOpen.value = false
})

onMounted(async () => {
  if (sessionUser.value) return
  try {
    sessionUser.value = await api.getMe()
  } catch {
    sessionUser.value = null
  }
})

async function logout() {
  try {
    await $fetch('/auth/logout', { method: 'POST' })
  } catch {
    // Session cookies are cleared server-side; continue locally regardless.
  }
  const tokenCookie = useCookie('gvp_studio_token')
  tokenCookie.value = null
  sessionUser.value = null
  await navigateTo('/login')
}
</script>

<template>
  <div class="flex min-h-screen flex-col bg-[var(--gvp-surface)] text-[var(--gvp-text)]">
    <StudioHeader :session-user="sessionUser" @toggle-drawer="mobileNavOpen = !mobileNavOpen" @toggle-collapse="collapsed = !collapsed" @logout="logout" />
    <div v-if="mobileNavOpen" class="fixed inset-0 z-40 bg-[var(--gvp-scrim)] md:hidden" aria-hidden="true" @click="mobileNavOpen = false" />
    <div class="flex flex-1">
      <StudioSidebar :collapsed="collapsed" :mobile-open="mobileNavOpen" :session-user="sessionUser" />
      <main class="min-w-0 flex-1 bg-[var(--gvp-surface)] p-4 sm:p-6 lg:p-8">
        <div class="mx-auto max-w-7xl">
          <div v-if="!sessionUser && route.path !== '/login'" class="mb-6 flex items-center justify-between rounded-xl border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-4 py-2.5 text-xs text-[var(--gvp-warning)]">
            <div class="flex items-center gap-2">
              <span class="font-semibold">로그인이 필요합니다</span>
              <span aria-hidden="true">|</span>
              <span>실제 채널 데이터를 보려면 로그인하세요.</span>
            </div>
            <NuxtLink to="/login" class="rounded border border-[var(--gvp-warning)] bg-[var(--gvp-surface)] px-2.5 py-1 font-mono text-[11px] font-medium hover:bg-[var(--gvp-warning-soft)]">
              LOGIN
            </NuxtLink>
          </div>
          <slot />
          <footer class="mt-10 border-t border-[var(--gvp-border)] pt-4 text-center text-xs text-[var(--gvp-text-muted)]">
            <span>vravra studio</span>
          </footer>
        </div>
      </main>
    </div>
  </div>
</template>
