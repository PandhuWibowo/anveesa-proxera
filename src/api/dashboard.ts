import { api } from './client'
import type { DashboardStats, TrafficPoint } from '../types'

export const dashboardApi = {
  stats: () => api.get<DashboardStats>('/dashboard/stats'),

  traffic: (hours = 24) =>
    api.get<TrafficPoint[]>(`/dashboard/traffic?hours=${hours}`),
}
