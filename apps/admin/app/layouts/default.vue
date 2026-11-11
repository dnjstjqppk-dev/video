<script setup lang="ts">
import { UserRole } from '@gvp/shared-types'

const route = useRoute()
const config = useRuntimeConfig()

const activeAdminRole = useState<UserRole>('admin-role', () => UserRole.SUPER_ADMIN)

const navLinks = [
  { path: '/', label: '플랫폼 통합 관제 (Operations)', icon: 'M3 3h7v7H3V3zm11 0h7v7h-7V3zm0 11h7v7h-7v-7zM3 14h7v7H3v-7z' },
  { path: '/ads', label: '광고 송출 진단 (Ad Delivery)', icon: 'M3 20h18M6 16v-4m4 4V8m4 8v-6m4 6V4' },
  { path: '/users', label: '사용자 & RBAC 거버넌스 (Users)', icon: 'M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2m8-10a4 4 0 100-8 4 4 0 000 8z' },
  { path: '/moderation', label: '신뢰 및 안전 큐 (Moderation)', icon: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' },
  { path: '/finance', label: '정산 승인 & 유닛 이코노믹스 (Finance)', icon: 'M12 2v20M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6' },
  { path: '/policies', label: '정책 버전 & 킬스위치 (Policies)', icon: 'M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4' },
  { path: '/infrastructure', label: '멀티 리전 K8s & DR 관제 (Infra)', icon: 'M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col bg-white text-slate-900">
    <!-- High-Security Super Admin Top Bar -->
    <header class="sticky top-0 z-40 flex h-14 items-center justify-between border-b border-slate-200 bg-white px-4 sm:px-6">
      <div class="flex items-center gap-2.5">
        <span class="flex h-7 w-7 items-center justify-center rounded-md bg-slate-900 text-xs font-bold text-white">
          A
        </span>
        <div class="flex items-center gap-2">
          <span class="text-sm font-bold tracking-tight text-slate-900">Super Admin Governance Console</span>
          <span class="rounded-md border border-amber-200 bg-amber-50 px-2 py-0.5 font-mono text-[10px] font-semibold text-amber-800">
            DEMO CONSOLE PREVIEW
          </span>
        </div>
      </div>

      <div class="flex items-center gap-2.5 text-xs">
        <span class="text-slate-500">Demo Role Preview:</span>
        <select
          v-model="activeAdminRole"
          class="h-8 rounded-md border border-slate-300 bg-white px-2.5 font-mono text-xs font-medium text-slate-900 focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
        >
          <option :value="UserRole.SUPER_ADMIN">SUPER_ADMIN</option>
          <option :value="UserRole.PLATFORM_ADMIN">PLATFORM_ADMIN</option>
          <option :value="UserRole.FINANCE_ADMIN">FINANCE_ADMIN</option>
          <option :value="UserRole.MODERATOR">TRUST_SAFETY_MODERATOR</option>
          <option :value="UserRole.SUPPORT_AGENT">SUPPORT_AGENT</option>
        </select>
        <span class="rounded-md border border-green-200 bg-green-50 px-2.5 py-1 font-mono text-[11px] font-medium text-green-700">
          Read-Only Sandbox
        </span>
      </div>
    </header>

    <div class="flex flex-1">
      <aside class="w-60 shrink-0 border-r border-slate-200 bg-slate-50 p-3">
        <nav class="space-y-0.5">
          <NuxtLink
            v-for="item in navLinks"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-2.5 rounded-md px-3 py-2 text-xs font-medium transition-colors"
            :class="
              route.path === item.path
                ? 'bg-slate-900 text-white font-semibold'
                : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
            "
          >
            <svg class="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" :d="item.icon" />
            </svg>
            <span class="truncate">{{ item.label }}</span>
          </NuxtLink>
        </nav>

        <div class="mt-6 rounded-lg border border-slate-200 bg-white p-3 text-[11px] text-slate-600">
          <p class="font-semibold text-slate-900">Immutable Audit Trail Active</p>
          <p class="mt-1 text-slate-500 leading-relaxed">
            Every administrative mutation is cryptographically signed and appended to PostgreSQL 18 + ClickHouse audit logs.
          </p>
        </div>
      </aside>

      <main class="min-w-0 flex-1 bg-white p-4 sm:p-6 lg:p-8">
        <div class="mx-auto max-w-7xl">
          <div class="mb-6 flex items-center justify-between rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-xs text-amber-800">
            <div class="flex items-center gap-2">
              <span class="font-semibold">어드민 관제 데모 샌드박스 (Demo Sandbox Mode)</span>
              <span class="text-amber-400">|</span>
              <span>실서비스 인증 게이트웨이 연동 전 아키텍처 시연용 프리뷰 화면입니다. 실제 제어 명령 및 킬스위치는 비활성화되어 있습니다.</span>
            </div>
            <span class="rounded border border-amber-300 bg-white px-2 py-0.5 font-mono text-[10px] font-semibold text-amber-800">SIMULATION ONLY</span>
          </div>
          <slot />
          <footer class="mt-10 border-t border-slate-200 pt-4 text-center text-xs text-slate-400">
            <span>StreamVerse Admin Console · Commit <span class="font-mono text-slate-600">3f8c1ef</span> · Architecture Prototype</span>
          </footer>
        </div>
      </main>
    </div>
  </div>
</template>
