<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const activeSection = ref('getting-started')

const sections = [
  { id: 'getting-started',  label: 'Getting Started' },
  { id: 'architecture',     label: 'Architecture' },
  { id: 'adapters',         label: 'Proxy Adapters' },
  { id: 'configuration',    label: 'Configuration Guide' },
  { id: 'encryption',       label: 'Encryption' },
  { id: 'docker',           label: 'Docker' },
  { id: 'troubleshooting',  label: 'Troubleshooting' },
  { id: 'changelog',        label: 'Changelog' },
]

function scrollTo(id: string) {
  activeSection.value = id
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

// Track active section on scroll
let observer: IntersectionObserver | null = null
onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      for (const e of entries) {
        if (e.isIntersecting) activeSection.value = e.target.id
      }
    },
    { rootMargin: '-20% 0px -70% 0px' },
  )
  sections.forEach(s => {
    const el = document.getElementById(s.id)
    if (el) observer!.observe(el)
  })
})
onUnmounted(() => observer?.disconnect())
</script>

<template>
  <div class="docs-layout">
    <!-- Sidebar TOC -->
    <aside class="docs-toc">
      <div class="docs-toc-title">On this page</div>
      <nav>
        <button
          v-for="s in sections"
          :key="s.id"
          class="docs-toc-item"
          :class="{ active: activeSection === s.id }"
          @click="scrollTo(s.id)"
        >
          {{ s.label }}
        </button>
      </nav>
    </aside>

    <!-- Main content -->
    <main class="docs-content">

      <!-- ── Getting Started ─────────────────────────────────────────── -->
      <section id="getting-started" class="docs-section">
        <h1 class="docs-h1">Proxera Documentation</h1>
        <p class="docs-lead">Unified reverse proxy management for NGINX, Traefik, Caddy, and HAProxy.</p>

        <h2 class="docs-h2">Getting Started</h2>

        <h3 class="docs-h3">Prerequisites</h3>
        <ul class="docs-list">
          <li>Go 1.22+</li>
          <li>Node 18+ / Bun 1.0+</li>
          <li>OpenSSL (for key generation)</li>
        </ul>

        <h3 class="docs-h3">Quick Setup</h3>
        <pre class="docs-code"><code><span class="c"># Clone and install</span>
git clone https://github.com/anveesa/proxera
cd anveesa-proxera

<span class="c"># First-time setup (generates encryption key, installs deps)</span>
make setup

