<script setup lang="ts">
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
</script>

<template>
  <div :class="[appStore.theme === 'dark' ? 'dark bg-[#0F0F0F] text-[#F1F1F1]' : 'bg-[#FFFFFF] text-[#0F0F0F]', 'min-h-screen antialiased selection:bg-[#FF0033] selection:text-white']">
    <NuxtRouteAnnouncer />
    <NuxtLoadingIndicator color="#FF0033" :height="3" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
