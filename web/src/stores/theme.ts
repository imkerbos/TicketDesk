import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'theme-mode'

export const useThemeStore = defineStore('theme', () => {
  // 用户选择的模式
  const mode = ref<ThemeMode>('system')
  // 系统当前偏好
  const systemPrefersDark = ref(false)

  // 实际生效的主题
  const isDark = computed(() => {
    if (mode.value === 'system') return systemPrefersDark.value
    return mode.value === 'dark'
  })

  // 应用主题到 DOM
  const applyTheme = () => {
    if (isDark.value) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  // 切换模式（循环：light → dark → system）
  const toggleMode = () => {
    const order: ThemeMode[] = ['light', 'dark', 'system']
    const currentIndex = order.indexOf(mode.value)
    mode.value = order[(currentIndex + 1) % order.length]
  }

  // 设置模式
  const setMode = (newMode: ThemeMode) => {
    mode.value = newMode
  }

  // 初始化
  const init = () => {
    // 读取 localStorage
    const stored = localStorage.getItem(STORAGE_KEY) as ThemeMode | null
    if (stored && ['light', 'dark', 'system'].includes(stored)) {
      mode.value = stored
    }

    // 监听系统偏好变化
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    systemPrefersDark.value = mediaQuery.matches
    mediaQuery.addEventListener('change', (e) => {
      systemPrefersDark.value = e.matches
    })

    // 立即应用
    applyTheme()
  }

  // 监听变化并持久化
  watch(mode, (newMode) => {
    localStorage.setItem(STORAGE_KEY, newMode)
    applyTheme()
  })

  watch(isDark, () => {
    applyTheme()
  })

  return {
    mode,
    isDark,
    toggleMode,
    setMode,
    init,
  }
})
