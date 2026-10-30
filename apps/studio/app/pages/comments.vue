<script setup lang="ts">
useSeoMeta({
  title: 'Channel Comments & Moderation — Creator Studio'
})

interface ModeratedComment {
  id: string
  author: string
  avatarUrl: string
  videoTitle: string
  content: string
  date: string
  likes: number
  status: 'published' | 'held' | 'spam'
  isHearted?: boolean
  isPinned?: boolean
}

const activeTab = ref<'published' | 'held' | 'spam'>('published')
const searchQuery = ref('')
const blockedWords = ref(['scam_link', 'free_crypto', 'telegram_bot', 'phishing'])
const newBlockedWord = ref('')

const comments = ref<ModeratedComment[]>([
  {
    id: 'c1',
    author: 'Sarah Chen (@sarahchen_dev)',
    avatarUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100',
    videoTitle: 'Building a 10M Concurrent Viewer Global Video Platform',
    content: 'The double-entry ledger walkthrough saved our team weeks of design review. Subscribed!',
    date: '2026-10-22',
    likes: 195,
    status: 'published',
    isHearted: true,
    isPinned: false
  },
  {
    id: 'c2',
    author: 'Minjun Kim (@kim_viewer)',
    avatarUrl: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100',
    videoTitle: 'Multimodal Video Indexing with Audio-Visual Fingerprints',
    content: 'OpenSearch kNN 벡터 인덱싱과 오디오 지문 추출 파이프라인 지연시간은 어느 정도인가요?',
    date: '2026-10-24',
    likes: 38,
    status: 'published',
    isHearted: false,
    isPinned: false
  },
  {
    id: 'c3',
    author: 'CryptoBot99 (@free_tokens_xyz)',
    avatarUrl: 'https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=100',
    videoTitle: 'Building a 10M Concurrent Viewer Global Video Platform',
    content: 'Claim 100 USDT now by joining our telegram_bot at bit.ly/scam_link!',
    date: '2026-10-25',
    likes: 0,
    status: 'spam',
    isHearted: false,
    isPinned: false
  },
  {
    id: 'c4',
    author: 'Alex Rivera (@alex_cloud)',
    avatarUrl: 'https://images.unsplash.com/photo-1566492031773-4f4e44671857?w=100',
    videoTitle: 'Building a 10M Concurrent Viewer Global Video Platform',
    content: 'Does the Cloudflare R2 multipart worker support S3 SSE-C customer keys?',
    date: '2026-10-26',
    likes: 12,
    status: 'held',
    isHearted: false,
    isPinned: false
  }
])

