import type { ProxyServer, RouteRule, LogEntry, Alert, TrafficPoint, User, DashboardStats } from '../types'

// ─── Mock Servers ─────────────────────────────────────────────
export const mockServers: ProxyServer[] = [
  {
    id: 's1',
    name: 'prod-nginx-01',
    host: '10.0.1.10',
    port: 80,
    proxyType: 'nginx',
    connectionType: 'ssh',
    status: 'online',
    uptime: '42d 7h 23m',
    version: '1.25.3',
    activeConnections: 1842,
    requestsPerSec: 3240,
    errorRate: 0.12,
    cpuUsage: 18,
    memUsage: 34,
    tags: ['production', 'us-east'],
    createdAt: '2024-01-15T10:00:00Z',
    lastChecked: new Date().toISOString(),
    location: 'US East (N. Virginia)',
    description: 'Primary edge NGINX serving api.example.com',
    sshUser: 'ubuntu',
    sshKeyPath: '~/.ssh/prod-key.pem',
  },
  {
    id: 's2',
    name: 'prod-traefik-01',
    host: '10.0.1.20',
    port: 8080,
    proxyType: 'traefik',
    connectionType: 'api',
    status: 'online',
    uptime: '15d 2h 11m',
    version: 'v3.1.0',
    activeConnections: 567,
    requestsPerSec: 890,
    errorRate: 0.31,
    cpuUsage: 8,
    memUsage: 22,
    tags: ['production', 'k8s'],
    createdAt: '2024-03-10T08:00:00Z',
    lastChecked: new Date().toISOString(),
    location: 'US East (N. Virginia)',
    description: 'Kubernetes ingress Traefik',
    apiUrl: 'http://10.0.1.20:8080/api',
    apiToken: 'tr_prod_xxx',
  },
  {
    id: 's3',
    name: 'staging-caddy-01',
    host: '10.0.2.10',
    port: 443,
    proxyType: 'caddy',
    connectionType: 'api',
    status: 'online',
    uptime: '3d 14h 52m',
    version: '2.7.6',
    activeConnections: 89,
    requestsPerSec: 120,
    errorRate: 0.05,
    cpuUsage: 3,
    memUsage: 12,
    tags: ['staging', 'eu-west'],
    createdAt: '2024-05-01T12:00:00Z',
    lastChecked: new Date().toISOString(),
    location: 'EU West (Ireland)',
    description: 'Staging Caddy with auto-TLS',
    apiUrl: 'http://10.0.2.10:2019/api',
  },
  {
    id: 's4',
    name: 'prod-haproxy-lb',
    host: '10.0.1.5',
    port: 9000,
    proxyType: 'haproxy',
    connectionType: 'api',
    status: 'warning',
    uptime: '7d 3h 0m',
    version: '2.8.5',
    activeConnections: 3201,
    requestsPerSec: 5800,
    errorRate: 1.82,
    cpuUsage: 67,
    memUsage: 78,
    tags: ['production', 'lb', 'us-east'],
    createdAt: '2023-11-20T00:00:00Z',
    lastChecked: new Date().toISOString(),
    location: 'US East (N. Virginia)',
    description: 'Load balancer fronting the prod NGINX pool',
    apiUrl: 'http://10.0.1.5:9000/v2',
    apiToken: 'ha_prod_xxx',
  },
  {
    id: 's5',
    name: 'dev-envoy-01',
    host: '10.0.3.10',
    port: 9901,
    proxyType: 'other',
    connectionType: 'api',
    status: 'offline',
    uptime: '0d 0h 0m',
    version: '1.29.0',
    activeConnections: 0,
    requestsPerSec: 0,
    errorRate: 0,
    cpuUsage: 0,
    memUsage: 0,
    tags: ['dev', 'service-mesh'],
    createdAt: '2024-06-01T00:00:00Z',
    lastChecked: new Date().toISOString(),
    location: 'Local Dev',
    description: 'Envoy proxy for local service mesh testing',
    apiUrl: 'http://10.0.3.10:9901',
  },
]

