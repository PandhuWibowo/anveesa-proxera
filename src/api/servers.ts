import { api } from './client'
import type { ProxyServer, ServerMetrics, ProxyConfig } from '../types'

export interface CreateServerPayload {
  name: string
  host: string
  port?: number
  proxyType: string
  connectionType: string
  location?: string
  description?: string
  tags?: string[]
  sshUser?: string
  sshKey?: string
  apiUrl?: string
  apiToken?: string
}

export interface UpdateServerPayload extends Partial<CreateServerPayload> {}

export interface ServerHealthResponse {
  serverId: string
  status: string
  latencyMs: number
  checkedAt: string
}

export interface ConfigValidation {
  isValid: boolean
  errors: string[]
}

export const serversApi = {
  list: (params?: { type?: string; status?: string }) => {
    const qs = params ? '?' + new URLSearchParams(params as Record<string, string>).toString() : ''
    return api.get<ProxyServer[]>(`/servers${qs}`)
  },

  get: (id: string) => api.get<ProxyServer>(`/servers/${id}`),

  create: (payload: CreateServerPayload) => api.post<ProxyServer>('/servers', payload),

  update: (id: string, payload: CreateServerPayload) =>
    api.put<ProxyServer>(`/servers/${id}`, payload),

  patch: (id: string, payload: UpdateServerPayload) =>
    api.patch<ProxyServer>(`/servers/${id}`, payload),

  delete: (id: string) => api.delete<{ message: string }>(`/servers/${id}`),

  health: (id: string) => api.get<ServerHealthResponse>(`/servers/${id}/health`),

  metrics: (id: string) => api.get<ServerMetrics>(`/servers/${id}/metrics`),

  getConfig: (id: string) => api.get<ProxyConfig>(`/servers/${id}/config`),

  putConfig: (id: string, content: string) =>
    api.put<ConfigValidation>(`/servers/${id}/config`, { content }),

  reload: (id: string) => api.post<{ message: string }>(`/servers/${id}/reload`),
}
