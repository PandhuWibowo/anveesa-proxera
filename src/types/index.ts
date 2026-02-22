// ─── Proxy Types ─────────────────────────────────────────────
export type ProxyType = 'nginx' | 'traefik' | 'caddy' | 'haproxy' | 'other'
export type ConnectionType = 'ssh' | 'api'
export type ServerStatus = 'online' | 'offline' | 'warning' | 'unknown'
export type AlertSeverity = 'critical' | 'warning' | 'info'
export type AlertStatus = 'active' | 'resolved' | 'acknowledged'
export type UserRole = 'admin' | 'operator' | 'viewer'

// ─── Server / Proxy ───────────────────────────────────────────
export interface ProxyServer {
  id: string
  name: string
  host: string
  port: number
  proxyType: ProxyType
  connectionType: ConnectionType
  status: ServerStatus
  uptime: string
  version: string
  activeConnections: number
  requestsPerSec: number
  errorRate: number
  cpuUsage: number
  memUsage: number
  tags: string[]
  createdAt: string
  lastChecked: string
  location?: string
  description?: string
  // SSH config
  sshUser?: string
  sshKeyPath?: string
  // API config
  apiUrl?: string
  apiToken?: string
}

// ─── Route / Rule ─────────────────────────────────────────────
export interface RouteRule {
  id: string
  serverId: string
  serverName: string
  name: string
  enabled: boolean
  matchHost?: string
  matchPath?: string
  matchMethod?: string
  targetUpstream: string
  loadBalancingMethod: 'round_robin' | 'least_conn' | 'ip_hash' | 'random'
  sslEnabled: boolean
  sslCertExpiry?: string
  middlewares: string[]
  priority: number
  createdAt: string
}

// ─── Log Entry ────────────────────────────────────────────────
export interface LogEntry {
  id: string
  serverId: string
  serverName: string
  proxyType: ProxyType
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
  remoteAddr?: string
  method?: string
  path?: string
  statusCode?: number
  responseTime?: number
  bytesOut?: number
}

// ─── Alert ───────────────────────────────────────────────────
export interface Alert {
  id: string
  serverId?: string
  serverName?: string
  severity: AlertSeverity
  status: AlertStatus
  title: string
  message: string
  timestamp: string
  resolvedAt?: string
  category: 'downtime' | 'config' | 'ssl' | 'performance' | 'security'
}

// ─── Metrics ─────────────────────────────────────────────────
export interface ServerMetrics {
  serverId: string
  timestamp: string
  requestsPerSec: number
  activeConnections: number
  errorRate: number
  p50Latency: number
  p95Latency: number
  p99Latency: number
  cpuUsage: number
  memUsage: number
  networkIn: number
  networkOut: number
}

export interface TrafficPoint {
  time: string
  requests: number
  errors: number
  latency: number
}

// ─── Config ──────────────────────────────────────────────────
export interface ProxyConfig {
  serverId: string
  serverName: string
  proxyType: ProxyType
  content: string
  format: 'nginx' | 'yaml' | 'caddyfile' | 'haproxy'
  lastModified: string
  isValid: boolean
  validationErrors: string[]
}

// ─── User / Auth ──────────────────────────────────────────────
export interface User {
  id: string
  name: string
  email: string
  role: UserRole
  avatar?: string
  lastLogin: string
  twoFactorEnabled: boolean
  createdAt: string
}

// ─── Dashboard Stats ──────────────────────────────────────────
export interface DashboardStats {
  totalServers: number
  onlineServers: number
  offlineServers: number
  totalRoutes: number
  activeAlerts: number
  totalRequestsToday: number
  avgErrorRate: number
  avgLatency: number
}
