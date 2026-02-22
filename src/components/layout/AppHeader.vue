<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useTheme } from '../../composables/useTheme'

const route = useRoute()
const { theme, toggleTheme } = useTheme()

const pageTitles: Record<string, { title: string; desc: string }> = {
  dashboard:  { title: 'Dashboard',     desc: 'Overview of all proxy servers and system health' },
  servers:    { title: 'Servers',       desc: 'Manage and configure proxy server connections' },
  monitoring: { title: 'Monitoring',    desc: 'Real-time health and performance metrics' },
  logs:       { title: 'Log Explorer',  desc: 'Aggregated logs from all connected proxies' },
  config:     { title: 'Configuration', desc: 'View and edit proxy configurations' },
  routes:     { title: 'Route Builder', desc: 'Manage routing rules, load balancing, and SSL' },
  analytics:  { title: 'Analytics',     desc: 'Traffic insights and performance trends' },
  alerts:     { title: 'Alerts',        desc: 'System alerts and notification history' },
  settings:   { title: 'Settings',      desc: 'User management and platform settings' },
  docs:       { title: 'Documentation', desc: 'API reference, guides, and deployment instructions' },
}

const current = () => {
  const name = String(route.name || '')
  return pageTitles[name] || { title: name, desc: '' }
}
</script>

<template>
  <header class="app-header">
    <div style="flex:1;min-width:0;">
      <div style="font-size:14px;font-weight:600;color:var(--color-text);letter-spacing:-0.01em;line-height:1.2;">
        {{ current().title }}
      </div>
      <div v-if="current().desc" style="font-size:11.5px;color:var(--color-text-3);margin-top:1px;">
        {{ current().desc }}
      </div>
    </div>

    <!-- Search -->
    <div style="position:relative;width:220px;flex-shrink:0;">
      <svg style="position:absolute;left:9px;top:50%;transform:translateY(-50%);width:13px;height:13px;color:var(--color-text-3);"
           viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
      </svg>
      <input class="input" placeholder="Search servers, routes…" style="padding-left:30px;height:32px;font-size:12.5px;" />
    </div>

    <!-- Theme toggle -->
    <button class="btn btn-ghost btn-icon" @click="toggleTheme" :data-tooltip="theme === 'dark' ? 'Light mode' : 'Dark mode'">
      <svg v-if="theme === 'dark'" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/>
        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
        <line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/>
        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
      </svg>
      <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/>
      </svg>
    </button>

    <!-- Notifications -->
    <button class="btn btn-ghost btn-icon" style="position:relative;" data-tooltip="Alerts">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/>
        <path d="M13.73 21a2 2 0 01-3.46 0"/>
      </svg>
      <span style="position:absolute;top:4px;right:4px;width:7px;height:7px;background:var(--color-error);border-radius:50%;border:1.5px solid var(--color-surface);"></span>
    </button>
  </header>
</template>
