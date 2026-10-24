<script setup lang="ts">
</script>

<template>
  <div class="dark min-h-screen bg-[#0B0E17] text-[#F1F5F9] antialiased">
    <NuxtRouteAnnouncer />
    <NuxtLoadingIndicator color="#3EA6FF" :height="3" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
