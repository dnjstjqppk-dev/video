<script setup lang="ts">
</script>

<template>
  <div class="dark min-h-screen bg-[#0D0F14] text-[#F3F4F6] antialiased">
    <NuxtRouteAnnouncer />
    <NuxtLoadingIndicator color="#FF0033" :height="3" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