<span class="c"># Start backend + frontend together</span>
make dev</code></pre>

        <h3 class="docs-h3">Environment Variables</h3>
        <div class="docs-table-wrap">
          <table class="docs-table">
            <thead>
              <tr><th>Variable</th><th>Required</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
              <tr><td><code>PROXERA_ENCRYPTION_KEY</code></td><td><span class="badge-req">required</span></td><td>—</td><td>64-char hex AES-256 key. Generate: <code>openssl rand -hex 32</code></td></tr>
              <tr><td><code>PORT</code></td><td>—</td><td><code>8080</code></td><td>Backend HTTP port</td></tr>
              <tr><td><code>DATABASE_PATH</code></td><td>—</td><td><code>./data/proxera.db</code></td><td>SQLite database file path</td></tr>
              <tr><td><code>ALLOW_ORIGINS</code></td><td>—</td><td><code>http://localhost:5173</code></td><td>Comma-separated CORS allowed origins</td></tr>
              <tr><td><code>ENVIRONMENT</code></td><td>—</td><td><code>development</code></td><td><code>development</code> or <code>production</code></td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ── Architecture ───────────────────────────────────────────── -->
      <section id="architecture" class="docs-section">
        <h2 class="docs-h2">Architecture</h2>
        <p class="docs-p">Proxera is a two-tier application: a Go backend exposing a REST + WebSocket API, and a Vue 3 SPA frontend.</p>

        <pre class="docs-code"><code>Browser
  │
  ├── GET/POST /api/v1/*   →  Gin HTTP router
  ├── WS  /ws              →  WebSocket hub (real-time metrics + alerts)
  └── GET /api/v1/servers/:id/logs  →  SSE log stream

Gin router
  ├── handlers/servers.go  →  proxy/manager.go  →  ProxyAdapter
  ├── handlers/routes.go   →  SQLite (GORM)
  ├── handlers/alerts.go   →  SQLite (GORM) + WS broadcast
  ├── handlers/dashboard.go→  SQLite aggregates
  └── handlers/ws.go       →  WSHub goroutine

ProxyAdapter implementations:
  nginx.go   → SSH  (golang.org/x/crypto/ssh)
  traefik.go → REST (net/http)
  caddy.go   → REST (net/http, Admin API)
  haproxy.go → REST (Stats CSV + Data Plane API)</code></pre>

        <h3 class="docs-h3">Data flow</h3>
        <ul class="docs-list">
          <li>All server credentials (SSH keys, API tokens) are AES-256-GCM encrypted before writing to SQLite.</li>
          <li>SSH connections are pooled per server with a 10-minute idle eviction and 30-second keepalive.</li>
          <li>The WebSocket hub maintains per-client subscription maps — clients only receive metrics for servers they subscribe to.</li>
        </ul>
      </section>

      <!-- ── Proxy Adapters ─────────────────────────────────────────── -->
      <section id="adapters" class="docs-section">
        <h2 class="docs-h2">Proxy Adapters</h2>

        <div class="adapters-grid">
          <div class="adapter-card">
            <div class="adapter-header">
              <span class="adapter-dot" style="background:#009639;"></span>
              <strong>NGINX</strong>
              <code class="adapter-conn">SSH</code>
            </div>
            <ul class="docs-list small">
              <li>Metrics via <code>/nginx_status</code> stub_status</li>
              <li>Config read from <code>/etc/nginx/nginx.conf</code></li>
              <li>Config write + <code>nginx -t</code> validation</li>
              <li>Reload via <code>nginx -s reload</code></li>
              <li>Log streaming via <code>tail -F</code></li>
            </ul>
          </div>

          <div class="adapter-card">
            <div class="adapter-header">
              <span class="adapter-dot" style="background:#24a1c1;"></span>
              <strong>Traefik</strong>
              <code class="adapter-conn">REST</code>
            </div>
            <ul class="docs-list small">
              <li>Health via <code>GET /ping</code></li>
              <li>Metrics via <code>GET /api/overview</code></li>
              <li>Config read via <code>GET /api/rawdata</code></li>
              <li>Config write: not supported</li>
              <li>Log streaming: not supported</li>
            </ul>
          </div>

          <div class="adapter-card">
            <div class="adapter-header">
              <span class="adapter-dot" style="background:#00adef;"></span>
              <strong>Caddy</strong>
              <code class="adapter-conn">REST</code>
            </div>
            <ul class="docs-list small">
              <li>Health via <code>GET /config/</code></li>
              <li>Config read/write via Admin API</li>
              <li>Apply config via <code>POST /load</code></li>
              <li>JSON validation before applying</li>
              <li>Log streaming: not supported</li>
            </ul>
          </div>

          <div class="adapter-card">
            <div class="adapter-header">
              <span class="adapter-dot" style="background:#e2001a;"></span>
              <strong>HAProxy</strong>
              <code class="adapter-conn">REST</code>
            </div>
            <ul class="docs-list small">
              <li>Metrics via CSV stats endpoint</li>
              <li>Config via Data Plane API</li>
              <li>Parses <code>scur</code> and <code>req_tot</code> columns</li>
              <li>Config write: not supported</li>
              <li>Log streaming: not supported</li>
            </ul>
          </div>
        </div>
      </section>

      <!-- ── Configuration Guide ───────────────────────────────────── -->
      <section id="configuration" class="docs-section">
        <h2 class="docs-h2">Configuration Guide</h2>
        <p class="docs-p">This guide walks through configuring Proxera from first launch — adding servers, defining routes, tuning alerts, and adjusting platform settings.</p>

        <h3 class="docs-h3">1. Adding a Proxy Server</h3>
        <p class="docs-p">Navigate to <strong>Servers → Add Server</strong>. Fill in the connection details:</p>
        <ul class="docs-list">
          <li><strong>Server Name</strong> — a human-readable label (e.g. <code>prod-nginx-01</code>)</li>
          <li><strong>Host / IP</strong> — reachable address from the Proxera backend</li>
          <li><strong>Port</strong> — SSH port (default <code>22</code>) or API port</li>
          <li><strong>Proxy Type</strong> — NGINX, Traefik, Caddy, HAProxy, or Other</li>
          <li><strong>Connection Type</strong> — <code>ssh</code> or <code>api</code></li>
        </ul>
        <div class="docs-callout info">
          <strong>SSH connection</strong> — paste the private key PEM content directly into the SSH Key field. Proxera stores it AES-256-GCM encrypted; the plaintext is never written to disk.
        </div>

        <h3 class="docs-h3">2. Managing Routes</h3>
        <p class="docs-p">Routes define how traffic is forwarded from a domain/path to an upstream service. Go to <strong>Routes → New Route</strong>:</p>
        <ul class="docs-list">
          <li><strong>Match Host</strong> — domain the rule applies to (e.g. <code>api.example.com</code>)</li>
          <li><strong>Match Path</strong> — path prefix or glob (e.g. <code>/v1/*</code>). Leave blank to match all paths.</li>
          <li><strong>Target Upstream</strong> — backend URL (e.g. <code>http://10.0.1.5:3000</code>)</li>
          <li><strong>Load Balancing</strong> — <code>round_robin</code>, <code>least_conn</code>, <code>ip_hash</code>, or <code>random</code></li>
          <li><strong>SSL</strong> — enable to terminate TLS at the proxy</li>
          <li><strong>Priority</strong> — lower number = higher precedence when multiple routes match</li>
        </ul>
        <div class="docs-callout warn">
          <strong>Route propagation</strong> — after saving, the route is persisted in SQLite. To apply it to the live proxy config, use <strong>Configuration → Push Config</strong> for the target server.
        </div>

        <h3 class="docs-h3">3. Alert Rules</h3>
        <p class="docs-p">Alerts are triggered automatically by the monitoring engine. You can also define custom thresholds under <strong>Alerts → New Alert Rule</strong>:</p>
        <ul class="docs-list">
          <li><strong>Metric</strong> — CPU usage, memory, error rate, or latency</li>
          <li><strong>Threshold</strong> — numeric value that triggers the alert</li>
          <li><strong>Severity</strong> — <code>info</code>, <code>warning</code>, or <code>critical</code></li>
          <li><strong>Cooldown</strong> — minimum minutes between repeated alerts for the same server</li>
        </ul>

        <h3 class="docs-h3">4. Platform Settings</h3>
        <p class="docs-p">Under <strong>Settings</strong> you can configure:</p>
        <div class="docs-table-wrap">
          <table class="docs-table">
            <thead><tr><th>Setting</th><th>Description</th></tr></thead>
            <tbody>
              <tr><td>Theme</td><td>Light / dark mode preference, persisted in localStorage</td></tr>
              <tr><td>Metrics Interval</td><td>How often the dashboard polls for live metrics (5 s – 60 s)</td></tr>
              <tr><td>SSH Pool Timeout</td><td>Idle eviction time for pooled SSH connections (default 10 min)</td></tr>
              <tr><td>Notification Channel</td><td>Webhook URL for outbound alert notifications (Slack, Teams, etc.)</td></tr>
              <tr><td>Retention Period</td><td>How many days of query logs and alert history to keep in SQLite</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ── Encryption ─────────────────────────────────────────────── -->
      <section id="encryption" class="docs-section">
        <h2 class="docs-h2">Encryption</h2>
        <p class="docs-p">Sensitive credentials are encrypted with AES-256-GCM before being written to SQLite.</p>

        <h3 class="docs-h3">Encrypted fields</h3>
        <ul class="docs-list">
          <li><code>ssh_key_enc</code> — SSH private key PEM content</li>
          <li><code>api_token_enc</code> — API token for REST-based adapters</li>
        </ul>

        <h3 class="docs-h3">Key generation</h3>
        <pre class="docs-code"><code><span class="c"># Generate a 32-byte (256-bit) key as a 64-char hex string</span>
openssl rand -hex 32

<span class="c"># Set in server/.env</span>
PROXERA_ENCRYPTION_KEY=903e03cc...</code></pre>

        <div class="docs-callout warn">
          <strong>Key rotation</strong> — Changing the encryption key will make all existing encrypted credentials unreadable. Re-enter all server credentials after rotating the key.
        </div>
      </section>

      <!-- ── Docker ─────────────────────────────────────────────────── -->
      <section id="docker" class="docs-section">
        <h2 class="docs-h2">Docker</h2>

        <h3 class="docs-h3">Production deployment</h3>
        <pre class="docs-code"><code><span class="c"># Build frontend</span>
bun run build

<span class="c"># Start full stack (backend + nginx frontend)</span>
docker-compose up --build

<span class="c"># Or via Makefile</span>
make docker-up</code></pre>

        <h3 class="docs-h3">Services</h3>
        <div class="docs-table-wrap">
          <table class="docs-table">
            <thead><tr><th>Service</th><th>Image</th><th>Port</th><th>Notes</th></tr></thead>
            <tbody>
              <tr><td><code>backend</code></td><td><code>distroless/static-debian12</code></td><td>8080 (internal)</td><td>Pure-Go binary, ~8MB, SQLite on <code>/data</code> volume</td></tr>
              <tr><td><code>frontend</code></td><td><code>nginx:alpine</code></td><td>80</td><td>Serves Vite dist, proxies <code>/api</code> and <code>/ws</code> to backend</td></tr>
            </tbody>
          </table>
        </div>

        <h3 class="docs-h3">Dockerfile stages</h3>
        <pre class="docs-code"><code><span class="c"># Stage 1 — Build (golang:1.22-alpine)</span>
CGO_ENABLED=0 go build -ldflags="-w -s" -o proxera .

<span class="c"># Stage 2 — Runtime (distroless/static-debian12:nonroot)</span>
<span class="c"># Final image: ~8MB, no shell, non-root user</span></code></pre>
      </section>

      <!-- ── Troubleshooting ────────────────────────────────────────── -->
      <section id="troubleshooting" class="docs-section">
        <h2 class="docs-h2">Troubleshooting</h2>

        <h3 class="docs-h3">SSH connection refused</h3>
        <ul class="docs-list">
          <li>Confirm the backend container/process can reach the target host on the SSH port.</li>
          <li>Verify the SSH user has permission to read <code>/etc/nginx/nginx.conf</code> and run <code>nginx -s reload</code>.</li>
          <li>Check that the private key PEM was pasted without leading/trailing whitespace.</li>
          <li>If using a passphrase-protected key, remove the passphrase first — Proxera does not support encrypted keys.</li>
        </ul>
        <pre class="docs-code"><code><span class="c"># Test SSH access from the machine running Proxera</span>
ssh -i /tmp/test_key ubuntu@10.0.1.10 "nginx -v"</code></pre>

        <h3 class="docs-h3">NGINX adapter returns stale metrics</h3>
        <ul class="docs-list">
          <li>Ensure the <code>stub_status</code> module is enabled and the endpoint is accessible from localhost.</li>
          <li>Add to your <code>nginx.conf</code>:</li>
        </ul>
        <pre class="docs-code"><code>server {
  listen 127.0.0.1:8888;
  location /nginx_status {
    stub_status;
    allow 127.0.0.1;
    deny all;
  }
}</code></pre>

        <h3 class="docs-h3">Config validation fails on push</h3>
        <ul class="docs-list">
          <li>Proxera runs <code>nginx -t</code> before applying. The raw error output is returned in the API response — check the Configuration page for the inline error message.</li>
          <li>Common causes: unclosed <code>{ }</code> blocks, missing semicolons, or referencing an upstream that doesn't exist.</li>
        </ul>
        <div class="docs-callout warn">
          <strong>No partial rollback</strong> — if a config push fails validation, the existing live config is left untouched. Fix the error in the editor and retry.
        </div>

        <h3 class="docs-h3">Traefik / Caddy shows "not supported"</h3>
        <p class="docs-p">Not all adapters support every feature. See the <a class="docs-link" href="#adapters">Proxy Adapters</a> section for a capability matrix. Features marked "not supported" are intentionally disabled to avoid inconsistent behaviour across adapter types.</p>

        <h3 class="docs-h3">Database locked error on startup</h3>
        <ul class="docs-list">
          <li>SQLite allows only one writer at a time. Make sure no other Proxera process is running against the same <code>.db</code> file.</li>
          <li>If running in Docker, ensure the <code>/data</code> volume is not bind-mounted to a network filesystem (NFS/CIFS) — use a local volume instead.</li>
        </ul>

        <h3 class="docs-h3">Frontend shows blank page after upgrade</h3>
        <ul class="docs-list">
          <li>Clear the browser cache or do a hard reload (<code>Ctrl + Shift + R</code>).</li>
          <li>If running behind a reverse proxy, ensure it is not caching the <code>index.html</code> response — set <code>Cache-Control: no-store</code> for HTML files.</li>
        </ul>
      </section>

      <!-- ── Changelog ──────────────────────────────────────────────── -->
      <section id="changelog" class="docs-section">
        <h2 class="docs-h2">Changelog</h2>

        <div class="changelog-entry">
          <div class="changelog-header">
            <span class="changelog-version">v0.3.0</span>
            <span class="changelog-date">2025-06-10</span>
            <span class="changelog-tag added">minor</span>
          </div>
          <ul class="docs-list">
            <li>Added Analytics view with traffic trend charts and error-rate breakdowns</li>
            <li>Route Builder now supports <code>ip_hash</code> and <code>random</code> load balancing methods</li>
            <li>Alert bulk-acknowledge action added to Alerts page</li>
            <li>Dark / light theme toggle persisted per-user in localStorage</li>
          </ul>
        </div>

        <div class="changelog-entry">
          <div class="changelog-header">
            <span class="changelog-version">v0.2.0</span>
            <span class="changelog-date">2025-04-22</span>
            <span class="changelog-tag added">minor</span>
          </div>
          <ul class="docs-list">
            <li>Caddy adapter: config read/write via Admin API (<code>POST /load</code>)</li>
            <li>HAProxy adapter: CSV stats parsing for <code>scur</code> and <code>req_tot</code></li>
            <li>Real-time monitoring view with 2-second live sparkline charts</li>
            <li>SSH connection pool with 10-minute idle eviction and 30-second keepalive</li>
            <li>AES-256-GCM encryption for all stored credentials</li>
          </ul>
        </div>

        <div class="changelog-entry">
          <div class="changelog-header">
            <span class="changelog-version">v0.1.0</span>
            <span class="changelog-date">2025-02-14</span>
            <span class="changelog-tag">initial</span>
          </div>
          <ul class="docs-list">
            <li>Initial release — NGINX (SSH) and Traefik (REST) adapters</li>
            <li>Server and route CRUD with SQLite persistence via GORM</li>
            <li>Dashboard with request volume, latency, and alert summary</li>
            <li>Go + Vue 3 + Vite project scaffold with Docker multi-stage build</li>
          </ul>
        </div>
      </section>

    </main>
  </div>
</template>

<style scoped>
/* ─── Layout ──────────────────────────────────────────────────────── */
.docs-layout {
  display: flex;
  gap: 0;
  min-height: 100%;
  position: relative;
}

