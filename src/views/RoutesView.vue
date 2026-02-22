<script setup lang="ts">
import { ref, computed } from 'vue'
import { mockRoutes, mockServers } from '../composables/useMockData'
import type { RouteRule } from '../types'

const search = ref('')
const showAddModal = ref(false)
const selectedRoute = ref<RouteRule | null>(null)

const routes = ref<RouteRule[]>([...mockRoutes])

const filtered = computed(() => routes.value.filter(r => {
  const q = search.value.toLowerCase()
  return !q || r.name.toLowerCase().includes(q) || r.matchHost?.includes(q) || r.serverName.toLowerCase().includes(q)
}))

function toggleRoute(id: string) {
  const r = routes.value.find(r => r.id === id)
  if (r) r.enabled = !r.enabled
}

function lbLabel(m: string) {
  return { round_robin: 'Round Robin', least_conn: 'Least Conn', ip_hash: 'IP Hash', random: 'Random' }[m] || m
}

function certDaysLeft(expiry?: string) {
  if (!expiry) return null
  const d = Math.ceil((new Date(expiry).getTime() - Date.now()) / 86400000)
  return d
}

const newRoute = ref<Partial<RouteRule>>({
  loadBalancingMethod: 'round_robin',
  sslEnabled: true,
  enabled: true,
  middlewares: [],
})
</script>

