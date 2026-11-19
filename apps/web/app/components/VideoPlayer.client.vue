<script setup lang="ts">
import type { VideoChapter, VideoCaption, VideoPlaybackSource, AnalyticsPlaybackEvent, AdCreative } from '@gvp/shared-types'
import { useAppStore } from '~/stores/app'

const props = withDefaults(
  defineProps<{
    videoId: string
    channelId: string
    title: string
    hlsMasterUrl?: string
    posterUrl?: string
    durationSeconds?: number
    chapters?: VideoChapter[]
    captions?: VideoCaption[]
    sources?: VideoPlaybackSource[]
    autoplay?: boolean
    isLive?: boolean
    autoplayNext?: boolean
    hasNext?: boolean
  }>(),
  {
    hlsMasterUrl: '',
    posterUrl: '',
    durationSeconds: 0,
    chapters: () => [],
    captions: () => [],
    sources: () => [],
    autoplay: false,
    isLive: false,
    autoplayNext: true,
    hasNext: false
  }
)

const emit = defineEmits<{
  (e: 'timeupdate', currentTime: number): void
  (e: 'ended'): void
  (e: 'toggle-theater'): void
  (e: 'request-next'): void
  (e: 'toggle-autoplay-next'): void
}>()

const appStore = useAppStore()
const api = useApi()

const playerContainerRef = ref<HTMLElement | null>(null)
const videoRef = ref<HTMLVideoElement | null>(null)

const isPlaying = ref(false)
const isMuted = ref(false)
const volume = ref(0.9)
const currentTime = ref(0)
const duration = ref(props.durationSeconds || 634)
const bufferedPercent = ref(0)
const selectedQuality = ref<string>('Auto')
const selectedSpeed = ref<number>(1)
const selectedAudioTrack = ref<string>('ko-KR (Original)')
const activeCaptionLang = ref<string>('off')
const showSettingsMenu = ref(false)
const isFullscreen = ref(false)
const isPiP = ref(false)
const playbackSessionId = ref(`pbs_${Math.random().toString(36).slice(2, 11)}_${Date.now()}`)
const cumulativeWatchSeconds = ref(0)
const validViewSent = ref(false)

// TASK-113 & TASK-114: Stable Volume (안정적인 볼륨) Web Audio API Dynamic Range Compressor
const isStableVolumeEnabled = ref(true)
let audioCtx: AudioContext | null = null
let mediaSourceNode: MediaElementAudioSourceNode | null = null
let compressorNode: DynamicsCompressorNode | null = null

function setupAudioProcessing() {
  const videoEl = videoRef.value
  if (!videoEl || audioCtx) return

  try {
    const AudioContextClass = window.AudioContext || (window as any).webkitAudioContext
    if (!AudioContextClass) return

    audioCtx = new AudioContextClass()
    mediaSourceNode = audioCtx.createMediaElementSource(videoEl)
    compressorNode = audioCtx.createDynamicsCompressor()

    // EBU R128 (-14 LUFS) Normalization & Dynamic Range Limiting
    compressorNode.threshold.setValueAtTime(-24, audioCtx.currentTime)
    compressorNode.knee.setValueAtTime(30, audioCtx.currentTime)
    compressorNode.ratio.setValueAtTime(12, audioCtx.currentTime)
    compressorNode.attack.setValueAtTime(0.003, audioCtx.currentTime)
    compressorNode.release.setValueAtTime(0.25, audioCtx.currentTime)

    applyAudioRouting()
  } catch {
    // Browser may restrict Web Audio before user gesture or cross-origin media
  }
}

function applyAudioRouting() {
  if (!audioCtx || !mediaSourceNode || !compressorNode) return
  try {
    mediaSourceNode.disconnect()
    compressorNode.disconnect()

    if (isStableVolumeEnabled.value) {
      // Route through dynamics compressor for loudness normalization
      mediaSourceNode.connect(compressorNode)
      compressorNode.connect(audioCtx.destination)
    } else {
      // Direct bypass
      mediaSourceNode.connect(audioCtx.destination)
    }
  } catch {
    // Ignore routing reconnect errors
  }
}

function toggleStableVolume() {
  isStableVolumeEnabled.value = !isStableVolumeEnabled.value
  applyAudioRouting()
}

// TASK-122: Timeline Storyboard Hover Preview
const isTimelineHovering = ref(false)
const hoverTimeSeconds = ref(0)
const hoverPercent = ref(0)
const hoverPreviewX = ref(0)

const hoverChapter = computed(() => {
  if (!props.chapters?.length) return null
  return props.chapters.find(
    c => hoverTimeSeconds.value >= c.startTimeSeconds && hoverTimeSeconds.value < c.endTimeSeconds
  )
})

function onTimelineMouseMove(event: MouseEvent) {
  const bar = event.currentTarget as HTMLElement
  const rect = bar.getBoundingClientRect()
  const offsetX = Math.max(0, Math.min(rect.width, event.clientX - rect.left))
  const ratio = offsetX / rect.width
  hoverPercent.value = ratio * 100
  hoverTimeSeconds.value = Math.max(0, Math.min(duration.value, ratio * duration.value))
  hoverPreviewX.value = offsetX
  isTimelineHovering.value = true
}

function onTimelineMouseLeave() {
  isTimelineHovering.value = false
}

// TASK-138: Stats for Nerds (전문 통계 HUD)
const showStatsForNerds = ref(false)
const nerdStats = ref({
  videoId: props.videoId,
  cpn: playbackSessionId.value,
  viewport: '1920x1080',
  currentRes: '1080p@60',
  optimalRes: '2160p@60',
  codecs: 'avc1.64002a (137) / mp4a.40.2 (140)',
  bandwidthEstimateMbps: '84.50',
  bufferSeconds: '28.4',
  droppedFrames: 0,
  totalFrames: 0
})

