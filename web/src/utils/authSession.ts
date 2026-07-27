import { clearAuth, getRefreshToken, updateTokens } from './token'

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
 * 用 refreshToken 换新 accessToken。并发调用会复用同一次请求。
 * 成功返回 true；无 refresh / 失败返回 false（不自动跳登录）。
 */
export async function tryRefreshAccessToken(): Promise<boolean> {
  if (refreshPromise) return refreshPromise
  refreshPromise = (async () => {
    const refreshToken = getRefreshToken()
    if (!refreshToken) return false
    try {
      const res = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken }),
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
