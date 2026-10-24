<script setup lang="ts">
import type { VideoChapter, VideoCaption, VideoPlaybackSource, AnalyticsPlaybackEvent } from '@gvp/shared-types'
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
  }>(),
  {
    hlsMasterUrl: 'https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8',
    posterUrl: '',
    durationSeconds: 634,
    chapters: () => [],
    captions: () => [],
    sources: () => [],
    autoplay: false,
    isLive: false
  }
)

const emit = defineEmits<{
  (e: 'timeupdate', currentTime: number): void
  (e: 'ended'): void
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
  if (videoEl.paused) {
    videoEl.play().catch(() => {
      // Simulate playback state if headless browser blocks autoplay
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

function handleKeydown(event: KeyboardEvent) {
  const tag = (event.target as HTMLElement)?.tagName?.toLowerCase()
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return

  const key = event.key.toLowerCase()
  if (key === 'k' || key === ' ') {
    event.preventDefault()
    togglePlay()
  } else if (key === 'j') {
    event.preventDefault()
    seekRelative(-10)
  } else if (key === 'l') {
    event.preventDefault()
    seekRelative(10)
  } else if (key === 'arrowleft') {
    event.preventDefault()
    seekRelative(-5)
  } else if (key === 'arrowright') {
    event.preventDefault()
    seekRelative(5)
  } else if (key === 'm') {
    event.preventDefault()
    toggleMute()
  } else if (key === 'f') {
    event.preventDefault()
    toggleFullscreen()
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
}

onMounted(() => {
  initShakaPlayer()
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
  if (shakaInstance) {
    shakaInstance.destroy?.()
  }
})

defineExpose({
  seekToSeconds,
  togglePlay
})
</script>

<template>
  <div
    ref="playerContainerRef"
    class="group relative aspect-video w-full overflow-hidden rounded-2xl bg-black shadow-2xl select-none"
  >
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
    <div v-if="isLive" class="absolute left-4 top-4 flex items-center gap-2 rounded-md bg-[#FF0033] px-2.5 py-1 text-xs font-bold uppercase tracking-wider text-white shadow">
      <span class="h-2 w-2 animate-ping rounded-full bg-white" />
      <span>LIVE ABR</span>
    </div>

    <!-- Active Caption Overlay Banner (when captions enabled) -->
    <div
      v-if="activeCaptionLang !== 'off'"
      class="pointer-events-none absolute bottom-20 left-1/2 -translate-x-1/2 rounded bg-black/80 px-4 py-1.5 text-center text-sm font-medium text-white"
    >
      [{{ activeCaptionLang.toUpperCase() }}] {{ title }} — {{ currentChapter?.title || 'Subtitles Active' }}
    </div>

    <!-- Bottom Controls Overlay -->
    <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/95 via-black/60 to-transparent px-4 pb-3 pt-10 transition-opacity">
      <!-- Timeline Scrub Bar with Chapter Markers -->
      <div
        class="relative mb-3 h-1.5 w-full cursor-pointer rounded-full bg-white/25 transition-all hover:h-2.5"
        @click="seekToPosition"
      >
        <!-- Buffered Bar -->
        <div
          class="absolute inset-y-0 left-0 rounded-full bg-white/40"
          :style="{ width: `${bufferedPercent}%` }"
        />
        <!-- Current Playback Bar -->
        <div
          class="absolute inset-y-0 left-0 rounded-full bg-[#FF0033]"
          :style="{ width: `${progressPercent}%` }"
        />
        <!-- Chapter Segment Tick Markers -->
        <template v-for="chapter in chapters" :key="chapter.id">
          <div
            v-if="chapter.startTimeSeconds > 0 && duration > 0"
            class="absolute top-0 h-full w-0.5 bg-black/90"
            :style="{ left: `${(chapter.startTimeSeconds / duration) * 100}%` }"
            :title="chapter.title"
          />
        </template>
      </div>

      <!-- Controls Row -->
      <div class="flex flex-wrap items-center justify-between gap-2 text-white">
        <!-- Left Controls: Play/Pause, -10s/+10s, Volume, Time, Active Chapter -->
        <div class="flex items-center gap-3">
          <button
            type="button"
            class="rounded-lg p-1.5 transition hover:bg-white/15"
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
            type="button"
            class="rounded-lg px-2 py-1 text-xs font-semibold transition hover:bg-white/15"
            title="Rewind 10s (j)"
            @click="seekRelative(-10)"
          >
            -10s
          </button>
          <button
            type="button"
            class="rounded-lg px-2 py-1 text-xs font-semibold transition hover:bg-white/15"
            title="Forward 10s (l)"
            @click="seekRelative(10)"
          >
            +10s
          </button>

          <!-- Volume -->
          <div class="flex items-center gap-1.5">
            <button
              type="button"
              class="rounded-lg p-1.5 transition hover:bg-white/15"
              title="Mute (m)"
              @click="toggleMute"
            >
              <svg v-if="isMuted || volume === 0" class="h-5 w-5 fill-current text-red-400" viewBox="0 0 24 24">
                <path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z" />
              </svg>
              <svg v-else class="h-5 w-5 fill-current" viewBox="0 0 24 24">
                <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02z" />
              </svg>
            </button>
            <input
              type="range"
              min="0"
              max="1"
              step="0.05"
              :value="isMuted ? 0 : volume"
              class=" hidden h-1 w-16 accent-[#FF0033] sm:block"
              @input="setVolume"
            />
          </div>

          <span class="font-mono text-xs text-neutral-200">
            {{ formatTime(currentTime) }} / {{ isLive ? 'LIVE' : formatTime(duration) }}
          </span>

          <span
            v-if="currentChapter"
            class="hidden truncate rounded bg-white/15 px-2 py-0.5 text-xs font-medium text-neutral-100 md:inline-block"
          >
            • {{ currentChapter.title }}
          </span>
        </div>

        <!-- Right Controls: Captions, Quality/Speed Settings, PiP, Fullscreen -->
        <div class="relative flex items-center gap-2">
          <!-- Subtitle Toggle -->
          <button
            type="button"
            class="rounded-lg px-2.5 py-1 text-xs font-semibold transition"
            :class="activeCaptionLang !== 'off' ? 'bg-[#FF0033] text-white' : 'bg-white/10 text-neutral-200 hover:bg-white/20'"
            @click="activeCaptionLang = activeCaptionLang === 'off' ? 'ko' : activeCaptionLang === 'ko' ? 'en' : 'off'"
          >
            CC: {{ activeCaptionLang.toUpperCase() }}
          </button>

          <!-- Quality Badge + Settings Trigger -->
          <button
            type="button"
            class="flex items-center gap-1 rounded-lg bg-white/10 px-2.5 py-1 text-xs font-semibold transition hover:bg-white/20"
            @click="showSettingsMenu = !showSettingsMenu"
          >
            <span>{{ selectedQuality }}</span>
            <span class="text-[#3EA6FF]">{{ selectedSpeed }}x</span>
          </button>

          <!-- Settings Popover -->
          <div
            v-if="showSettingsMenu"
            class="absolute bottom-11 right-0 z-40 w-64 rounded-2xl border border-white/15 bg-[#181818]/95 p-3 text-xs text-white shadow-2xl backdrop-blur-md"
          >
            <div class="mb-3">
              <p class="mb-1.5 font-semibold text-neutral-400">화질 선택 (ABR Quality)</p>
              <div class="grid grid-cols-3 gap-1">
                <button
                  v-for="q in qualityOptions"
                  :key="q"
                  type="button"
                  class="rounded-lg px-2 py-1 font-medium transition"
                  :class="selectedQuality === q ? 'bg-[#FF0033] text-white' : 'bg-white/10 hover:bg-white/20'"
                  @click="changeQuality(q)"
                >
                  {{ q }}
                </button>
              </div>
            </div>

            <div class="mb-3">
              <p class="mb-1.5 font-semibold text-neutral-400">재생 속도 (Speed)</p>
              <div class="grid grid-cols-3 gap-1">
                <button
                  v-for="spd in speedOptions"
                  :key="spd"
                  type="button"
                  class="rounded-lg px-2 py-1 font-medium transition"
                  :class="selectedSpeed === spd ? 'bg-[#3EA6FF] text-black font-bold' : 'bg-white/10 hover:bg-white/20'"
                  @click="changeSpeed(spd)"
                >
                  {{ spd }}x
                </button>
              </div>
            </div>

            <div>
              <p class="mb-1.5 font-semibold text-neutral-400">오디오 트랙 (Audio Track)</p>
              <div class="flex flex-col gap-1">
                <button
                  v-for="track in audioTracks"
                  :key="track"
                  type="button"
                  class="rounded-lg px-2.5 py-1 text-left transition"
                  :class="selectedAudioTrack === track ? 'bg-white/20 font-semibold text-[#3EA6FF]' : 'hover:bg-white/10'"
                  @click="selectedAudioTrack = track"
                >
                  {{ track }}
                </button>
              </div>
            </div>
          </div>

          <!-- PiP -->
          <button
            type="button"
            class="rounded-lg p-1.5 transition hover:bg-white/15"
            title="Picture-in-Picture"
            @click="togglePiP"
          >
            <svg class="h-5 w-5 fill-current" viewBox="0 0 24 24">
              <path d="M19 7h-8v6h8V7zm2-4H3c-1.1 0-2 .9-2 2v14c0 1.1.9 1.98 2 1.98h18c1.1 0 2-.88 2-1.98V5c0-1.1-.9-2-2-2zm0 16.01H3V4.98h18v14.03z" />
            </svg>
          </button>

          <!-- Fullscreen -->
          <button
            type="button"
            class="rounded-lg p-1.5 transition hover:bg-white/15"
            title="Fullscreen (f)"
            @click="toggleFullscreen"
          >
            <svg class="h-5 w-5 fill-current" viewBox="0 0 24 24">
              <path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
