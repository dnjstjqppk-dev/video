<script setup lang="ts">
const route = useRoute()

export type PartnerTenantType = 'ADVERTISER' | 'MCN' | 'RIGHTS_OWNER'

const activeTenant = useState<PartnerTenantType>('partner-tenant', () => 'ADVERTISER')

const tenantOrganizations: Record<PartnerTenantType, { id: string; name: string; badge: string }> = {
  ADVERTISER: {
    id: 'org_adv_samsung',
    name: 'Samsung Electronics Global Media Buying',
    badge: 'Enterprise Advertiser'
  },
  MCN: {
    id: 'org_mcn_sandbox',
    name: 'K-Creator MCN Alliance (42 Channels)',
    badge: 'Certified MCN Partner'
  },
  RIGHTS_OWNER: {
    id: 'org_rights_cjenm',
    name: 'CJ Music & Cinema Rights Holder Corp.',
    badge: 'Content ID Rights Owner'
  }
}

const navItems = [
  { path: '/', label: '파트너 종합 현황 (Overview)', icon: '🏢' },
  { path: '/campaigns', label: '광고 캠페인 콘솔 (Advertiser)', icon: '📣' },
  { path: '/mcn', label: 'MCN 소속 채널 관리 (MCN)', icon: '🤝' },
  { path: '/rights', label: '저작권 & 핑거프린트 (Rights ID)', icon: '🛡️' },
  { path: '/settlements', label: '복식부기 정산서 (Settlements)', icon: '🧾' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- Top Multi-Tenant Partner Header -->
    <header class="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-white/10 bg-[#111625]/95 px-6 backdrop-blur">
      <div class="flex items-center gap-3">
        <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-[#3EA6FF] font-black text-black">
          P
        </span>
        <div>
          <span class="text-base font-extrabold text-white">StreamVerse Partner Console</span>
          <span class="ml-2 rounded bg-[#3EA6FF]/20 px-2 py-0.5 text-[10px] font-bold text-[#3EA6FF]">
            {{ tenantOrganizations[activeTenant].badge }}
          </span>
        </div>
      </div>

      <!-- Multi-Tenant Organization Switcher -->
      <div class="flex items-center gap-3">
        <span class="text-xs text-neutral-400">Tenant Organization:</span>
        <div class="flex rounded-xl border border-white/15 bg-[#0B0E17] p-1 text-xs font-bold">
          <button
            type="button"
            class="rounded-lg px-3 py-1 transition"
            :class="activeTenant === 'ADVERTISER' ? 'bg-[#3EA6FF] text-black' : 'text-neutral-400 hover:text-white'"
            @click="activeTenant = 'ADVERTISER'"
          >
            Advertiser
          </button>
          <button
            type="button"
            class="rounded-lg px-3 py-1 transition"
            :class="activeTenant === 'MCN' ? 'bg-[#3EA6FF] text-black' : 'text-neutral-400 hover:text-white'"
            @click="activeTenant = 'MCN'"
          >
            MCN Agency
          </button>
          <button
            type="button"
            class="rounded-lg px-3 py-1 transition"
            :class="activeTenant === 'RIGHTS_OWNER' ? 'bg-[#3EA6FF] text-black' : 'text-neutral-400 hover:text-white'"
            @click="activeTenant = 'RIGHTS_OWNER'"
          >
            Rights Owner
          </button>
        </div>
      </div>
    </header>

    <div class="flex flex-1">
      <!-- Left Sidebar -->
      <aside class="w-64 shrink-0 border-r border-white/10 bg-[#111625] p-4">
        <div class="mb-6 rounded-2xl border border-white/10 bg-[#171E31] p-3.5 text-xs">
          <p class="text-[10px] font-bold uppercase tracking-wider text-[#3EA6FF]">Active Organization</p>
          <p class="mt-1 font-bold text-white">{{ tenantOrganizations[activeTenant].name }}</p>
          <p class="mt-0.5 font-mono text-[11px] text-neutral-400">ID: {{ tenantOrganizations[activeTenant].id }}</p>
        </div>

        <nav class="space-y-1">
          <NuxtLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-xs font-semibold transition"
            :class="
              route.path === item.path
                ? 'bg-[#3EA6FF] text-black font-bold shadow'
                : 'text-neutral-300 hover:bg-white/10 hover:text-white'
            "
          >
            <span>{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </NuxtLink>
        </nav>
      </aside>

      <!-- Main Partner Content -->
      <main class="min-w-0 flex-1 p-6 lg:p-8">
        <slot />
      </main>
    </div>
  </div>
</template>
