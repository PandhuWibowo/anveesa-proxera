<script setup lang="ts">
import { ref } from 'vue'
import { mockUsers } from '../composables/useMockData'
import { useTheme } from '../composables/useTheme'
import type { User } from '../types'

const { theme, setTheme } = useTheme()
const activeSection = ref('team')
const showInviteModal = ref(false)
const users = ref<User[]>([...mockUsers])

const inviteEmail = ref('')
const inviteRole = ref<'admin'|'operator'|'viewer'>('viewer')

function roleColor(role: string) {
  return { admin: 'badge-offline', operator: 'badge-warning', viewer: 'badge-online' }[role] || 'badge-idle'
}

function timeAgo(iso: string) {
  const d = Math.floor((Date.now() - new Date(iso).getTime()) / 86400000)
  if (d === 0) return 'today'
  if (d === 1) return 'yesterday'
  return `${d}d ago`
}

const notifSettings = ref({
  emailOnCritical: true,
  emailOnWarning: true,
  slackOnCritical: true,
  slackOnWarning: false,
  pagerdutyOnCritical: false,
})

const generalSettings = ref({
  healthCheckInterval: 30,
  logRetentionDays: 30,
  defaultLBMethod: 'round_robin',
  autoSSLRenew: true,
  metricsRetentionDays: 90,
})
</script>

