<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen bg-white text-slate-900 antialiased selection:bg-slate-900 selection:text-white">
    <NuxtRouteAnnouncer />
    <NuxtLoadingIndicator color="#0F172A" :height="2" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
