const AUTH_KEY = 'uc_auth_profile'

export interface StoredAuth {
  /** @deprecated cookie SSO — kept empty for profile-only storage */
  accessToken: string
  refreshToken?: string
  expiresAt: number
  user: { id: number; email: string; displayName: string; isPlatform: boolean }
  tenant: { id: number; companyId: number; name: string; code: string }
  permissions: string[]
  tenants: { id: number; companyId: number; name: string; code: string }[]
}

/** httpOnly Cookie 会话：JS 不可读 access/refresh */
export function getToken(): string | undefined {
  return undefined
}

export function getRefreshToken(): string | undefined {
  return undefined
}

export function saveAuth(auth: StoredAuth) {
  const profile: StoredAuth = {
    ...auth,
    accessToken: '',
    refreshToken: undefined,
  }
  localStorage.setItem(AUTH_KEY, JSON.stringify(profile))
}

export function updateTokens(_accessToken: string, expiresAt: number, _refreshToken?: string) {
  const current = loadAuth()
  if (current) {
    current.expiresAt = expiresAt
    current.accessToken = ''
    current.refreshToken = undefined
    localStorage.setItem(AUTH_KEY, JSON.stringify(current))
  }
}

export function loadAuth(): StoredAuth | null {
  const raw = localStorage.getItem(AUTH_KEY)
  if (!raw) return null
  try {
    const auth = JSON.parse(raw) as StoredAuth
    auth.accessToken = ''
    auth.refreshToken = undefined
    return auth
  } catch {
    return null
  }
}

export function clearAuth() {
  localStorage.removeItem(AUTH_KEY)
  // legacy keys
  localStorage.removeItem('uc_access_token')
  localStorage.removeItem('uc_refresh_token')
}

export function hasPerm(perms: string[] | undefined, code: string): boolean {
  if (!perms) return false
  return perms.includes(code) || perms.includes('*')
}
