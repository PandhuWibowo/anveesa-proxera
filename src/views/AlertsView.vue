<script setup lang="ts">
import { ref, computed } from 'vue'
import { mockAlerts } from '../composables/useMockData'
import type { Alert, AlertSeverity, AlertStatus } from '../types'

const alerts = ref<Alert[]>([...mockAlerts])
const filterStatus  = ref<'all'|AlertStatus>('all')
const filterSeverity = ref<'all'|AlertSeverity>('all')
const filterCategory = ref('all')

const filtered = computed(() => alerts.value.filter(a => {
  const matchStatus   = filterStatus.value === 'all' || a.status === filterStatus.value
  const matchSeverity = filterSeverity.value === 'all' || a.severity === filterSeverity.value
  const matchCategory = filterCategory.value === 'all' || a.category === filterCategory.value
  return matchStatus && matchSeverity && matchCategory
}))

const counts = computed(() => ({
  active: alerts.value.filter(a=>a.status==='active').length,
  acknowledged: alerts.value.filter(a=>a.status==='acknowledged').length,
  resolved: alerts.value.filter(a=>a.status==='resolved').length,
}))

function ackAlert(id: string) {
  const a = alerts.value.find(a => a.id === id)
  if (a) a.status = 'acknowledged'
}
function resolveAlert(id: string) {
  const a = alerts.value.find(a => a.id === id)
  if (a) { a.status = 'resolved'; a.resolvedAt = new Date().toISOString() }
}

function severityBadge(s: AlertSeverity) {
  return { critical: 'badge-offline', warning: 'badge-warning', info: 'badge-info' }[s]
}

function timeAgo(iso: string) {
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  if (m < 1440) return `${Math.floor(m/60)}h ago`
  return `${Math.floor(m/1440)}d ago`
}

