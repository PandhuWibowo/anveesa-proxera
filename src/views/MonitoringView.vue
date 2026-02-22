<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { mockServers } from '../composables/useMockData'

const selected = ref(mockServers[0]!.id)
const server = computed(() => mockServers.find(s => s.id === selected.value) ?? mockServers[0]!)

// Simulated live metrics history
const history = ref<{ time: string; cpu: number; mem: number; rps: number; latency: number }[]>([])

function addPoint() {
  const s = server.value
  const now = new Date()
  const time = `${String(now.getHours()).padStart(2,'0')}:${String(now.getMinutes()).padStart(2,'0')}:${String(now.getSeconds()).padStart(2,'0')}`
  const vary = (base: number, pct: number) => Math.max(0, Math.min(100, base + (Math.random()-0.5)*2*pct))
  history.value.push({
    time,
    cpu: vary(s.cpuUsage, 8),
    mem: vary(s.memUsage, 3),
    rps: Math.max(0, s.requestsPerSec + (Math.random()-0.5)*200),
    latency: Math.max(5, 38 + (Math.random()-0.5)*30),
  })
  if (history.value.length > 30) history.value.shift()
}

let iv: number
onMounted(() => {
  for (let i = 0; i < 20; i++) addPoint()
  iv = setInterval(addPoint, 2000)
})
onUnmounted(() => clearInterval(iv))

const latest = computed(() => history.value[history.value.length - 1] || { cpu: 0, mem: 0, rps: 0, latency: 0 })

function sparkPath(data: number[], maxVal: number, w = 200, h = 60) {
  if (data.length < 2) return ''
  const pts = data.map((v, i) => {
    const x = (i / (data.length - 1)) * w
    const y = h - (v / maxVal) * (h - 4) - 2
    return `${x},${y}`
  })
  return 'M' + pts.join('L')
}

const cpuMax  = computed(() => Math.max(...history.value.map(h => h.cpu), 20))
const memMax  = computed(() => Math.max(...history.value.map(h => h.mem), 20))
const rpsMax  = computed(() => Math.max(...history.value.map(h => h.rps), 100))
const latMax  = computed(() => Math.max(...history.value.map(h => h.latency), 50))
</script>

