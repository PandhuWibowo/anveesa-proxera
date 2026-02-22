<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { mockServers, mockAlerts, dashboardStats, generateTrafficData } from '../composables/useMockData'

const stats = dashboardStats
const recentAlerts = computed(() => mockAlerts.filter(a => a.status === 'active').slice(0, 4))
const trafficData = ref(generateTrafficData(12))

// Simulate live updates
let interval: number
onMounted(() => {
  interval = setInterval(() => {
    trafficData.value = generateTrafficData(12)
  }, 5000)
})
onUnmounted(() => clearInterval(interval))

const serversByType = computed(() => {
  const map: Record<string, number> = {}
  mockServers.forEach(s => { map[s.proxyType] = (map[s.proxyType] || 0) + 1 })
  return Object.entries(map)
})

function fmtNum(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}

function timeAgo(iso: string) {
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  return `${Math.floor(m / 60)}h ago`
}

function severityClass(s: string) {
  return s === 'critical' ? 'badge-offline' : s === 'warning' ? 'badge-warning' : 'badge-info'
}

const maxBar = computed(() => Math.max(...trafficData.value.map(d => d.requests), 1))
</script>

<template>
  <div class="page animate-slide">
    <!-- Stats Row -->
    <div class="grid-4">
      <div class="metric-card">
        <div class="metric-label">Total Servers</div>
        <div class="metric-value">{{ stats.totalServers }}</div>
        <div class="metric-sub">
          <span class="badge badge-online" style="font-size:11px;padding:1px 7px;">{{ stats.onlineServers }} online</span>
          <span class="badge badge-offline" style="font-size:11px;padding:1px 7px;margin-left:4px;">{{ stats.offlineServers }} offline</span>
        </div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Requests Today</div>
        <div class="metric-value">{{ fmtNum(stats.totalRequestsToday) }}</div>
        <div class="metric-sub metric-trend up">↑ 12.4% vs yesterday</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Avg Latency</div>
        <div class="metric-value">{{ stats.avgLatency }}<span style="font-size:16px;font-weight:500;">ms</span></div>
        <div class="metric-sub metric-trend up">↑ Stable (p95: 112ms)</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Active Alerts</div>
        <div class="metric-value" :style="stats.activeAlerts > 0 ? 'color:var(--color-error)' : ''">{{ stats.activeAlerts }}</div>
        <div class="metric-sub" style="color:var(--color-error)">{{ stats.activeAlerts }} require attention</div>
      </div>
    </div>

    <div style="display:grid;grid-template-columns:2fr 1fr;gap:16px;">
      <!-- Traffic Chart -->
      <div class="card">
        <div class="card-header">
          <div class="card-title">Request Volume (12h)</div>
          <div class="flex gap-2 items-center">
            <span class="badge badge-online" style="font-size:10.5px;">Live</span>
            <div class="tab-list">
              <div class="tab-item active">12h</div>
              <div class="tab-item">24h</div>
              <div class="tab-item">7d</div>
            </div>
          </div>
        </div>
        <div class="card-body" style="padding-top:12px;">
          <!-- Bar chart -->
          <div style="display:flex;align-items:flex-end;gap:4px;height:140px;">
            <div v-for="(d, i) in trafficData" :key="i"
                 style="flex:1;display:flex;flex-direction:column;align-items:center;gap:4px;">
              <div style="width:100%;display:flex;flex-direction:column;gap:2px;flex:1;justify-content:flex-end;">
                <div :style="{
                  width: '100%',
                  height: Math.max(2, (d.requests / maxBar) * 120) + 'px',
                  background: 'var(--color-primary)',
                  borderRadius: '3px 3px 0 0',
                  opacity: 0.85,
                  transition: 'height 0.5s ease',
                }" />
              </div>
              <div style="font-size:9px;color:var(--color-text-3);white-space:nowrap;">{{ d.time }}</div>
            </div>
          </div>
          <div class="flex gap-4 items-center" style="margin-top:12px;padding-top:12px;border-top:1px solid var(--color-border);">
            <div class="flex gap-2 items-center text-sm text-muted">
              <div style="width:10px;height:10px;background:var(--color-primary);border-radius:2px;opacity:.85;"></div>
              Requests
            </div>
            <div class="flex gap-2 items-center text-sm text-muted">
              <div style="width:10px;height:10px;background:var(--color-error);border-radius:2px;opacity:.85;"></div>
              Errors
            </div>
          </div>
        </div>
      </div>

      <!-- Server Status -->
      <div class="card">
        <div class="card-header">
          <div class="card-title">Server Status</div>
        </div>
        <div style="padding:8px 0;">
          <div v-for="server in mockServers" :key="server.id"
               style="display:flex;align-items:center;gap:10px;padding:9px 16px;transition:background var(--t);"
               onmouseover="this.style.background='var(--color-surface-2)'"
               onmouseout="this.style.background=''"
          >
            <div :class="['proxy-badge', `proxy-${server.proxyType}`]" style="font-size:10px;min-width:56px;justify-content:center;">
              {{ server.proxyType.toUpperCase() }}
            </div>
            <div style="flex:1;min-width:0;">
              <div style="font-size:12.5px;font-weight:500;color:var(--color-text);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">
                {{ server.name }}
              </div>
              <div style="font-size:11px;color:var(--color-text-3);">{{ server.requestsPerSec.toLocaleString() }} req/s</div>
            </div>
            <div :class="['badge', server.status === 'online' ? 'badge-online' : server.status === 'warning' ? 'badge-warning' : 'badge-offline']" style="font-size:10.5px;">
              {{ server.status }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;">
      <!-- Active Alerts -->
      <div class="card">
        <div class="card-header">
          <div class="card-title">Active Alerts</div>
          <RouterLink to="/alerts" class="btn btn-ghost btn-sm">View all</RouterLink>
        </div>
        <div style="padding:4px 0;">
          <div v-for="alert in recentAlerts" :key="alert.id"
               style="display:flex;align-items:flex-start;gap:12px;padding:12px 16px;border-bottom:1px solid var(--color-border);">
            <div :class="['badge', severityClass(alert.severity)]" style="flex-shrink:0;margin-top:1px;">
              {{ alert.severity }}
            </div>
            <div style="flex:1;min-width:0;">
              <div style="font-size:12.5px;font-weight:500;color:var(--color-text);line-height:1.3;">{{ alert.title }}</div>
              <div style="font-size:11.5px;color:var(--color-text-3);margin-top:2px;">{{ alert.serverName || 'System' }} · {{ timeAgo(alert.timestamp) }}</div>
            </div>
          </div>
          <div v-if="recentAlerts.length === 0" class="empty-state" style="padding:24px;">
            <div class="empty-state-title">No active alerts</div>
          </div>
        </div>
      </div>

      <!-- Proxy Distribution -->
      <div class="card">
        <div class="card-header">
          <div class="card-title">Proxy Distribution</div>
        </div>
        <div class="card-body" style="display:flex;flex-direction:column;gap:12px;">
          <div v-for="[type, count] in serversByType" :key="type"
               style="display:flex;align-items:center;gap:12px;">
            <div :class="['proxy-badge', `proxy-${type}`]" style="min-width:70px;justify-content:center;">
              {{ type.toUpperCase() }}
            </div>
            <div style="flex:1;">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: (count / stats.totalServers * 100) + '%' }" />
              </div>
            </div>
            <div style="font-size:12px;color:var(--color-text-2);min-width:18px;text-align:right;">{{ count }}</div>
          </div>

          <div style="border-top:1px solid var(--color-border);padding-top:12px;margin-top:4px;">
            <div style="font-size:12px;color:var(--color-text-2);">Total routes managed</div>
            <div style="font-size:20px;font-weight:700;color:var(--color-text);margin-top:2px;">{{ stats.totalRoutes }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
