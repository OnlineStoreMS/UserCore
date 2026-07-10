import client, { unwrap, type PageData } from './client'

export interface UserRow {
  id: number
  email: string
  displayName: string
  phone: string
  status: number
  isPlatform: boolean
  roles: { id: number; name: string; code: string }[]
}

export interface RoleRow {
  id: number
  code: string
  name: string
  description: string
  isBuiltin: boolean
  permissions: string[]
}

export interface PermissionRow {
  code: string
  name: string
  appCode: string
}

export interface TenantRow {
  id: number
  companyId: number
  companyName?: string
  name: string
  code: string
  status: number
  remark: string
}

export interface CompanyRow {
  id: number
  name: string
  code: string
  status: number
  remark: string
  tenantCount: number
}

export async function fetchUsers(params?: { page?: number; pageSize?: number; keyword?: string }) {
  const res = await client.get('/users', { params })
  return unwrap<PageData<UserRow>>(res)
}

export async function createUser(data: {
  email: string
  password: string
  displayName: string
  phone?: string
  roleIds?: number[]
}) {
  const res = await client.post('/users', data)
  return unwrap<UserRow>(res)
}

export async function updateUser(id: number, data: Record<string, unknown>) {
  const res = await client.put(`/users/${id}`, data)
  return unwrap<UserRow>(res)
}

export async function removeUser(id: number) {
  await client.delete(`/users/${id}`)
}

export async function fetchRoles() {
  const res = await client.get('/roles')
  return unwrap<RoleRow[]>(res)
}

export async function createRole(data: {
  code: string
  name: string
  description?: string
  permissions?: string[]
}) {
  const res = await client.post('/roles', data)
  return unwrap<RoleRow>(res)
}

export async function updateRole(id: number, data: Record<string, unknown>) {
  const res = await client.put(`/roles/${id}`, data)
  return unwrap<RoleRow>(res)
}

export async function deleteRole(id: number) {
  await client.delete(`/roles/${id}`)
}

export async function fetchPermissions() {
  const res = await client.get('/permissions')
  return unwrap<PermissionRow[]>(res)
}

export async function fetchTenants(params?: { page?: number; pageSize?: number; keyword?: string }) {
  const res = await client.get('/tenants', { params })
  return unwrap<PageData<TenantRow>>(res)
}

export async function createTenant(data: {
  companyId: number
  name: string
  code: string
  remark?: string
}) {
  const res = await client.post('/tenants', data)
  return unwrap<TenantRow>(res)
}

export async function updateTenant(id: number, data: Record<string, unknown>) {
  const res = await client.put(`/tenants/${id}`, data)
  return unwrap<TenantRow>(res)
}

export async function fetchCompanies(params?: { page?: number; pageSize?: number; keyword?: string }) {
  const res = await client.get('/companies', { params })
  return unwrap<PageData<CompanyRow>>(res)
}

export async function createCompany(data: { name: string; code: string; remark?: string }) {
  const res = await client.post('/companies', data)
  return unwrap<CompanyRow>(res)
}

export async function updateCompany(id: number, data: Record<string, unknown>) {
  const res = await client.put(`/companies/${id}`, data)
  return unwrap<CompanyRow>(res)
}