<template>
  <div class="page animate-slide">
    <div class="page-header">
      <div>
        <div class="page-title">Real-time Monitoring</div>
        <div class="page-description">Live health and performance metrics — updates every 2s</div>
      </div>
      <div class="flex gap-2 items-center">
        <span class="badge badge-online" style="font-size:11px;">
          <div class="badge-dot" style="background:var(--color-success);animation:pulse 1s infinite;"></div>
          Live
        </span>
      </div>
    </div>

    <!-- Server Selector -->
    <div class="flex gap-2" style="overflow-x:auto;padding-bottom:2px;">
      <button v-for="s in mockServers" :key="s.id"
              class="btn"
              :class="selected === s.id ? 'btn-primary' : 'btn-secondary'"
              @click="selected = s.id; history = []"
              style="flex-shrink:0;">
        <div :class="['proxy-badge', `proxy-${s.proxyType}`]" style="font-size:10px;padding:1px 6px;">
          {{ s.proxyType.toUpperCase() }}
        </div>
        {{ s.name }}
        <div :style="{
          width:'7px',height:'7px',borderRadius:'50%',flexShrink:0,
          background: s.status==='online'?'var(--color-success)':s.status==='warning'?'var(--color-warning)':'var(--color-error)'
        }"></div>
      </button>
    </div>

    <!-- Current Server Info Banner -->
    <div class="card">
      <div style="padding:14px 18px;display:flex;align-items:center;gap:20px;flex-wrap:wrap;">
        <div>
          <div style="font-size:12.5px;font-weight:600;color:var(--color-text);">{{ server.name }}</div>
          <div style="font-size:11px;color:var(--color-text-3);">{{ server.host }}:{{ server.port }} · {{ server.location }}</div>
        </div>
        <div style="display:flex;gap:20px;flex-wrap:wrap;">
          <div>
            <div style="font-size:10px;text-transform:uppercase;letter-spacing:.05em;color:var(--color-text-3);">Version</div>
            <div style="font-size:12.5px;font-weight:500;color:var(--color-text);font-family:var(--font-mono);">{{ server.version }}</div>
          </div>
          <div>
            <div style="font-size:10px;text-transform:uppercase;letter-spacing:.05em;color:var(--color-text-3);">Uptime</div>
            <div style="font-size:12.5px;font-weight:500;color:var(--color-text);">{{ server.uptime }}</div>
          </div>
          <div>
            <div style="font-size:10px;text-transform:uppercase;letter-spacing:.05em;color:var(--color-text-3);">Status</div>
            <div :class="['badge', server.status==='online'?'badge-online':server.status==='warning'?'badge-warning':'badge-offline']" style="margin-top:2px;">
              {{ server.status }}
            </div>
          </div>
          <div>
            <div style="font-size:10px;text-transform:uppercase;letter-spacing:.05em;color:var(--color-text-3);">Connection</div>
            <div style="font-size:12.5px;font-weight:500;color:var(--color-text);text-transform:uppercase;">{{ server.connectionType }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Live Metric Cards -->
    <div class="grid-4">
      <div class="metric-card">
        <div class="metric-label">CPU Usage</div>
        <div class="metric-value" :style="{ color: latest.cpu>80?'var(--color-error)':latest.cpu>60?'var(--color-warning)':'var(--color-text)' }">
          {{ latest.cpu.toFixed(1) }}<span style="font-size:15px;font-weight:500;">%</span>
        </div>
        <div style="margin-top:6px;">
          <svg width="100%" height="44" viewBox="0 0 200 44" preserveAspectRatio="none" style="display:block;">
            <defs>
              <linearGradient id="cpuGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="var(--color-primary)" stop-opacity="0.3"/>
                <stop offset="100%" stop-color="var(--color-primary)" stop-opacity="0"/>
              </linearGradient>
            </defs>
            <path :d="sparkPath(history.map(h=>h.cpu), cpuMax) + `L200,44 L0,44Z`" fill="url(#cpuGrad)"/>
            <path :d="sparkPath(history.map(h=>h.cpu), cpuMax)" fill="none" stroke="var(--color-primary)" stroke-width="1.5"/>
          </svg>
        </div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Memory Usage</div>
        <div class="metric-value" :style="{ color: latest.mem>85?'var(--color-error)':latest.mem>70?'var(--color-warning)':'var(--color-text)' }">
          {{ latest.mem.toFixed(1) }}<span style="font-size:15px;font-weight:500;">%</span>
        </div>
        <div style="margin-top:6px;">
          <svg width="100%" height="44" viewBox="0 0 200 44" preserveAspectRatio="none" style="display:block;">
            <path :d="sparkPath(history.map(h=>h.mem), memMax)" fill="none" stroke="var(--color-info)" stroke-width="1.5"/>
          </svg>
        </div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Requests / sec</div>
        <div class="metric-value">{{ Math.round(latest.rps).toLocaleString() }}</div>
        <div style="margin-top:6px;">
          <svg width="100%" height="44" viewBox="0 0 200 44" preserveAspectRatio="none" style="display:block;">
            <path :d="sparkPath(history.map(h=>h.rps), rpsMax)" fill="none" stroke="var(--color-success)" stroke-width="1.5"/>
          </svg>
        </div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Avg Latency</div>
        <div class="metric-value">{{ Math.round(latest.latency) }}<span style="font-size:15px;font-weight:500;">ms</span></div>
        <div style="margin-top:6px;">
          <svg width="100%" height="44" viewBox="0 0 200 44" preserveAspectRatio="none" style="display:block;">
            <path :d="sparkPath(history.map(h=>h.latency), latMax)" fill="none" stroke="var(--color-warning)" stroke-width="1.5"/>
          </svg>
        </div>
      </div>
    </div>

    <!-- Connections Table -->
    <div class="card">
      <div class="card-header">
        <div class="card-title">Active Connections</div>
        <div style="font-size:22px;font-weight:700;color:var(--color-text);">{{ server.activeConnections.toLocaleString() }}</div>
      </div>
      <div class="card-body">
        <div style="display:flex;flex-direction:column;gap:12px;">
          <div v-for="label in [
            { label: 'HTTP/2', pct: 62 },
            { label: 'HTTP/1.1', pct: 28 },
            { label: 'gRPC', pct: 7 },
            { label: 'WebSocket', pct: 3 },
          ]" :key="label.label" style="display:flex;align-items:center;gap:12px;">
            <div style="font-size:12.5px;color:var(--color-text-2);width:90px;">{{ label.label }}</div>
            <div class="progress-bar" style="flex:1;height:6px;">
              <div class="progress-fill" :style="{ width: label.pct + '%' }" />
            </div>
            <div style="font-size:12px;color:var(--color-text-2);width:36px;text-align:right;">{{ label.pct }}%</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Health Checks -->
    <div class="card">
      <div class="card-header"><div class="card-title">Health Endpoints</div></div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Endpoint</th>
              <th>Method</th>
              <th>Status</th>
              <th>Response Time</th>
              <th>Last Checked</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="ep in [
              { path: '/health', method: 'GET', code: 200, ms: 2, ok: true },
              { path: '/metrics', method: 'GET', code: 200, ms: 5, ok: true },
              { path: '/ready', method: 'GET', code: 200, ms: 3, ok: true },
              { path: '/nginx_status', method: 'GET', code: 200, ms: 1, ok: true },
            ]" :key="ep.path">
              <td><code style="font-size:12px;font-family:var(--font-mono);">{{ ep.path }}</code></td>
              <td><span style="font-size:11px;font-family:var(--font-mono);color:var(--color-info);">{{ ep.method }}</span></td>
              <td>
                <span :class="['badge', ep.ok ? 'badge-online' : 'badge-offline']">{{ ep.code }}</span>
              </td>
              <td style="font-family:var(--font-mono);font-size:12px;">{{ ep.ms }}ms</td>
              <td style="color:var(--color-text-3);font-size:12px;">just now</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
