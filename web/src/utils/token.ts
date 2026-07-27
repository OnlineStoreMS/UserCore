const TOKEN_KEY = 'uc_access_token'
const REFRESH_KEY = 'uc_refresh_token'
const AUTH_KEY = 'uc_auth_profile'

export interface StoredAuth {
  accessToken: string
  refreshToken?: string
  expiresAt: number
  user: { id: number; email: string; displayName: string; isPlatform: boolean }
  tenant: { id: number; companyId: number; name: string; code: string }
  permissions: string[]
  tenants: { id: number; companyId: number; name: string; code: string }[]
}

export function getToken(): string | undefined {
  return localStorage.getItem(TOKEN_KEY) || undefined
}

export function getRefreshToken(): string | undefined {
  return localStorage.getItem(REFRESH_KEY) || undefined
}

export function saveAuth(auth: StoredAuth) {
  localStorage.setItem(TOKEN_KEY, auth.accessToken)
  localStorage.setItem(AUTH_KEY, JSON.stringify(auth))
  if (auth.refreshToken) {
    localStorage.setItem(REFRESH_KEY, auth.refreshToken)
  }
}

export function updateTokens(accessToken: string, expiresAt: number, refreshToken?: string) {
  localStorage.setItem(TOKEN_KEY, accessToken)
  if (refreshToken) {
    localStorage.setItem(REFRESH_KEY, refreshToken)
  }
  const current = loadAuth()
  if (current) {
    current.accessToken = accessToken
    current.expiresAt = expiresAt
    if (refreshToken) current.refreshToken = refreshToken
    localStorage.setItem(AUTH_KEY, JSON.stringify(current))
  }
}

export function loadAuth(): StoredAuth | null {
  const raw = localStorage.getItem(AUTH_KEY)
  if (!raw) return null
  try {
    const auth = JSON.parse(raw) as StoredAuth
    if (!auth.refreshToken) {
      auth.refreshToken = getRefreshToken()
    }
    return auth
  } catch {
    return null
  }
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_KEY)
  localStorage.removeItem(AUTH_KEY)
}

export function hasPerm(perms: string[] | undefined, code: string): boolean {
  if (!perms) return false
  return perms.includes(code) || perms.includes('*')
}