.docs-toc {
  width: 200px;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  height: calc(100vh - 57px);
  overflow-y: auto;
  padding: 24px 0 24px 4px;
  border-right: 1px solid var(--color-border);
}

.docs-toc-title {
  font-size: 10.5px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--color-text-3);
  padding: 0 12px 10px;
}

.docs-toc-item {
  display: block;
  width: 100%;
  text-align: left;
  background: none;
  border: none;
  padding: 5px 12px;
  font-size: 12.5px;
  color: var(--color-text-2);
  cursor: pointer;
  border-radius: 5px;
  transition: color 0.15s, background 0.15s;
  line-height: 1.4;
}
.docs-toc-item:hover {
  color: var(--color-text);
  background: var(--color-hover);
}
.docs-toc-item.active {
  color: var(--color-primary);
  font-weight: 500;
  background: var(--color-primary-alpha);
}

.docs-content {
  flex: 1;
  min-width: 0;
  padding: 32px 40px 80px;
  max-width: 820px;
}

/* ─── Typography ─────────────────────────────────────────────────── */
.docs-section {
  margin-bottom: 56px;
  scroll-margin-top: 24px;
}

.docs-h1 {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text);
  letter-spacing: -0.02em;
  margin: 0 0 8px;
  line-height: 1.2;
}

