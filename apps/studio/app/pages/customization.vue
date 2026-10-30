<script setup lang="ts">
useSeoMeta({
  title: 'Channel Customization — Layout, Branding & Basic Info'
})

const activeTab = ref<'layout' | 'branding' | 'info'>('layout')

const channelName = ref('GVP Tech Engineering')
const channelHandle = ref('@gvptech')
const description = ref(
  'Official engineering channel of StreamVerse Global Video Platform. Sharing deep dives on Go 1.24, Nuxt 4, Flutter 3, OpenSearch, and distributed HLS transcoding.'
)
const contactEmail = ref('engineering@gvp.media')
const bannerUrl = ref('https://images.unsplash.com/photo-1518770660439-4636190af475?w=1600&q=80')
const avatarUrl = ref('https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&q=80')
const trailerVideoId = ref('vid_arch_2026')
const featuredVideoId = ref('vid_ai_agents')

const links = ref([
  { title: 'Official Docs & GitHub', url: 'https://github.com/dnjstjqppk-dev/video' },
  { title: 'Technical Architecture Blog', url: 'https://blog.gvp.media' }
])

const saveSuccess = ref(false)

function saveCustomization() {
  saveSuccess.value = true
  setTimeout(() => {
    saveSuccess.value = false
  }, 2500)
}
</script>

<template>
  <div class="max-w-5xl">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 맞춤설정 (Channel Customization)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Customize channel homepage layout, branding assets, custom handle, and spotlight trailers
        </p>
      </div>

      <button
        type="button"
        class="rounded-xl bg-[#FF0033] px-5 py-2 text-xs font-bold text-white shadow transition hover:bg-[#D9002B]"
        @click="saveCustomization"
      >
        {{ saveSuccess ? '✓ Saved!' : '게시 (Publish Changes)' }}
      </button>
    </div>

    <!-- Navigation Tabs -->
    <div class="mt-6 flex gap-6 border-b border-white/10 text-sm font-semibold">
      <button
        type="button"
        class="border-b-2 pb-3 transition"
        :class="activeTab === 'layout' ? 'border-white text-white' : 'border-transparent text-neutral-400 hover:text-white'"
        @click="activeTab = 'layout'"
      >
        레이아웃 (Layout)
      </button>
      <button
        type="button"
        class="border-b-2 pb-3 transition"
        :class="activeTab === 'branding' ? 'border-white text-white' : 'border-transparent text-neutral-400 hover:text-white'"
        @click="activeTab = 'branding'"
      >
        브랜딩 (Branding)
      </button>
      <button
        type="button"
        class="border-b-2 pb-3 transition"
        :class="activeTab === 'info' ? 'border-white text-white' : 'border-transparent text-neutral-400 hover:text-white'"
        @click="activeTab = 'info'"
      >
        기본 정보 (Basic Info)
      </button>
    </div>

    <!-- Tab 1: Layout -->
    <div v-if="activeTab === 'layout'" class="mt-6 space-y-6">
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="text-sm font-bold text-white">동영상 스포트라이트 (Video Spotlight)</h2>
        <p class="mt-1 text-xs text-neutral-400">
          Add a video to the top of your channel homepage for specific audience segments
        </p>

        <div class="mt-4 space-y-4">
          <div class="flex items-center justify-between rounded-xl border border-white/10 bg-white/5 p-4 text-xs">
            <div>
              <p class="font-bold text-white">비구독자용 채널 예고편 (Channel trailer for people who haven't subscribed)</p>
              <p class="mt-0.5 text-neutral-400">Selected: Building a 10M Concurrent Viewer Global Video Platform</p>
            </div>
            <select v-model="trailerVideoId" class="rounded-lg border border-white/15 bg-[#1B1F2B] px-3 py-1.5 text-xs text-white">
              <option value="vid_arch_2026">Building a 10M Concurrent Viewer Global Video Platform</option>
              <option value="vid_ai_agents">Multimodal Video Indexing with Audio-Visual Fingerprints</option>
            </select>
          </div>

          <div class="flex items-center justify-between rounded-xl border border-white/10 bg-white/5 p-4 text-xs">
            <div>
              <p class="font-bold text-white">재방문 구독자용 추천 동영상 (Featured video for returning subscribers)</p>
              <p class="mt-0.5 text-neutral-400">Selected: Multimodal Video Indexing with Audio-Visual Fingerprints</p>
            </div>
            <select v-model="featuredVideoId" class="rounded-lg border border-white/15 bg-[#1B1F2B] px-3 py-1.5 text-xs text-white">
              <option value="vid_ai_agents">Multimodal Video Indexing with Audio-Visual Fingerprints</option>
              <option value="vid_arch_2026">Building a 10M Concurrent Viewer Global Video Platform</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab 2: Branding -->
    <div v-else-if="activeTab === 'branding'" class="mt-6 space-y-6">
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="text-sm font-bold text-white">사진 및 프로필 이미지 (Picture)</h2>
        <div class="mt-4 flex items-center gap-6">
          <img :src="avatarUrl" alt="Avatar" class="h-24 w-24 rounded-full border-2 border-white/20 object-cover" />
          <div class="text-xs text-neutral-300">
            <p>At least 98 x 98 pixels and 4MB or less. Use PNG or GIF (no animations).</p>
            <input
              v-model="avatarUrl"
              type="text"
              placeholder="Avatar image URL"
              class="mt-2 w-full max-w-md rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-1.5 text-white outline-none"
            />
          </div>
        </div>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="text-sm font-bold text-white">배너 이미지 (Banner Image)</h2>
        <div class="mt-4">
          <div class="h-32 w-full overflow-hidden rounded-xl border border-white/15 bg-black">
            <img :src="bannerUrl" alt="Banner" class="h-full w-full object-cover" />
          </div>
          <input
            v-model="bannerUrl"
            type="text"
            placeholder="Banner image URL (Recommended: 2048 x 1152 pixels)"
            class="mt-3 w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-1.5 text-xs text-white outline-none"
          />
        </div>
      </div>
    </div>

    <!-- Tab 3: Basic Info -->
    <div v-else class="mt-6 space-y-6">
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5 space-y-4 text-xs">
        <div>
          <label class="block font-bold text-white">채널 이름 (Channel Name)</label>
          <input
            v-model="channelName"
            type="text"
            class="mt-1 w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-2 text-white outline-none focus:border-[#3EA6FF]"
          />
        </div>

        <div>
          <label class="block font-bold text-white">핸들 (Handle)</label>
          <input
            v-model="channelHandle"
            type="text"
            class="mt-1 w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-2 text-white outline-none focus:border-[#3EA6FF]"
          />
        </div>

        <div>
          <label class="block font-bold text-white">설명 (Description)</label>
          <textarea
            v-model="description"
            rows="4"
            class="mt-1 w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-2 text-white outline-none focus:border-[#3EA6FF]"
          />
        </div>

        <div>
          <label class="block font-bold text-white">비즈니스 문의 이메일 (Contact Email)</label>
          <input
            v-model="contactEmail"
            type="email"
            class="mt-1 w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-2 text-white outline-none focus:border-[#3EA6FF]"
          />
        </div>
      </div>
    </div>
  </div>
</template>
