<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { mockLogs } from '../composables/useMockData'
import type { LogEntry } from '../types'

const search = ref('')
const levelFilter = ref<'all'|'info'|'warn'|'error'>('all')
const serverFilter = ref('all')
const streaming = ref(true)
const logs = ref<LogEntry[]>([...mockLogs])

const servers = ['all', ...new Set(mockLogs.map(l => l.serverName))]

const filtered = computed(() => {
  return logs.value.filter(l => {
    const q = search.value.toLowerCase()
    const matchSearch = !q || l.message.toLowerCase().includes(q) || l.serverName.toLowerCase().includes(q)
    const matchLevel  = levelFilter.value === 'all' || l.level === levelFilter.value
    const matchServer = serverFilter.value === 'all' || l.serverName === serverFilter.value
    return matchSearch && matchLevel && matchServer
  })
})

const logMessages = [
  { level: 'info' as const, msg: 'GET /api/v1/health 200 1ms' },
  { level: 'info' as const, msg: 'POST /api/v1/auth/token 200 32ms' },
  { level: 'warn' as const, msg: 'slow upstream response 850ms' },
  { level: 'error' as const, msg: 'upstream 10.0.2.3 connection reset' },
  { level: 'info' as const, msg: 'TLS session resumed for client 203.x.x.x' },
  { level: 'info' as const, msg: 'GET /api/v1/users/42 200 15ms' },
]

let iv: number
let logId = logs.value.length
onMounted(() => {
  if (streaming.value) {
    iv = setInterval(() => {
      if (!streaming.value) return
      const m = logMessages[Math.floor(Math.random() * logMessages.length)]!
      const serverIdx = Math.floor(Math.random() * 4)
      const serverNames = ['prod-nginx-01','prod-traefik-01','staging-caddy-01','prod-haproxy-lb'] as const
      const proxyTypes  = ['nginx','traefik','caddy','haproxy'] as const
      logs.value.unshift({
        id: `live-${++logId}`,
        serverId: `s${serverIdx+1}`,
        serverName: serverNames[serverIdx]!,
        proxyType: proxyTypes[serverIdx]!,
        timestamp: new Date().toISOString(),
        level: m.level,
        message: m.msg,
      })
      if (logs.value.length > 200) logs.value.pop()
    }, 1500)
  }
})
onUnmounted(() => clearInterval(iv))

function fmtTime(iso: string) {
  const d = new Date(iso)
  return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}:${String(d.getSeconds()).padStart(2,'0')}.${String(d.getMilliseconds()).padStart(3,'0')}`
}

function levelClass(level: string) {
  return { info: 'log-info', warn: 'log-warn', error: 'log-error', debug: 'log-info' }[level] || ''
}
function levelTextClass(level: string) {
  return { info: 'info', warn: 'warn', error: 'error', debug: 'info' }[level] || 'info'
}
const counts = computed(() => ({
  all: logs.value.length,
  info: logs.value.filter(l=>l.level==='info').length,
  warn: logs.value.filter(l=>l.level==='warn').length,
  error: logs.value.filter(l=>l.level==='error').length,
}))
</script>

<template>
  <div class="page animate-slide">
    <div class="page-header">
      <div>
        <div class="page-title">Log Explorer</div>
        <div class="page-description">Aggregated real-time logs from all connected proxy servers</div>
      </div>
      <div class="flex gap-2 items-center">
        <button class="btn" :class="streaming ? 'btn-primary' : 'btn-secondary'" @click="streaming = !streaming">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle v-if="streaming" cx="12" cy="12" r="10"/><rect v-if="streaming" x="9" y="9" width="6" height="6"/>
            <polygon v-else points="5 3 19 12 5 21 5 3"/>
          </svg>
          {{ streaming ? 'Pause' : 'Resume' }} Stream
        </button>
        <button class="btn btn-secondary">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          Export
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="card" style="padding:12px 16px;">
      <div class="flex gap-3 items-center flex-wrap">
        <div style="position:relative;flex:1;min-width:200px;max-width:340px;">
          <svg style="position:absolute;left:9px;top:50%;transform:translateY(-50%);width:13px;height:13px;color:var(--color-text-3);"
               viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
          </svg>
          <input v-model="search" class="input" placeholder="Filter logs…" style="padding-left:30px;height:32px;font-size:12.5px;" />
        </div>
        <div class="tab-list">
          <div class="tab-item" :class="{ active: levelFilter==='all' }" @click="levelFilter='all'">All ({{ counts.all }})</div>
          <div class="tab-item" :class="{ active: levelFilter==='info' }" @click="levelFilter='info'" style="color:var(--color-info)">Info</div>
          <div class="tab-item" :class="{ active: levelFilter==='warn' }" @click="levelFilter='warn'" style="color:var(--color-warning)">Warn ({{ counts.warn }})</div>
          <div class="tab-item" :class="{ active: levelFilter==='error' }" @click="levelFilter='error'" style="color:var(--color-error)">Error ({{ counts.error }})</div>
        </div>
        <select class="input" v-model="serverFilter" style="height:32px;font-size:12.5px;width:auto;min-width:160px;">
          <option v-for="s in servers" :key="s" :value="s">{{ s === 'all' ? 'All servers' : s }}</option>
        </select>
        <button class="btn btn-ghost btn-sm" @click="logs = [...mockLogs]">Clear live</button>
      </div>
    </div>

    <!-- Log output -->
    <div class="card" style="overflow:hidden;">
      <div style="background:var(--color-surface-2);padding:8px 14px;border-bottom:1px solid var(--color-border);display:flex;align-items:center;gap:12px;">
        <div style="font-size:11px;font-weight:600;color:var(--color-text-3);text-transform:uppercase;letter-spacing:.06em;">Log Output</div>
        <div style="margin-left:auto;font-size:11px;color:var(--color-text-3);">{{ filtered.length }} entries</div>
        <span v-if="streaming" class="badge badge-online" style="font-size:10px;">
          <div class="badge-dot" style="background:var(--color-success);animation:pulse 1s infinite;"></div>
          Streaming
        </span>
      </div>
      <div style="height:520px;overflow-y:auto;background:var(--color-surface);">
        <div v-for="log in filtered" :key="log.id" :class="['log-line', levelClass(log.level)]">
          <span class="log-time">{{ fmtTime(log.timestamp) }}</span>
          <span :class="['proxy-badge', `proxy-${log.proxyType}`]" style="font-size:9.5px;padding:1px 5px;flex-shrink:0;align-self:flex-start;margin-top:1px;">
            {{ log.proxyType.toUpperCase() }}
          </span>
          <span class="log-level" :class="levelTextClass(log.level)">{{ log.level.toUpperCase() }}</span>
          <span class="log-msg">{{ log.serverName }} — {{ log.message }}</span>
        </div>
        <div v-if="filtered.length === 0" class="empty-state" style="padding:32px;">
          <div class="empty-state-title">No matching log entries</div>
          <div class="empty-state-desc">Try adjusting your filters</div>
        </div>
      </div>
    </div>
  </div>
</template>