.docs-lead {
  font-size: 14.5px;
  color: var(--color-text-2);
  margin: 0 0 28px;
  line-height: 1.6;
}

.docs-h2 {
  font-size: 17px;
  font-weight: 650;
  color: var(--color-text);
  letter-spacing: -0.01em;
  margin: 0 0 14px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--color-border);
}

.docs-h3 {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
  margin: 20px 0 8px;
  letter-spacing: -0.005em;
}

.docs-p {
  font-size: 13.5px;
  color: var(--color-text-2);
  line-height: 1.7;
  margin: 0 0 14px;
}

.docs-list {
  font-size: 13px;
  color: var(--color-text-2);
  line-height: 1.7;
  padding-left: 20px;
  margin: 0 0 14px;
}
.docs-list.small {
  font-size: 12px;
  margin: 0;
}
.docs-list li {
  margin-bottom: 3px;
}

code {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
  background: var(--color-hover);
  color: var(--color-primary);
  padding: 1px 5px;
  border-radius: 4px;
  border: 1px solid var(--color-border);
}

/* ─── Code blocks ────────────────────────────────────────────────── */
.docs-code {
  background: var(--color-surface-2, var(--color-hover));
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px 18px;
  overflow-x: auto;
  margin: 10px 0 18px;
  font-size: 12px;
  line-height: 1.65;
}
.docs-code code {
  background: none;
  border: none;
  padding: 0;
  color: var(--color-text);
  font-size: inherit;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}
