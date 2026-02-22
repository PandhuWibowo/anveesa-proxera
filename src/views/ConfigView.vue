<script setup lang="ts">
import { ref, computed } from 'vue'
import { mockServers, mockNginxConfig, mockTraefikConfig } from '../composables/useMockData'
import type { ProxyServer } from '../types'

const selected = ref('s1')
const activeTab = ref<'view' | 'edit' | 'diff'>('view')
const isEditing = ref(false)
const editContent = ref('')

const server = computed<ProxyServer>(() => mockServers.find(s => s.id === selected.value) ?? mockServers[0]!)

const configMap: Record<string, { content: string; format: string }> = {
  s1: { content: mockNginxConfig, format: 'nginx' },
  s2: { content: mockTraefikConfig, format: 'yaml' },
  s3: { content: `# Caddyfile — staging-caddy-01
# Auto-TLS via Let's Encrypt

staging.example.com {
    tls ops@example.com

    encode gzip

    reverse_proxy /api/* {
        to http://staging-app:5173
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
    }

    handle_errors {
        respond "Server Error" 500
    }

    log {
        output file /var/log/caddy/staging.log
        format json
    }
}`, format: 'caddyfile' },
  s4: { content: `# HAProxy Configuration — prod-haproxy-lb
global
    log /dev/log local0
    maxconn 100000
    user haproxy
    group haproxy
    daemon
    stats socket /run/haproxy/admin.sock mode 660 level admin

defaults
    log     global
    mode    http
    option  httplog
    option  dontlognull
    timeout connect 5s
    timeout client  30s
    timeout server  30s
    errorfile 400 /etc/haproxy/errors/400.http
    errorfile 503 /etc/haproxy/errors/503.http

frontend http_in
    bind *:80
    bind *:443 ssl crt /etc/ssl/certs/example.com.pem
    redirect scheme https if !{ ssl_fc }
    default_backend nginx_pool

backend nginx_pool
    balance leastconn
    option forwardfor
    option http-server-close
    server nginx1 10.0.1.10:80 check inter 2s
    server nginx2 10.0.1.11:80 check inter 2s
    server nginx3 10.0.1.12:80 check inter 2s backup

listen stats
    bind *:9000
    stats enable
    stats uri /stats
    stats refresh 10s
    stats auth admin:changeme`, format: 'haproxy' },
  s5: { content: `# Envoy Bootstrap Config — dev-envoy-01
admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901

static_resources:
  listeners:
  - name: listener_http
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 8080
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          route_config:
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: service_backend
  clusters:
  - name: service_backend
    type: STRICT_DNS
    load_assignment:
      cluster_name: service_backend
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: backend
                port_value: 8080`, format: 'yaml' },
}

const config = computed<{ content: string; format: string }>(() => configMap[selected.value] ?? { content: '', format: 'nginx' })

function startEdit() {
  editContent.value = config.value.content
  isEditing.value = true
  activeTab.value = 'edit'
}

function copyConfig() {
  navigator.clipboard?.writeText(config.value.content)
}