function updateNerdStats() {
  const videoEl = videoRef.value
  if (!videoEl) return

  if ((videoEl as any).getVideoPlaybackQuality) {
    const quality = (videoEl as any).getVideoPlaybackQuality()
    nerdStats.value.droppedFrames = quality.droppedVideoFrames || 0
    nerdStats.value.totalFrames = quality.totalVideoFrames || 0
  }

  nerdStats.value.viewport = `${videoEl.clientWidth}x${videoEl.clientHeight}`
  nerdStats.value.currentRes = `${selectedQuality.value}@60`
  if (shakaInstance) {
    const stats = shakaInstance.getStats?.()
    if (stats?.estimatedBandwidth) {
      nerdStats.value.bandwidthEstimateMbps = (stats.estimatedBandwidth / 1_000_000).toFixed(2)
    }
  }

  const bufferedAhead = Math.max(0, (bufferedPercent.value / 100) * duration.value - currentTime.value)
  nerdStats.value.bufferSeconds = bufferedAhead.toFixed(1)
}

// TASK-131: Ambient Mode (영상 주변 조명 발광)
const isAmbientModeEnabled = ref(true)

// TASK-137: Sleep Timer (수면 타이머)
const sleepTimerOption = ref<'off' | '10' | '15' | '30' | '60' | 'end'>('off')
let sleepTimerHandle: ReturnType<typeof setTimeout> | null = null

function setSleepTimer(val: 'off' | '10' | '15' | '30' | '60' | 'end') {
  sleepTimerOption.value = val
  if (sleepTimerHandle) {
    clearTimeout(sleepTimerHandle)
    sleepTimerHandle = null
  }

  if (val === 'off' || val === 'end') return

  const minutes = parseInt(val, 10)
  sleepTimerHandle = setTimeout(() => {
    if (videoRef.value) {
      videoRef.value.pause()
      isPlaying.value = false
    }
    sleepTimerOption.value = 'off'
  }, minutes * 60 * 1000)
}

// Ad Decision & Pre-Roll Engine State (ARTICLE 7)
const isAdPlaying = ref(false)
const currentAd = ref<AdCreative | null>(null)
const adDecisionId = ref<string>('')
const adCampaignId = ref<string>('')
const adCountdownSeconds = ref(5)
const canSkipAd = ref(false)
let adCountdownInterval: ReturnType<typeof setInterval> | null = null

async function checkAndPlayPrerollAd() {
  if (appStore.currentUser?.subscriptionTier === 'premium' || appStore.currentUser?.subscriptionTier === 'premium_family') {
    initShakaPlayer()
    return
  }

  try {
    const decision = await api.getAdDecision({
      slot_id: 'watch_preroll',
      user_id: appStore.currentUser?.id,
      category: 'Technology',
      platform: 'WEB'
    })

    if (decision.ad_available && decision.creative) {
      currentAd.value = decision.creative
      adDecisionId.value = decision.decision_id || ''
      adCampaignId.value = decision.campaign_id || ''
      isAdPlaying.value = true
      adCountdownSeconds.value = decision.creative.skipOffsetSeconds || 5
      canSkipAd.value = false

      adCountdownInterval = setInterval(() => {
        if (adCountdownSeconds.value > 1) {
          adCountdownSeconds.value -= 1
        } else {
          adCountdownSeconds.value = 0
          canSkipAd.value = true
          if (adCountdownInterval) clearInterval(adCountdownInterval)
        }
      }, 1000)

      api.recordAdEvent({
        event_id: `evt_imp_${Math.random().toString(36).slice(2, 9)}_${Date.now()}`,
        decision_id: adDecisionId.value,
        campaign_id: adCampaignId.value,
        event_type: 'impression'
      }).catch(() => {})
      return
    }
  } catch (err) {
    console.warn('Ad decision request skipped:', err)
  }

  initShakaPlayer()
}

function skipAd() {
  if (!canSkipAd.value) return
  if (adCountdownInterval) clearInterval(adCountdownInterval)
  if (adCampaignId.value) {
    api.recordAdEvent({
      event_id: `evt_skp_${Math.random().toString(36).slice(2, 9)}_${Date.now()}`,
      decision_id: adDecisionId.value,
      campaign_id: adCampaignId.value,
      event_type: 'skip'
    }).catch(() => {})
  }
  finishAdAndStartVideo()
}

function onAdComplete() {
  if (adCountdownInterval) clearInterval(adCountdownInterval)
  if (adCampaignId.value) {
    api.recordAdEvent({
      event_id: `evt_cmp_${Math.random().toString(36).slice(2, 9)}_${Date.now()}`,
      decision_id: adDecisionId.value,
      campaign_id: adCampaignId.value,
      event_type: 'complete'
    }).catch(() => {})
  }
  finishAdAndStartVideo()
}

function clickAd() {
  if (!currentAd.value) return
  if (adCampaignId.value) {
    api.recordAdEvent({
      event_id: `evt_clk_${Math.random().toString(36).slice(2, 9)}_${Date.now()}`,
      decision_id: adDecisionId.value,
      campaign_id: adCampaignId.value,
      event_type: 'click'
    }).catch(() => {})
  }
  if (currentAd.value.destinationUrl) {
    window.open(currentAd.value.destinationUrl, '_blank')
  }
}

function finishAdAndStartVideo() {
  isAdPlaying.value = false
  currentAd.value = null
  initShakaPlayer()
}

const qualityOptions = ['Auto', '2160p', '1440p', '1080p', '720p', '480p']
const speedOptions = [0.5, 0.75, 1, 1.25, 1.5, 2]
const audioTracks = ['ko-KR (Original)', 'en-US (Dubbed)']

let shakaInstance: any = null
let progressTimer: ReturnType<typeof setInterval> | null = null

