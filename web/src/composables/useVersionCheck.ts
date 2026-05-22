import { ref, onMounted, onUnmounted } from 'vue'

const CHECK_INTERVAL = 60 * 60 * 1000 // 1 小时
const INITIAL_DELAY = 30 * 1000 // 首次检查延迟 30 秒

export function useVersionCheck() {
  const hasNewVersion = ref(false)
  let timer: ReturnType<typeof setInterval> | null = null
  let initialTimer: ReturnType<typeof setTimeout> | null = null

  const check = async () => {
    try {
      // 添加时间戳避免缓存
      const res = await fetch(`/version.json?t=${Date.now()}`)
      if (!res.ok) return
      const data = await res.json()
      if (data.version && data.version !== __APP_VERSION__) {
        hasNewVersion.value = true
        stop()
      }
    } catch {
      // 版本检查为辅助功能，静默处理
    }
  }

  const stop = () => {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    if (initialTimer) {
      clearTimeout(initialTimer)
      initialTimer = null
    }
  }

  onMounted(() => {
    // 仅生产环境启用
    if (!import.meta.env.PROD) return
    initialTimer = setTimeout(() => {
      check()
      timer = setInterval(check, CHECK_INTERVAL)
    }, INITIAL_DELAY)
  })

  onUnmounted(stop)

  const reload = () => {
    window.location.reload()
  }

  return { hasNewVersion, reload }
}
