<script setup lang="ts">
import { ref, computed } from 'vue'
import { generateTrafficData, mockServers } from '../composables/useMockData'

const period = ref<'24h' | '7d' | '30d'>('24h')
const selectedServer = ref('all')

const traffic24h = generateTrafficData(24)
const traffic7d   = Array.from({ length: 7 }, (_, i) => {
  const d = new Date(); d.setDate(d.getDate() - 6 + i)
  const label = d.toLocaleDateString('en', { weekday: 'short' })
  const base = 80000 + Math.random() * 60000
  return { time: label, requests: Math.floor(base), errors: Math.floor(base * 0.005), latency: 30 + Math.random() * 40 }
})
const traffic30d = Array.from({ length: 30 }, (_, i) => {
  const d = new Date(); d.setDate(d.getDate() - 29 + i)
  const label = `${d.getMonth()+1}/${d.getDate()}`
  const base = 80000 + Math.random() * 100000
  return { time: label, requests: Math.floor(base), errors: Math.floor(base * 0.006), latency: 28 + Math.random() * 50 }
})

const chartData = computed(() => period.value === '24h' ? traffic24h : period.value === '7d' ? traffic7d : traffic30d)
const maxReq = computed(() => Math.max(...chartData.value.map(d => d.requests), 1))

const totalRequests = computed(() => chartData.value.reduce((a, d) => a + d.requests, 0))
const totalErrors   = computed(() => chartData.value.reduce((a, d) => a + d.errors, 0))
const avgLatency    = computed(() => Math.round(chartData.value.reduce((a, d) => a + d.latency, 0) / chartData.value.length))
const errorRate     = computed(() => ((totalErrors.value / totalRequests.value) * 100).toFixed(2))

function fmtNum(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K'
  return String(n)
}

const topEndpoints = [
  { path: 'GET /api/v1/users',       requests: 142300, errorRate: 0.1,  avgMs: 22  },
  { path: 'POST /api/v1/auth/login', requests: 98700,  errorRate: 2.1,  avgMs: 145 },
  { path: 'GET /api/v1/products',    requests: 87200,  errorRate: 0.3,  avgMs: 38  },
  { path: 'PUT /api/v1/cart',        requests: 65400,  errorRate: 0.8,  avgMs: 62  },
  { path: 'GET /api/v1/health',      requests: 420000, errorRate: 0.0,  avgMs: 2   },
  { path: 'POST /api/v1/orders',     requests: 32100,  errorRate: 1.4,  avgMs: 210 },
  { path: 'GET /api/v1/search',      requests: 28900,  errorRate: 0.4,  avgMs: 88  },
  { path: 'DELETE /api/v1/session',  requests: 18200,  errorRate: 0.0,  avgMs: 8   },
]

const statusCodes = [
  { code: '2xx', count: 4210000, pct: 87.3, color: 'var(--color-success)' },
  { code: '3xx', count: 310000,  pct: 6.4,  color: 'var(--color-info)'    },
  { code: '4xx', count: 240000,  pct: 5.0,  color: 'var(--color-warning)' },
  { code: '5xx', count: 63000,   pct: 1.3,  color: 'var(--color-error)'   },
]

const maxEndpointReq = Math.max(...topEndpoints.map(e => e.requests))
</script>