// ─── Mock Routes ──────────────────────────────────────────────
export const mockRoutes: RouteRule[] = [
  {
    id: 'r1', serverId: 's1', serverName: 'prod-nginx-01',
    name: 'API Gateway Route',
    enabled: true,
    matchHost: 'api.example.com',
    matchPath: '/v1/*',
    targetUpstream: 'http://backend-pool:3000',
    loadBalancingMethod: 'round_robin',
    sslEnabled: true,
    sslCertExpiry: '2025-03-15',
    middlewares: ['rate-limit', 'auth-jwt', 'cors'],
    priority: 100,
    createdAt: '2024-01-16T00:00:00Z',
  },
  {
    id: 'r2', serverId: 's1', serverName: 'prod-nginx-01',
    name: 'Static Assets',
    enabled: true,
    matchHost: 'static.example.com',
    matchPath: '/assets/*',
    targetUpstream: 'http://cdn-backend:8080',
    loadBalancingMethod: 'least_conn',
    sslEnabled: true,
    sslCertExpiry: '2025-03-15',
    middlewares: ['cache', 'gzip'],
    priority: 90,
    createdAt: '2024-01-16T01:00:00Z',
  },
  {
    id: 'r3', serverId: 's2', serverName: 'prod-traefik-01',
    name: 'Dashboard Service',
    enabled: true,
    matchHost: 'dashboard.k8s.example.com',
    targetUpstream: 'http://dashboard-svc:80',
    loadBalancingMethod: 'round_robin',
    sslEnabled: true,
    sslCertExpiry: '2025-08-20',
    middlewares: ['auth-basic', 'redirect-https'],
    priority: 80,
    createdAt: '2024-03-11T00:00:00Z',
  },
  {
    id: 'r4', serverId: 's3', serverName: 'staging-caddy-01',
    name: 'Staging App',
    enabled: true,
    matchHost: 'staging.example.com',
    matchPath: '/*',
    targetUpstream: 'http://staging-app:5173',
    loadBalancingMethod: 'round_robin',
    sslEnabled: true,
    sslCertExpiry: '2025-06-01',
    middlewares: ['cors'],
    priority: 70,
    createdAt: '2024-05-02T00:00:00Z',
  },
  {
    id: 'r5', serverId: 's4', serverName: 'prod-haproxy-lb',
    name: 'TCP Passthrough DB',
    enabled: false,
    matchHost: 'db.internal',
    targetUpstream: 'tcp://postgres-pool:5432',
    loadBalancingMethod: 'least_conn',
    sslEnabled: false,
    middlewares: [],
    priority: 60,
    createdAt: '2024-02-01T00:00:00Z',
  },
]

// ─── Mock Logs ────────────────────────────────────────────────
function randLog(): LogEntry[] {
  const msgs = [
    { level: 'info' as const, msg: 'GET /api/v1/users 200 12ms' },
    { level: 'info' as const, msg: 'POST /api/v1/auth/login 200 45ms' },
    { level: 'info' as const, msg: 'GET /api/v1/metrics 200 5ms' },
    { level: 'warn' as const, msg: 'upstream response time 1200ms exceeds threshold' },
    { level: 'error' as const, msg: 'upstream server 10.0.1.11:3001 connection refused' },
    { level: 'info' as const, msg: 'TLS handshake completed with api.example.com' },
    { level: 'warn' as const, msg: 'rate limit approached: 480/500 req/s' },
    { level: 'error' as const, msg: 'GET /api/v1/data 502 upstream timeout' },
    { level: 'info' as const, msg: 'health check /health 200 2ms' },
    { level: 'info' as const, msg: 'DELETE /api/v1/session 204 8ms' },
  ]
  const servers = [
    { id: 's1', name: 'prod-nginx-01', type: 'nginx' as const },
    { id: 's2', name: 'prod-traefik-01', type: 'traefik' as const },
    { id: 's3', name: 'staging-caddy-01', type: 'caddy' as const },
    { id: 's4', name: 'prod-haproxy-lb', type: 'haproxy' as const },
  ]
  return Array.from({ length: 80 }, (_, i) => {
    const m = msgs[i % msgs.length]!
    const s = servers[i % servers.length]!
    const d = new Date(Date.now() - i * 3200)
    return {
      id: `l${i}`,
      serverId: s.id,
      serverName: s.name,
      proxyType: s.type,
      timestamp: d.toISOString(),
      level: m.level,
      message: m.msg,
    }
  })
}
export const mockLogs: LogEntry[] = randLog()

// ─── Mock Alerts ──────────────────────────────────────────────
export const mockAlerts: Alert[] = [
  {
    id: 'a1',
    serverId: 's4', serverName: 'prod-haproxy-lb',
    severity: 'critical', status: 'active',
    title: 'High error rate detected',
    message: 'Error rate on prod-haproxy-lb has exceeded 1.5% threshold (currently 1.82%)',
    timestamp: new Date(Date.now() - 600000).toISOString(),
    category: 'performance',
  },
  {
    id: 'a2',
    serverId: 's5', serverName: 'dev-envoy-01',
    severity: 'critical', status: 'active',
    title: 'Server unreachable',
    message: 'dev-envoy-01 has been offline for more than 10 minutes. No health check response.',
    timestamp: new Date(Date.now() - 3600000).toISOString(),
    category: 'downtime',
  },
  {
    id: 'a3',
    serverId: 's1', serverName: 'prod-nginx-01',
    severity: 'warning', status: 'active',
    title: 'SSL certificate expiring soon',
    message: 'Certificate for api.example.com expires in 22 days (2025-03-15). Renewal required.',
    timestamp: new Date(Date.now() - 7200000).toISOString(),
    category: 'ssl',
  },
  {
    id: 'a4',
    serverId: 's4', serverName: 'prod-haproxy-lb',
    severity: 'warning', status: 'active',
    title: 'High CPU usage',
    message: 'CPU usage on prod-haproxy-lb is at 67%, approaching the 75% warning threshold.',
    timestamp: new Date(Date.now() - 1800000).toISOString(),
    category: 'performance',
  },
  {
    id: 'a5',
    serverId: 's2', serverName: 'prod-traefik-01',
    severity: 'info', status: 'resolved',
    title: 'Configuration reloaded',
    message: 'Traefik configuration was reloaded successfully after route update.',
    timestamp: new Date(Date.now() - 86400000).toISOString(),
    resolvedAt: new Date(Date.now() - 86000000).toISOString(),
    category: 'config',
  },
  {
    id: 'a6',
    severity: 'warning', status: 'acknowledged',
    title: 'New proxy server added without SSL',
    message: 'dev-envoy-01 was added without TLS configuration. Connections are unencrypted.',
    timestamp: new Date(Date.now() - 172800000).toISOString(),
    category: 'security',
  },
]

