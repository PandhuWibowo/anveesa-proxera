<script setup lang="ts">
import { useRoute } from 'vue-router'
import AppSidebar from './components/layout/AppSidebar.vue'
import AppHeader from './components/layout/AppHeader.vue'

const route = useRoute()
</script>

<template>
  <div class="app-shell">
    <AppSidebar />
    <div class="app-main">
      <AppHeader :title="String(route.meta.title || '')" />
      <main class="app-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