function syntaxHighlight(code: string) {
  return code
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/(#[^\n]*)/g, '<span style="color:var(--color-text-3)">$1</span>')
    .replace(/\b(server_name|location|proxy_pass|ssl_certificate|listen|upstream|backend|frontend)\b/g,
      '<span style="color:var(--color-primary)">$1</span>')
    .replace(/\b(http|https|ssl|tcp)\b/g, '<span style="color:var(--color-info)">$1</span>')
    .replace(/(:\s*)(\d+)/g, '$1<span style="color:var(--color-warning)">$2</span>')
}
</script>

<template>
  <div class="page animate-slide">
    <div class="page-header">
      <div>
        <div class="page-title">Configuration</div>
        <div class="page-description">View and edit proxy configurations with syntax highlighting and validation</div>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-secondary" @click="startEdit">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z"/>
          </svg>
          Edit Config
        </button>
        <button class="btn btn-secondary">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          Download
        </button>
      </div>
    </div>

    <div style="display:grid;grid-template-columns:220px 1fr;gap:16px;align-items:start;">
      <!-- Server List -->
      <div class="card">
        <div class="card-header" style="padding:12px 14px;">
          <div class="card-title">Servers</div>
        </div>
        <div style="padding:6px 0;">
          <div v-for="s in mockServers" :key="s.id"
               @click="selected = s.id; isEditing = false; activeTab = 'view'"
               style="display:flex;align-items:center;gap:8px;padding:8px 12px;cursor:pointer;transition:background var(--t);"
               :style="{ background: selected===s.id ? 'var(--color-primary-alpha)' : '' }"
               onmouseover="if(this.style.background==='') this.style.background='var(--color-surface-2)'"
               onmouseout="this.style.background=selected===this.dataset.id?'var(--color-primary-alpha)':''"
          >
            <div :class="['proxy-badge', `proxy-${s.proxyType}`]" style="font-size:9.5px;padding:1px 5px;">
              {{ s.proxyType.toUpperCase() }}
            </div>
            <div style="flex:1;min-width:0;">
              <div style="font-size:12px;font-weight:500;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;"
                   :style="{ color: selected===s.id ? 'var(--color-primary)' : 'var(--color-text)' }">
                {{ s.name }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Config Panel -->
      <div class="card">
        <div class="card-header">
          <div style="display:flex;align-items:center;gap:10px;">
            <div :class="['proxy-badge', `proxy-${server.proxyType}`]">{{ server.proxyType.toUpperCase() }}</div>
            <div class="card-title">{{ server.name }}</div>
            <span class="badge badge-info" style="font-size:10px;">{{ config.format }}</span>
          </div>
          <div class="tab-list">
            <div class="tab-item" :class="{ active: activeTab==='view' }" @click="activeTab='view'; isEditing=false">View</div>
            <div class="tab-item" :class="{ active: activeTab==='edit' }" @click="startEdit">Edit</div>
            <div class="tab-item" :class="{ active: activeTab==='diff' }" @click="activeTab='diff'">Diff</div>
          </div>
        </div>

        <!-- Validation Banner -->
        <div v-if="isEditing" style="padding:8px 16px;background:var(--color-success-alpha);border-bottom:1px solid var(--color-border);display:flex;align-items:center;gap:8px;">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="var(--color-success)" stroke-width="2">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          <span style="font-size:12px;color:var(--color-success);font-weight:500;">Configuration is valid — no errors found</span>
        </div>

        <!-- View -->
        <div v-if="activeTab==='view'" style="position:relative;">
          <div style="display:flex;align-items:center;gap:8px;padding:6px 14px;background:var(--color-surface-2);border-bottom:1px solid var(--color-border);">
            <div style="display:flex;gap:4px;">
              <div style="width:10px;height:10px;border-radius:50%;background:#ef4444;"></div>
              <div style="width:10px;height:10px;border-radius:50%;background:#f59e0b;"></div>
              <div style="width:10px;height:10px;border-radius:50%;background:#22c55e;"></div>
            </div>
            <div style="font-size:11px;color:var(--color-text-3);font-family:var(--font-mono);">{{ server.name }}.conf</div>
            <button class="btn btn-ghost btn-sm" style="margin-left:auto;font-size:11px;" @click="copyConfig">
              Copy
            </button>
          </div>
          <div style="overflow:auto;max-height:600px;">
            <div style="display:flex;">
              <!-- Line numbers -->
              <div style="padding:12px 8px;background:var(--color-surface-2);border-right:1px solid var(--color-border);text-align:right;user-select:none;min-width:40px;">
                <div v-for="(_, i) in config.content.split('\n')" :key="i"
                     style="font-family:var(--font-mono);font-size:11.5px;color:var(--color-text-3);line-height:1.7;">
                  {{ i + 1 }}
                </div>
              </div>
              <pre style="flex:1;padding:12px 16px;font-family:var(--font-mono);font-size:12px;line-height:1.7;color:var(--color-text-2);overflow:visible;background:var(--color-surface);margin:0;"
                   v-html="syntaxHighlight(config.content)"></pre>
            </div>
          </div>
        </div>

        <!-- Edit -->
        <div v-else-if="activeTab==='edit'">
          <textarea
            v-model="editContent"
            class="input"
            style="height:560px;border:none;border-radius:0;resize:none;font-family:var(--font-mono);font-size:12px;line-height:1.7;padding:14px 16px;"
            spellcheck="false"
          />
          <div style="padding:12px 16px;border-top:1px solid var(--color-border);display:flex;gap:8px;justify-content:flex-end;">
            <button class="btn btn-secondary" @click="isEditing=false;activeTab='view'">Discard</button>
            <button class="btn btn-primary">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              Validate & Save
            </button>
          </div>
        </div>

        <!-- Diff -->
        <div v-else-if="activeTab==='diff'">
          <div style="padding:24px;text-align:center;color:var(--color-text-3);">
            <div style="font-size:32px;margin-bottom:8px;">⚡</div>
            <div style="font-size:13.5px;font-weight:600;color:var(--color-text-2);">No pending changes</div>
            <div style="font-size:12.5px;margin-top:4px;">Edit the configuration to see a diff comparison</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
