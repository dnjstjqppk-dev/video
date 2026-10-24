<script setup lang="ts">
import { VideoProcessingStatus, VideoVisibility } from '@gvp/shared-types'

useSeoMeta({
  title: 'Content Manager — Videos, Shorts, Live & Playlists'
})

interface StudioContentRow {
  id: string
  title: string
  type: 'video' | 'shorts' | 'live' | 'playlist'
  visibility: VideoVisibility
  status: VideoProcessingStatus
  copyright: 'None' | 'Revenue Share' | 'Blocked'
  monetized: boolean
  views: number
  comments: number
  selected: boolean
}

const activeTab = ref<'all' | 'video' | 'shorts' | 'live' | 'playlist'>('all')

const items = ref<StudioContentRow[]>([
  {
    id: 'vid_arch_2026',
    title: 'Building a 10M Concurrent Viewer Global Video Platform with Go 1.24, Nuxt 4 & ClickHouse',
    type: 'video',
    visibility: VideoVisibility.PUBLIC,
    status: VideoProcessingStatus.READY,
    copyright: 'None',
    monetized: true,
    views: 412900,
    comments: 1842,
    selected: false
  },
  {
    id: 'short_01',
    title: '60초 만에 이해하는 HLS ABR 화질 자동 전환 원리 ⚡',
    type: 'shorts',
    visibility: VideoVisibility.PUBLIC,
    status: VideoProcessingStatus.READY,
    copyright: 'None',
    monetized: true,
    views: 1420000,
    comments: 912,
    selected: false
  },
  {
    id: 'vid_seoul_4k',
    title: '[4K 60fps HDR] Cyberpunk Rainy Night Walk in Gangnam & Hongdae, Seoul 2026',
    type: 'video',
    visibility: VideoVisibility.PUBLIC,
    status: VideoProcessingStatus.READY,
    copyright: 'Revenue Share',
    monetized: true,
    views: 1294000,
    comments: 4510,
    selected: false
  },
  {
    id: 'live_main_01',
    title: '[LIVE VOD] 2026 Global Architecture Summit — Keynote & Live Q&A',
    type: 'live',
    visibility: VideoVisibility.MEMBERS_ONLY,
    status: VideoProcessingStatus.READY,
    copyright: 'None',
    monetized: true,
    views: 84200,
    comments: 3120,
    selected: false
  }
])

const filteredItems = computed(() =>
  items.value.filter(i => (activeTab.value === 'all' ? true : i.type === activeTab.value))
)

const selectedCount = computed(() => items.value.filter(i => i.selected).length)

function applyBulkVisibility(vis: VideoVisibility) {
  items.value.forEach(item => {
    if (item.selected) item.visibility = vis
  })
}

function toggleBulkMonetization(enable: boolean) {
  items.value.forEach(item => {
    if (item.selected) item.monetized = enable
  })
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 콘텐츠 관리 (Content Manager)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Manage visibility, monetization, subtitles, and copyright claims across Videos, Shorts, Live VODs, and Playlists
        </p>
      </div>

      <!-- Bulk Actions Toolbar -->
      <div v-if="selectedCount > 0" class="flex items-center gap-2 rounded-xl border border-[#3EA6FF]/40 bg-[#182330] px-4 py-2 text-xs">
        <span class="font-bold text-white">{{ selectedCount }} selected</span>
        <button
          type="button"
          class="rounded-lg bg-white/10 px-2.5 py-1 text-white hover:bg-white/20"
          @click="applyBulkVisibility(VideoVisibility.PUBLIC)"
        >
          Set Public
        </button>
        <button
          type="button"
          class="rounded-lg bg-white/10 px-2.5 py-1 text-white hover:bg-white/20"
          @click="applyBulkVisibility(VideoVisibility.PRIVATE)"
        >
          Set Private
        </button>
        <button
          type="button"
          class="rounded-lg bg-emerald-500/20 px-2.5 py-1 font-bold text-emerald-300"
          @click="toggleBulkMonetization(true)"
        >
          Enable Monetization
        </button>
      </div>
    </div>

    <!-- Filter Tabs -->
    <div class="mt-6 flex gap-4 border-b border-white/10 text-xs font-bold">
      <button
        v-for="tab in (['all', 'video', 'shorts', 'live', 'playlist'] as const)"
        :key="tab"
        type="button"
        class="border-b-2 pb-3 uppercase transition"
        :class="activeTab === tab ? 'border-[#FF0033] text-white' : 'border-transparent text-neutral-400 hover:text-white'"
        @click="activeTab = tab"
      >
        {{ tab }}
      </button>
    </div>

    <!-- Content Table -->
    <div class="mt-4 overflow-x-auto rounded-2xl border border-white/10 bg-[#151822]">
      <table class="w-full text-left text-xs">
        <thead class="border-b border-white/10 bg-white/5 uppercase text-neutral-400">
          <tr>
            <th class="p-3.5">Select</th>
            <th class="p-3.5">Video / Content</th>
            <th class="p-3.5">Visibility</th>
            <th class="p-3.5">Monetization</th>
            <th class="p-3.5">Copyright</th>
            <th class="p-3.5">Views</th>
            <th class="p-3.5">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-white/10">
          <tr v-for="row in filteredItems" :key="row.id" class="hover:bg-white/5">
            <td class="p-3.5">
              <input v-model="row.selected" type="checkbox" class="rounded accent-[#FF0033]" />
            </td>
            <td class="p-3.5">
              <p class="font-bold text-white">{{ row.title }}</p>
              <span class="mt-0.5 inline-block rounded bg-white/10 px-1.5 py-0.5 text-[10px] uppercase text-neutral-300">
                {{ row.type }} · {{ row.status }}
              </span>
            </td>
            <td class="p-3.5">
              <select
                v-model="row.visibility"
                class="rounded-lg border border-white/15 bg-[#0D0F14] px-2.5 py-1 text-xs text-white"
              >
                <option :value="VideoVisibility.PUBLIC">Public</option>
                <option :value="VideoVisibility.UNLISTED">Unlisted</option>
                <option :value="VideoVisibility.PRIVATE">Private</option>
                <option :value="VideoVisibility.MEMBERS_ONLY">Members Only</option>
              </select>
            </td>
            <td class="p-3.5">
              <button
                type="button"
                class="rounded-full px-3 py-1 font-bold"
                :class="row.monetized ? 'bg-emerald-500/20 text-emerald-400' : 'bg-neutral-700 text-neutral-300'"
                @click="row.monetized = !row.monetized"
              >
                {{ row.monetized ? '$ ON' : 'OFF' }}
              </button>
            </td>
            <td class="p-3.5">
              <span
                class="rounded px-2 py-0.5 font-semibold"
                :class="row.copyright === 'None' ? 'text-emerald-400' : 'bg-amber-500/20 text-amber-300'"
              >
                {{ row.copyright }}
              </span>
            </td>
            <td class="p-3.5 font-mono text-white">{{ row.views.toLocaleString() }}</td>
            <td class="p-3.5">
              <NuxtLink
                :to="`/editor/${row.id}`"
                class="rounded-lg bg-[#3EA6FF]/20 px-3 py-1.5 font-bold text-[#3EA6FF] hover:bg-[#3EA6FF]/30"
              >
                Timeline Editor
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