<template>
  <div class="page animate-slide">
    <div class="page-header">
      <div>
        <div class="page-title">Route Builder</div>
        <div class="page-description">Manage routing rules, load balancing, and SSL across all proxies</div>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-primary" @click="showAddModal = true">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          Add Route
        </button>
      </div>
    </div>

    <!-- Search -->
    <div style="position:relative;max-width:320px;">
      <svg style="position:absolute;left:9px;top:50%;transform:translateY(-50%);width:13px;height:13px;color:var(--color-text-3);"
           viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
      </svg>
      <input v-model="search" class="input" placeholder="Search routes…" style="padding-left:30px;height:34px;" />
    </div>

    <!-- Routes Table -->
    <div class="card">
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Enabled</th>
              <th>Name</th>
              <th>Server</th>
              <th>Match</th>
              <th>Upstream</th>
              <th>LB Method</th>
              <th>SSL</th>
              <th>Middlewares</th>
              <th>Priority</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="route in filtered" :key="route.id">
              <td>
                <div @click="toggleRoute(route.id)"
                     style="width:34px;height:18px;border-radius:99px;position:relative;cursor:pointer;transition:background var(--t);"
                     :style="{ background: route.enabled ? 'var(--color-primary)' : 'var(--color-surface-3)' }">
                  <div style="position:absolute;top:2px;width:14px;height:14px;border-radius:50%;background:#fff;transition:left var(--t);box-shadow:0 1px 3px rgba(0,0,0,.2);"
                       :style="{ left: route.enabled ? '18px' : '2px' }"></div>
                </div>
              </td>
              <td>
                <div style="font-weight:500;color:var(--color-text);font-size:13px;">{{ route.name }}</div>
              </td>
              <td>
                <div style="display:flex;align-items:center;gap:6px;">
                  <div :class="['proxy-badge', `proxy-${mockServers.find(s=>s.id===route.serverId)?.proxyType||'other'}`]" style="font-size:9.5px;padding:1px 5px;">
                    {{ (mockServers.find(s=>s.id===route.serverId)?.proxyType||'other').toUpperCase() }}
                  </div>
                  <span style="font-size:12px;color:var(--color-text-2);">{{ route.serverName }}</span>
                </div>
              </td>
              <td>
                <div style="font-family:var(--font-mono);font-size:11.5px;">
                  <span v-if="route.matchHost" style="color:var(--color-primary);">{{ route.matchHost }}</span>
                  <span v-if="route.matchPath" style="color:var(--color-text-2);">{{ route.matchPath }}</span>
                </div>
              </td>
              <td>
                <div style="font-family:var(--font-mono);font-size:11.5px;color:var(--color-text-2);max-width:180px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">
                  {{ route.targetUpstream }}
                </div>
              </td>
              <td>
                <span class="badge badge-info" style="font-size:10.5px;">{{ lbLabel(route.loadBalancingMethod) }}</span>
              </td>
              <td>
                <template v-if="route.sslEnabled">
                  <div style="display:flex;align-items:center;gap:6px;">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--color-success)" stroke-width="2">
                      <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/>
                    </svg>
                    <span v-if="route.sslCertExpiry" style="font-size:11px;"
                          :style="{ color: (certDaysLeft(route.sslCertExpiry)||999) < 30 ? 'var(--color-warning)' : 'var(--color-text-3)' }">
                      {{ certDaysLeft(route.sslCertExpiry) }}d
                    </span>
                  </div>
                </template>
                <span v-else style="font-size:12px;color:var(--color-text-3);">—</span>
              </td>
              <td>
                <div style="display:flex;gap:4px;flex-wrap:wrap;">
                  <span v-for="m in route.middlewares" :key="m" class="badge badge-idle" style="font-size:10px;padding:1px 6px;">{{ m }}</span>
                  <span v-if="route.middlewares.length===0" style="font-size:12px;color:var(--color-text-3);">none</span>
                </div>
              </td>
              <td style="font-size:12px;color:var(--color-text-3);">{{ route.priority }}</td>
              <td>
                <div style="display:flex;gap:4px;">
                  <button class="btn btn-ghost btn-sm btn-icon" @click="selectedRoute = route" data-tooltip="Edit">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z"/>
                    </svg>
                  </button>
                  <button class="btn btn-danger btn-sm btn-icon" data-tooltip="Delete">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/>
                      <path d="M10 11v6"/><path d="M14 11v6"/>
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- SSL Certificate Overview -->
    <div class="card">
      <div class="card-header"><div class="card-title">SSL Certificate Status</div></div>
      <div class="card-body">
        <div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:12px;">
          <div v-for="route in routes.filter(r=>r.sslEnabled && r.sslCertExpiry)" :key="route.id"
               style="background:var(--color-surface-2);border:1px solid var(--color-border);border-radius:var(--r);padding:12px;">
            <div style="font-size:12.5px;font-weight:500;color:var(--color-text);">{{ route.matchHost || route.name }}</div>
            <div style="font-size:11.5px;color:var(--color-text-3);margin-top:2px;">{{ route.serverName }}</div>
            <div style="margin-top:8px;display:flex;align-items:center;justify-content:space-between;">
              <span :class="['badge', (certDaysLeft(route.sslCertExpiry)||999) < 30 ? 'badge-warning' : 'badge-online']" style="font-size:11px;">
                {{ certDaysLeft(route.sslCertExpiry) }} days left
              </span>
              <span style="font-size:11px;color:var(--color-text-3);">{{ route.sslCertExpiry }}</span>
            </div>
            <div style="margin-top:8px;">
              <div class="progress-bar">
                <div class="progress-fill"
                     :class="{ warn: (certDaysLeft(route.sslCertExpiry)||999) < 30 }"
                     :style="{ width: Math.min(100, ((certDaysLeft(route.sslCertExpiry)||0) / 365) * 100) + '%' }" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Route Modal -->
    <Teleport to="body">
      <div v-if="showAddModal" class="modal-backdrop" @click.self="showAddModal=false">
        <div class="modal" style="max-width:540px;">
          <div class="modal-header">
            <div class="modal-title">Create Route Rule</div>
            <button class="btn btn-ghost btn-icon" @click="showAddModal=false">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label class="form-label">Route Name</label>
              <input class="input" v-model="newRoute.name" placeholder="e.g. API Gateway v2" />
            </div>
            <div class="form-group">
              <label class="form-label">Server</label>
              <select class="input" v-model="newRoute.serverId">
                <option v-for="s in mockServers" :key="s.id" :value="s.id">{{ s.name }}</option>
              </select>
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
              <div class="form-group">
                <label class="form-label">Match Host</label>
                <input class="input" v-model="newRoute.matchHost" placeholder="api.example.com" />
              </div>
              <div class="form-group">
                <label class="form-label">Match Path</label>
                <input class="input" v-model="newRoute.matchPath" placeholder="/api/*" />
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Target Upstream</label>
              <input class="input" v-model="newRoute.targetUpstream" placeholder="http://backend:3000" />
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
              <div class="form-group">
                <label class="form-label">Load Balancing</label>
                <select class="input" v-model="newRoute.loadBalancingMethod">
                  <option value="round_robin">Round Robin</option>
                  <option value="least_conn">Least Connections</option>
                  <option value="ip_hash">IP Hash</option>
                  <option value="random">Random</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Priority</label>
                <input class="input" v-model.number="newRoute.priority" type="number" placeholder="100" />
              </div>
            </div>
            <div class="flex gap-3 items-center">
              <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:13px;color:var(--color-text-2);">
                <input type="checkbox" v-model="newRoute.sslEnabled" />
                Enable SSL/TLS
              </label>
              <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:13px;color:var(--color-text-2);">
                <input type="checkbox" v-model="newRoute.enabled" />
                Enable route
              </label>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showAddModal=false">Cancel</button>
            <button class="btn btn-primary" @click="showAddModal=false">Create Route</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
