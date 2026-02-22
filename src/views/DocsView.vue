<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const activeSection = ref('getting-started')

const sections = [
  { id: 'getting-started',  label: 'Getting Started' },
  { id: 'architecture',     label: 'Architecture' },
  { id: 'api-servers',      label: 'API — Servers' },
  { id: 'api-routes',       label: 'API — Routes' },
  { id: 'api-alerts',       label: 'API — Alerts' },
  { id: 'api-dashboard',    label: 'API — Dashboard' },
  { id: 'websocket',        label: 'WebSocket' },
  { id: 'sse-logs',         label: 'SSE Log Stream' },
  { id: 'adapters',         label: 'Proxy Adapters' },
  { id: 'encryption',       label: 'Encryption' },
  { id: 'docker',           label: 'Docker' },
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

      <!-- ── API Servers ────────────────────────────────────────────── -->
      <section id="api-servers" class="docs-section">
        <h2 class="docs-h2">API — Servers</h2>
        <p class="docs-p">Base path: <code>/api/v1/servers</code></p>

        <div class="docs-table-wrap">
          <table class="docs-table">
            <thead><tr><th>Method</th><th>Path</th><th>Description</th></tr></thead>
            <tbody>
              <tr><td><span class="method get">GET</span></td><td><code>/servers</code></td><td>List all. Query: <code>?type=nginx</code> <code>?status=online</code></td></tr>
              <tr><td><span class="method post">POST</span></td><td><code>/servers</code></td><td>Create server</td></tr>
              <tr><td><span class="method get">GET</span></td><td><code>/servers/:id</code></td><td>Get single server</td></tr>
              <tr><td><span class="method put">PUT</span></td><td><code>/servers/:id</code></td><td>Full update</td></tr>
              <tr><td><span class="method patch">PATCH</span></td><td><code>/servers/:id</code></td><td>Partial update</td></tr>
              <tr><td><span class="method del">DELETE</span></td><td><code>/servers/:id</code></td><td>Soft delete, evicts SSH pool</td></tr>
              <tr><td><span class="method get">GET</span></td><td><code>/servers/:id/health</code></td><td>Live health check + latency</td></tr>
              <tr><td><span class="method get">GET</span></td><td><code>/servers/:id/metrics</code></td><td>Live metrics from adapter</td></tr>
              <tr><td><span class="method get">GET</span></td><td><code>/servers/:id/config</code></td><td>Fetch proxy config text</td></tr>
              <tr><td><span class="method put">PUT</span></td><td><code>/servers/:id/config</code></td><td>Write + validate config</td></tr>
              <tr><td><span class="method post">POST</span></td><td><code>/servers/:id/reload</code></td><td>Graceful reload signal</td></tr>
              <tr><td><span class="method get">GET</span></td><td><code>/servers/:id/logs</code></td><td>SSE log stream</td></tr>
            </tbody>
          </table>
        </div>

        <h3 class="docs-h3">Create Server payload</h3>
        <pre class="docs-code"><code>{
  <span class="k">"name"</span>:           <span class="s">"prod-nginx"</span>,
  <span class="k">"host"</span>:           <span class="s">"10.0.1.10"</span>,
  <span class="k">"port"</span>:           22,
  <span class="k">"proxyType"</span>:      <span class="s">"nginx"</span>,       <span class="c">// nginx | traefik | caddy | haproxy | other</span>
  <span class="k">"connectionType"</span>: <span class="s">"ssh"</span>,         <span class="c">// ssh | api</span>
  <span class="k">"sshUser"</span>:        <span class="s">"ubuntu"</span>,
  <span class="k">"sshKey"</span>:         <span class="s">"-----BEGIN..."</span>, <span class="c">// stored AES-256-GCM encrypted</span>
  <span class="k">"tags"</span>:           [<span class="s">"production"</span>, <span class="s">"us-east"</span>],
  <span class="k">"location"</span>:       <span class="s">"us-east-1"</span>,
  <span class="k">"description"</span>:    <span class="s">"Primary load balancer"</span>
}</code></pre>

        <div class="docs-callout info">
          <strong>Token masking</strong> — API responses never return raw tokens. They are masked as <code>***last4</code>.
        </div>
      </section>

      <!-- ── API Routes ─────────────────────────────────────────────── -->
      <section id="api-routes" class="docs-section">
        <h2 class="docs-h2">API — Routes</h2>
        <p class="docs-p">Base path: <code>/api/v1/routes</code></p>

        <div class="docs-table-wrap">
          <table class="docs-table">
            <thead><tr><th>Method</th><th>Path</th><th>Description</th></tr></thead>
            <tbody>
              <tr><td><span class="method get">GET</span></td><td><code>/routes</code></td><td>List. Query: <code>?serverId=</code></td></tr>
              <tr><td><span class="method post">POST</span></td><td><code>/routes</code></td><td>Create route</td></tr>
              <tr><td><span class="method get">GET</span></td><td><code>/routes/:id</code></td><td>Get single</td></tr>
              <tr><td><span class="method put">PUT</span></td><td><code>/routes/:id</code></td><td>Full update</td></tr>
              <tr><td><span class="method patch">PATCH</span></td><td><code>/routes/:id</code></td><td>Partial update</td></tr>
              <tr><td><span class="method del">DELETE</span></td><td><code>/routes/:id</code></td><td>Soft delete</td></tr>
              <tr><td><span class="method post">POST</span></td><td><code>/routes/:id/toggle</code></td><td>Flip <code>enabled</code> bool</td></tr>
            </tbody>
          </table>
        </div>

        <h3 class="docs-h3">Create Route payload</h3>
        <pre class="docs-code"><code>{
  <span class="k">"serverId"</span>:            <span class="s">"uuid"</span>,
  <span class="k">"name"</span>:               <span class="s">"api-route"</span>,
  <span class="k">"enabled"</span>:            true,
  <span class="k">"matchHost"</span>:          <span class="s">"api.example.com"</span>,
  <span class="k">"matchPath"</span>:          <span class="s">"/v1/*"</span>,
  <span class="k">"targetUpstream"</span>:     <span class="s">"http://backend:3000"</span>,
  <span class="k">"loadBalancingMethod"</span>: <span class="s">"round_robin"</span>, <span class="c">// round_robin | least_conn | ip_hash | random</span>
  <span class="k">"sslEnabled"</span>:         true,
  <span class="k">"middlewares"</span>:        [<span class="s">"rate-limit"</span>, <span class="s">"auth"</span>],
  <span class="k">"priority"</span>:           10
}</code></pre>
      </section>

      <!-- ── API Alerts ─────────────────────────────────────────────── -->
      <section id="api-alerts" class="docs-section">
        <h2 class="docs-h2">API — Alerts</h2>
        <p class="docs-p">Base path: <code>/api/v1/alerts</code></p>

        <div class="docs-table-wrap">
          <table class="docs-table">
            <thead><tr><th>Method</th><th>Path</th><th>Description</th></tr></thead>
            <tbody>
              <tr><td><span class="method get">GET</span></td><td><code>/alerts</code></td><td>List. Query: <code>?status=active</code> <code>?severity=critical</code></td></tr>
              <tr><td><span class="method post">POST</span></td><td><code>/alerts</code></td><td>Create alert (also broadcasts via WS)</td></tr>
              <tr><td><span class="method get">GET</span></td><td><code>/alerts/:id</code></td><td>Get single</td></tr>
              <tr><td><span class="method patch">PATCH</span></td><td><code>/alerts/:id</code></td><td>Update status / severity</td></tr>
              <tr><td><span class="method del">DELETE</span></td><td><code>/alerts/:id</code></td><td>Delete</td></tr>
              <tr><td><span class="method post">POST</span></td><td><code>/alerts/:id/acknowledge</code></td><td>Set status → acknowledged</td></tr>
              <tr><td><span class="method post">POST</span></td><td><code>/alerts/:id/resolve</code></td><td>Set status → resolved + resolvedAt</td></tr>
              <tr><td><span class="method post">POST</span></td><td><code>/alerts/bulk/acknowledge</code></td><td>Body: <code>{"ids":["…"]}</code></td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ── API Dashboard ──────────────────────────────────────────── -->
      <section id="api-dashboard" class="docs-section">
        <h2 class="docs-h2">API — Dashboard</h2>

        <div class="docs-table-wrap">
          <table class="docs-table">
            <thead><tr><th>Method</th><th>Path</th><th>Description</th></tr></thead>
            <tbody>
              <tr><td><span class="method get">GET</span></td><td><code>/dashboard/stats</code></td><td>Server counts, alert counts, avg metrics</td></tr>
              <tr><td><span class="method get">GET</span></td><td><code>/dashboard/traffic</code></td><td>Traffic points. Query: <code>?hours=24</code> (max 168)</td></tr>
            </tbody>
          </table>
        </div>

        <h3 class="docs-h3">Stats response</h3>
        <pre class="docs-code"><code>{
  <span class="k">"totalServers"</span>:       4,
  <span class="k">"onlineServers"</span>:      3,
  <span class="k">"offlineServers"</span>:     1,
  <span class="k">"totalRoutes"</span>:        18,
  <span class="k">"activeAlerts"</span>:       2,
  <span class="k">"totalRequestsToday"</span>: 482910,
  <span class="k">"avgErrorRate"</span>:       0.8,
  <span class="k">"avgLatency"</span>:         24.3
}</code></pre>
      </section>

      <!-- ── WebSocket ──────────────────────────────────────────────── -->
      <section id="websocket" class="docs-section">
        <h2 class="docs-h2">WebSocket</h2>
        <p class="docs-p">Connect to <code>ws://localhost:8080/ws</code> for real-time metrics and alert pushes.</p>

        <h3 class="docs-h3">Subscribe to server metrics</h3>
        <pre class="docs-code"><code><span class="c">// Send from client:</span>
{
  <span class="k">"type"</span>: <span class="s">"subscribe"</span>,
  <span class="k">"payload"</span>: {
    <span class="k">"serverIds"</span>: [<span class="s">"uuid-1"</span>, <span class="s">"uuid-2"</span>],
    <span class="k">"channel"</span>: <span class="s">"metrics"</span>
  }
}

<span class="c">// Server broadcasts:</span>
{ <span class="k">"type"</span>: <span class="s">"metrics"</span>,       <span class="k">"payload"</span>: ServerMetrics }
{ <span class="k">"type"</span>: <span class="s">"alert"</span>,         <span class="k">"payload"</span>: Alert }
{ <span class="k">"type"</span>: <span class="s">"status_change"</span>, <span class="k">"payload"</span>: { <span class="k">"serverId"</span>: <span class="s">"…"</span>, <span class="k">"status"</span>: <span class="s">"offline"</span> } }</code></pre>

        <h3 class="docs-h3">Frontend usage</h3>
        <pre class="docs-code"><code><span class="k">import</span> { ProxeraWebSocket } <span class="k">from</span> <span class="s">'@/api/ws'</span>

<span class="k">const</span> ws = <span class="k">new</span> ProxeraWebSocket()

ws.subscribe([<span class="s">'server-uuid'</span>])

<span class="k">const</span> off = ws.on(<span class="s">'metrics'</span>, (payload) =&gt; {
  console.log(payload.requestsPerSec)
})

<span class="c">// Cleanup</span>
off()
ws.destroy()</code></pre>

        <div class="docs-callout info">
          <strong>Auto-reconnect</strong> — <code>ProxeraWebSocket</code> reconnects automatically with exponential backoff (up to 10 attempts). Subscriptions are re-sent after reconnect.
        </div>
      </section>

      <!-- ── SSE Logs ───────────────────────────────────────────────── -->
      <section id="sse-logs" class="docs-section">
        <h2 class="docs-h2">SSE Log Stream</h2>
        <p class="docs-p">Real-time log streaming via Server-Sent Events.</p>

        <pre class="docs-code"><code>GET /api/v1/servers/:id/logs
Content-Type: text/event-stream

retry: 3000

id: l1
event: log
data: {"id":"l1","serverId":"uuid","level":"info","message":"GET /health 200 2ms","timestamp":"…"}

event: heartbeat
data: {"timestamp":"…"}</code></pre>

        <h3 class="docs-h3">Consuming in JavaScript</h3>
        <pre class="docs-code"><code><span class="k">const</span> es = <span class="k">new</span> EventSource(<span class="s">`/api/v1/servers/<span class="k">${id}</span>/logs`</span>)

es.addEventListener(<span class="s">'log'</span>, (e) =&gt; {
  <span class="k">const</span> entry = JSON.parse(e.data)
  console.log(entry.level, entry.message)
})

<span class="c">// Close when done</span>
es.close()</code></pre>

        <div class="docs-callout warn">
          <strong>NGINX only</strong> — Log streaming via SSH <code>tail -F</code> is supported for NGINX. Traefik, Caddy, and HAProxy adapters return <code>501 Not Supported</code> for this endpoint.
        </div>
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

/* ─── Method badges ──────────────────────────────────────────────── */
.method {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  font-family: 'JetBrains Mono', monospace;
  padding: 2px 6px;
  border-radius: 4px;
  letter-spacing: 0.03em;
}
.method.get   { background: rgba(91,199,130,0.15); color: #5bc782; }
.method.post  { background: rgba(99,169,255,0.15); color: #63a9ff; }
.method.put   { background: rgba(255,171,64,0.15); color: #ffab40; }
.method.patch { background: rgba(180,130,255,0.15); color: #b482ff; }
.method.del   { background: rgba(255,99,99,0.15);  color: #ff6363; }

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

@media (max-width: 900px) {
  .docs-toc { display: none; }
  .docs-content { padding: 24px 20px 60px; }
  .adapters-grid { grid-template-columns: 1fr; }
}
</style>