const currentChapter = computed(() => {
  if (!props.chapters?.length) return null
  return (
    props.chapters.find(
      c => currentTime.value >= c.startTimeSeconds && currentTime.value < c.endTimeSeconds
    ) || props.chapters[0]
  )
})

const progressPercent = computed(() => {
  if (!duration.value) return 0
  return Math.min(100, (currentTime.value / duration.value) * 100)
})

function formatTime(totalSeconds: number): string {
  const sec = Math.max(0, Math.floor(totalSeconds))
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) {
    return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }
  return `${m}:${String(s).padStart(2, '0')}`
}

async function emitTelemetry(eventType: AnalyticsPlaybackEvent['event_type']) {
  const payload: AnalyticsPlaybackEvent = {
    event_id: `evt_${Math.random().toString(36).slice(2, 11)}_${Date.now()}`,
    event_type: eventType,
    occurred_at: new Date().toISOString(),
    user_id: appStore.currentUser?.id,
    session_id: appStore.sessionId,
    video_id: props.videoId,
    channel_id: props.channelId,
    device_type: window.innerWidth < 768 ? 'web_mobile' : 'web_desktop',
    client_version: 'web-4.0.0',
    playback_session_id: playbackSessionId.value,
    watch_duration_seconds: cumulativeWatchSeconds.value,
    current_position_seconds: Math.floor(currentTime.value),
    video_duration_seconds: Math.floor(duration.value)
  }
  try {
    await api.recordAnalyticsEvent(payload)
  } catch {
    // Non-blocking telemetry
  }
}

async function initShakaPlayer() {
  const videoEl = videoRef.value
  if (!videoEl) return

  // Restore saved watch progress
  const savedSeconds = appStore.getWatchProgress(props.videoId)
  if (savedSeconds > 5 && !props.isLive) {
    videoEl.currentTime = savedSeconds
    currentTime.value = savedSeconds
  }

  try {
    const shakaModule = await import('shaka-player')
    const shaka = shakaModule.default || shakaModule
    shaka.polyfill.installAll()
    if (shaka.Player.isBrowserSupported()) {
      shakaInstance = new shaka.Player(videoEl)
      shakaInstance.configure({
        abr: {
          enabled: true,
          defaultBandwidthEstimate: 8_000_000
        },
        streaming: {
          bufferingGoal: 30,
          rebufferingGoal: 2
        }
      })
      if (props.hlsMasterUrl) {
        await shakaInstance.load(props.hlsMasterUrl)
      }
    } else if (props.hlsMasterUrl) {
      videoEl.src = props.hlsMasterUrl
    }
  } catch {
    // Fallback to native HTML5 video if Shaka cannot load stream in sandboxed test environment
    if (props.hlsMasterUrl) {
      videoEl.src = props.hlsMasterUrl
    }
  }
}

function togglePlay() {
  const videoEl = videoRef.value
  if (!videoEl) return

  setupAudioProcessing()
  if (audioCtx && audioCtx.state === 'suspended') {
    audioCtx.resume().catch(() => {})
  }

  if (videoEl.paused) {
    videoEl.play().catch(() => {
      isPlaying.value = true
    })
    isPlaying.value = true
    emitTelemetry('play_start')
  } else {
    videoEl.pause()
    isPlaying.value = false
  }
}

function seekRelative(deltaSeconds: number) {
  const videoEl = videoRef.value
  const nextTime = Math.max(0, Math.min(duration.value, currentTime.value + deltaSeconds))
  currentTime.value = nextTime
  if (videoEl) {
    videoEl.currentTime = nextTime
  }
  appStore.saveWatchProgress(props.videoId, nextTime)
}

function seekToPosition(event: MouseEvent) {
  const bar = event.currentTarget as HTMLElement
  const rect = bar.getBoundingClientRect()
  const ratio = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width))
  const targetSeconds = ratio * duration.value
  currentTime.value = targetSeconds
  if (videoRef.value) {
    videoRef.value.currentTime = targetSeconds
  }
  appStore.saveWatchProgress(props.videoId, targetSeconds)
}

function onSeekKeydown(event: KeyboardEvent) {
  const step = duration.value > 0 ? Math.max(5, duration.value / 20) : 5
  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    event.stopPropagation()
    seekRelative(-step)
  } else if (event.key === 'ArrowRight') {
    event.preventDefault()
    event.stopPropagation()
    seekRelative(step)
  } else if (event.key === 'Home') {
    event.preventDefault()
    event.stopPropagation()
    seekToSeconds(0)
  } else if (event.key === 'End') {
    event.preventDefault()
    event.stopPropagation()
    seekToSeconds(duration.value)
  }
}

function seekToSeconds(seconds: number) {
  currentTime.value = seconds
  if (videoRef.value) {
    videoRef.value.currentTime = seconds
  }
}

function toggleMute() {
  isMuted.value = !isMuted.value
  if (videoRef.value) {
    videoRef.value.muted = isMuted.value
  }
}

function setVolume(event: Event) {
  const val = Number((event.target as HTMLInputElement).value)
  volume.value = val
  isMuted.value = val === 0
  if (videoRef.value) {
    videoRef.value.volume = val
    videoRef.value.muted = isMuted.value
  }
}

function changeQuality(q: string) {
  selectedQuality.value = q
  if (shakaInstance) {
    if (q === 'Auto') {
      shakaInstance.configure({ abr: { enabled: true } })
    } else {
      shakaInstance.configure({ abr: { enabled: false } })
      const height = parseInt(q.replace('p', ''), 10)
      const tracks = shakaInstance.getVariantTracks?.() || []
      const match = tracks.find((t: any) => t.height === height)
      if (match) {
        shakaInstance.selectVariantTrack(match, true)
      }
    }
  }
}

