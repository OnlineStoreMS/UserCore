const TOKEN_KEY = 'uc_access_token'
const AUTH_KEY = 'uc_auth_profile'

export interface StoredAuth {
  accessToken: string
  expiresAt: number
  user: { id: number; email: string; displayName: string; isPlatform: boolean }
  tenant: { id: number; companyId: number; name: string; code: string }
  permissions: string[]
  tenants: { id: number; companyId: number; name: string; code: string }[]
}

export function getToken(): string | undefined {
  return localStorage.getItem(TOKEN_KEY) || undefined
}

export function saveAuth(auth: StoredAuth) {
  localStorage.setItem(TOKEN_KEY, auth.accessToken)
  localStorage.setItem(AUTH_KEY, JSON.stringify(auth))
}

export function loadAuth(): StoredAuth | null {
  const raw = localStorage.getItem(AUTH_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as StoredAuth
  } catch {
    return null
  }
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(AUTH_KEY)
}

export function hasPerm(perms: string[] | undefined, code: string): boolean {
  if (!perms) return false
  return perms.includes(code) || perms.includes('*')
}
