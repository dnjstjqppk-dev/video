<script setup lang="ts">
import GvpBanner from '@gvp/ui/src/components/GvpBanner.vue'
import GvpTextInput from '@gvp/ui/src/components/GvpTextInput.vue'
import { useStudioApi } from '~/composables/useStudioApi'

useSeoMeta({
  title: 'Channel Customization — Branding & Basic Info'
})

const api = useStudioApi()

const channelName = ref('')
const channelHandle = ref('')
const description = ref('')
const avatarUrl = ref('')
const bannerUrl = ref('')
const subscriberCount = ref(0)
const videoCount = ref(0)
const channelId = ref('')

const saving = ref(false)
const saveSuccess = ref(false)
const saveError = ref<string | null>(null)

const { data, pending, error, refresh } = await useAsyncData('studio-channel', async () => {
  const dashboard = await api.getStudioDashboard()
  return dashboard.channel
})

watch(data, (ch) => {
  if (!ch) return
  channelId.value = ch.id
  channelName.value = ch.name
  channelHandle.value = ch.handle
  description.value = ch.description || ''
  avatarUrl.value = ch.avatarUrl || ''
  bannerUrl.value = ch.bannerUrl || ''
  subscriberCount.value = ch.subscriberCount
  videoCount.value = ch.videoCount
}, { immediate: true })

async function saveCustomization() {
  if (!channelId.value || saving.value) return
  saving.value = true
  saveSuccess.value = false
  saveError.value = null
  try {
    const updated = await api.updateChannel(channelId.value, {
      name: channelName.value.trim(),
      description: description.value.trim(),
      avatarUrl: avatarUrl.value.trim() || undefined,
      bannerUrl: bannerUrl.value.trim() || undefined,
    })
    channelName.value = updated.name
    description.value = updated.description || ''
    avatarUrl.value = updated.avatarUrl || ''
    bannerUrl.value = updated.bannerUrl || ''
    saveSuccess.value = true
    setTimeout(() => { saveSuccess.value = false }, 2500)
  } catch (e: unknown) {
    saveError.value = e instanceof Error ? e.message : '저장에 실패했습니다.'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="max-w-5xl space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-5">
      <div>
        <h1 class="text-xl font-semibold tracking-tight text-slate-900">채널 맞춤설정 (Channel Customization)</h1>
        <p class="mt-1 text-xs text-slate-500">
          채널 이름·소개·프로필·배너를 변경하면 시청자 페이지에 즉시 반영됩니다
        </p>
      </div>

      <button
        type="button"
        :disabled="saving || pending || !!error"
        class="inline-flex items-center justify-center h-9 px-4 text-sm font-medium text-white bg-slate-900 rounded-md hover:bg-slate-800 transition-colors disabled:opacity-50"
        @click="saveCustomization"
      >
        {{ saving ? '저장 중…' : saveSuccess ? 'Saved' : '게시 (Publish Changes)' }}
      </button>
    </div>

    <GvpBanner v-if="saveError" tone="danger" dismissible @dismiss="saveError = null">
      {{ saveError }}
    </GvpBanner>

    <div v-if="pending" class="space-y-6" role="status" aria-label="불러오는 중">
      <div class="h-20 animate-pulse rounded-lg bg-slate-100" />
      <div class="h-44 animate-pulse rounded-lg bg-slate-100" />
      <div class="h-56 animate-pulse rounded-lg bg-slate-100" />
    </div>

    <GvpBanner v-else-if="error" tone="danger">
      채널 정보를 불러오지 못했습니다. <button type="button" class="font-semibold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <div v-else class="space-y-6">
      <!-- Channel Stats -->
      <div class="flex flex-wrap items-center gap-3 text-xs tabular-nums" role="status">
        <span class="rounded-md border border-slate-200 bg-slate-100 px-3 py-1.5 text-slate-600">
          핸들 <strong class="font-semibold text-slate-900">{{ channelHandle }}</strong>
        </span>
        <span class="rounded-md border border-slate-200 bg-slate-100 px-3 py-1.5 text-slate-600">
          구독자 <strong class="font-semibold text-slate-900">{{ subscriberCount.toLocaleString() }}</strong>
        </span>
        <span class="rounded-md border border-slate-200 bg-slate-100 px-3 py-1.5 text-slate-600">
          영상 <strong class="font-semibold text-slate-900">{{ videoCount.toLocaleString() }}</strong>개
        </span>
      </div>

      <!-- Branding -->
      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <h2 class="text-sm font-semibold text-slate-900">사진 및 프로필 이미지 (Picture)</h2>
        <div class="mt-4 flex flex-wrap items-center gap-6">
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            alt="채널 프로필 미리보기"
            class="h-24 w-24 rounded-full border border-slate-200 object-cover"
          />
          <span
            v-else
            class="flex h-24 w-24 items-center justify-center rounded-full border border-slate-200 bg-slate-100 text-2xl font-semibold text-slate-500"
            aria-hidden="true"
          >
            ?
          </span>
          <div class="min-w-0 flex-1 text-xs text-slate-600">
            <p class="tabular-nums">98 x 98 픽셀 이상, 4MB 이하의 PNG·JPEG 권장</p>
            <div class="mt-2 max-w-md">
              <GvpTextInput v-model="avatarUrl" input-id="ch-avatar" label="프로필 이미지 URL" placeholder="https://…" />
            </div>
          </div>
        </div>
      </div>

      <div class="rounded-lg border border-slate-200 bg-white p-5">
        <h2 class="text-sm font-semibold text-slate-900">배너 이미지 (Banner Image)</h2>
        <div class="mt-4">
          <div class="h-32 w-full overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
            <img v-if="bannerUrl" :src="bannerUrl" alt="채널 배너 미리보기" class="h-full w-full object-cover" />
            <p v-else class="flex h-full items-center justify-center text-xs text-slate-400">배너가 등록되지 않았습니다</p>
          </div>
          <div class="mt-3">
            <GvpTextInput v-model="bannerUrl" input-id="ch-banner" label="배너 이미지 URL (권장 2048 x 1152)" placeholder="https://…" />
          </div>
        </div>
      </div>

      <!-- Basic Info -->
      <div class="space-y-4 rounded-lg border border-slate-200 bg-white p-5 text-xs">
        <GvpTextInput v-model="channelName" input-id="ch-name" label="채널 이름 (Channel Name)" />
        <div>
          <label for="ch-desc" class="mb-1.5 block text-xs font-medium text-slate-700">설명 (Description)</label>
          <textarea
            id="ch-desc"
            v-model="description"
            rows="4"
            class="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 outline-none focus:border-slate-900"
          />
        </div>
      </div>
    </div>
  </div>
</template>
