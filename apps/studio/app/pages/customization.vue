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
  <div class="max-w-5xl">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white">채널 맞춤설정 (Channel Customization)</h1>
        <p class="mt-1 text-xs text-neutral-400">
          채널 이름·소개·프로필·배너를 변경하면 시청자 페이지에 즉시 반영됩니다
        </p>
      </div>

      <button
        type="button"
        :disabled="saving || pending || !!error"
        class="rounded-xl bg-[#FF0033] px-5 py-2 text-xs font-bold text-white shadow transition hover:bg-[#D9002B] disabled:opacity-50"
        @click="saveCustomization"
      >
        {{ saving ? '저장 중…' : saveSuccess ? '✓ Saved!' : '게시 (Publish Changes)' }}
      </button>
    </div>

    <GvpBanner v-if="saveError" tone="danger" dismissible class="mt-4" @dismiss="saveError = null">
      {{ saveError }}
    </GvpBanner>

    <div v-if="pending" class="mt-6 space-y-6" role="status" aria-label="불러오는 중">
      <div class="h-24 animate-pulse rounded-2xl bg-white/5" />
      <div class="h-48 animate-pulse rounded-2xl bg-white/5" />
      <div class="h-64 animate-pulse rounded-2xl bg-white/5" />
    </div>

    <GvpBanner v-else-if="error" tone="danger" class="mt-6">
      채널 정보를 불러오지 못했습니다. <button type="button" class="font-bold underline" @click="refresh()">다시 시도</button>
    </GvpBanner>

    <div v-else class="mt-6 space-y-6">
      <!-- Channel Stats -->
      <div class="flex flex-wrap items-center gap-3 text-xs" role="status">
        <span class="rounded-full bg-white/5 px-3 py-1.5 text-neutral-300">핸들 <strong class="text-white">{{ channelHandle }}</strong></span>
        <span class="rounded-full bg-white/5 px-3 py-1.5 text-neutral-300">구독자 <strong class="text-white">{{ subscriberCount.toLocaleString() }}</strong></span>
        <span class="rounded-full bg-white/5 px-3 py-1.5 text-neutral-300">영상 <strong class="text-white">{{ videoCount.toLocaleString() }}</strong>개</span>
      </div>

      <!-- Branding -->
      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="text-sm font-bold text-white">사진 및 프로필 이미지 (Picture)</h2>
        <div class="mt-4 flex flex-wrap items-center gap-6">
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            alt="채널 프로필 미리보기"
            class="h-24 w-24 rounded-full border-2 border-white/20 object-cover"
          />
          <span
            v-else
            class="flex h-24 w-24 items-center justify-center rounded-full bg-white/5 text-2xl font-black text-neutral-500"
            aria-hidden="true"
          >
            ?
          </span>
          <div class="min-w-0 flex-1 text-xs text-neutral-300">
            <p>98 x 98 픽셀 이상, 4MB 이하의 PNG·JPEG 권장</p>
            <div class="mt-2 max-w-md">
              <GvpTextInput v-model="avatarUrl" input-id="ch-avatar" label="프로필 이미지 URL" placeholder="https://…" />
            </div>
          </div>
        </div>
      </div>

      <div class="rounded-2xl border border-white/10 bg-[#151822] p-5">
        <h2 class="text-sm font-bold text-white">배너 이미지 (Banner Image)</h2>
        <div class="mt-4">
          <div class="h-32 w-full overflow-hidden rounded-xl border border-white/15 bg-black">
            <img v-if="bannerUrl" :src="bannerUrl" alt="채널 배너 미리보기" class="h-full w-full object-cover" />
            <p v-else class="flex h-full items-center justify-center text-xs text-neutral-500">배너가 등록되지 않았습니다</p>
          </div>
          <div class="mt-3">
            <GvpTextInput v-model="bannerUrl" input-id="ch-banner" label="배너 이미지 URL (권장 2048 x 1152)" placeholder="https://…" />
          </div>
        </div>
      </div>

      <!-- Basic Info -->
      <div class="space-y-4 rounded-2xl border border-white/10 bg-[#151822] p-5 text-xs">
        <GvpTextInput v-model="channelName" input-id="ch-name" label="채널 이름 (Channel Name)" />
        <div>
          <label for="ch-desc" class="mb-1.5 block text-[13px] font-bold text-neutral-200">설명 (Description)</label>
          <textarea
            id="ch-desc"
            v-model="description"
            rows="4"
            class="w-full rounded-lg border border-white/15 bg-[#0D0F14] px-3 py-2 text-white outline-none focus:border-[#3EA6FF]"
          />
        </div>
      </div>
    </div>
  </div>
</template>
