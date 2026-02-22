<script setup lang="ts">
import { ref, computed } from 'vue'
import { mockServers } from '../composables/useMockData'
import type { ProxyServer, ProxyType } from '../types'

const search = ref('')
const filterType = ref<ProxyType | 'all'>('all')
const showAddModal = ref(false)

const filtered = computed(() => {
  return mockServers.filter(s => {
    const q = search.value.toLowerCase()
    const matchSearch = !q || s.name.includes(q) || s.host.includes(q) || s.location?.toLowerCase().includes(q)
    const matchType = filterType.value === 'all' || s.proxyType === filterType.value
    return matchSearch && matchType
  })
})

const newServer = ref<Partial<ProxyServer>>({
  proxyType: 'nginx',
  connectionType: 'ssh',
  port: 80,
})

function statusDot(status: string) {
  if (status === 'online') return 'badge-online'
  if (status === 'warning') return 'badge-warning'
  return 'badge-offline'
}

function proxyTypeLabel(t: string) {
  return { nginx: 'NGINX', traefik: 'Traefik', caddy: 'Caddy', haproxy: 'HAProxy', other: 'Other' }[t] || t
}
</script>

<template>
  <div class="page animate-slide">
    <!-- Header -->
    <div class="page-header">
      <div>
        <div class="page-title">Proxy Servers</div>
        <div class="page-description">{{ mockServers.length }} servers registered · {{ mockServers.filter(s=>s.status==='online').length }} online</div>
      </div>
      <div class="page-actions">
        <button class="btn btn-primary" @click="showAddModal = true">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          Add Server
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex gap-3 items-center">
      <div style="position:relative;flex:1;max-width:320px;">
        <svg style="position:absolute;left:9px;top:50%;transform:translateY(-50%);width:13px;height:13px;color:var(--color-text-3);"
             viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
        </svg>
        <input v-model="search" class="input" placeholder="Search servers…" style="padding-left:30px;height:34px;" />
      </div>
      <div class="tab-list">
        <div class="tab-item" :class="{ active: filterType === 'all' }" @click="filterType = 'all'">All</div>
        <div class="tab-item" :class="{ active: filterType === 'nginx' }" @click="filterType = 'nginx'">NGINX</div>
        <div class="tab-item" :class="{ active: filterType === 'traefik' }" @click="filterType = 'traefik'">Traefik</div>
        <div class="tab-item" :class="{ active: filterType === 'caddy' }" @click="filterType = 'caddy'">Caddy</div>
        <div class="tab-item" :class="{ active: filterType === 'haproxy' }" @click="filterType = 'haproxy'">HAProxy</div>
      </div>
    </div>

    <!-- Server Grid -->
    <div class="grid-2" style="grid-template-columns:repeat(auto-fill,minmax(340px,1fr));">
      <div v-for="server in filtered" :key="server.id" class="card" style="display:flex;flex-direction:column;gap:0;">
        <!-- Card Header -->
        <div style="padding:14px 16px;border-bottom:1px solid var(--color-border);display:flex;align-items:center;gap:10px;">
          <div :class="['proxy-badge', `proxy-${server.proxyType}`]">{{ proxyTypeLabel(server.proxyType) }}</div>
          <div style="flex:1;min-width:0;">
            <div style="font-size:13.5px;font-weight:600;color:var(--color-text);truncate;">{{ server.name }}</div>
            <div style="font-size:11px;color:var(--color-text-3);">{{ server.host }}:{{ server.port }}</div>
          </div>
          <div :class="['badge', statusDot(server.status)]">
            <div class="badge-dot" :style="{
              background: server.status==='online'?'var(--color-success)':server.status==='warning'?'var(--color-warning)':'var(--color-error)',
              animation: server.status==='online' ? 'pulse 2s infinite' : 'none'
            }"></div>
            {{ server.status }}
          </div>
        </div>

        <!-- Metrics Row -->
        <div style="display:grid;grid-template-columns:repeat(3,1fr);gap:0;border-bottom:1px solid var(--color-border);">
          <div style="padding:10px 14px;text-align:center;border-right:1px solid var(--color-border);">
            <div style="font-size:15px;font-weight:700;color:var(--color-text);">{{ server.requestsPerSec.toLocaleString() }}</div>
            <div style="font-size:10px;color:var(--color-text-3);text-transform:uppercase;letter-spacing:.04em;">req/s</div>
          </div>
          <div style="padding:10px 14px;text-align:center;border-right:1px solid var(--color-border);">
            <div style="font-size:15px;font-weight:700;color:var(--color-text);">{{ server.activeConnections.toLocaleString() }}</div>
            <div style="font-size:10px;color:var(--color-text-3);text-transform:uppercase;letter-spacing:.04em;">connections</div>
          </div>
          <div style="padding:10px 14px;text-align:center;">
            <div style="font-size:15px;font-weight:700;" :style="{ color: server.errorRate > 1 ? 'var(--color-error)' : 'var(--color-text)' }">{{ server.errorRate.toFixed(2) }}%</div>
            <div style="font-size:10px;color:var(--color-text-3);text-transform:uppercase;letter-spacing:.04em;">errors</div>
          </div>
        </div>

        <!-- Resources -->
        <div style="padding:12px 14px;display:flex;flex-direction:column;gap:8px;">
          <div style="display:flex;align-items:center;gap:10px;">
            <div style="font-size:11px;color:var(--color-text-3);width:28px;">CPU</div>
            <div class="progress-bar" style="flex:1;">
              <div class="progress-fill" :class="{ warn: server.cpuUsage>60, error: server.cpuUsage>80 }"
                   :style="{ width: server.cpuUsage + '%' }" />
            </div>
            <div style="font-size:11.5px;color:var(--color-text-2);width:32px;text-align:right;">{{ server.cpuUsage }}%</div>
          </div>
          <div style="display:flex;align-items:center;gap:10px;">
            <div style="font-size:11px;color:var(--color-text-3);width:28px;">MEM</div>
            <div class="progress-bar" style="flex:1;">
              <div class="progress-fill" :class="{ warn: server.memUsage>70, error: server.memUsage>85 }"
                   :style="{ width: server.memUsage + '%' }" />
            </div>
            <div style="font-size:11.5px;color:var(--color-text-2);width:32px;text-align:right;">{{ server.memUsage }}%</div>
          </div>
        </div>

        <!-- Footer -->
        <div style="padding:10px 14px;border-top:1px solid var(--color-border);display:flex;align-items:center;gap:8px;">
          <div style="font-size:11px;color:var(--color-text-3);">
            <span>v{{ server.version }}</span>
            <span style="margin:0 6px;">·</span>
            <span>↑ {{ server.uptime }}</span>
          </div>
          <div style="margin-left:auto;display:flex;gap:6px;">
            <button class="btn btn-ghost btn-sm btn-icon" data-tooltip="View config">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/>
              </svg>
            </button>
            <button class="btn btn-ghost btn-sm btn-icon" data-tooltip="Monitoring">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
              </svg>
            </button>
            <button class="btn btn-ghost btn-sm btn-icon" data-tooltip="Edit server">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-if="filtered.length === 0" class="card">
      <div class="empty-state">
        <div class="empty-state-icon">🖥️</div>
        <div class="empty-state-title">No servers found</div>
        <div class="empty-state-desc">Try adjusting your search or filters</div>
      </div>
    </div>

    <!-- Add Server Modal -->
    <Teleport to="body">
      <div v-if="showAddModal" class="modal-backdrop" @click.self="showAddModal = false">
        <div class="modal">
          <div class="modal-header">
            <div class="modal-title">Add Proxy Server</div>
            <button class="btn btn-ghost btn-icon" @click="showAddModal = false">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label class="form-label">Server Name</label>
              <input class="input" v-model="newServer.name" placeholder="e.g. prod-nginx-02" />
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
              <div class="form-group">
                <label class="form-label">Host / IP</label>
                <input class="input" v-model="newServer.host" placeholder="10.0.1.10" />
              </div>
              <div class="form-group">
                <label class="form-label">Port</label>
                <input class="input" v-model.number="newServer.port" type="number" placeholder="80" />
              </div>
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
              <div class="form-group">
                <label class="form-label">Proxy Type</label>
                <select class="input" v-model="newServer.proxyType">
                  <option value="nginx">NGINX</option>
                  <option value="traefik">Traefik</option>
                  <option value="caddy">Caddy</option>
                  <option value="haproxy">HAProxy</option>
                  <option value="other">Other</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Connection</label>
                <select class="input" v-model="newServer.connectionType">
                  <option value="ssh">SSH</option>
                  <option value="api">REST API</option>
                </select>
              </div>
            </div>
            <template v-if="newServer.connectionType === 'ssh'">
              <div class="form-group">
                <label class="form-label">SSH Username</label>
                <input class="input" v-model="newServer.sshUser" placeholder="ubuntu" />
              </div>
              <div class="form-group">
                <label class="form-label">SSH Key Path</label>
                <input class="input" v-model="newServer.sshKeyPath" placeholder="~/.ssh/id_rsa" />
                <div class="form-hint">Path to private key on your machine</div>
              </div>
            </template>
            <template v-if="newServer.connectionType === 'api'">
              <div class="form-group">
                <label class="form-label">API Base URL</label>
                <input class="input" v-model="newServer.apiUrl" placeholder="http://10.0.1.20:8080/api" />
              </div>
              <div class="form-group">
                <label class="form-label">API Token</label>
                <input class="input" v-model="newServer.apiToken" type="password" placeholder="Bearer token" />
              </div>
            </template>
            <div class="form-group">
              <label class="form-label">Description (optional)</label>
              <input class="input" v-model="newServer.description" placeholder="Brief description of this server" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showAddModal = false">Cancel</button>
            <button class="btn btn-primary" @click="showAddModal = false">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              Connect Server
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