<template>
  <div class="page animate-slide">
    <div class="page-header">
      <div>
        <div class="page-title">Traffic Analytics</div>
        <div class="page-description">Request volume, latency, error rates, and top endpoints</div>
      </div>
      <div class="flex gap-2 items-center">
        <select class="input" v-model="selectedServer" style="height:34px;font-size:12.5px;width:auto;">
          <option value="all">All Servers</option>
          <option v-for="s in mockServers" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
        <div class="tab-list">
          <div class="tab-item" :class="{ active: period==='24h' }" @click="period='24h'">24h</div>
          <div class="tab-item" :class="{ active: period==='7d' }"  @click="period='7d'">7d</div>
          <div class="tab-item" :class="{ active: period==='30d' }" @click="period='30d'">30d</div>
        </div>
        <button class="btn btn-secondary btn-sm">Export CSV</button>
      </div>
    </div>

    <!-- Summary Stats -->
    <div class="grid-4">
      <div class="metric-card">
        <div class="metric-label">Total Requests</div>
        <div class="metric-value">{{ fmtNum(totalRequests) }}</div>
        <div class="metric-sub metric-trend up">↑ 14.2% vs previous</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Total Errors</div>
        <div class="metric-value" style="color:var(--color-error);">{{ fmtNum(totalErrors) }}</div>
        <div class="metric-sub">{{ errorRate }}% error rate</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Avg Latency</div>
        <div class="metric-value">{{ avgLatency }}<span style="font-size:16px;font-weight:500;">ms</span></div>
        <div class="metric-sub metric-trend up">p95: 112ms · p99: 280ms</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Bandwidth Out</div>
        <div class="metric-value">4.2<span style="font-size:16px;font-weight:500;">TB</span></div>
        <div class="metric-sub metric-trend up">↑ 8.7% vs previous</div>
      </div>
    </div>

    <!-- Request Volume Chart -->
    <div class="card">
      <div class="card-header">
        <div class="card-title">Request Volume</div>
      </div>
      <div class="card-body">
        <div style="display:flex;align-items:flex-end;gap:3px;height:160px;overflow:hidden;">
          <div v-for="(d, i) in chartData" :key="i"
               style="flex:1;display:flex;flex-direction:column;align-items:center;gap:3px;min-width:0;">
            <div style="width:100%;display:flex;flex-direction:column;gap:1px;justify-content:flex-end;flex:1;">
              <!-- error bar -->
              <div :style="{
                width:'100%',height:Math.max(1,(d.errors/maxReq)*148)+'px',
                background:'var(--color-error)',borderRadius:'2px 2px 0 0',opacity:.7
              }"></div>
              <!-- request bar -->
              <div :style="{
                width:'100%',height:Math.max(2,((d.requests-d.errors)/maxReq)*148)+'px',
                background:'var(--color-primary)',borderRadius:'2px 2px 0 0',opacity:.85
              }"></div>
            </div>
            <div style="font-size:9px;color:var(--color-text-3);white-space:nowrap;overflow:hidden;max-width:100%;text-align:center;">
              {{ d.time }}
            </div>
          </div>
        </div>
        <div class="flex gap-4 items-center" style="margin-top:12px;padding-top:12px;border-top:1px solid var(--color-border);">
          <div class="flex gap-2 items-center text-sm text-muted">
            <div style="width:10px;height:10px;background:var(--color-primary);border-radius:2px;opacity:.85;"></div>
            Requests
          </div>
          <div class="flex gap-2 items-center text-sm text-muted">
            <div style="width:10px;height:10px;background:var(--color-error);border-radius:2px;opacity:.7;"></div>
            Errors
          </div>
          <div class="ml-auto text-sm text-muted">Total: {{ fmtNum(totalRequests) }} requests</div>
        </div>
      </div>
    </div>

    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;">
      <!-- Top Endpoints -->
      <div class="card">
        <div class="card-header"><div class="card-title">Top Endpoints</div></div>
        <div style="padding:8px 0;">
          <div v-for="ep in topEndpoints" :key="ep.path"
               style="padding:9px 16px;border-bottom:1px solid var(--color-border);">
            <div style="display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:5px;">
              <div style="font-family:var(--font-mono);font-size:11.5px;color:var(--color-text);flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">
                {{ ep.path }}
              </div>
              <div style="display:flex;gap:10px;flex-shrink:0;">
                <span style="font-size:11.5px;color:var(--color-text-2);">{{ fmtNum(ep.requests) }}</span>
                <span style="font-size:11.5px;" :style="{ color: ep.errorRate>1?'var(--color-error)':'var(--color-text-3)' }">
                  {{ ep.errorRate }}% err
                </span>
                <span style="font-size:11.5px;color:var(--color-text-3);">{{ ep.avgMs }}ms</span>
              </div>
            </div>
            <div class="progress-bar" style="height:3px;">
              <div class="progress-fill" :style="{ width: (ep.requests/maxEndpointReq*100)+'%' }" />
            </div>
          </div>
        </div>
      </div>

      <!-- Status Code Distribution -->
      <div class="card">
        <div class="card-header"><div class="card-title">Status Code Distribution</div></div>
        <div class="card-body" style="display:flex;flex-direction:column;gap:16px;">
          <!-- Horizontal stack -->
          <div style="height:32px;border-radius:var(--r);overflow:hidden;display:flex;">
            <div v-for="s in statusCodes" :key="s.code"
                 :style="{ width: s.pct+'%', background: s.color, opacity: 0.8 }"></div>
          </div>

          <div style="display:flex;flex-direction:column;gap:10px;">
            <div v-for="s in statusCodes" :key="s.code" style="display:flex;align-items:center;gap:12px;">
              <div style="width:8px;height:8px;border-radius:50%;flex-shrink:0;" :style="{ background: s.color }"></div>
              <div style="width:32px;font-family:var(--font-mono);font-size:12.5px;font-weight:600;" :style="{ color: s.color }">{{ s.code }}</div>
              <div class="progress-bar" style="flex:1;">
                <div class="progress-fill" :style="{ width: s.pct+'%', background: s.color }" />
              </div>
              <div style="font-size:12px;color:var(--color-text-2);width:36px;text-align:right;">{{ s.pct }}%</div>
              <div style="font-size:11.5px;color:var(--color-text-3);width:52px;text-align:right;">{{ fmtNum(s.count) }}</div>
            </div>
          </div>

          <div style="border-top:1px solid var(--color-border);padding-top:12px;">
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
              <div>
                <div style="font-size:10.5px;text-transform:uppercase;letter-spacing:.05em;color:var(--color-text-3);">Success Rate</div>
                <div style="font-size:20px;font-weight:700;color:var(--color-success);margin-top:2px;">93.7%</div>
              </div>
              <div>
                <div style="font-size:10.5px;text-transform:uppercase;letter-spacing:.05em;color:var(--color-text-3);">5xx Rate</div>
                <div style="font-size:20px;font-weight:700;color:var(--color-error);margin-top:2px;">1.3%</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Latency Percentiles -->
    <div class="card">
      <div class="card-header"><div class="card-title">Latency Percentiles (per server)</div></div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Server</th>
              <th>Type</th>
              <th>p50</th>
              <th>p75</th>
              <th>p90</th>
              <th>p95</th>
              <th>p99</th>
              <th>p99.9</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in mockServers.filter(s=>s.status!=='offline')" :key="s.id">
              <td style="font-weight:500;color:var(--color-text);">{{ s.name }}</td>
              <td><div :class="['proxy-badge', `proxy-${s.proxyType}`]" style="font-size:10px;padding:1px 5px;">{{ s.proxyType.toUpperCase() }}</div></td>
              <td style="font-family:var(--font-mono);font-size:12px;">12ms</td>
              <td style="font-family:var(--font-mono);font-size:12px;">28ms</td>
              <td style="font-family:var(--font-mono);font-size:12px;">58ms</td>
              <td style="font-family:var(--font-mono);font-size:12px;color:var(--color-warning);">112ms</td>
              <td style="font-family:var(--font-mono);font-size:12px;color:var(--color-error);">280ms</td>
              <td style="font-family:var(--font-mono);font-size:12px;color:var(--color-error);">1.2s</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
