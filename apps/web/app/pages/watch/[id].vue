<script setup lang="ts">
import type { Video, Comment } from '@gvp/shared-types'
import { VideoProcessingStatus, VideoType, VideoVisibility, VideoQuality } from '@gvp/shared-types'
import { formatRelativeTime, formatTimecode, formatViewCount } from '@gvp/ui'
import { useAppStore } from '~/stores/app'

const route = useRoute()
const videoId = computed(() => String(route.params.id || 'vid_arch_2026'))
const appStore = useAppStore()
const api = useApi()
const config = useRuntimeConfig()

const playerRef = ref<{ seekToSeconds: (sec: number) => void; getCurrentTimeSeconds: () => number } | null>(null)
const theater = ref(false)
const showShareModal = ref(false)
const shareStartSeconds = ref(0)
const autoplayNext = ref(true)

const { data: video, pending: videoPending, error: videoError } = await useAsyncData(`video-${videoId.value}`, async () => {
  return await api.getVideo(videoId.value)
})

// The template v-if guards on video; the assertion keeps bindings typed.
const activeVideo = computed(() => video.value as Video)

// SSR SEO Meta + Open Graph + Twitter Card
useSeoMeta({
  title: () => activeVideo.value?.title || 'vravra',
  description: () => (activeVideo.value?.description || '').slice(0, 160),
  ogTitle: () => activeVideo.value?.title || 'vravra',
  ogDescription: () => (activeVideo.value?.description || '').slice(0, 160),
  ogImage: () => activeVideo.value?.thumbnailUrl,
  ogType: 'video.other',
  ogUrl: () => `${config.public.siteUrl}/watch/${videoId.value}`,
  twitterCard: 'summary_large_image',
  twitterTitle: () => activeVideo.value?.title || 'vravra',
  twitterDescription: () => (activeVideo.value?.description || '').slice(0, 160),
  twitterImage: () => activeVideo.value?.thumbnailUrl
})

// Schema.org VideoObject JSON-LD Structured Data
useHead({
  script: [
    {
      type: 'application/ld+json',
      innerHTML: computed(() => {
        // Fetch failures (unknown id, API outage) must render the error
        // block, not crash SSR. Emit a minimal WebPage schema instead of
        // a VideoObject when there is no video to describe.
        const v = video.value
        if (!v) {
          return JSON.stringify({
            '@context': 'https://schema.org',
            '@type': 'WebPage',
            name: 'vravra',
            url: `${config.public.siteUrl}/watch/${videoId.value}`
          })
        }
        return JSON.stringify({
          '@context': 'https://schema.org',
          '@type': 'VideoObject',
          name: v.title,
          description: v.description,
          thumbnailUrl: [v.thumbnailUrl],
          uploadDate: v.publishedAt || v.createdAt,
          duration: `PT${Math.floor(v.durationSeconds / 60)}M${v.durationSeconds % 60}S`,
          contentUrl: v.hlsMasterUrl,
          embedUrl: `${config.public.siteUrl}/watch/${v.id}`,
          interactionStatistic: {
            '@type': 'InteractionCounter',
            interactionType: { '@type': 'WatchAction' },
            userInteractionCount: v.viewCount
          },
          author: {
            '@type': 'Person',
            name: v.channel?.name || 'vravra 크리에이터'
          }
        })
      })
    }
  ]
})

const subscriberCount = ref(0)
const isLiked = ref(false)
const likeCount = ref(0)
const actionError = ref<string | null>(null)
watch(video, (v) => {
  if (!v) return
  subscriberCount.value = v.channel?.subscriberCount || 0
  likeCount.value = v.likeCount || 0
}, { immediate: true })
const descriptionExpanded = ref(false)
const showReportModal = ref(false)
const reportReason = ref('copyright_or_policy')
const reportSubmitted = ref(false)
const reportSubmitting = ref(false)
const reportError = ref<string | null>(null)
const reportCaseId = ref<string | null>(null)

function closeReportModal() {
  showReportModal.value = false
  reportSubmitted.value = false
  reportError.value = null
}

function onReportModalKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closeReportModal()
}

async function submitReport() {
  if (reportSubmitting.value) return
  reportSubmitting.value = true
  reportError.value = null
  try {
    const created = await api.reportContent({
      targetType: 'video',
      targetId: videoId.value,
      reason: reportReason.value,
      policyCategory: 'general'
    })
    reportCaseId.value = created.id
    reportSubmitted.value = true
  } catch (e: unknown) {
    reportError.value = e instanceof Error ? e.message : '신고 접수에 실패했습니다.'
  } finally {
    reportSubmitting.value = false
  }
}

function fmtDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleDateString(appStore.locale === 'ko' ? 'ko-KR' : 'en-US')
}

const newCommentText = ref('')
const replyTargetId = ref<string | null>(null)
const replyText = ref('')

const commentsError = ref<string | null>(null)
const { data: commentsList } = await useAsyncData(`comments-${videoId.value}`, async () => {
  try {
    const res = await api.getVideoComments(videoId.value)
    return res.items || []
  } catch (e: unknown) {
    commentsError.value = e instanceof Error ? e.message : 'Failed to load comments'
    return []
  }
})

const comments = ref<Comment[]>(commentsList.value || [])
const isDisliked = ref(false)
const commentSort = ref<'top' | 'newest'>('top')
const expandedReplies = ref<Record<string, boolean>>({})

function toggleDislike() {
  isDisliked.value = !isDisliked.value
  if (isDisliked.value && isLiked.value) {
    isLiked.value = false
    likeCount.value = Math.max(0, likeCount.value - 1)
  }
}

function toggleExpandReplies(commentId: string) {
  expandedReplies.value[commentId] = !expandedReplies.value[commentId]
}

const sortedComments = computed(() => {
  const list = [...comments.value]
  if (commentSort.value === 'newest') {
    return list.sort((a, b) => {
      if (a.isPinned !== b.isPinned) return a.isPinned ? -1 : 1
      return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    })
  }
  return list.sort((a, b) => {
    if (a.isPinned !== b.isPinned) return a.isPinned ? -1 : 1
    return (b.likeCount || 0) - (a.likeCount || 0)
  })
})

function onSubscribeChange(v: { subscribed: boolean; count: number }) {
  subscriberCount.value = v.count
}

async function toggleLike() {
  try {
    if (isLiked.value) {
      const res = await api.unlikeVideo(activeVideo.value.id)
      isLiked.value = res.liked
      likeCount.value = res.likeCount
    } else {
      const res = await api.likeVideo(activeVideo.value.id)
      isLiked.value = res.liked
      likeCount.value = res.likeCount
    }
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Like update failed'
  }
}

const { data: relatedData } = await useAsyncData(`related-${videoId.value}`, async () => {
  if (!video.value?.channelId) return []
  try {
    const res = await api.getChannelVideos(video.value.channelId, 1, 10)
    return (res.items || []).filter(v => v.id !== videoId.value).slice(0, 6)
  } catch {
    return []
  }
})
const relatedVideos = computed(() => relatedData.value || [])
const upNext = computed(() => relatedVideos.value[0] || null)

function openShare() {
  shareStartSeconds.value = playerRef.value?.getCurrentTimeSeconds() || 0
  showShareModal.value = true
}

function playNext() {
  if (upNext.value) navigateTo(`/watch/${upNext.value.id}`)
}

function onPlayerEnded() {
  if (autoplayNext.value && upNext.value) navigateTo(`/watch/${upNext.value.id}`)
}

function toggleAutoplayNext() {
  autoplayNext.value = !autoplayNext.value
  try { localStorage.setItem('vravra_autoplay_next', autoplayNext.value ? '1' : '0') } catch { /* quota */ }
}

onMounted(() => {
  try {
    autoplayNext.value = localStorage.getItem('vravra_autoplay_next') !== '0'
  } catch { /* default on */ }
  const t = Number(route.query.t)
  if (Number.isFinite(t) && t > 0) {
    setTimeout(() => playerRef.value?.seekToSeconds(t), 800)
  }
})

async function submitComment() {
  const text = newCommentText.value.trim()
  if (!text) return
  commentsError.value = null
  try {
    const created = await api.addComment(activeVideo.value.id, text)
    comments.value.unshift(created)
    newCommentText.value = ''
  } catch (e: unknown) {
    commentsError.value = e instanceof Error ? e.message : '댓글 작성에 실패했습니다.'
  }
}

async function submitReply(parentId: string) {
  const text = replyText.value.trim()
  if (!text) return
  let replyObj: Comment
  try {
    replyObj = await api.addComment(activeVideo.value.id, text, parentId)
  } catch (e: unknown) {
    commentsError.value = e instanceof Error ? e.message : '답글 작성에 실패했습니다.'
    return
  }

  const parent = comments.value.find(c => c.id === parentId)
  if (parent) {
    if (!parent.replies) parent.replies = []
    parent.replies.push(replyObj)
    parent.replyCount = parent.replies.length
    expandedReplies.value[parentId] = true
  } else {
    comments.value.push(replyObj)
  }

  replyText.value = ''
  replyTargetId.value = null
}