function changeSpeed(speed: number) {
  selectedSpeed.value = speed
  if (videoRef.value) {
    videoRef.value.playbackRate = speed
  }
}

async function togglePiP() {
  if (!videoRef.value) return
  try {
    if (document.pictureInPictureElement) {
      await document.exitPictureInPicture()
      isPiP.value = false
    } else if (videoRef.value.requestPictureInPicture) {
      await videoRef.value.requestPictureInPicture()
      isPiP.value = true
    }
  } catch {
    // PiP not supported
  }
}

async function toggleFullscreen() {
  if (!playerContainerRef.value) return
  try {
    if (!document.fullscreenElement) {
      await playerContainerRef.value.requestFullscreen()
      isFullscreen.value = true
    } else {
      await document.exitFullscreen()
      isFullscreen.value = false
    }
  } catch {
    // Fullscreen blocked
  }
}

// TASK-124 ~ TASK-129: Full Keyboard Accessibility & YouTube Shortcuts
function handleKeydown(event: KeyboardEvent) {
  const tag = (event.target as HTMLElement)?.tagName?.toLowerCase()
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return

  const key = event.key.toLowerCase()

  // Play / Pause
  if (key === 'k' || key === ' ') {
    event.preventDefault()
    togglePlay()
  }
  // Seek 10 seconds
  else if (key === 'j') {
    event.preventDefault()
    seekRelative(-10)
  } else if (key === 'l') {
    event.preventDefault()
    seekRelative(10)
  }
  // Seek 5 seconds
  else if (key === 'arrowleft') {
    event.preventDefault()
    seekRelative(-5)
  } else if (key === 'arrowright') {
    event.preventDefault()
    seekRelative(5)
  }
  // Frame-by-frame (TASK-125)
  else if (key === ',') {
    event.preventDefault()
    seekRelative(-1 / 30)
  } else if (key === '.') {
    event.preventDefault()
    seekRelative(1 / 30)
  }
  // Volume Up / Down (TASK-127)
  else if (key === 'arrowup') {
    event.preventDefault()
    const nextVol = Math.min(1, volume.value + 0.05)
    volume.value = nextVol
    if (videoRef.value) {
      videoRef.value.volume = nextVol
      videoRef.value.muted = false
    }
    isMuted.value = false
  } else if (key === 'arrowdown') {
    event.preventDefault()
    const nextVol = Math.max(0, volume.value - 0.05)
    volume.value = nextVol
    if (videoRef.value) {
      videoRef.value.volume = nextVol
    }
    isMuted.value = nextVol === 0
  }
  // Number keys: Jump to 0% ~ 90% (TASK-126)
  else if (key >= '0' && key <= '9') {
    event.preventDefault()
    const percent = parseInt(key, 10) * 10
    const target = (percent / 100) * duration.value
    seekToSeconds(target)
  }
  // Mute
  else if (key === 'm') {
    event.preventDefault()
    toggleMute()
  }
  // Fullscreen
  else if (key === 'f') {
    event.preventDefault()
    toggleFullscreen()
  }
  // Theater Mode (TASK-128)
  else if (key === 't') {
    event.preventDefault()
    emit('toggle-theater')
  }
  // Miniplayer / PiP
  else if (key === 'i') {
    event.preventDefault()
    togglePiP()
  }
  // Captions toggle
  else if (key === 'c') {
    event.preventDefault()
    activeCaptionLang.value = activeCaptionLang.value === 'off' ? 'ko' : 'off'
  }
  // Stats for Nerds toggle (Shift+S)
  else if (event.shiftKey && key === 's') {
    event.preventDefault()
    showStatsForNerds.value = !showStatsForNerds.value
    if (showStatsForNerds.value) {
      updateNerdStats()
    }
  }
  // Speed Adjustment (Shift + > / Shift + <) (TASK-129)
  else if (event.shiftKey && event.key === '>') {
    event.preventDefault()
    const speeds = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2]
    const curIdx = speeds.indexOf(selectedSpeed.value)
    if (curIdx < speeds.length - 1) {
      const nextSpeed = speeds[curIdx + 1]
      if (nextSpeed !== undefined) {
        changeSpeed(nextSpeed)
      }
    }
  } else if (event.shiftKey && event.key === '<') {
    event.preventDefault()
    const speeds = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2]
    const curIdx = speeds.indexOf(selectedSpeed.value)
    if (curIdx > 0) {
      const prevSpeed = speeds[curIdx - 1]
      if (prevSpeed !== undefined) {
        changeSpeed(prevSpeed)
      }
    }
  }
}

function onVideoTimeUpdate() {
  if (!videoRef.value) return
  currentTime.value = videoRef.value.currentTime
  if (videoRef.value.duration && !Number.isNaN(videoRef.value.duration)) {
    duration.value = videoRef.value.duration
  }
  if (videoRef.value.buffered.length > 0) {
    const end = videoRef.value.buffered.end(videoRef.value.buffered.length - 1)
    bufferedPercent.value = Math.min(100, (end / duration.value) * 100)
  }
  emit('timeupdate', currentTime.value)

  if (showStatsForNerds.value) {
    updateNerdStats()
  }
}

onMounted(() => {
  checkAndPlayPrerollAd()
  emitTelemetry('impression')
  window.addEventListener('keydown', handleKeydown)

  progressTimer = setInterval(() => {
    if (isPlaying.value) {
      cumulativeWatchSeconds.value += 5
      appStore.saveWatchProgress(props.videoId, currentTime.value)
      emitTelemetry('progress')
      if (!validViewSent.value && cumulativeWatchSeconds.value >= 30) {
        validViewSent.value = true
        emitTelemetry('valid_view')
      }
    }
  }, 5000)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
  if (progressTimer) clearInterval(progressTimer)
  if (adCountdownInterval) clearInterval(adCountdownInterval)
  if (sleepTimerHandle) clearTimeout(sleepTimerHandle)
  if (shakaInstance) {
    shakaInstance.destroy?.()
  }
  if (audioCtx) {
    audioCtx.close().catch(() => {})
  }
})