.docs-code .c  { color: var(--color-text-3); font-style: italic; }
.docs-code .k  { color: var(--color-primary); }
.docs-code .s  { color: #7ec894; }

/* ─── Tables ─────────────────────────────────────────────────────── */
.docs-table-wrap {
  overflow-x: auto;
  margin-bottom: 18px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
}

.docs-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12.5px;
}
.docs-table thead tr {
  background: var(--color-hover);
}
.docs-table th {
  padding: 8px 14px;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-2);
  border-bottom: 1px solid var(--color-border);
  white-space: nowrap;
}
.docs-table td {
  padding: 8px 14px;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text);
  vertical-align: top;
  line-height: 1.5;
}
.docs-table tr:last-child td {
  border-bottom: none;
}
.docs-table tr:hover td {
  background: var(--color-hover);
}

/* ─── Callouts ───────────────────────────────────────────────────── */
.docs-callout {
  padding: 10px 14px;
  border-radius: 7px;
  font-size: 12.5px;
  line-height: 1.6;
  margin: 10px 0 18px;
  border-left: 3px solid;
}
.docs-callout.info {
  background: var(--color-primary-alpha);
  border-color: var(--color-primary);
  color: var(--color-text-2);
}
.docs-callout.warn {
  background: rgba(255,171,64,0.08);
  border-color: #ffab40;
  color: var(--color-text-2);
}
.docs-callout strong {
  color: var(--color-text);
}

