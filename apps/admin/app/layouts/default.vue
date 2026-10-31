<script setup lang="ts">
import { UserRole } from '@gvp/shared-types'

const route = useRoute()
const config = useRuntimeConfig()

const activeAdminRole = useState<UserRole>('admin-role', () => UserRole.SUPER_ADMIN)

const navLinks = [
  { path: '/', label: '플랫폼 통합 관제 (Operations)', icon: '🛰️' },
  { path: '/ads', label: '광고 송출 진단 (Ad Delivery)', icon: '🎯' },
  { path: '/users', label: '사용자 & RBAC 거버넌스 (Users)', icon: '👥' },
  { path: '/moderation', label: '신뢰 및 안전 큐 (Moderation)', icon: '🛡️' },
  { path: '/finance', label: '정산 승인 & 유닛 이코노믹스 (Finance)', icon: '🏦' },
  { path: '/policies', label: '정책 버전 & 킬스위치 (Policies)', icon: '⚙️' },
  { path: '/infrastructure', label: '멀티 리전 K8s & DR 관제 (Infra)', icon: '🌐' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- High-Security Super Admin Top Bar -->
    <header class="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-red-500/25 bg-[#10131C]/95 px-6 backdrop-blur">
      <div class="flex items-center gap-3">
        <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-red-600 font-black text-white shadow">
          A
        </span>
        <div>
          <span class="text-base font-extrabold text-white">Super Admin Governance Console</span>
          <span class="ml-2 rounded bg-amber-500/20 px-2 py-0.5 font-mono text-[10px] font-bold text-amber-400">
            DEMO CONSOLE PREVIEW
          </span>
        </div>
      </div>

      <div class="flex items-center gap-3 text-xs">
        <span class="text-neutral-400">Demo Role Preview:</span>
        <select
          v-model="activeAdminRole"
          class="rounded-xl border border-white/15 bg-[#090B10] px-3 py-1.5 font-mono text-xs font-bold text-emerald-400 outline-none"
        >
          <option :value="UserRole.SUPER_ADMIN">SUPER_ADMIN</option>
          <option :value="UserRole.PLATFORM_ADMIN">PLATFORM_ADMIN</option>
          <option :value="UserRole.FINANCE_ADMIN">FINANCE_ADMIN</option>
          <option :value="UserRole.MODERATOR">TRUST_SAFETY_MODERATOR</option>
          <option :value="UserRole.SUPPORT_AGENT">SUPPORT_AGENT</option>
        </select>
        <span class="rounded-lg bg-emerald-500/15 px-2.5 py-1 font-mono text-[11px] text-emerald-400">
          Read-Only Sandbox
        </span>
      </div>
    </header>

    <div class="flex flex-1">
      <aside class="w-64 shrink-0 border-r border-white/10 bg-[#10131C] p-4">
        <nav class="space-y-1">
          <NuxtLink
            v-for="item in navLinks"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-xs font-semibold transition"
            :class="
              route.path === item.path
                ? 'bg-red-600 text-white font-bold shadow'
                : 'text-neutral-300 hover:bg-white/10 hover:text-white'
            "
          >
            <span>{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </NuxtLink>
        </nav>

        <div class="mt-8 rounded-2xl border border-red-500/25 bg-red-950/20 p-3.5 text-[11px] text-red-200">
          <p class="font-bold">🔒 Immutable Audit Trail Active</p>
          <p class="mt-1 text-neutral-400">
            Every administrative mutation is cryptographically signed and appended to PostgreSQL 18 + ClickHouse audit logs.
          </p>
        </div>
      </aside>

            <main class="min-w-0 flex-1 p-6 lg:p-8">
        <div class="mb-6 flex items-center justify-between rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-2.5 text-xs text-amber-300">
          <div class="flex items-center gap-2">
            <span class="font-bold">⚠️ 어드민 관제 데모 샌드박스 (Demo Sandbox Mode)</span>
            <span class="text-neutral-400">|</span>
            <span class="text-neutral-300">실서비스 인증 게이트웨이 연동 전 아키텍처 시연용 프리뷰 화면입니다. 실제 제어 명령 및 킬스위치는 비활성화되어 있습니다.</span>
          </div>
          <span class="rounded bg-amber-500/20 px-2 py-0.5 font-mono text-[10px] text-amber-200">SIMULATION ONLY</span>
        </div>
        <slot />
        <footer class="mt-12 border-t border-white/5 pt-4 text-center text-[11px] text-neutral-400">
          <span>StreamVerse Admin Console · Commit <span class="font-mono text-neutral-300">3f8c1ef</span> · Architecture Prototype</span>
        </footer>
      </main>
    </div>
  </div>
</template>
