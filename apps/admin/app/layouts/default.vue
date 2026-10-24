<script setup lang="ts">
import { UserRole } from '@gvp/shared-types'

const route = useRoute()

const activeAdminRole = useState<UserRole>('admin-role', () => UserRole.SUPER_ADMIN)

const navLinks = [
  { path: '/', label: '플랫폼 통합 관제 (Operations)', icon: '🛰️' },
  { path: '/users', label: '사용자 & RBAC 거버넌스 (Users)', icon: '👥' },
  { path: '/moderation', label: '신뢰 및 안전 큐 (Moderation)', icon: '🛡️' },
  { path: '/finance', label: '정산 승인 & 유닛 이코노믹스 (Finance)', icon: '🏦' },
  { path: '/policies', label: '정책 버전 & 킬스위치 (Policies)', icon: '⚙️' }
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
          <span class="ml-2 rounded bg-red-500/20 px-2 py-0.5 font-mono text-[10px] font-bold text-red-400">
            MFA + IP ALLOWLIST ENFORCED
          </span>
        </div>
      </div>

      <div class="flex items-center gap-3 text-xs">
        <span class="text-neutral-400">Active Admin Role:</span>
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
          Session TTL: 14m 52s
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
        <slot />
      </main>
    </div>
  </div>
</template>