/* ─── Adapter cards ──────────────────────────────────────────────── */
.adapters-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 18px;
}

.adapter-card {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 14px;
  background: var(--color-surface);
}

.adapter-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 13px;
  color: var(--color-text);
}

.adapter-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.adapter-conn {
  margin-left: auto;
  font-size: 10px;
  font-weight: 600;
  background: var(--color-hover);
  border: 1px solid var(--color-border);
  color: var(--color-text-2);
  padding: 1px 6px;
  border-radius: 4px;
}

/* ─── Inline link ─────────────────────────────────────────────────── */
.docs-link {
  color: var(--color-primary);
  text-decoration: none;
  font-size: inherit;
}
.docs-link:hover { text-decoration: underline; }

/* ─── Changelog ──────────────────────────────────────────────────── */
.changelog-entry {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 14px 16px;
  margin-bottom: 14px;
  background: var(--color-surface);
}

.changelog-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.changelog-version {
  font-size: 13.5px;
  font-weight: 700;
  color: var(--color-text);
  font-family: 'JetBrains Mono', monospace;
  letter-spacing: -0.02em;
}

.changelog-date {
  font-size: 12px;
  color: var(--color-text-3);
}

.changelog-tag {
  margin-left: auto;
  font-size: 10px;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 99px;
  background: var(--color-surface-2);
  color: var(--color-text-3);
  border: 1px solid var(--color-border);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.changelog-tag.added {
  background: var(--color-primary-alpha);
  color: var(--color-primary);
  border-color: var(--color-primary-alpha2);
}

/* ─── Required badge ─────────────────────────────────────────────── */
.badge-req {
  display: inline-block;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(255,99,99,0.12);
  color: #ff6363;
}

@media (max-width: 900px) {
  .docs-toc { display: none; }
  .docs-content { padding: 24px 20px 60px; }
  .adapters-grid { grid-template-columns: 1fr; }
}
</style>
