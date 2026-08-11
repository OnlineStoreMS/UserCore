import { loadAuth } from './token'
import { isAuthExpired, shouldRefreshSoon, tryRefreshAccessToken } from './authSession'

const RENEW_BEFORE_MS = 5 * 60 * 1000
const MIN_DELAY_MS = 15 * 1000

let timer: ReturnType<typeof setTimeout> | null = null
let started = false

function clearTimer() {
  if (timer) {
    clearTimeout(timer)
    timer = null
  }
}

async function renewIfNeeded() {
  const auth = loadAuth()
  if (!auth?.expiresAt) {
    // 无本地过期时间时仍尝试 cookie 续期（例如仅 cookie 会话）
    await tryRefreshAccessToken()
    scheduleNext()
    return
  }
  if (isAuthExpired(auth.expiresAt) || shouldRefreshSoon(auth.expiresAt, RENEW_BEFORE_MS)) {
    const ok = await tryRefreshAccessToken()
    if (!ok) return
  }
  scheduleNext()
}

function scheduleNext() {
  clearTimer()
  const auth = loadAuth()
  const exp = auth?.expiresAt
  if (!exp) {
    timer = setTimeout(() => {
      void renewIfNeeded()
    }, RENEW_BEFORE_MS)
    return
  }
  const delay = Math.max(exp * 1000 - Date.now() - RENEW_BEFORE_MS, MIN_DELAY_MS)
  timer = setTimeout(() => {
    void renewIfNeeded()
  }, delay)
}

function onVisibility() {
  if (document.visibilityState === 'visible') {
    void renewIfNeeded()
  }
}

/** 启动静默续期：到期前 5 分钟刷新；切回标签页时再检查一次 */
export function startTokenKeepAlive() {
  if (started) {
    scheduleNext()
    return
  }
  started = true
  document.addEventListener('visibilitychange', onVisibility)
  void renewIfNeeded()
}

export function stopTokenKeepAlive() {
  clearTimer()
  if (started) {
    document.removeEventListener('visibilitychange', onVisibility)
    started = false
  }
}
