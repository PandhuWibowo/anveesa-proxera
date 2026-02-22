import { api } from './client'
import type { RouteRule } from '../types'

export interface CreateRoutePayload {
  serverId: string
  name: string
  enabled?: boolean
  matchHost?: string
  matchPath?: string
  matchMethod?: string
  targetUpstream: string
  loadBalancingMethod?: string
  sslEnabled?: boolean
  sslCertExpiry?: string
  middlewares?: string[]
  priority?: number
}

export interface UpdateRoutePayload extends Partial<Omit<CreateRoutePayload, 'serverId'>> {}

export const routesApi = {
  list: (params?: { serverId?: string }) => {
    const qs = params?.serverId ? `?serverId=${params.serverId}` : ''
    return api.get<RouteRule[]>(`/routes${qs}`)
  },

  get: (id: string) => api.get<RouteRule>(`/routes/${id}`),

  create: (payload: CreateRoutePayload) => api.post<RouteRule>('/routes', payload),

  update: (id: string, payload: CreateRoutePayload) =>
    api.put<RouteRule>(`/routes/${id}`, payload),

  patch: (id: string, payload: UpdateRoutePayload) =>
    api.patch<RouteRule>(`/routes/${id}`, payload),

  delete: (id: string) => api.delete<{ message: string }>(`/routes/${id}`),

  toggle: (id: string) =>
    api.post<{ id: string; enabled: boolean }>(`/routes/${id}/toggle`),
}
