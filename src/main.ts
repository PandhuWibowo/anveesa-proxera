import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import './assets/styles/globals.css'

// Init theme before mount to prevent flash
const saved = localStorage.getItem('proxera_theme') || 'dark'
document.documentElement.setAttribute('data-theme', saved)

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