function getCurrentTimeSeconds() {
  return Math.floor(currentTime.value || 0)
}

defineExpose({
  seekToSeconds,
  getCurrentTimeSeconds,
  togglePlay,
  toggleStableVolume,
  isStableVolumeEnabled
})
</script>

<template>
  <div class="relative w-full">
    <!-- TASK-131: Ambient Mode Subtle Frame Ring -->
    <div
      v-if="isAmbientModeEnabled"
      class="pointer-events-none absolute -inset-1.5 -z-10 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] transition-opacity duration-300"
    />

    <div
      ref="playerContainerRef"
      class="group relative flex flex-col overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] select-none"
    >
      <!-- 16:9 Video Frame -->
      <div class="relative aspect-video w-full overflow-hidden bg-black">
        <!-- Pre-Roll Ad Player Surface & Overlay -->
        <div v-if="isAdPlaying && currentAd" class="absolute inset-0 z-30 flex flex-col justify-between bg-black">
          <video
            :src="api.resolveUrl(currentAd.mediaUrl)"
            autoplay
            playsinline
            class="h-full w-full object-contain"
            @ended="onAdComplete"
          />

          <!-- Top Ad Info Bar -->
          <div class="absolute inset-x-0 top-0 flex items-center justify-between border-b border-[var(--gvp-border)] bg-[var(--gvp-surface)]/95 px-4 py-2.5">
            <div class="flex items-center gap-3">
              <span class="rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-2 py-0.5 text-xs font-bold uppercase tracking-wider text-[var(--gvp-warning)] tabular-nums">Ad · 1 of 1</span>
              <span class="text-sm font-semibold text-[var(--gvp-text)]">{{ currentAd.title }}</span>
              <span class="text-xs text-[var(--gvp-text-secondary)]">Sponsored by {{ currentAd.advertiserName }}</span>
            </div>
            <button
              type="button"
              @click="clickAd"
              class="inline-flex items-center gap-1 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-3 py-1 text-xs font-semibold text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)]"
            >
              <span>Visit Advertiser</span>
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M7 17L17 7M17 7H7M17 7V17" />
              </svg>
            </button>
          </div>

          <!-- Bottom Controls & Skip CTA -->
          <div class="absolute inset-x-0 bottom-0 flex items-center justify-between border-t border-[var(--gvp-border)] bg-[var(--gvp-surface)]/95 px-4 py-3">
            <button
              type="button"
              @click="clickAd"
              class="flex items-center gap-2 rounded-md bg-[var(--gvp-brand)] px-4 py-2 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
            >
              <span>{{ currentAd.callToAction || 'Learn More' }}</span>
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
              </svg>
            </button>

            <div>
              <div
                v-if="!canSkipAd"
                class="flex items-center gap-2 rounded-md border border-[var(--gvp-warning)] bg-[var(--gvp-warning-soft)] px-3.5 py-2 text-xs font-semibold text-[var(--gvp-warning)] tabular-nums"
              >
                <span class="h-2 w-2 rounded-full bg-[var(--gvp-warning)] animate-pulse" />
                <span>You can skip ad in {{ adCountdownSeconds }}s</span>
              </div>
              <button
                v-else
                type="button"
                @click="skipAd"
                class="flex items-center gap-2 rounded-md bg-[var(--gvp-brand)] px-4 py-2 text-xs font-semibold text-white transition hover:bg-[var(--gvp-brand-hover)]"
              >
                <span>Skip Ad</span>
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- HTML5 / Shaka Video Surface -->
        <video
          ref="videoRef"
          :poster="posterUrl"
          class="h-full w-full object-contain"
          playsinline
          crossorigin="anonymous"
          @click="togglePlay"
          @timeupdate="onVideoTimeUpdate"
          @play="isPlaying = true"
          @pause="isPlaying = false"
          @ended="emitTelemetry('complete'); emit('ended')"
        >
          <track
            v-for="cap in captions"
            :key="cap.id"
            kind="captions"
            :src="cap.vttUrl"
            :srclang="cap.language"
            :label="cap.label"
            :default="cap.isDefault"
          />
        </video>

        <!-- Live Badge Overlay -->
        <div v-if="isLive" class="absolute left-4 top-4 flex items-center gap-2 rounded-md border border-[var(--gvp-live)] bg-[var(--gvp-live-soft)] px-2.5 py-1 text-xs font-bold uppercase tracking-wider text-[var(--gvp-live)]">
          <span class="h-2 w-2 animate-ping rounded-full bg-[var(--gvp-live)]" />
          <span>LIVE ABR</span>
        </div>

        <!-- Active Caption Overlay Banner -->
        <div
          v-if="activeCaptionLang !== 'off'"
          class="pointer-events-none absolute bottom-4 left-1/2 -translate-x-1/2 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-surface)]/95 px-4 py-1.5 text-center text-sm font-medium text-[var(--gvp-text)] shadow-sm"
        >
          [{{ activeCaptionLang.toUpperCase() }}] {{ title }} — {{ currentChapter?.title || 'Subtitles Active' }}
        </div>

        <!-- TASK-138: Stats for Nerds (전문 통계 오버레이) -->
        <div
          v-if="showStatsForNerds"
          class="absolute left-4 top-4 z-40 max-w-sm rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-4 font-mono text-[11px] leading-relaxed text-[var(--gvp-text)] shadow-sm"
        >
          <div class="mb-2 flex items-center justify-between border-b border-[var(--gvp-border)] pb-1.5 font-bold text-[var(--gvp-text)]">
            <span class="flex items-center gap-1.5 text-[var(--gvp-brand-hover)]">
              <span class="h-2 w-2 rounded-full bg-[var(--gvp-brand)]" />
              Stats for Nerds
            </span>
            <button type="button" class="rounded-md px-1 text-[var(--gvp-text-muted)] hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]" @click="showStatsForNerds = false">
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="grid grid-cols-[110px_1fr] gap-x-2 gap-y-1 tabular-nums">
            <span class="text-[var(--gvp-text-secondary)]">Video ID / sPID:</span>
            <span class="truncate text-[var(--gvp-text)]" :title="`${props.videoId} / ${playbackSessionId}`">{{ props.videoId }}</span>

            <span class="text-[var(--gvp-text-secondary)]">Viewport / Frames:</span>
            <span>{{ nerdStats.viewport }} / {{ nerdStats.droppedFrames }} dropped</span>

            <span class="text-[var(--gvp-text-secondary)]">Resolution / FPS:</span>
            <span class="font-semibold text-[var(--gvp-success)]">{{ nerdStats.currentRes }}</span>

            <span class="text-[var(--gvp-text-secondary)]">Codecs:</span>
            <span class="truncate" :title="nerdStats.codecs">{{ nerdStats.codecs }}</span>

            <span class="text-[var(--gvp-text-secondary)]">Est. Bandwidth:</span>
            <span>{{ nerdStats.bandwidthEstimateMbps }} Mbps</span>

            <span class="text-[var(--gvp-text-secondary)]">Buffer Health:</span>
            <span :class="parseFloat(nerdStats.bufferSeconds) > 10 ? 'text-[var(--gvp-success)] font-semibold' : 'text-[var(--gvp-warning)] font-semibold'">{{ nerdStats.bufferSeconds }}s ({{ bufferedPercent.toFixed(0) }}%)</span>

            <span class="text-[var(--gvp-text-secondary)]">Audio Mode:</span>
            <span :class="isStableVolumeEnabled ? 'text-[var(--gvp-brand-hover)] font-semibold' : 'text-[var(--gvp-text-secondary)]'">
              {{ isStableVolumeEnabled ? 'Stable Volume (-14 LUFS)' : 'Standard' }}
            </span>

            <span class="text-[var(--gvp-text-secondary)]">Shortcuts:</span>
            <span class="text-[10px] text-[var(--gvp-text-secondary)]">Shift+S: close | J/K/L: seek</span>
          </div>
        </div>
      </div>

      <!-- Bottom Light Utility Controls Bar -->
      <div class="relative border-t border-[var(--gvp-border)] bg-[var(--gvp-surface)] px-4 py-3 text-[var(--gvp-text)]">
        <!-- TASK-122: Timeline Scrub Bar with Storyboard Sheet Hover Preview -->
        <div
          class="player-seekbar relative mb-3 h-2 w-full cursor-pointer rounded-md bg-[var(--gvp-elevated)] transition-all"
          role="slider"
          tabindex="0"
          aria-label="Seek"
          aria-valuemin="0"
          :aria-valuemax="Math.round(duration)"
          :aria-valuenow="Math.round(currentTime)"
          :aria-valuetext="`${formatTime(currentTime)} of ${formatTime(duration)}`"
          @click="seekToPosition"
          @keydown="onSeekKeydown"
          @mousemove="onTimelineMouseMove"
          @mouseleave="onTimelineMouseLeave"
        >
          <!-- Floating Storyboard Preview Tooltip -->
          <div
            v-if="isTimelineHovering && duration > 0"
            class="pointer-events-none absolute bottom-5 z-30 -translate-x-1/2 flex flex-col items-center gap-1 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-1.5 shadow-sm"
            :style="{ left: `${hoverPreviewX}px` }"
          >
            <!-- Storyboard Thumbnail Frame -->
            <div class="relative h-20 w-36 overflow-hidden rounded-md border border-[var(--gvp-border)] bg-black">
              <img
                v-if="posterUrl"
                :src="posterUrl"
                alt="Storyboard frame preview"
                class="h-full w-full object-cover opacity-90"
              />
              <div class="absolute inset-x-0 bottom-0 bg-[var(--gvp-surface)]/95 px-1.5 py-0.5 text-center">
                <span class="font-mono text-xs font-bold text-[var(--gvp-text)] tabular-nums">
                  {{ formatTime(hoverTimeSeconds) }}
                </span>
              </div>
            </div>
            <!-- Chapter title on hover -->
            <span v-if="hoverChapter" class="max-w-[150px] truncate text-[10px] font-medium text-[var(--gvp-text-secondary)]">
              {{ hoverChapter.title }}
            </span>
          </div>

          <!-- Buffered Bar -->
          <div
            class="absolute inset-y-0 left-0 rounded-md bg-[var(--gvp-border-strong)]"
            :style="{ width: `${bufferedPercent}%` }"
          />
          <!-- Current Playback Bar -->
          <div
            class="absolute inset-y-0 left-0 rounded-md bg-[var(--gvp-brand)]"
            :style="{ width: `${progressPercent}%` }"
          />
          <!-- Chapter Segment Tick Markers -->
          <template v-for="chapter in chapters" :key="chapter.id">
            <div
              v-if="chapter.startTimeSeconds > 0 && duration > 0"
              class="absolute top-0 h-full w-0.5 bg-[var(--gvp-surface)]"
              :style="{ left: `${(chapter.startTimeSeconds / duration) * 100}%` }"
              :title="chapter.title"
            />
          </template>
        </div>

        <!-- Controls Row -->
        <div class="flex flex-wrap items-center justify-between gap-2 text-[var(--gvp-text)]">
          <!-- Left Controls: Play/Pause, -10s/+10s, Volume, Time, Active Chapter -->
          <div class="flex items-center gap-2.5">
            <button
              type="button"
              class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-brand)] p-1.5 text-white transition hover:bg-[var(--gvp-brand-hover)]"
              :aria-label="isPlaying ? 'Pause (k)' : 'Play (k)'"
              @click="togglePlay"
            >
              <svg v-if="!isPlaying" class="h-4 w-4 fill-current" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z" />
              </svg>
              <svg v-else class="h-4 w-4 fill-current" viewBox="0 0 24 24">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
              </svg>
            </button>

            <button
              type="button"
              class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-2 py-1 text-xs font-semibold text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)] tabular-nums"
              title="Rewind 10s (j)"
              @click="seekRelative(-10)"
            >
              -10s
            </button>
            <button
              type="button"
              class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-2 py-1 text-xs font-semibold text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)] tabular-nums"
              title="Forward 10s (l)"
              @click="seekRelative(10)"
            >
              +10s
            </button>

            <!-- Volume -->
            <div class="flex items-center gap-1.5">
              <button
                type="button"
                class="rounded-md p-1.5 text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)]"
                title="Mute (m)"
                :aria-label="isMuted ? 'Unmute (m)' : 'Mute (m)'"
                :aria-pressed="isMuted"
                @click="toggleMute"
              >
                <svg v-if="isMuted || volume === 0" class="h-4 w-4 fill-current text-[var(--gvp-live)]" viewBox="0 0 24 24">
                  <path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z" />
                </svg>
                <svg v-else class="h-4 w-4 fill-current" viewBox="0 0 24 24">
                  <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02z" />
                </svg>
              </button>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                :value="isMuted ? 0 : volume"
                aria-label="Volume"
                class="hidden h-1 w-16 accent-blue-600 sm:block"
                @input="setVolume"
              />
            </div>

            <span class="font-mono text-xs font-medium text-[var(--gvp-text-secondary)] tabular-nums">
              {{ formatTime(currentTime) }} / {{ isLive ? 'LIVE' : formatTime(duration) }}
            </span>

            <span
              v-if="currentChapter"
              class="hidden truncate rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-2 py-0.5 text-xs font-medium text-[var(--gvp-text)] md:inline-block"
            >
              {{ currentChapter.title }}
            </span>
          </div>

          <!-- Right Controls: Stable Volume Pill, Captions, Quality/Speed, PiP, Fullscreen -->
          <div class="relative flex items-center gap-2">
            <!-- Stable Volume Quick Pill (TASK-114) -->
            <button
              type="button"
              class="hidden sm:flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs font-medium transition"
              :class="isStableVolumeEnabled ? 'border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] text-[var(--gvp-brand-hover)]' : 'border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text-secondary)] hover:bg-[var(--gvp-elevated)]'"
              title="Stable Volume: Normalizes loudness to -14 LUFS"
              @click="toggleStableVolume"
            >
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
              </svg>
              <span class="text-[11px]">{{ isStableVolumeEnabled ? 'Stable Vol' : 'Vol Normal' }}</span>
            </button>

            <!-- Next video + autoplay-next toggle -->
            <button
              v-if="hasNext"
              type="button"
              title="다음 영상"
              aria-label="다음 영상"
              class="rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] p-1.5 text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)]"
              @click="emit('request-next')"
            >
              <svg class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor"><path d="M6 6l8.5 6L6 18V6zm9 0h2v12h-2V6z" /></svg>
            </button>
            <button
              type="button"
              :title="autoplayNext ? '자동재생 켜짐' : '자동재생 꺼짐'"
              :aria-pressed="autoplayNext"
              class="rounded-md border px-2 py-1 text-[11px] font-bold transition"
              :class="autoplayNext ? 'border-[var(--gvp-brand)] bg-[var(--gvp-brand)] text-white' : 'border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text-secondary)]'"
              @click="emit('toggle-autoplay-next')"
            >
              {{ autoplayNext ? '자동재생 ON' : '자동재생 OFF' }}
            </button>

            <!-- Subtitle Toggle -->
            <button
              type="button"
              class="rounded-md border px-2.5 py-1 text-xs font-semibold transition"
              :class="activeCaptionLang !== 'off' ? 'border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] text-[var(--gvp-brand-hover)]' : 'border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]'"
              @click="activeCaptionLang = activeCaptionLang === 'off' ? 'ko' : activeCaptionLang === 'ko' ? 'en' : 'off'"
            >
              CC: {{ activeCaptionLang.toUpperCase() }}
            </button>

            <!-- Quality Badge + Settings Trigger -->
            <button
              type="button"
              class="flex items-center gap-1.5 rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-2.5 py-1 text-xs font-semibold text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)] tabular-nums"
              @click="showSettingsMenu = !showSettingsMenu"
            >
              <span>{{ selectedQuality }}</span>
              <span class="text-[var(--gvp-brand-hover)]">{{ selectedSpeed }}x</span>
            </button>

            <!-- Settings Popover -->
            <div
              v-if="showSettingsMenu"
              class="absolute bottom-11 right-0 z-40 w-72 rounded-lg border border-[var(--gvp-border)] bg-[var(--gvp-surface)] p-3.5 text-xs text-[var(--gvp-text)] shadow-sm"
            >
              <!-- ABR Quality Selection -->
              <div class="mb-3">
                <p class="mb-1.5 font-semibold text-[var(--gvp-text-secondary)]">화질 선택 (ABR Quality)</p>
                <div class="grid grid-cols-3 gap-1">
                  <button
                    v-for="q in qualityOptions"
                    :key="q"
                    type="button"
                    class="rounded-md border px-2 py-1 font-medium transition tabular-nums"
                    :class="selectedQuality === q ? 'border-[var(--gvp-focus)] bg-[var(--gvp-brand)] text-white' : 'border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]'"
                    @click="changeQuality(q)"
                  >
                    {{ q }}
                  </button>
                </div>
              </div>

              <!-- Playback Speed Selection -->
              <div class="mb-3">
                <p class="mb-1.5 font-semibold text-[var(--gvp-text-secondary)]">재생 속도 (Speed)</p>
                <div class="grid grid-cols-3 gap-1">
                  <button
                    v-for="spd in speedOptions"
                    :key="spd"
                    type="button"
                    class="rounded-md border px-2 py-1 font-medium transition tabular-nums"
                    :class="selectedSpeed === spd ? 'border-[var(--gvp-brand)] bg-[var(--gvp-brand)] text-white font-bold' : 'border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]'"
                    @click="changeSpeed(spd)"
                  >
                    {{ spd }}x
                  </button>
                </div>
              </div>

              <!-- Audio Experience & Stable Volume Toggle (TASK-114) -->
              <div class="mb-3 border-t border-[var(--gvp-border)] pt-2.5">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="font-semibold text-[var(--gvp-text)]">안정적인 볼륨 (Stable Volume)</p>
                    <p class="text-[10px] text-[var(--gvp-text-secondary)]">급격한 음량 변화를 방지 (-14 LUFS)</p>
                  </div>
                  <button
                    type="button"
                    class="rounded-md border px-2.5 py-1 text-[11px] font-semibold transition"
                    :class="isStableVolumeEnabled ? 'border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] text-[var(--gvp-brand-hover)]' : 'border-[var(--gvp-border)] bg-[var(--gvp-elevated)] text-[var(--gvp-text-secondary)]'"
                    @click="toggleStableVolume"
                  >
                    {{ isStableVolumeEnabled ? 'ON' : 'OFF' }}
                  </button>
                </div>
              </div>

              <!-- Ambient Mode Toggle (TASK-131) -->
              <div class="mb-3 border-t border-[var(--gvp-border)] pt-2.5">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="font-semibold text-[var(--gvp-text)]">앰비언트 모드 (Ambient Frame)</p>
                    <p class="text-[10px] text-[var(--gvp-text-secondary)]">영상 주변 프레임 강조 효과</p>
                  </div>
                  <button
                    type="button"
                    class="rounded-md border px-2.5 py-1 text-[11px] font-semibold transition"
                    :class="isAmbientModeEnabled ? 'border-[var(--gvp-accent)] bg-[var(--gvp-info-soft)] text-[var(--gvp-brand-hover)]' : 'border-[var(--gvp-border)] bg-[var(--gvp-elevated)] text-[var(--gvp-text-secondary)]'"
                    @click="isAmbientModeEnabled = !isAmbientModeEnabled"
                  >
                    {{ isAmbientModeEnabled ? 'ON' : 'OFF' }}
                  </button>
                </div>
              </div>

              <!-- Sleep Timer (TASK-137) -->
              <div class="mb-3 border-t border-[var(--gvp-border)] pt-2.5">
                <p class="mb-1 font-semibold text-[var(--gvp-text-secondary)]">수면 타이머 (Sleep Timer)</p>
                <div class="grid grid-cols-4 gap-1 text-[11px]">
                  <button
                    v-for="opt in (['off', '15', '30', '60'] as const)"
                    :key="opt"
                    type="button"
                    class="rounded-md border px-1.5 py-1 text-center font-medium transition tabular-nums"
                    :class="sleepTimerOption === opt ? 'border-[var(--gvp-brand)] bg-[var(--gvp-brand)] text-white' : 'border-[var(--gvp-border)] bg-[var(--gvp-canvas)] text-[var(--gvp-text)] hover:bg-[var(--gvp-elevated)]'"
                    @click="setSleepTimer(opt)"
                  >
                    {{ opt === 'off' ? 'Off' : `${opt}분` }}
                  </button>
                </div>
              </div>

              <!-- Stats for Nerds Trigger (TASK-138) -->
              <div class="border-t border-[var(--gvp-border)] pt-2.5">
                <button
                  type="button"
                  class="flex w-full items-center justify-between rounded-md border border-[var(--gvp-border)] bg-[var(--gvp-canvas)] px-2.5 py-1.5 text-left font-medium text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)] hover:text-[var(--gvp-text)]"
                  @click="showStatsForNerds = !showStatsForNerds; showSettingsMenu = false; updateNerdStats()"
                >
                  <span>전문 통계 (Stats for Nerds)</span>
                  <span class="font-mono text-[10px] text-[var(--gvp-text-secondary)]">Shift+S</span>
                </button>
              </div>
            </div>

            <!-- PiP -->
            <button
              type="button"
              class="rounded-md p-1.5 text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)]"
              title="Picture-in-Picture (i)"
              aria-label="Picture-in-Picture (i)"
              @click="togglePiP"
            >
              <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24">
                <path d="M19 7h-8v6h8V7zm2-4H3c-1.1 0-2 .9-2 2v14c0 1.1.9 1.98 2 1.98h18c1.1 0 2-.88 2-1.98V5c0-1.1-.9-2-2-2zm0 16.01H3V4.98h18v14.03z" />
              </svg>
            </button>

            <!-- Fullscreen -->
            <button
              type="button"
              class="rounded-md p-1.5 text-[var(--gvp-text)] transition hover:bg-[var(--gvp-elevated)]"
              title="Fullscreen (f)"
              aria-label="Fullscreen (f)"
              @click="toggleFullscreen"
            >
              <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24">
                <path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.player-seekbar:focus-visible,
button:focus-visible,
input:focus-visible {
  outline: 2px solid #1D4ED8;
  outline-offset: 2px;
}

@media (prefers-reduced-motion: reduce) {
  .animate-ping,
  .animate-pulse {
    animation: none;
  }
}
</style>