const filteredComments = computed(() => {
  return comments.value.filter(c => {
    if (c.status !== activeTab.value) return false
    if (!searchQuery.value) return true
    return (
      c.content.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      c.author.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  })
})

function approveComment(commentId: string) {
  const item = comments.value.find(c => c.id === commentId)
  if (item) item.status = 'published'
}

function removeComment(commentId: string) {
  comments.value = comments.value.filter(c => c.id !== commentId)
}

function toggleHeart(item: ModeratedComment) {
  item.isHearted = !item.isHearted
}

function togglePin(item: ModeratedComment) {
  item.isPinned = !item.isPinned
}

function addBlockedWord() {
  const word = newBlockedWord.value.trim().toLowerCase()
  if (word && !blockedWords.value.includes(word)) {
    blockedWords.value.push(word)
    newBlockedWord.value = ''
  }
}

function removeBlockedWord(word: string) {
  blockedWords.value = blockedWords.value.filter(w => w !== word)
}
</script>

<template>
  <div class="max-w-6xl">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 댓글 및 검토 관리 (Comments & Mentions)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          Moderate comments, respond to viewers, manage automated word filters, and approve held comments
        </p>
      </div>
    </div>

    <!-- Tabs & Filter Bar -->
    <div class="mt-6 flex flex-wrap items-center justify-between gap-4 border-b border-white/10 pb-3">
      <div class="flex gap-6 text-sm font-semibold">
        <button
          type="button"
          class="border-b-2 pb-3 transition"
          :class="activeTab === 'published' ? 'border-white text-white' : 'border-transparent text-neutral-400 hover:text-white'"
          @click="activeTab = 'published'"
        >
          게시된 댓글 (Published)
        </button>
        <button
          type="button"
          class="border-b-2 pb-3 transition"
          :class="activeTab === 'held' ? 'border-white text-white' : 'border-transparent text-neutral-400 hover:text-white'"
          @click="activeTab = 'held'"
        >
          검토 대기 중 (Held for review)
        </button>
        <button
          type="button"
          class="border-b-2 pb-3 transition"
          :class="activeTab === 'spam' ? 'border-white text-white' : 'border-transparent text-neutral-400 hover:text-white'"
          @click="activeTab = 'spam'"
        >
          스팸 의심 (Likely spam)
        </button>
      </div>

      <input
        v-model="searchQuery"
        type="text"
        placeholder="Search comments or author..."
        class="w-64 rounded-xl border border-white/15 bg-[#151822] px-3 py-1.5 text-xs text-white outline-none focus:border-[#3EA6FF]"
      />
    </div>

    <!-- Comments List -->
    <div class="mt-6 space-y-4">
      <div
        v-for="cmt in filteredComments"
        :key="cmt.id"
        class="flex items-start gap-4 rounded-2xl border border-white/10 bg-[#151822] p-4 text-xs"
      >
        <img :src="cmt.avatarUrl" :alt="cmt.author" class="h-10 w-10 shrink-0 rounded-full object-cover" />
        <div class="flex-1">
          <div class="flex items-center gap-2">
            <span class="font-bold text-white">{{ cmt.author }}</span>
            <span class="text-neutral-400">· {{ cmt.date }}</span>
            <span v-if="cmt.isPinned" class="rounded bg-white/10 px-1.5 py-0.5 text-[10px] text-neutral-300">
              📌 Pinned
            </span>
            <span v-if="cmt.isHearted" class="text-rose-400">❤️</span>
          </div>

          <p class="mt-0.5 text-[11px] text-[#3EA6FF]">on: {{ cmt.videoTitle }}</p>
          <p class="mt-2 text-sm text-neutral-200">{{ cmt.content }}</p>

          <!-- Actions -->
          <div class="mt-3 flex items-center gap-3">
            <button
              v-if="cmt.status === 'held' || cmt.status === 'spam'"
              type="button"
              class="rounded-lg bg-emerald-600 px-3 py-1 font-bold text-white hover:bg-emerald-500"
              @click="approveComment(cmt.id)"
            >
              ✓ 승인 (Approve)
            </button>
            <button
              v-if="cmt.status === 'published'"
              type="button"
              class="rounded-lg bg-white/10 px-2.5 py-1 text-white hover:bg-white/20"
              @click="toggleHeart(cmt)"
            >
              {{ cmt.isHearted ? '❤️ 하트 취소' : '🤍 하트' }}
            </button>
            <button
              v-if="cmt.status === 'published'"
              type="button"
              class="rounded-lg bg-white/10 px-2.5 py-1 text-white hover:bg-white/20"
              @click="togglePin(cmt)"
            >
              {{ cmt.isPinned ? '📌 고정 해제' : '📌 상단 고정' }}
            </button>
            <button
              type="button"
              class="rounded-lg bg-red-600/20 px-2.5 py-1 font-bold text-red-400 hover:bg-red-600/40"
              @click="removeComment(cmt.id)"
            >
              🗑️ 삭제 (Delete)
            </button>
          </div>
        </div>
      </div>

      <div v-if="filteredComments.length === 0" class="rounded-2xl border border-white/10 bg-[#151822] p-8 text-center text-neutral-400">
        해당 탭에 표시할 댓글이 없습니다.
      </div>
    </div>

    <!-- Automated Moderation & Blocked Words Settings -->
    <div class="mt-8 rounded-2xl border border-white/10 bg-[#151822] p-5">
      <h2 class="text-sm font-bold text-white">자동 금지어 및 차단 단어 필터 (Automated Word Filter)</h2>
      <p class="mt-1 text-xs text-neutral-400">
        Comments containing these words will be automatically held for review or rejected
      </p>

      <form class="mt-4 flex gap-2" @submit.prevent="addBlockedWord">
        <input
          v-model="newBlockedWord"
          type="text"
          placeholder="Add blocked word or URL pattern..."
          class="flex-1 rounded-xl border border-white/15 bg-[#0D0F14] px-3 py-2 text-xs text-white outline-none focus:border-[#3EA6FF]"
        />
        <button
          type="submit"
          class="rounded-xl bg-[#3EA6FF] px-4 py-2 text-xs font-bold text-black hover:bg-[#65B8FF]"
        >
          추가 (Add)
        </button>
      </form>

      <div class="mt-3 flex flex-wrap gap-2">
        <span
          v-for="word in blockedWords"
          :key="word"
          class="flex items-center gap-1.5 rounded-full bg-white/10 px-3 py-1 text-xs text-neutral-200"
        >
          <span>{{ word }}</span>
          <button type="button" class="text-neutral-400 hover:text-white" @click="removeBlockedWord(word)">
            ✕
          </button>
        </span>
      </div>
    </div>
  </div>
</template>
