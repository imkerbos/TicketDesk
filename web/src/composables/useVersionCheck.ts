import { ref, onMounted, onUnmounted } from 'vue'

// 兜底定时（4h）+ 事件触发（visibilitychange / focus / online / WS 故障）
// 节流：同一时间窗口内多次 trigger 只发一次请求
// mismatch 时显示倒计时 banner，到时自动 reload，可点击「先不刷新」取消
const FALLBACK_INTERVAL = 4 * 60 * 60 * 1000 // 4 小时
const INITIAL_DELAY = 3 * 1000 // 启动后 3 秒
const TRIGGER_THROTTLE_MS = 30 * 1000 // 事件型最小间隔 30 秒
const AUTO_RELOAD_COUNTDOWN = 5 // 倒计时秒数

// 模块级状态（单例），多个使用方共享同一份
const hasNewVersion = ref(false)
const countdown = ref(AUTO_RELOAD_COUNTDOWN)
const dismissed = ref(false)

let lastCheckAt = 0
let fallbackTimer: ReturnType<typeof setInterval> | null = null
let initialTimer: ReturnType<typeof setTimeout> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null
let mounted = false

async function doCheck() {
  // 已检测到新版本 / 已取消 / 节流期内 → 跳过
  if (hasNewVersion.value) return
  const now = Date.now()
  if (now - lastCheckAt < TRIGGER_THROTTLE_MS) return
  lastCheckAt = now

  try {
    const res = await fetch(`/version.json?t=${now}`, { cache: 'no-store' })
    if (!res.ok) return
    const data = await res.json()
    if (data.version && data.version !== __APP_VERSION__) {
      hasNewVersion.value = true
      stopFallback()
      if (!dismissed.value) {
        startCountdown()
      }
    }
  } catch {
    // 版本检查为辅助功能，静默处理
  }
}

// 暴露给外部模块（如 WS onclose）的触发器；自动节流
export function triggerVersionCheck() {
  void doCheck()
}

function startCountdown() {
  stopCountdown()
  countdown.value = AUTO_RELOAD_COUNTDOWN
  countdownTimer = setInterval(() => {
    if (dismissed.value) {
      stopCountdown()
      return
    }
    countdown.value -= 1
    if (countdown.value <= 0) {
      stopCountdown()
      window.location.reload()
    }
  }, 1000)
}

function stopCountdown() {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

function stopFallback() {
  if (fallbackTimer) {
    clearInterval(fallbackTimer)
    fallbackTimer = null
  }
  if (initialTimer) {
    clearTimeout(initialTimer)
    initialTimer = null
  }
}

function onVisibilityChange() {
  if (document.visibilityState === 'visible') triggerVersionCheck()
}

function onFocus() {
  triggerVersionCheck()
}

function onOnline() {
  triggerVersionCheck()
}

export function useVersionCheck() {
  const reload = () => {
    window.location.reload()
  }
  const dismiss = () => {
    dismissed.value = true
    stopCountdown()
  }

  onMounted(() => {
    // 仅生产环境启用
    if (!import.meta.env.PROD) return
    if (mounted) return
    mounted = true

    initialTimer = setTimeout(() => {
      void doCheck()
      fallbackTimer = setInterval(doCheck, FALLBACK_INTERVAL)
    }, INITIAL_DELAY)

    window.addEventListener('visibilitychange', onVisibilityChange)
    window.addEventListener('focus', onFocus)
    window.addEventListener('online', onOnline)
  })

  onUnmounted(() => {
    // App 卸载（一般不会发生），清理监听
    window.removeEventListener('visibilitychange', onVisibilityChange)
    window.removeEventListener('focus', onFocus)
    window.removeEventListener('online', onOnline)
    stopFallback()
    stopCountdown()
    mounted = false
  })

  return { hasNewVersion, countdown, dismissed, reload, dismiss }
}