function categoryIcon(cat: string) {
  const icons: Record<string, string> = {
    downtime:    `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>`,
    performance: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>`,
    ssl:         `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>`,
    config:      `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>`,
    security:    `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>`,
  }
  return icons[cat] || icons.config
}

const categories = ['all', 'downtime', 'performance', 'ssl', 'config', 'security']
</script>

<template>
  <div class="page animate-slide">
    <div class="page-header">
      <div>
        <div class="page-title">Alerts</div>
        <div class="page-description">Monitor critical events across all proxy servers</div>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-secondary" @click="alerts.forEach(a=>{if(a.status==='active')a.status='acknowledged'})">
          Acknowledge All
        </button>
      </div>
    </div>

    <!-- Summary -->
    <div class="grid-3">
      <div class="metric-card" style="border-color:var(--color-error-alpha);">
        <div class="metric-label" style="color:var(--color-error);">Active</div>
        <div class="metric-value" style="color:var(--color-error);">{{ counts.active }}</div>
        <div class="metric-sub">Require immediate attention</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Acknowledged</div>
        <div class="metric-value">{{ counts.acknowledged }}</div>
        <div class="metric-sub">Being investigated</div>
      </div>
      <div class="metric-card" style="border-color:var(--color-success-alpha);">
        <div class="metric-label" style="color:var(--color-success);">Resolved</div>
        <div class="metric-value" style="color:var(--color-success);">{{ counts.resolved }}</div>
        <div class="metric-sub">In the last 7 days</div>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex gap-3 items-center flex-wrap">
      <div class="tab-list">
        <div class="tab-item" :class="{ active: filterStatus==='all' }" @click="filterStatus='all'">All</div>
        <div class="tab-item" :class="{ active: filterStatus==='active' }" @click="filterStatus='active'" style="color:var(--color-error)">Active</div>
        <div class="tab-item" :class="{ active: filterStatus==='acknowledged' }" @click="filterStatus='acknowledged'">Acknowledged</div>
        <div class="tab-item" :class="{ active: filterStatus==='resolved' }" @click="filterStatus='resolved'" style="color:var(--color-success)">Resolved</div>
      </div>
      <div class="tab-list">
        <div class="tab-item" :class="{ active: filterSeverity==='all' }" @click="filterSeverity='all'">Any</div>
        <div class="tab-item" :class="{ active: filterSeverity==='critical' }" @click="filterSeverity='critical'" style="color:var(--color-error)">Critical</div>
        <div class="tab-item" :class="{ active: filterSeverity==='warning' }" @click="filterSeverity='warning'" style="color:var(--color-warning)">Warning</div>
        <div class="tab-item" :class="{ active: filterSeverity==='info' }" @click="filterSeverity='info'" style="color:var(--color-info)">Info</div>
      </div>
      <select class="input" v-model="filterCategory" style="height:32px;font-size:12.5px;width:auto;">
        <option v-for="c in categories" :key="c" :value="c">{{ c === 'all' ? 'All categories' : c }}</option>
      </select>
    </div>

    <!-- Alert List -->
    <div class="card" style="overflow:hidden;">
      <div v-for="alert in filtered" :key="alert.id"
           style="padding:16px 18px;border-bottom:1px solid var(--color-border);display:flex;align-items:flex-start;gap:14px;"
           :style="{ opacity: alert.status==='resolved' ? 0.65 : 1 }">
        <!-- Icon -->
        <div style="width:34px;height:34px;border-radius:var(--r);display:flex;align-items:center;justify-content:center;flex-shrink:0;"
             :style="{
               background: alert.severity==='critical'?'var(--color-error-alpha)':alert.severity==='warning'?'var(--color-warning-alpha)':'var(--color-info-alpha)',
               color: alert.severity==='critical'?'var(--color-error)':alert.severity==='warning'?'var(--color-warning)':'var(--color-info)',
             }">
          <span style="width:16px;height:16px;" v-html="categoryIcon(alert.category)"></span>
        </div>

        <!-- Content -->
        <div style="flex:1;min-width:0;">
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:3px;">
            <div :class="['badge', severityBadge(alert.severity)]" style="font-size:10.5px;">{{ alert.severity }}</div>
            <div style="font-size:12px;color:var(--color-text-3);">{{ alert.category }}</div>
            <div v-if="alert.status==='acknowledged'" class="badge badge-warning" style="font-size:10px;">acknowledged</div>
            <div v-if="alert.status==='resolved'" class="badge badge-online" style="font-size:10px;">resolved</div>
          </div>
          <div style="font-size:13.5px;font-weight:600;color:var(--color-text);line-height:1.3;">{{ alert.title }}</div>
          <div style="font-size:12.5px;color:var(--color-text-2);margin-top:4px;line-height:1.5;">{{ alert.message }}</div>
          <div style="display:flex;gap:12px;margin-top:6px;">
            <span style="font-size:11.5px;color:var(--color-text-3);">{{ alert.serverName || 'System' }}</span>
            <span style="font-size:11.5px;color:var(--color-text-3);">{{ timeAgo(alert.timestamp) }}</span>
            <span v-if="alert.resolvedAt" style="font-size:11.5px;color:var(--color-success);">
              Resolved {{ timeAgo(alert.resolvedAt) }}
            </span>
          </div>
        </div>

        <!-- Actions -->
        <div v-if="alert.status !== 'resolved'" style="display:flex;gap:6px;flex-shrink:0;">
          <button v-if="alert.status==='active'" class="btn btn-secondary btn-sm" @click="ackAlert(alert.id)">
            Acknowledge
          </button>
          <button class="btn btn-ghost btn-sm" @click="resolveAlert(alert.id)">
            Resolve
          </button>
        </div>
      </div>

      <div v-if="filtered.length === 0" class="empty-state">
        <div class="empty-state-icon">✅</div>
        <div class="empty-state-title">No alerts</div>
        <div class="empty-state-desc">All systems are operating normally</div>
      </div>
    </div>

    <!-- Alert Rules -->
    <div class="card">
      <div class="card-header">
        <div class="card-title">Alert Rules</div>
        <button class="btn btn-primary btn-sm">
          <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          New Rule
        </button>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Rule Name</th>
              <th>Condition</th>
              <th>Threshold</th>
              <th>Severity</th>
              <th>Notify</th>
              <th>Enabled</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in [
              { name: 'Server Down',        cond: 'Health check failed',        thresh: '3 consecutive',    sev: 'critical', notify: 'Email, Slack' },
              { name: 'High Error Rate',    cond: 'Error rate exceeds threshold', thresh: '> 1.5%',          sev: 'critical', notify: 'Email, PagerDuty' },
              { name: 'SSL Expiry Warning', cond: 'Certificate expires soon',    thresh: '< 30 days',       sev: 'warning',  notify: 'Email' },
              { name: 'High CPU',           cond: 'CPU usage',                   thresh: '> 75%',           sev: 'warning',  notify: 'Slack' },
              { name: 'High Memory',        cond: 'Memory usage',                thresh: '> 85%',           sev: 'warning',  notify: 'Slack' },
              { name: 'Config Changed',     cond: 'Configuration file modified', thresh: 'Any change',      sev: 'info',     notify: 'Email' },
            ]" :key="rule.name">
              <td style="font-weight:500;color:var(--color-text);">{{ rule.name }}</td>
              <td style="font-size:12.5px;color:var(--color-text-2);">{{ rule.cond }}</td>
              <td style="font-family:var(--font-mono);font-size:12px;color:var(--color-text-2);">{{ rule.thresh }}</td>
              <td>
                <span :class="['badge', rule.sev==='critical'?'badge-offline':rule.sev==='warning'?'badge-warning':'badge-info']" style="font-size:10.5px;">{{ rule.sev }}</span>
              </td>
              <td style="font-size:12px;color:var(--color-text-3);">{{ rule.notify }}</td>
              <td>
                <div style="width:34px;height:18px;border-radius:99px;background:var(--color-primary);position:relative;cursor:pointer;">
                  <div style="position:absolute;top:2px;left:18px;width:14px;height:14px;border-radius:50%;background:#fff;box-shadow:0 1px 3px rgba(0,0,0,.2);"></div>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