async function handleLikeComment(cmt: Comment) {
  try {
    const res = await api.likeComment(cmt.id)
    cmt.likeCount = res.likeCount
  } catch (e: unknown) {
    commentsError.value = e instanceof Error ? e.message : '댓글 좋아요에 실패했습니다.'
  }
}

async function handleHeartComment(cmt: Comment) {
  try {
    const res = await api.heartComment(cmt.id)
    cmt.isHeartedByCreator = res.isHearted
  } catch (e: unknown) {
    commentsError.value = e instanceof Error ? e.message : '하트 표시에 실패했습니다.'
  }
}

async function handlePinComment(cmt: Comment) {
  try {
    const res = await api.pinComment(cmt.id)
    cmt.isPinned = res.isPinned
  } catch (e: unknown) {
    commentsError.value = e instanceof Error ? e.message : '댓글 고정에 실패했습니다.'
  }
}

async function removeComment(commentId: string, parentId?: string) {
  try {
    await api.deleteComment(commentId)
  } catch (e: unknown) {
    commentsError.value = e instanceof Error ? e.message : '댓글 삭제에 실패했습니다.'
    return
  }
  if (parentId) {
    const parent = comments.value.find(c => c.id === parentId)
    if (parent && parent.replies) {
      parent.replies = parent.replies.filter(r => r.id !== commentId)
      parent.replyCount = parent.replies.length
    }
  } else {
    comments.value = comments.value.filter(c => c.id !== commentId)
  }
}

</script>

