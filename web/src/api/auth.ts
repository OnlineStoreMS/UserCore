import client, { unwrap } from './client'

export interface LoginRequest {
  email: string
  password: string
  tenantId?: number
}

export interface TenantBrief {
  id: number
  companyId: number
  name: string
  code: string
}

export interface UserProfile {
  id: number
  email: string
  displayName: string
  isPlatform: boolean
}

export interface LoginResponse {
  accessToken: string
  expiresAt: number
  user: UserProfile
  tenant: TenantBrief
  permissions: string[]
  tenants?: TenantBrief[]
}

export interface AppItem {
  id: number
  code: string
  name: string
  description: string
  icon: string
  url: string
  sort: number
}

export async function login(data: LoginRequest) {
  const res = await client.post('/auth/login', data)
  return unwrap<LoginResponse>(res)
}

export async function fetchMe() {
  const res = await client.get('/auth/me')
  return unwrap<LoginResponse>(res)
}

export async function switchTenant(tenantId: number) {
  const res = await client.post('/auth/switch-tenant', { tenantId })
  return unwrap<LoginResponse>(res)
}

export async function fetchApps() {
  const res = await client.get('/apps')
  return unwrap<AppItem[]>(res)
}
