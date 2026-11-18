<script setup lang="ts">
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
