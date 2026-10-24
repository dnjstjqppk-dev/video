<script setup lang="ts">
</script>

<template>
  <div class="dark min-h-screen bg-[#090B10] text-[#F8FAFC] antialiased">
    <NuxtRouteAnnouncer />
    <NuxtLoadingIndicator color="#EF4444" :height="3" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
