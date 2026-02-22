import { ref, watchEffect } from 'vue'

const STORAGE_KEY = 'proxera_theme'
type Theme = 'light' | 'dark'

const theme = ref<Theme>((localStorage.getItem(STORAGE_KEY) as Theme) || 'dark')

watchEffect(() => {
  document.documentElement.setAttribute('data-theme', theme.value)
  localStorage.setItem(STORAGE_KEY, theme.value)
})

export function useTheme() {
  function toggleTheme() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }
  function setTheme(t: Theme) {
    theme.value = t
  }
  return { theme, toggleTheme, setTheme }
}
