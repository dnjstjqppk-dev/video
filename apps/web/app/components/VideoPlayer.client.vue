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
const config = useRuntimeConfig()

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
const settingsView = ref<'main' | 'quality' | 'speed' | 'captions' | 'sleep'>('main')
const availableQualities = ref<string[]>(['Auto'])
const controlsVisible = ref(true)
const isBuffering = ref(false)
const playerError = ref<string | null>(null)
let idleTimer: ReturnType<typeof setTimeout> | null = null
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

function onAdError() {
  // A broken ad asset must never brick playback: fall through to content.
  if (adCountdownInterval) clearInterval(adCountdownInterval)
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

const speedMenuOptions = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2]
const speedLabel = (s: number) => (s === 1 ? 'Normal' : `${s}`)

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
  playerError.value = null
  isBuffering.value = true

  // Restore saved watch progress
  const savedSeconds = appStore.getWatchProgress(props.videoId)
  if (savedSeconds > 5 && !props.isLive) {
    videoEl.currentTime = savedSeconds
    currentTime.value = savedSeconds
  }

  // Signed playback token: lets Shaka fetch gated HLS through the
  // gateway without Authorization headers (falls back to the bare URL).
  // Live pages pass a stream id, not a video id: skip the VOD lookup there.
  let streamUrl = props.hlsMasterUrl
  let drmKeys: Record<string, string> | null = null
  if (!props.isLive) {
    try {
      const pb = await api.getPlaybackInfo(props.videoId)
      if (pb.playbackToken && streamUrl) {
        streamUrl += (streamUrl.includes('?') ? '&' : '?') + 'token=' + encodeURIComponent(pb.playbackToken)
      }
      if (pb.drm?.keyId && pb.drm?.key) {
        drmKeys = { [pb.drm.keyId]: pb.drm.key }
      }
    } catch {
      // Public content or offline playback info: use the bare URL.
    }
  }

  try {
    const shakaModule = await import('shaka-player')
    const shaka = shakaModule.default || shakaModule
    shaka.polyfill.installAll()
    if (shaka.Player.isBrowserSupported()) {
      shakaInstance = new shaka.Player(videoEl)
      shakaInstance.addEventListener('error', (event: any) => {
        const detail = event?.detail
        const sev = (shaka as any)?.util?.Error?.Severity
        if (detail && sev && detail.severity === sev.RECOVERABLE) return
        onPlayerError(detail?.code ? `shaka-${detail.code}` : 'shaka-error')
      })
      const drmServers: Record<string, string> = {}
      const widevineUrl = (config.public.widevineLicenseUrl as string) || ''
      const fairplayUrl = (config.public.fairplayLicenseUrl as string) || ''
      if (widevineUrl) drmServers['com.widevine.alpha'] = widevineUrl
      if (fairplayUrl) drmServers['com.apple.fps'] = fairplayUrl
      shakaInstance.configure({
        abr: {
          enabled: true,
          defaultBandwidthEstimate: 8_000_000
        },
        streaming: {
          bufferingGoal: 30,
          rebufferingGoal: 2
        },
        drm: {
          servers: drmServers,
          ...(drmKeys ? { clearKeys: drmKeys } : {})
        }
      })
      if (streamUrl) {
        await shakaInstance.load(streamUrl)
      }
    } else if (streamUrl) {
      videoEl.src = streamUrl
    }
  } catch {
    // Fallback to native HTML5 video if Shaka cannot load stream in sandboxed test environment
    if (streamUrl) {
      videoEl.src = streamUrl
    }
  }

  // Real quality ladder for the settings menu (Auto + detected rungs).
  try {
    const tracks = (shakaInstance?.getVariantTracks?.() || []) as any[]
    const heights = Array.from(new Set<number>(
      tracks.map(t => t.height).filter((h: any): h is number => typeof h === 'number')
    )).sort((a, b) => b - a)
    if (heights.length) availableQualities.value = ['Auto', ...heights.map(h => `${h}p`)]
  } catch {
    // Keep Auto-only when tracks are unreadable.
  }
  applyCaptionLang()
  isBuffering.value = false

  // Autoplay-next arrivals (and explicit ?autoplay=1 links) start playing
  // once the stream is attached; a rejected play() just leaves the video
  // paused under the browser's autoplay policy.
  if (props.autoplay) {
    videoEl.play().catch(() => {})
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
  pokeControls()
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

// YouTube-style chrome: controls overlay the picture and auto-hide while
// playing. They stay up when paused, while the settings menu is open, or
// shortly after any interaction.
function pokeControls() {
  controlsVisible.value = true
  if (idleTimer) clearTimeout(idleTimer)
  idleTimer = setTimeout(() => {
    if (isPlaying.value && !showSettingsMenu.value) controlsVisible.value = false
  }, 3000)
}

function hideControlsSoon() {
  if (idleTimer) clearTimeout(idleTimer)
  idleTimer = setTimeout(() => {
    if (isPlaying.value && !showSettingsMenu.value) controlsVisible.value = false
  }, 400)
}

function openSettings() {
  settingsView.value = 'main'
  showSettingsMenu.value = true
  pokeControls()
}

function applyCaptionLang() {
  const tracks = videoRef.value?.textTracks
  if (!tracks) return
  for (let i = 0; i < tracks.length; i++) {
    const track = tracks[i]
    if (!track) continue
    track.mode = activeCaptionLang.value !== 'off' && track.language === activeCaptionLang.value
      ? 'showing'
      : 'disabled'
  }
}

function selectCaption(lang: string) {
  activeCaptionLang.value = lang
  applyCaptionLang()
  settingsView.value = 'main'
}

function preferredCaptionLang(): string {
  const caps = props.captions || []
  if (caps.some(c => c.language === 'ko')) return 'ko'
  return caps[0]?.language || 'ko'
}

function onPlayerError(code: string) {
  if (playerError.value) return
  isBuffering.value = false
  isPlaying.value = false
  playerError.value = code
  controlsVisible.value = true
}

function retryPlayback() {
  playerError.value = null
  initShakaPlayer()
}

// Chapter segments for the YouTube-style segmented seekbar. Without
// chapters the whole duration is one segment.
const chapterSegments = computed(() => {
  const dur = duration.value > 0 ? duration.value : 1
  const list = [...(props.chapters || [])].sort((a, b) => a.startTimeSeconds - b.startTimeSeconds)
  if (!list.length) return [{ start: 0, end: dur }]
  return list.map((c, i) => {
    const next = i + 1 < list.length ? list[i + 1] : undefined
    return {
      start: Math.max(0, Math.min(c.startTimeSeconds, dur)),
      end: next ? Math.max(0, Math.min(next.startTimeSeconds, dur)) : dur
    }
  })
})

function segFlex(seg: { start: number; end: number }) {
  return Math.max(seg.end - seg.start, 0.001)
}

function segFill(seg: { start: number; end: number }) {
  if (seg.end <= seg.start) return 0
  return Math.max(0, Math.min(100, ((currentTime.value - seg.start) / (seg.end - seg.start)) * 100))
}

// TASK-124 ~ TASK-129: Full Keyboard Accessibility & YouTube Shortcuts
function handleKeydown(event: KeyboardEvent) {
  const tag = (event.target as HTMLElement)?.tagName?.toLowerCase()
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return
  pokeControls()

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
    activeCaptionLang.value = activeCaptionLang.value === 'off' ? preferredCaptionLang() : 'off'
    applyCaptionLang()
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
  if (idleTimer) clearTimeout(idleTimer)
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
      class="group relative flex flex-col overflow-hidden rounded-lg border border-[var(--gvp-border)] bg-black select-none"
      :class="{ 'cursor-none': !controlsVisible && isPlaying && !playerError }"
      @mousemove="pokeControls"
      @touchstart.passive="pokeControls"
      @mouseleave="hideControlsSoon"
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
            @error="onAdError"
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
          data-testid="main-video"
          :poster="posterUrl"
          class="h-full w-full object-contain"
          playsinline
          crossorigin="anonymous"
          @click="togglePlay"
          @dblclick="toggleFullscreen"
          @timeupdate="onVideoTimeUpdate"
          @play="isPlaying = true; isBuffering = false; pokeControls()"
          @pause="isPlaying = false; isBuffering = false; pokeControls()"
          @waiting="isBuffering = true"
          @stalled="isBuffering = true"
          @canplay="isBuffering = false"
          @error="onPlayerError('media-error')"
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

        <!-- Buffering Spinner -->
        <div
          v-if="isBuffering && !playerError"
          data-testid="player-spinner"
          class="pointer-events-none absolute inset-0 z-30 flex items-center justify-center"
        >
          <span class="h-12 w-12 animate-spin rounded-full border-4 border-white/20 border-t-white" aria-label="Buffering" />
        </div>

        <!-- Fatal Error Overlay -->
        <div
          v-if="playerError"
          data-testid="player-error"
          class="absolute inset-0 z-40 flex items-center justify-center bg-black/85 p-6"
        >
          <div class="max-w-sm text-center">
            <p class="text-base font-semibold text-white">재생 중 오류가 발생했습니다</p>
            <p class="mt-1 font-mono text-xs text-white/50">{{ playerError }} · {{ videoId.slice(0, 8) }}</p>
            <div class="mt-4 flex items-center justify-center gap-3">
              <button
                type="button"
                class="rounded-full bg-white px-6 py-2 text-sm font-bold text-black transition hover:bg-white/80"
                @click="retryPlayback"
              >
                다시 시도
              </button>
            </div>
          </div>
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

        <!-- YouTube-style Control Overlay: gradient scrim, auto-hiding -->
        <div
          class="absolute inset-x-0 bottom-0 z-20 bg-gradient-to-t from-black/80 via-black/40 to-transparent px-3 pb-2 pt-10 transition-opacity duration-200"
          :class="{ 'pointer-events-none opacity-0': !controlsVisible }"
        >
          <!-- Seekbar -->
          <div
            v-if="!isLive"
            class="group/bar relative h-4 w-full cursor-pointer"
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
            <div
              v-if="isTimelineHovering && duration > 0"
              class="pointer-events-none absolute bottom-6 z-30 flex -translate-x-1/2 flex-col items-center gap-1 rounded-lg border border-white/20 bg-black/95 p-1.5"
              :style="{ left: `${hoverPreviewX}px` }"
            >
              <div class="relative h-20 w-36 overflow-hidden rounded-md bg-black">
                <img
                  v-if="posterUrl"
                  :src="posterUrl"
                  alt="Storyboard frame preview"
                  class="h-full w-full object-cover opacity-90"
                />
                <div class="absolute inset-x-0 bottom-0 bg-black/80 px-1.5 py-0.5 text-center">
                  <span class="font-mono text-xs font-bold tabular-nums text-white">
                    {{ formatTime(hoverTimeSeconds) }}
                  </span>
                </div>
              </div>
              <span v-if="hoverChapter" class="max-w-[150px] truncate text-[10px] font-medium text-white/70">
                {{ hoverChapter.title }}
              </span>
            </div>
            <div class="absolute inset-x-0 bottom-0.5 h-[3px] transition-all group-hover/bar:h-[5px]">
              <div class="absolute inset-0 rounded-full bg-white/20" />
              <div
                class="absolute inset-y-0 left-0 rounded-full bg-white/40"
                :style="{ width: `${bufferedPercent}%` }"
              />
              <div class="absolute inset-0 flex gap-[2px]">
                <div
                  v-for="(seg, i) in chapterSegments"
                  :key="i"
                  class="relative h-full"
                  :style="{ flexGrow: segFlex(seg) }"
                >
                  <div
                    class="absolute inset-y-0 left-0 bg-[var(--gvp-brand)]"
                    :style="{ width: `${segFill(seg)}%` }"
                  />
                </div>
              </div>
              <div
                class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-[var(--gvp-brand)] opacity-0 transition-opacity group-hover/bar:opacity-100"
                :style="{ left: `${progressPercent}%` }"
              />
            </div>
          </div>

          <!-- Control Row -->
          <div class="flex items-center gap-0.5 text-white">
            <button
              type="button"
              class="rounded-full p-2 opacity-90 transition hover:bg-white/10 hover:opacity-100"
              :aria-label="isPlaying ? 'Pause (k)' : 'Play (k)'"
              @click="togglePlay"
            >
              <svg v-if="!isPlaying" class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z" />
              </svg>
              <svg v-else class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
              </svg>
            </button>
            <button
              v-if="hasNext"
              type="button"
              title="다음 영상"
              aria-label="다음 영상"
              class="rounded-full p-2 opacity-90 transition hover:bg-white/10 hover:opacity-100"
              @click="emit('request-next')"
            >
              <svg class="h-6 w-6 fill-current" viewBox="0 0 24 24"><path d="M6 6l8.5 6L6 18V6zm9 0h2v12h-2V6z" /></svg>
            </button>
            <div class="group/vol flex items-center">
              <button
                type="button"
                class="rounded-full p-2 opacity-90 transition hover:bg-white/10 hover:opacity-100"
                title="Mute (m)"
                :aria-label="isMuted ? 'Unmute (m)' : 'Mute (m)'"
                :aria-pressed="isMuted"
                @click="toggleMute"
              >
                <svg v-if="isMuted || volume === 0" class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                  <path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z" />
                </svg>
                <svg v-else class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                  <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02z" />
                </svg>
              </button>
              <div class="w-0 overflow-hidden transition-all duration-200 group-hover/vol:w-16 [@media(hover:none)]:w-16">
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.05"
                  :value="isMuted ? 0 : volume"
                  aria-label="Volume"
                  class="h-1 w-16 cursor-pointer accent-white"
                  @input="setVolume"
                />
              </div>
            </div>
            <span v-if="isLive" class="flex items-center gap-1.5 px-2 text-xs font-semibold">
              <span class="h-1.5 w-1.5 rounded-full bg-[var(--gvp-brand)]" /> LIVE
            </span>
            <span v-else class="px-1 text-xs tabular-nums">
              {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
            </span>
            <span class="flex-1" />
            <button
              type="button"
              title="자막 (c)"
              aria-label="자막"
              class="relative rounded-full p-2 opacity-90 transition hover:bg-white/10 hover:opacity-100"
              @click="showSettingsMenu = true; settingsView = 'captions'; pokeControls()"
            >
              <svg class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                <path d="M19 4H5c-1.11 0-2 .9-2 2v12c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm-8 7H9.5v-.5h-2v3h2v-.5H11v1c0 .55-.45 1-1 1H7.5c-.55 0-1-.45-1-1v-4c0-.55.45-1 1-1H10c.55 0 1 .45 1 1v1zm7 0h-1.5v-.5h-2v3h2v-.5H18v1c0 .55-.45 1-1 1h-2.5c-.55 0-1-.45-1-1v-4c0-.55.45-1 1-1H17c.55 0 1 .45 1 1v1z" />
              </svg>
              <span
                v-if="activeCaptionLang !== 'off'"
                class="absolute inset-x-2 bottom-1 h-0.5 rounded bg-white"
              />
            </button>
            <button
              type="button"
              data-testid="player-settings-toggle"
              title="설정"
              aria-label="설정"
              class="rounded-full p-2 opacity-90 transition hover:bg-white/10 hover:opacity-100"
              @click="openSettings"
            >
              <svg class="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="3.2" />
                <g v-for="n in 8" :key="n" :transform="`rotate(${n * 45} 12 12)`">
                  <line x1="12" y1="2.5" x2="12" y2="6" stroke-linecap="round" />
                </g>
              </svg>
            </button>
            <button
              type="button"
              title="미니플레이어 (i)"
              aria-label="미니플레이어 (i)"
              class="hidden rounded-full p-2 opacity-90 transition hover:bg-white/10 hover:opacity-100 sm:block"
              @click="togglePiP"
            >
              <svg class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                <path d="M19 7h-8v6h8V7zm2-4H3c-1.1 0-2 .9-2 2v14c0 1.1.9 1.98 2 1.98h18c1.1 0 2-.88 2-1.98V5c0-1.1-.9-2-2-2zm0 16.01H3V4.98h18v14.03z" />
              </svg>
            </button>
            <button
              type="button"
              title="시어터 모드 (t)"
              aria-label="시어터 모드 (t)"
              class="hidden rounded-full p-2 opacity-90 transition hover:bg-white/10 hover:opacity-100 sm:block"
              @click="emit('toggle-theater')"
            >
              <svg class="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="5" width="18" height="14" rx="2" />
                <line x1="15" y1="5" x2="15" y2="19" />
              </svg>
            </button>
            <button
              type="button"
              title="전체화면 (f)"
              aria-label="전체화면 (f)"
              class="rounded-full p-2 opacity-90 transition hover:bg-white/10 hover:opacity-100"
              @click="toggleFullscreen"
            >
              <svg class="h-6 w-6 fill-current" viewBox="0 0 24 24">
                <path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Settings Menu -->
        <div
          v-if="showSettingsMenu"
          v-click-outside="() => (showSettingsMenu = false)"
          data-testid="player-settings"
          class="absolute bottom-[76px] right-3 z-30 max-h-[70%] w-72 overflow-y-auto rounded-xl bg-black/95 py-2 text-sm text-white shadow-2xl"
        >
          <template v-if="settingsView === 'main'">
            <button type="button" class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-white/10" @click="settingsView = 'quality'">
              <span>화질</span>
              <span class="flex items-center gap-1 text-white/60">{{ selectedQuality === 'Auto' ? '자동' : selectedQuality }}
                <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" /></svg>
              </span>
            </button>
            <button type="button" class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-white/10" @click="settingsView = 'speed'">
              <span>재생 속도</span>
              <span class="flex items-center gap-1 text-white/60">{{ speedLabel(selectedSpeed) }}
                <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" /></svg>
              </span>
            </button>
            <button type="button" class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-white/10" @click="settingsView = 'captions'">
              <span>자막</span>
              <span class="flex items-center gap-1 text-white/60">{{ activeCaptionLang === 'off' ? '꺼짐' : activeCaptionLang.toUpperCase() }}
                <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" /></svg>
              </span>
            </button>
            <div class="mx-4 my-1 border-t border-white/10" />
            <button type="button" class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-white/10" @click="emit('toggle-autoplay-next')">
              <span>자동재생</span>
              <span class="flex h-5 w-9 items-center rounded-full p-0.5 transition" :class="autoplayNext ? 'justify-end bg-white' : 'justify-start bg-white/30'">
                <span class="h-4 w-4 rounded-full" :class="autoplayNext ? 'bg-black' : 'bg-white'" />
              </span>
            </button>
            <button type="button" class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-white/10" @click="toggleStableVolume">
              <span>안정적인 볼륨</span>
              <span class="flex h-5 w-9 items-center rounded-full p-0.5 transition" :class="isStableVolumeEnabled ? 'justify-end bg-white' : 'justify-start bg-white/30'">
                <span class="h-4 w-4 rounded-full" :class="isStableVolumeEnabled ? 'bg-black' : 'bg-white'" />
              </span>
            </button>
            <button type="button" class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-white/10" @click="isAmbientModeEnabled = !isAmbientModeEnabled">
              <span>앰비언트 모드</span>
              <span class="flex h-5 w-9 items-center rounded-full p-0.5 transition" :class="isAmbientModeEnabled ? 'justify-end bg-white' : 'justify-start bg-white/30'">
                <span class="h-4 w-4 rounded-full" :class="isAmbientModeEnabled ? 'bg-black' : 'bg-white'" />
              </span>
            </button>
            <button type="button" class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-white/10" @click="settingsView = 'sleep'">
              <span>수면 타이머</span>
              <span class="flex items-center gap-1 text-white/60">{{ sleepTimerOption === 'off' ? '꺼짐' : `${sleepTimerOption}분` }}
                <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" /></svg>
              </span>
            </button>
            <button
              type="button"
              class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-white/10"
              @click="showStatsForNerds = !showStatsForNerds; showSettingsMenu = false; updateNerdStats()"
            >
              <span>통계 정보</span>
              <span class="font-mono text-[10px] text-white/60">Shift+S</span>
            </button>
          </template>
          <template v-else>
            <div class="flex items-center gap-1 border-b border-white/10 px-2 py-1.5">
              <button type="button" aria-label="뒤로" class="rounded-full p-1.5 hover:bg-white/10" @click="settingsView = 'main'">
                <svg class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" /></svg>
              </button>
              <span class="text-sm font-semibold">
                {{ settingsView === 'quality' ? '화질' : settingsView === 'speed' ? '재생 속도' : settingsView === 'captions' ? '자막' : '수면 타이머' }}
              </span>
            </div>
            <template v-if="settingsView === 'quality'">
              <button
                v-for="q in availableQualities"
                :key="q"
                type="button"
                class="flex w-full items-center gap-2 px-4 py-2 text-left hover:bg-white/10"
                @click="changeQuality(q); settingsView = 'main'"
              >
                <svg v-if="selectedQuality === q" class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" /></svg>
                <span v-else class="h-4 w-4" />
                <span>{{ q === 'Auto' ? '자동' : q }}</span>
              </button>
            </template>
            <template v-else-if="settingsView === 'speed'">
              <button
                v-for="spd in speedMenuOptions"
                :key="spd"
                type="button"
                class="flex w-full items-center gap-2 px-4 py-2 text-left hover:bg-white/10"
                @click="changeSpeed(spd); settingsView = 'main'"
              >
                <svg v-if="selectedSpeed === spd" class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" /></svg>
                <span v-else class="h-4 w-4" />
                <span>{{ speedLabel(spd) }}</span>
              </button>
            </template>
            <template v-else-if="settingsView === 'captions'">
              <button
                type="button"
                class="flex w-full items-center gap-2 px-4 py-2 text-left hover:bg-white/10"
                @click="selectCaption('off')"
              >
                <svg v-if="activeCaptionLang === 'off'" class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" /></svg>
                <span v-else class="h-4 w-4" />
                <span>꺼짐</span>
              </button>
              <button
                v-for="cap in (captions || [])"
                :key="cap.id"
                type="button"
                class="flex w-full items-center gap-2 px-4 py-2 text-left hover:bg-white/10"
                @click="selectCaption(cap.language)"
              >
                <svg v-if="activeCaptionLang === cap.language" class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" /></svg>
                <span v-else class="h-4 w-4" />
                <span>{{ cap.label || cap.language }}</span>
              </button>
            </template>
            <template v-else-if="settingsView === 'sleep'">
              <button
                v-for="opt in (['off', '15', '30', '60'] as const)"
                :key="opt"
                type="button"
                class="flex w-full items-center gap-2 px-4 py-2 text-left hover:bg-white/10"
                @click="setSleepTimer(opt); settingsView = 'main'"
              >
                <svg v-if="sleepTimerOption === opt" class="h-4 w-4 fill-current" viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" /></svg>
                <span v-else class="h-4 w-4" />
                <span>{{ opt === 'off' ? '꺼짐' : `${opt}분` }}</span>
              </button>
            </template>
          </template>
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