// ─── Mock Traffic ─────────────────────────────────────────────
export function generateTrafficData(points = 24): TrafficPoint[] {
  return Array.from({ length: points }, (_, i) => {
    const h = (new Date().getHours() - (points - 1 - i) + 24) % 24
    const base = h >= 9 && h <= 18 ? 3000 : 800
    const rnd = (min: number, max: number) => Math.floor(Math.random() * (max - min) + min)
    return {
      time: `${String(h).padStart(2, '0')}:00`,
      requests: rnd(base * 0.8, base * 1.2),
      errors: rnd(0, Math.floor(base * 0.02)),
      latency: rnd(18, 90),
    }
  })
}

// ─── Mock Users ───────────────────────────────────────────────
export const mockUsers: User[] = [
  {
    id: 'u1', name: 'Alex Johnson', email: 'alex@example.com',
    role: 'admin', lastLogin: new Date(Date.now() - 3600000).toISOString(),
    twoFactorEnabled: true, createdAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'u2', name: 'Maria Garcia', email: 'maria@example.com',
    role: 'operator', lastLogin: new Date(Date.now() - 86400000).toISOString(),
    twoFactorEnabled: true, createdAt: '2024-02-15T00:00:00Z',
  },
  {
    id: 'u3', name: 'James Lee', email: 'james@example.com',
    role: 'viewer', lastLogin: new Date(Date.now() - 604800000).toISOString(),
    twoFactorEnabled: false, createdAt: '2024-04-01T00:00:00Z',
  },
]

// ─── Dashboard Stats ──────────────────────────────────────────
export const dashboardStats: DashboardStats = {
  totalServers: 5,
  onlineServers: 3,
  offlineServers: 1,
  totalRoutes: 5,
  activeAlerts: 4,
  totalRequestsToday: 4_821_340,
  avgErrorRate: 0.46,
  avgLatency: 38,
}

// ─── Mock Configs ─────────────────────────────────────────────
export const mockNginxConfig = `# NGINX Configuration — prod-nginx-01
# Generated by Proxera

worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /run/nginx.pid;

events {
    worker_connections 4096;
    multi_accept on;
    use epoll;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" $request_time';
    access_log /var/log/nginx/access.log main;

    # Performance
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    gzip on;
    gzip_types text/plain application/json application/javascript;

    # Upstream pools
    upstream backend_pool {
        least_conn;
        server 10.0.2.1:3000;
        server 10.0.2.2:3000;
        server 10.0.2.3:3000;
        keepalive 32;
    }

    # API Gateway
    server {
        listen 443 ssl http2;
        server_name api.example.com;

        ssl_certificate     /etc/ssl/api.example.com.crt;
        ssl_certificate_key /etc/ssl/api.example.com.key;
        ssl_protocols       TLSv1.2 TLSv1.3;
        ssl_ciphers         HIGH:!aNULL:!MD5;

        location /v1/ {
            proxy_pass http://backend_pool;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_read_timeout 60s;
            proxy_connect_timeout 10s;
        }

        location /health {
            return 200 'OK';
            add_header Content-Type text/plain;
        }
    }

    # Redirect HTTP → HTTPS
    server {
        listen 80;
        server_name api.example.com;
        return 301 https://$host$request_uri;
    }
}`

export const mockTraefikConfig = `# Traefik Static Configuration — prod-traefik-01
# Generated by Proxera

global:
  checkNewVersion: false
  sendAnonymousUsage: false

log:
  level: INFO
  format: json

accessLog:
  format: json
  fields:
    defaultMode: keep
    headers:
      defaultMode: redact

api:
  dashboard: true
  insecure: false

metrics:
  prometheus:
    addEntryPointsLabels: true
    addServicesLabels: true
    buckets:
      - 0.1
      - 0.3
      - 1.2
      - 5.0

entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entryPoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"
    http:
      tls:
        certResolver: letsencrypt

certificatesResolvers:
  letsencrypt:
    acme:
      email: ops@example.com
      storage: /data/acme.json
      tlsChallenge: {}

providers:
  kubernetesIngress:
    allowExternalNameServices: true
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false`
