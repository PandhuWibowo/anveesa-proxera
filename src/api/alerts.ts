import { api } from './client'
import type { Alert } from '../types'

export interface CreateAlertPayload {
  serverId?: string
  serverName?: string
  severity: 'critical' | 'warning' | 'info'
  title: string
  message?: string
  category?: string
}

export interface UpdateAlertPayload {
  status?: 'active' | 'resolved' | 'acknowledged'
  severity?: 'critical' | 'warning' | 'info'
  title?: string
  message?: string
}

export const alertsApi = {
  list: (params?: { status?: string; severity?: string; serverId?: string }) => {
    const qs = params ? '?' + new URLSearchParams(params as Record<string, string>).toString() : ''
    return api.get<Alert[]>(`/alerts${qs}`)
  },

  get: (id: string) => api.get<Alert>(`/alerts/${id}`),

  create: (payload: CreateAlertPayload) => api.post<Alert>('/alerts', payload),

  update: (id: string, payload: UpdateAlertPayload) =>
    api.patch<Alert>(`/alerts/${id}`, payload),

  delete: (id: string) => api.delete<{ message: string }>(`/alerts/${id}`),

  acknowledge: (id: string) => api.post<Alert>(`/alerts/${id}/acknowledge`),

  resolve: (id: string) => api.post<Alert>(`/alerts/${id}/resolve`),

  bulkAcknowledge: (ids: string[]) =>
    api.post<{ acknowledged: number }>('/alerts/bulk/acknowledge', { ids }),
}
