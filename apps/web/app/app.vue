<script setup lang="ts">
import { useAppStore } from '~/stores/app'

const appStore = useAppStore()
</script>

<template>
  <div class="min-h-screen bg-[var(--gvp-surface)] text-[var(--gvp-text)] antialiased selection:bg-[var(--gvp-brand)] selection:text-white">
    <NuxtRouteAnnouncer />
    <NuxtLoadingIndicator color="#0F172A" :height="2" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
