import { clearAuth } from './token'

let handlingUnauthorized = false

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
