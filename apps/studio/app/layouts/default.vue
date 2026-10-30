<script setup lang="ts">
const route = useRoute()
const config = useRuntimeConfig()

const currency = useState<'KRW' | 'USD'>('studio-currency', () => 'KRW')
const selectedRole = useState<string>('studio-role', () => 'Channel Owner (Full Access)')

const navLinks = [
  { path: '/', label: '대시보드 (Dashboard)', icon: '📊' },
  { path: '/content', label: '콘텐츠 관리 (Content)', icon: '🎬' },
  { path: '/upload', label: '멀티파트 업로드 (Upload)', icon: '⬆️' },
  { path: '/editor/vid_arch_2026', label: '타임라인 에디터 (Editor)', icon: '✂️' },
  { path: '/analytics', label: '채널 분석 (Analytics)', icon: '📈' },
  { path: '/monetization', label: '수익 창출 & 정산 (Earn)', icon: '💰' },
  { path: '/customization', label: '채널 맞춤설정 (Customization)', icon: '🎨' },
  { path: '/comments', label: '댓글 및 검토 (Comments)', icon: '💬' },
  { path: '/subtitles', label: '자막 관리 (Subtitles)', icon: '📝' },
  { path: '/copyright', label: '저작권 및 이의제기 (Copyright)', icon: '⚖️' },
  { path: '/live', label: '라이브 관제실 (Live Room)', icon: '📡' }
]
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- Top Studio Header -->
    <header class="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-white/10 bg-[#12151E]/95 px-6 backdrop-blur">
      <div class="flex items-center gap-4">
        <NuxtLink to="/" class="flex items-center gap-2.5">
          <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-[#FF0033] font-black text-white">
            S
          </span>
          <div>
            <span class="text-base font-extrabold tracking-tight text-white">Creator Studio</span>
            <span class="ml-2 rounded bg-emerald-500/20 px-1.5 py-0.5 text-[10px] font-bold text-emerald-400">
              PARTNER VERIFIED
            </span>
          </div>
        </NuxtLink>
      </div>

      <div class="flex items-center gap-3">
        <!-- Channel Role Selector (Owner, Manager, Editor, Analyst) -->
        <select
          v-model="selectedRole"
          class="rounded-xl border border-white/15 bg-[#1B1F2B] px-3 py-1.5 text-xs font-medium text-neutral-200 outline-none"
        >
          <option>Channel Owner (Full Access)</option>
          <option>Channel Manager</option>
          <option>Channel Video Editor</option>
          <option>Channel Analyst (Read-Only)</option>
        </select>

        <!-- Currency Switcher -->
        <button
          type="button"
          class="rounded-xl border border-white/15 bg-[#1B1F2B] px-3 py-1.5 text-xs font-bold text-[#3EA6FF] hover:bg-white/10"
          @click="currency = currency === 'KRW' ? 'USD' : 'KRW'"
        >
          {{ currency === 'KRW' ? '₩ KRW' : '$ USD' }}
        </button>

        <NuxtLink
          to="/upload"
          class="flex items-center gap-1.5 rounded-xl bg-[#FF0033] px-4 py-2 text-xs font-bold text-white shadow transition hover:bg-[#D9002B]"
        >
          <span>＋</span>
          <span>새 동영상 업로드 (Upload)</span>
        </NuxtLink>

        <a
          :href="config.public.webUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="rounded-xl border border-white/15 px-3 py-2 text-xs font-semibold text-neutral-300 hover:bg-white/10"
        >
          Public Channel ↗
        </a>
      </div>
    </header>

    <div class="flex flex-1">
      <!-- Studio Left Navigation Sidebar -->
      <aside class="w-64 shrink-0 border-r border-white/10 bg-[#12151E] p-4">
        <!-- Creator Channel Badge -->
        <div class="mb-6 flex items-center gap-3 rounded-2xl border border-white/10 bg-[#181C28] p-3">
          <img
            src="https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=120&q=80"
            alt="GVP Tech Engineering"
            class="h-11 w-11 rounded-full object-cover ring-2 ring-[#FF0033]"
          />
          <div class="min-w-0">
            <p class="truncate text-xs font-bold text-white">GVP Tech Engineering</p>
            <p class="text-[11px] text-neutral-400">284,000 subscribers</p>
          </div>
        </div>

        <nav class="space-y-1">
          <NuxtLink
            v-for="item in navLinks"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-xs font-semibold transition"
            :class="
              route.path === item.path
                ? 'bg-[#FF0033] text-white shadow'
                : 'text-neutral-300 hover:bg-white/10 hover:text-white'
            "
          >
            <span class="text-base">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </NuxtLink>
        </nav>

        <div class="mt-8 rounded-2xl border border-white/10 bg-[#181C28] p-3.5 text-xs">
          <p class="font-bold text-white">Subtitles & Copyright Status</p>
          <p class="mt-1 text-[11px] text-emerald-400">✓ 0 Active Copyright Strikes</p>
          <p class="mt-0.5 text-[11px] text-neutral-400">✓ Auto-captions (ko/en) enabled</p>
        </div>
      </aside>

      <!-- Main Studio Workspace -->
      <main class="min-w-0 flex-1 p-6 lg:p-8">
        <slot />
      </main>
    </div>
  </div>
</template>