<template>
  <div class="page animate-slide">
    <div class="page-title" style="margin-bottom:-8px;">Settings</div>

    <div style="display:grid;grid-template-columns:200px 1fr;gap:16px;align-items:start;">
      <!-- Sidebar Nav -->
      <div class="card" style="padding:8px 0;">
        <div v-for="item in [
          { key:'team',    label:'Team & Users',     icon:'👥' },
          { key:'notif',   label:'Notifications',    icon:'🔔' },
          { key:'general', label:'General',          icon:'⚙️' },
          { key:'theme',   label:'Appearance',       icon:'🎨' },
          { key:'api',     label:'API Keys',         icon:'🔑' },
          { key:'audit',   label:'Audit Log',        icon:'📋' },
        ]" :key="item.key"
          @click="activeSection = item.key"
          class="nav-item"
          :class="{ active: activeSection === item.key }"
          style="margin:0 8px;border-radius:var(--r-sm);"
        >
          <span>{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </div>
      </div>

      <!-- Settings Panel -->
      <div>
        <!-- Team & Users -->
        <div v-if="activeSection==='team'" class="card">
          <div class="card-header">
            <div class="card-title">Team Members</div>
            <button class="btn btn-primary btn-sm" @click="showInviteModal=true">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
              </svg>
              Invite User
            </button>
          </div>

          <!-- RBAC Info -->
          <div style="padding:14px 18px;border-bottom:1px solid var(--color-border);display:flex;gap:12px;flex-wrap:wrap;">
            <div v-for="role in [
              { role:'admin', desc:'Full access. Can manage users, servers, config, and settings.' },
              { role:'operator', desc:'Can manage servers and routes. Cannot manage users or settings.' },
              { role:'viewer', desc:'Read-only access. Can view dashboards, logs, and analytics.' },
            ]" :key="role.role"
              style="flex:1;min-width:160px;background:var(--color-surface-2);border-radius:var(--r);padding:10px 12px;"
            >
              <div :class="['badge', roleColor(role.role)]" style="font-size:10.5px;margin-bottom:5px;">{{ role.role }}</div>
              <div style="font-size:11.5px;color:var(--color-text-2);line-height:1.5;">{{ role.desc }}</div>
            </div>
          </div>

          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>User</th>
                  <th>Role</th>
                  <th>Last Login</th>
                  <th>2FA</th>
                  <th>Joined</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="user in users" :key="user.id">
                  <td>
                    <div style="display:flex;align-items:center;gap:10px;">
                      <div style="width:30px;height:30px;border-radius:50%;background:var(--color-primary-alpha);display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;color:var(--color-primary);flex-shrink:0;">
                        {{ user.name.split(' ').map(n=>n[0]).join('') }}
                      </div>
                      <div>
                        <div style="font-size:13px;font-weight:500;color:var(--color-text);">{{ user.name }}</div>
                        <div style="font-size:11.5px;color:var(--color-text-3);">{{ user.email }}</div>
                      </div>
                    </div>
                  </td>
                  <td>
                    <select class="input" :value="user.role" style="height:28px;font-size:12px;width:100px;">
                      <option value="admin">Admin</option>
                      <option value="operator">Operator</option>
                      <option value="viewer">Viewer</option>
                    </select>
                  </td>
                  <td style="font-size:12.5px;color:var(--color-text-3);">{{ timeAgo(user.lastLogin) }}</td>
                  <td>
                    <span v-if="user.twoFactorEnabled" class="badge badge-online" style="font-size:10px;">Enabled</span>
                    <span v-else class="badge badge-warning" style="font-size:10px;">Disabled</span>
                  </td>
                  <td style="font-size:12px;color:var(--color-text-3);">{{ new Date(user.createdAt).toLocaleDateString() }}</td>
                  <td>
                    <button class="btn btn-danger btn-sm btn-icon" data-tooltip="Remove user">
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/>
                      </svg>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Notifications -->
        <div v-else-if="activeSection==='notif'" class="card">
          <div class="card-header"><div class="card-title">Notification Channels</div></div>
          <div class="card-body" style="display:flex;flex-direction:column;gap:20px;">
            <!-- Email -->
            <div>
              <div style="font-size:13.5px;font-weight:600;color:var(--color-text);margin-bottom:10px;">Email</div>
              <div style="display:flex;flex-direction:column;gap:10px;">
                <label v-for="(label, key) in { emailOnCritical: 'Alert on Critical events', emailOnWarning: 'Alert on Warning events' }"
                       :key="key" style="display:flex;align-items:center;justify-content:space-between;cursor:pointer;">
                  <span style="font-size:13px;color:var(--color-text-2);">{{ label }}</span>
                  <div @click="notifSettings[key as keyof typeof notifSettings] = !notifSettings[key as keyof typeof notifSettings]"
                       style="width:34px;height:18px;border-radius:99px;position:relative;cursor:pointer;transition:background var(--t);"
                       :style="{ background: notifSettings[key as keyof typeof notifSettings] ? 'var(--color-primary)' : 'var(--color-surface-3)' }">
                    <div style="position:absolute;top:2px;width:14px;height:14px;border-radius:50%;background:#fff;transition:left var(--t);box-shadow:0 1px 3px rgba(0,0,0,.2);"
                         :style="{ left: notifSettings[key as keyof typeof notifSettings] ? '18px' : '2px' }"></div>
                  </div>
                </label>
                <div class="form-group" style="max-width:360px;">
                  <label class="form-label">Recipient email</label>
                  <input class="input" value="ops@example.com" />
                </div>
              </div>
            </div>

            <div class="divider"></div>

            <!-- Slack -->
            <div>
              <div style="font-size:13.5px;font-weight:600;color:var(--color-text);margin-bottom:10px;">Slack</div>
              <div style="display:flex;flex-direction:column;gap:10px;">
                <label v-for="(label, key) in { slackOnCritical: 'Notify on Critical events', slackOnWarning: 'Notify on Warning events' }"
                       :key="key" style="display:flex;align-items:center;justify-content:space-between;cursor:pointer;">
                  <span style="font-size:13px;color:var(--color-text-2);">{{ label }}</span>
                  <div @click="notifSettings[key as keyof typeof notifSettings] = !notifSettings[key as keyof typeof notifSettings]"
                       style="width:34px;height:18px;border-radius:99px;position:relative;cursor:pointer;transition:background var(--t);"
                       :style="{ background: notifSettings[key as keyof typeof notifSettings] ? 'var(--color-primary)' : 'var(--color-surface-3)' }">
                    <div style="position:absolute;top:2px;width:14px;height:14px;border-radius:50%;background:#fff;transition:left var(--t);box-shadow:0 1px 3px rgba(0,0,0,.2);"
                         :style="{ left: notifSettings[key as keyof typeof notifSettings] ? '18px' : '2px' }"></div>
                  </div>
                </label>
                <div class="form-group" style="max-width:360px;">
                  <label class="form-label">Webhook URL</label>
                  <input class="input" placeholder="https://hooks.slack.com/services/…" />
                </div>
              </div>
            </div>

            <div style="display:flex;justify-content:flex-end;gap:8px;padding-top:8px;border-top:1px solid var(--color-border);">
              <button class="btn btn-primary">Save Notification Settings</button>
            </div>
          </div>
        </div>

        <!-- General -->
        <div v-else-if="activeSection==='general'" class="card">
          <div class="card-header"><div class="card-title">General Settings</div></div>
          <div class="card-body" style="display:flex;flex-direction:column;gap:16px;max-width:480px;">
            <div class="form-group">
              <label class="form-label">Health Check Interval</label>
              <div style="display:flex;align-items:center;gap:8px;">
                <input class="input" type="number" v-model="generalSettings.healthCheckInterval" style="width:100px;" />
                <span style="font-size:13px;color:var(--color-text-3);">seconds</span>
              </div>
              <div class="form-hint">How often to ping each server's health endpoint</div>
            </div>
            <div class="form-group">
              <label class="form-label">Log Retention</label>
              <div style="display:flex;align-items:center;gap:8px;">
                <input class="input" type="number" v-model="generalSettings.logRetentionDays" style="width:100px;" />
                <span style="font-size:13px;color:var(--color-text-3);">days</span>
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Metrics Retention</label>
              <div style="display:flex;align-items:center;gap:8px;">
                <input class="input" type="number" v-model="generalSettings.metricsRetentionDays" style="width:100px;" />
                <span style="font-size:13px;color:var(--color-text-3);">days</span>
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Default LB Method</label>
              <select class="input" v-model="generalSettings.defaultLBMethod" style="width:220px;">
                <option value="round_robin">Round Robin</option>
                <option value="least_conn">Least Connections</option>
                <option value="ip_hash">IP Hash</option>
              </select>
            </div>
            <label style="display:flex;align-items:center;gap:10px;cursor:pointer;">
              <input type="checkbox" v-model="generalSettings.autoSSLRenew" />
              <div>
                <div style="font-size:13px;font-weight:500;color:var(--color-text);">Auto SSL Renewal</div>
                <div style="font-size:12px;color:var(--color-text-3);">Automatically trigger cert renewal 30 days before expiry</div>
              </div>
            </label>
            <div style="padding-top:8px;border-top:1px solid var(--color-border);">
              <button class="btn btn-primary">Save Settings</button>
            </div>
          </div>
        </div>

        <!-- Appearance -->
        <div v-else-if="activeSection==='theme'" class="card">
          <div class="card-header"><div class="card-title">Appearance</div></div>
          <div class="card-body">
            <div style="font-size:13px;font-weight:500;color:var(--color-text);margin-bottom:12px;">Theme</div>
            <div style="display:flex;gap:12px;">
              <div v-for="t in ['light','dark']" :key="t"
                   @click="setTheme(t as 'light'|'dark')"
                   style="cursor:pointer;border-radius:var(--r-lg);overflow:hidden;border:2px solid;transition:border-color var(--t);width:140px;"
                   :style="{ borderColor: theme===t ? 'var(--color-primary)' : 'var(--color-border)' }">
                <!-- Theme preview -->
                <div :style="{
                  height:'80px',
                  background: t==='dark'?'#141517':'#f8f8f7',
                  display:'flex',padding:'10px',gap:'8px'
                }">
                  <div :style="{ width:'40px',height:'100%',background:t==='dark'?'#1c1d21':'#fff',borderRadius:'6px' }"></div>
                  <div style="flex:1;display:flex;flex-direction:column;gap:6px;">
                    <div :style="{ height:'8px',background:t==='dark'?'#1c1d21':'#fff',borderRadius:'4px' }"></div>
                    <div :style="{ height:'8px',background:t==='dark'?'#1c1d21':'#fff',borderRadius:'4px',width:'70%' }"></div>
                    <div :style="{ height:'20px',background:'#3a9d8f',borderRadius:'4px',marginTop:'auto',opacity:.8 }"></div>
                  </div>
                </div>
                <div :style="{
                  padding:'8px 12px',
                  background: t==='dark'?'#1c1d21':'#fff',
                  display:'flex',alignItems:'center',justifyContent:'space-between'
                }">
                  <span style="font-size:12.5px;font-weight:500;color:var(--color-text);text-transform:capitalize;">{{ t }}</span>
                  <div v-if="theme===t" style="width:14px;height:14px;border-radius:50%;background:var(--color-primary);display:flex;align-items:center;justify-content:center;">
                    <svg width="8" height="8" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- API Keys -->
        <div v-else-if="activeSection==='api'" class="card">
          <div class="card-header">
            <div class="card-title">API Keys</div>
            <button class="btn btn-primary btn-sm">Generate Key</button>
          </div>
          <div class="table-wrap">
            <table>
              <thead><tr><th>Name</th><th>Key</th><th>Created</th><th>Last Used</th><th></th></tr></thead>
              <tbody>
                <tr v-for="k in [
                  { name:'CI/CD Pipeline', key:'prx_live_xxxx…4f8a', created:'2024-01-15', used:'today' },
                  { name:'Monitoring Bot', key:'prx_live_xxxx…9c3d', created:'2024-03-01', used:'2d ago' },
                ]" :key="k.name">
                  <td style="font-weight:500;color:var(--color-text);">{{ k.name }}</td>
                  <td><code style="font-family:var(--font-mono);font-size:12px;color:var(--color-text-3);">{{ k.key }}</code></td>
                  <td style="font-size:12px;color:var(--color-text-3);">{{ k.created }}</td>
                  <td style="font-size:12px;color:var(--color-text-3);">{{ k.used }}</td>
                  <td><button class="btn btn-danger btn-sm">Revoke</button></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Audit Log -->
        <div v-else-if="activeSection==='audit'" class="card">
          <div class="card-header"><div class="card-title">Audit Log</div></div>
          <div class="table-wrap">
            <table>
              <thead><tr><th>Timestamp</th><th>User</th><th>Action</th><th>Resource</th><th>IP</th></tr></thead>
              <tbody>
                <tr v-for="entry in [
                  { ts:'10:42:01', user:'Alex Johnson', action:'Update config', res:'prod-nginx-01', ip:'192.168.1.1' },
                  { ts:'10:38:22', user:'Maria Garcia', action:'Add route', res:'prod-traefik-01', ip:'192.168.1.2' },
                  { ts:'10:30:05', user:'Alex Johnson', action:'Acknowledge alert', res:'Alert #a1', ip:'192.168.1.1' },
                  { ts:'09:55:40', user:'System', action:'Health check failed', res:'dev-envoy-01', ip:'—' },
                  { ts:'09:20:18', user:'Alex Johnson', action:'Invite user', res:'james@example.com', ip:'192.168.1.1' },
                ]" :key="entry.ts">
                  <td style="font-family:var(--font-mono);font-size:12px;color:var(--color-text-3);">{{ entry.ts }}</td>
                  <td style="font-size:12.5px;color:var(--color-text);">{{ entry.user }}</td>
                  <td style="font-size:12.5px;color:var(--color-text-2);">{{ entry.action }}</td>
                  <td style="font-family:var(--font-mono);font-size:12px;color:var(--color-primary);">{{ entry.res }}</td>
                  <td style="font-family:var(--font-mono);font-size:12px;color:var(--color-text-3);">{{ entry.ip }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Invite Modal -->
    <Teleport to="body">
      <div v-if="showInviteModal" class="modal-backdrop" @click.self="showInviteModal=false">
        <div class="modal" style="max-width:400px;">
          <div class="modal-header">
            <div class="modal-title">Invite Team Member</div>
            <button class="btn btn-ghost btn-icon" @click="showInviteModal=false">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label class="form-label">Email Address</label>
              <input class="input" v-model="inviteEmail" type="email" placeholder="colleague@example.com" />
            </div>
            <div class="form-group">
              <label class="form-label">Role</label>
              <select class="input" v-model="inviteRole">
                <option value="admin">Admin</option>
                <option value="operator">Operator</option>
                <option value="viewer">Viewer</option>
              </select>
              <div class="form-hint">
                <template v-if="inviteRole==='admin'">Full access including user management and settings.</template>
                <template v-else-if="inviteRole==='operator'">Can manage servers, routes, and configs.</template>
                <template v-else>Read-only access to dashboards and logs.</template>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showInviteModal=false">Cancel</button>
            <button class="btn btn-primary" @click="showInviteModal=false">Send Invitation</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
