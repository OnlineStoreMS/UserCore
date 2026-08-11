import { clearAuth, updateTokens } from './token'

let handlingUnauthorized = false
let refreshPromise: Promise<boolean> | null = null

/** 登录失效：清本地凭证并跳转登录页 */
export function handleUnauthorized() {
  if (handlingUnauthorized) return
  if (window.location.pathname === '/login') return
  handlingUnauthorized = true
  clearAuth()
  const redirect = encodeURIComponent(window.location.pathname + window.location.search)
  window.location.assign(`/login?redirect=${redirect}`)
}

export function isAuthExpired(expiresAt?: number): boolean {
  if (!expiresAt) return false
  return expiresAt * 1000 <= Date.now()
}

/** 距过期不足 thresholdMs 时需要续期 */
export function shouldRefreshSoon(expiresAt?: number, thresholdMs = 5 * 60 * 1000): boolean {
  if (!expiresAt) return false
  return expiresAt * 1000 - Date.now() <= thresholdMs
}

/**
 * Cookie 会话静默续期（凭 uc_refresh httpOnly cookie）。
 * 并发调用复用同一次请求。
 */
export async function tryRefreshAccessToken(): Promise<boolean> {
  if (refreshPromise) return refreshPromise
  refreshPromise = (async () => {
    try {
      const res = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      })
      const body = await res.json()
      if (body.code !== 200 || !body.data?.accessToken) return false
      updateTokens(body.data.accessToken, body.data.expiresAt, body.data.refreshToken)
      return true
    } catch {
      return false
    } finally {
      refreshPromise = null
    }
  })()
  return refreshPromise
}

export async function logoutOnServer() {
  try {
    await fetch('/api/v1/auth/logout', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    })
  } catch {
    // ignore
  }
}
