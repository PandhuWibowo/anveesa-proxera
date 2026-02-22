import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/dashboard', name: 'dashboard', component: DashboardView },
    { path: '/servers', name: 'servers', component: () => import('../views/ServersView.vue') },
    { path: '/monitoring', name: 'monitoring', component: () => import('../views/MonitoringView.vue') },
    { path: '/logs', name: 'logs', component: () => import('../views/LogsView.vue') },
    { path: '/config', name: 'config', component: () => import('../views/ConfigView.vue') },
    { path: '/routes', name: 'routes', component: () => import('../views/RoutesView.vue') },
    { path: '/analytics', name: 'analytics', component: () => import('../views/AnalyticsView.vue') },
    { path: '/alerts', name: 'alerts', component: () => import('../views/AlertsView.vue') },
    { path: '/settings', name: 'settings', component: () => import('../views/SettingsView.vue') },
  ],
})

export default router