<template>
  <div class="mx-auto max-w-[1720px] bg-[var(--gvp-surface)] px-4 py-5 text-[var(--gvp-text)] lg:px-8">
    <div v-if="videoPending" class="animate-pulse">
      <div class="aspect-video w-full rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-elevated)]" />
      <div class="mt-4 h-6 w-2/3 rounded-md bg-[var(--gvp-elevated)]" />
    </div>
    <div v-else-if="videoError || !video" class="rounded-lg border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-12 text-center" role="alert">
      <p class="text-base font-semibold text-[var(--gvp-live)]">
        {{ appStore.locale === 'ko' ? '동영상을 불러오지 못했습니다.' : 'This video is unavailable.' }}
      </p>
      <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">{{ videoId }}</p>
    </div>
    <div v-else class="grid grid-cols-1 gap-6" :class="theater ? '' : 'xl:grid-cols-[minmax(0,1fr)_402px]'">
      <!-- Primary Column: Video Player, Metadata, Chapters, Comments -->
      <div class="min-w-0" :class="theater ? 'max-w-none' : 'max-w-[1280px]'">
        <ClientOnly>
          <VideoPlayer
            ref="playerRef"
            :video-id="activeVideo.id"
            :channel-id="activeVideo.channelId"
            :title="activeVideo.title"
            :hls-master-url="activeVideo.hlsMasterUrl || ''"
            :poster-url="activeVideo.thumbnailUrl"
            :duration-seconds="activeVideo.durationSeconds"
            :chapters="activeVideo.chapters"
            :captions="activeVideo.captions"
            :sources="activeVideo.sources"
            :autoplay-next="autoplayNext"
            :has-next="!!upNext"
            @ended="onPlayerEnded"
            @request-next="playNext"
            @toggle-autoplay-next="toggleAutoplayNext"
            @toggle-theater="theater = !theater"
          />
          <template #fallback>
            <div class="flex aspect-video w-full items-center justify-center rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-sm text-[var(--gvp-text-secondary)]">
              Loading Shaka HLS ABR Player...
            </div>
          </template>
        </ClientOnly>

        <!-- Video Title -->
        <h1 class="mt-4 text-lg font-bold leading-snug text-[var(--gvp-text)] sm:text-xl">
          {{ activeVideo.title }}
        </h1>
        <p v-if="actionError" class="mt-2 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-2 text-xs font-semibold text-[var(--gvp-live)]" role="alert">
          {{ actionError }}
        </p>

        <!-- Channel Bar & Action Buttons -->
        <div class="mt-3 flex flex-wrap items-center justify-between gap-4 border-b border-[var(--gvp-border)] pb-4">
          <div class="flex items-center gap-3">
            <NuxtLink :to="`/channels/${activeVideo.channelId}`">
              <img
                :src="activeVideo.channel?.avatarUrl || 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&q=80'"
                :alt="activeVideo.channel?.name"
                class="h-10 w-10 rounded-full border border-[var(--gvp-border)] object-cover"
              />
            </NuxtLink>
            <div>
              <NuxtLink :to="`/channels/${activeVideo.channelId}`" class="flex items-center gap-1 font-bold text-[var(--gvp-text)] hover:text-[var(--gvp-brand)]">
                <span>{{ activeVideo.channel?.name || 'vravra 크리에이터' }}</span>
                <svg v-if="activeVideo.channel?.isVerified" class="h-4 w-4 text-[var(--gvp-brand)]" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10 10-4.5 10-10S17.5 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
                </svg>
              </NuxtLink>
              <p class="text-xs text-[var(--gvp-text-secondary)] tabular-nums">
                {{ subscriberCount.toLocaleString() }} subscribers
              </p>
            </div>

            <WatchSubscribeButton
              :channel-id="activeVideo.channelId"
              :initial-count="activeVideo.channel?.subscriberCount || 0"
              @change="onSubscribeChange"
            />
          </div>

          <!-- Action Pills: Like/Dislike, Share, Save, Report -->
          <div class="flex flex-wrap items-center gap-2">
            <!-- Unified Like / Dislike Split Control -->
            <div class="flex items-center overflow-hidden rounded-full bg-[var(--gvp-elevated)]">
              <button
                type="button"
                class="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold transition tabular-nums active:scale-95"
                :class="isLiked ? 'bg-[var(--gvp-info-soft)] text-[var(--gvp-brand-hover)]' : 'text-[var(--gvp-text)] hover:bg-[var(--gvp-canvas)]'"
                @click="toggleLike"
              >
                <svg :key="String(isLiked)" class="h-4 w-4 transition-transform duration-150" :class="isLiked ? 'scale-125' : ''" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
                </svg>
                <span>{{ likeCount.toLocaleString() }}</span>
              </button>
              <div class="h-4 w-px bg-[var(--gvp-border-strong)]" />
              <button
                type="button"
                class="px-3 py-1.5 text-xs font-semibold transition"
                :class="isDisliked ? 'bg-[var(--gvp-info-soft)] text-[var(--gvp-brand-hover)]' : 'text-[var(--gvp-text)] hover:bg-[var(--gvp-canvas)]'"
                @click="toggleDislike"
                title="I dislike this"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M10 14H5.236a2 2 0 01-1.789-2.894l3.5-7A2 2 0 018.736 3h4.018a2 2 0 01.485.06l3.76.94m-7 10v5a2 2 0 002 2h.096c.5 0 .905-.405.905-.904 0-.715.211-1.413.608-2.008L17 13V4m-7 10h2m5-10h2a2 2 0 012 2v6a2 2 0 01-2 2h-2.5" />
                </svg>
              </button>
            </div>

            <button
              type="button"
              class="flex items-center gap-1.5 rounded-full bg-[var(--gvp-elevated)] px-3.5 py-1.5 text-xs font-semibold text-[var(--gvp-text)] transition hover:bg-[var(--gvp-canvas)]"
              @click="openShare"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
              </svg>
              <span>{{ appStore.locale === 'ko' ? '공유' : 'Share' }}</span>
            </button>

            <WatchSavePopover :video="activeVideo" />

            <button
              type="button"
              class="flex items-center gap-1.5 rounded-full bg-[var(--gvp-elevated)] px-3.5 py-1.5 text-xs font-semibold text-[var(--gvp-text-secondary)] transition hover:bg-[var(--gvp-canvas)] hover:text-[var(--gvp-text)]"
              @click="showReportModal = true"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 21v-4m0 0V5a2 2 0 012-2h6.5l1 1H21l-3 6 3 6h-8.5l-1-1H5a2 2 0 00-2 2zm9-13.5V9" />
              </svg>
              <span>{{ appStore.locale === 'ko' ? '신고' : 'Report' }}</span>
            </button>
          </div>
        </div>

        <!-- Expandable Description & Interactive Chapters -->
        <div class="mt-4 rounded-xl bg-[var(--gvp-canvas)] p-4 text-sm">
          <div class="flex flex-wrap items-center gap-2.5 font-semibold text-[var(--gvp-text)] tabular-nums">
            <span>{{ activeVideo.viewCount.toLocaleString() }} views</span>
            <span class="text-[var(--gvp-text-muted)]">•</span>
            <span>Published {{ fmtDate(activeVideo.publishedAt || activeVideo.createdAt) }}</span>
            <span
              v-for="tag in activeVideo.tags"
              :key="tag"
              class="rounded-md border border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] px-2 py-0.5 text-xs font-medium text-[var(--gvp-brand-hover)]"
            >
              #{{ tag }}
            </span>
          </div>

          <p class="mt-2 whitespace-pre-line text-[var(--gvp-text-secondary)]" :class="!descriptionExpanded ? 'line-clamp-3' : ''">
            {{ activeVideo.description }}
          </p>

          <!-- Chapters Timeline Strip -->
          <div v-if="activeVideo.chapters?.length" class="mt-4 border-t border-[var(--gvp-border)] pt-3">
            <h3 class="mb-2 text-xs font-bold uppercase tracking-wider text-[var(--gvp-text-secondary)]">
              {{ appStore.locale === 'ko' ? '챕터 바로가기 (Video Chapters)' : 'Video Chapters' }}
            </h3>
            <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <button
                v-for="chp in activeVideo.chapters"
                :key="chp.id"
                type="button"
                class="flex items-center justify-between rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-3 py-2 text-left text-xs transition hover:bg-[var(--gvp-canvas)]"
                @click="playerRef?.seekToSeconds(chp.startTimeSeconds)"
              >
                <span class="truncate font-medium text-[var(--gvp-text)]">{{ chp.title }}</span>
                <span class="ml-2 shrink-0 rounded-md border border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] px-1.5 py-0.5 font-mono text-[11px] font-semibold text-[var(--gvp-brand-hover)] tabular-nums">
                  {{ Math.floor(chp.startTimeSeconds / 60) }}:{{ String(chp.startTimeSeconds % 60).padStart(2, '0') }}
                </span>
              </button>
            </div>
          </div>

          <button
            type="button"
            class="mt-3 text-xs font-bold text-[var(--gvp-brand-hover)] hover:underline"
            @click="descriptionExpanded = !descriptionExpanded"
          >
            {{ descriptionExpanded ? (appStore.locale === 'ko' ? '간략히 보기' : 'Show less') : (appStore.locale === 'ko' ? '더보기' : 'Show more') }}
          </button>
        </div>

        <!-- Threaded Comments Section -->
        <section class="mt-6">
          <div class="flex flex-wrap items-center justify-between gap-4">
            <div class="flex items-center gap-6">
              <h2 class="text-base font-bold text-[var(--gvp-text)] tabular-nums">
                {{ comments.length.toLocaleString() }} {{ appStore.locale === 'ko' ? '개의 댓글' : 'Comments' }}
              </h2>
              <div class="flex items-center gap-1.5 text-xs font-medium">
                <button
                  type="button"
                  class="rounded-full px-3 py-1 transition"
                  :class="commentSort === 'top' ? 'bg-[var(--gvp-elevated)] font-semibold text-[var(--gvp-text)]' : 'text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]'"
                  @click="commentSort = 'top'"
                >
                  <span>{{ appStore.locale === 'ko' ? '인기순' : 'Top' }}</span>
                </button>
                <button
                  type="button"
                  class="rounded-full px-3 py-1 transition"
                  :class="commentSort === 'newest' ? 'bg-[var(--gvp-elevated)] font-semibold text-[var(--gvp-text)]' : 'text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]'"
                  @click="commentSort = 'newest'"
                >
                  <span>{{ appStore.locale === 'ko' ? '최신순' : 'Newest' }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- New Comment Input -->
          <form class="mt-4 flex items-start gap-3" @submit.prevent="submitComment">
            <img
              :src="appStore.currentUser?.avatarUrl || 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&q=80'"
              :alt="appStore.currentUser?.displayName"
              class="h-9 w-9 rounded-full border border-[var(--gvp-border)] object-cover"
            />
            <div class="flex-1">
              <input
                v-model="newCommentText"
                type="text"
                :placeholder="appStore.locale === 'ko' ? '공개 댓글 추가...' : 'Add a public comment...'"
                class="w-full rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-2 text-sm text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-brand)] focus:bg-[var(--gvp-surface)]"
              />
              <div class="mt-2 flex justify-end gap-2">
                <button
                  type="button"
                  class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-3 py-1.5 text-xs font-medium text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-canvas)]"
                  @click="newCommentText = ''"
                >
                  {{ appStore.locale === 'ko' ? '취소' : 'Cancel' }}
                </button>
                <button
                  type="submit"
                  class="rounded-md bg-[var(--gvp-brand)] px-4 py-1.5 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
                >
                  {{ appStore.locale === 'ko' ? '댓글 작성' : 'Comment' }}
                </button>
              </div>
            </div>
          </form>

          <p v-if="commentsError" class="mt-4 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-2 text-xs font-semibold text-[var(--gvp-live)]" role="alert">
            {{ commentsError }}
          </p>
          <div v-if="sortedComments.length === 0" class="mt-6 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-8 text-center">
            <p class="text-sm font-semibold text-[var(--gvp-text-secondary)]">
              {{ appStore.locale === 'ko' ? '아직 댓글이 없습니다. 첫 댓글을 남겨보세요.' : 'No comments yet. Start the conversation.' }}
            </p>
          </div>

          <!-- Comment Thread List -->
          <div v-else class="mt-6 space-y-6 divide-y divide-[var(--gvp-border)]">
            <article v-for="cmt in sortedComments" :key="cmt.id" class="flex items-start gap-3 pt-4 first:pt-0">
              <img
                :src="cmt.author?.avatarUrl || 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&q=80'"
                :alt="cmt.author?.displayName"
                class="h-9 w-9 shrink-0 rounded-full border border-[var(--gvp-border)] object-cover"
              />
              <div class="flex-1 text-sm">
                <!-- Header: Display name, Badges, Date -->
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-semibold text-[var(--gvp-text)]">{{ cmt.author?.displayName || 'User' }}</span>
                  <span
                    v-if="cmt.isChannelOwner || cmt.author?.isCreator"
                    class="rounded-md border border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] px-2 py-0.5 text-[10px] font-bold text-[var(--gvp-brand-hover)]"
                  >
                    {{ appStore.locale === 'ko' ? '크리에이터' : 'Creator' }}
                  </span>
                  <span
                    v-if="cmt.isPinned"
                    class="inline-flex items-center gap-1 rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-1.5 py-0.5 text-[10px] font-semibold text-[var(--gvp-warning)]"
                  >
                    <span>{{ appStore.locale === 'ko' ? '고정됨' : 'Pinned' }}</span>
                  </span>
                  <span class="text-xs text-[var(--gvp-text-muted)] tabular-nums">
                    {{ fmtDate(cmt.createdAt) }}
                  </span>
                </div>

                <!-- Body -->
                <p class="mt-1 text-[var(--gvp-text)]"><CommentText :text="cmt.content" @seek="(s) => playerRef?.seekToSeconds(s)" /></p>

                <!-- Actions: Like, Heart, Pin, Reply, Delete -->
                <div class="mt-2 flex flex-wrap items-center gap-4 text-xs text-[var(--gvp-text-secondary)]">
                  <button
                    type="button"
                    class="flex items-center gap-1 transition hover:text-[var(--gvp-text)] tabular-nums"
                    @click="handleLikeComment(cmt)"
                  >
                    <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
                    </svg>
                    <span>{{ cmt.likeCount || 0 }}</span>
                  </button>

                  <!-- Creator Heart indicator -->
                  <span
                    v-if="cmt.isHeartedByCreator"
                    class="inline-flex items-center gap-1 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-1.5 py-0.5 text-[10px] font-semibold text-[var(--gvp-live)]"
                    :title="appStore.locale === 'ko' ? '게시자가 하트를 눌렀습니다' : 'Hearted by creator'"
                  >
                    <svg class="h-3 w-3 fill-current" viewBox="0 0 24 24">
                      <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z" />
                    </svg>
                    <span>Creator Heart</span>
                  </span>

                  <!-- Channel Owner controls: Heart & Pin -->
                  <button
                    v-if="appStore.currentUser?.id && activeVideo.channel?.ownerId && appStore.currentUser.id === activeVideo.channel.ownerId"
                    type="button"
                    class="flex items-center gap-1 transition hover:text-[var(--gvp-text)]"
                    :class="cmt.isHeartedByCreator ? 'text-[var(--gvp-live)] font-semibold' : 'text-[var(--gvp-text-secondary)]'"
                    @click="handleHeartComment(cmt)"
                  >
                    <span>{{ appStore.locale === 'ko' ? '하트' : 'Heart' }}</span>
                  </button>
                  <button
                    v-if="appStore.currentUser?.id && activeVideo.channel?.ownerId && appStore.currentUser.id === activeVideo.channel.ownerId"
                    type="button"
                    class="flex items-center gap-1 transition hover:text-[var(--gvp-text)]"
                    :class="cmt.isPinned ? 'text-[var(--gvp-warning)] font-semibold' : 'text-[var(--gvp-text-secondary)]'"
                    @click="handlePinComment(cmt)"
                  >
                    <span>{{ cmt.isPinned ? (appStore.locale === 'ko' ? '고정 해제' : 'Unpin') : (appStore.locale === 'ko' ? '고정' : 'Pin') }}</span>
                  </button>

                  <button
                    type="button"
                    class="font-semibold transition hover:text-[var(--gvp-text)]"
                    @click="replyTargetId = replyTargetId === cmt.id ? null : cmt.id"
                  >
                    {{ appStore.locale === 'ko' ? '답글' : 'Reply' }}
                  </button>
                  <button
                    v-if="cmt.userId === appStore.currentUser?.id"
                    type="button"
                    class="text-[var(--gvp-live)] transition hover:underline"
                    @click="removeComment(cmt.id)"
                  >
                    {{ appStore.locale === 'ko' ? '삭제' : 'Delete' }}
                  </button>
                </div>

                <!-- Inline Reply Input Box -->
                <div v-if="replyTargetId === cmt.id" class="mt-3 flex items-center gap-2">
                  <input
                    v-model="replyText"
                    type="text"
                    :placeholder="appStore.locale === 'ko' ? '답글 작성...' : 'Write a reply...'"
                    class="flex-1 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-1.5 text-xs text-[var(--gvp-text)] outline-none focus:border-[var(--gvp-brand)] focus:bg-[var(--gvp-surface)]"
                    @keydown.enter="submitReply(cmt.id)"
                  />
                  <button
                    type="button"
                    class="rounded-md bg-[var(--gvp-brand)] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
                    @click="submitReply(cmt.id)"
                  >
                    {{ appStore.locale === 'ko' ? '등록' : 'Send' }}
                  </button>
                </div>

                <!-- Collapsible Nested Replies -->
                <div v-if="cmt.replies && cmt.replies.length > 0" class="mt-3">
                  <button
                    type="button"
                    class="flex items-center gap-1.5 text-xs font-bold text-[var(--gvp-brand-hover)] transition hover:underline tabular-nums"
                    @click="toggleExpandReplies(cmt.id)"
                  >
                    <span>{{ expandedReplies[cmt.id] ? '▲' : '▼' }}</span>
                    <span>
                      {{ cmt.replies.length }} {{ appStore.locale === 'ko' ? '개의 답글' : (cmt.replies.length === 1 ? 'reply' : 'replies') }}
                    </span>
                  </button>

                  <div v-if="expandedReplies[cmt.id]" class="mt-3 space-y-3 border-l-2 border-[var(--gvp-border)] pl-4">
                    <article v-for="reply in cmt.replies" :key="reply.id" class="flex items-start gap-2.5">
                      <img
                        :src="reply.author?.avatarUrl || 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&q=80'"
                        :alt="reply.author?.displayName"
                        class="h-7 w-7 shrink-0 rounded-full border border-[var(--gvp-border)] object-cover"
                      />
                      <div class="flex-1 text-xs">
                        <div class="flex items-center gap-2">
                          <span class="font-semibold text-[var(--gvp-text)]">{{ reply.author?.displayName || 'User' }}</span>
                          <span
                            v-if="reply.isChannelOwner || reply.author?.isCreator"
                            class="rounded-md border border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] px-1.5 py-0.5 text-[9px] font-bold text-[var(--gvp-brand-hover)]"
                          >
                            {{ appStore.locale === 'ko' ? '크리에이터' : 'Creator' }}
                          </span>
                          <span class="text-[var(--gvp-text-muted)] tabular-nums">
                            {{ fmtDate(reply.createdAt) }}
                          </span>
                        </div>
                        <p class="mt-0.5 text-[var(--gvp-text)]"><CommentText :text="reply.content" @seek="(s) => playerRef?.seekToSeconds(s)" /></p>
                        <div class="mt-1 flex items-center gap-3 text-[var(--gvp-text-secondary)]">
                          <button
                            type="button"
                            class="flex items-center gap-1 transition hover:text-[var(--gvp-text)] tabular-nums"
                            @click="handleLikeComment(reply)"
                          >
                            <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                              <path stroke-linecap="round" stroke-linejoin="round" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
                            </svg>
                            <span>{{ reply.likeCount || 0 }}</span>
                          </button>
                          <button
                            v-if="reply.userId === appStore.currentUser?.id"
                            type="button"
                            class="text-[var(--gvp-live)] hover:underline"
                            @click="removeComment(reply.id, cmt.id)"
                          >
                            {{ appStore.locale === 'ko' ? '삭제' : 'Delete' }}
                          </button>
                        </div>
                      </div>
                    </article>
                  </div>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>

      <!-- Right Sidebar: Related & Up Next Videos -->
      <aside v-if="!theater && relatedVideos.length">
        <h2 class="mb-3 text-xs font-bold uppercase tracking-wider text-[var(--gvp-text-secondary)]">
          {{ appStore.locale === 'ko' ? '같은 채널의 다른 동영상' : 'More from this channel' }}
        </h2>
        <div class="flex flex-col gap-3">
          <NuxtLink
            v-for="item in relatedVideos"
            :key="item.id"
            :to="`/watch/${item.id}`"
            class="group flex gap-3 rounded-xl p-1.5 transition hover:bg-[var(--gvp-elevated)]"
          >
            <div class="relative aspect-video w-44 shrink-0 overflow-hidden rounded-xl bg-[var(--gvp-elevated)]">
              <img :src="item.thumbnailUrl" :alt="item.title" class="h-full w-full object-cover" loading="lazy" />
              <span class="absolute bottom-1.5 right-1.5 rounded bg-black/80 px-1 py-0.5 text-[11px] font-medium text-white tabular-nums">
                {{ formatTimecode(item.durationSeconds) }}
              </span>
            </div>
            <div class="min-w-0 flex-1 py-0.5">
              <h3 class="line-clamp-2 text-xs font-semibold leading-snug text-[var(--gvp-text)]">
                {{ item.title }}
              </h3>
              <p v-if="item.channel?.name" class="mt-1 truncate text-[11px] text-[var(--gvp-text-secondary)]">{{ item.channel.name }}</p>
              <p class="text-[11px] text-[var(--gvp-text-secondary)] tabular-nums">{{ formatViewCount(item.viewCount, appStore.locale) }} · {{ formatRelativeTime(item.publishedAt || item.createdAt, appStore.locale) }}</p>
            </div>
          </NuxtLink>
        </div>
      </aside>
    </div>

    <!-- Report Modal -->
    <div
      v-if="showReportModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-[var(--gvp-scrim)] p-4"
      @mousedown.self="closeReportModal"
      @keydown="onReportModalKeydown"
    >
      <div
        class="w-full max-w-md rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-6 text-[var(--gvp-text)] shadow-sm"
        role="dialog"
        aria-modal="true"
        aria-labelledby="report-modal-title"
      >
        <h3 id="report-modal-title" class="text-base font-bold text-[var(--gvp-text)]">
          {{ appStore.locale === 'ko' ? '콘텐츠 신고 (Trust & Safety Report)' : 'Report Video to Trust & Safety' }}
        </h3>
        <p class="mt-1 text-xs text-[var(--gvp-text-secondary)]">
          Reports are queued directly to the Moderation Engine (`/v1/admin/moderation`).
        </p>
        <div class="mt-4 space-y-2 text-xs text-[var(--gvp-text)]">
          <label class="flex items-center gap-2">
            <input v-model="reportReason" type="radio" value="copyright_or_policy" />
            <span>Copyright infringement or unauthorized re-upload</span>
          </label>
          <label class="flex items-center gap-2">
            <input v-model="reportReason" type="radio" value="spam_or_misleading" />
            <span>Spam, scam, or deceptive metadata</span>
          </label>
          <label class="flex items-center gap-2">
            <input v-model="reportReason" type="radio" value="harmful_content" />
            <span>Harmful, hateful, or age-inappropriate content</span>
          </label>
        </div>
        <p v-if="reportSubmitted" class="mt-3 rounded-md border border-[var(--gvp-success)] bg-[var(--gvp-success-soft)] p-2 text-xs font-semibold text-[var(--gvp-success)]" role="status">
          Report {{ reportCaseId }} received. Our Trust & Safety team will review it shortly.
        </p>
        <p v-if="reportError" class="mt-3 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] p-2 text-xs font-semibold text-[var(--gvp-live)]" role="alert">
          {{ reportError }}
        </p>
        <div class="mt-5 flex justify-end gap-2">
          <button
            type="button"
            class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-4 py-1.5 text-xs font-semibold text-[var(--gvp-text)] hover:bg-[var(--gvp-canvas)]"
            @click="closeReportModal"
          >
            Close
          </button>
          <button
            type="button"
            :disabled="reportSubmitting || reportSubmitted"
            class="rounded-md bg-[var(--gvp-brand)] px-4 py-1.5 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)] disabled:opacity-50"
            @click="submitReport"
          >
            {{ reportSubmitting ? 'Submitting…' : 'Submit Report' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Share Modal -->
    <WatchShareModal
      v-if="showShareModal"
      :video-id="videoId"
      :start-seconds="shareStartSeconds"
      @close="showShareModal = false"
    />
  </div>
</template>
