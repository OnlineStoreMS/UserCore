import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchMe, type LoginResponse, type MeResponse } from '../api/auth'
import { clearAuth, loadAuth, saveAuth, type StoredAuth } from '../utils/token'

function toStoredAuth(data: LoginResponse | MeResponse, current?: StoredAuth | null): StoredAuth {
  return {
    accessToken: 'accessToken' in data ? data.accessToken : current!.accessToken,
    expiresAt: 'expiresAt' in data ? data.expiresAt : current!.expiresAt,
    user: data.user,
    tenant: data.tenant,
    permissions: data.permissions,
    tenants: data.tenants?.length ? data.tenants : [data.tenant],
  }
}

export const useAuthStore = defineStore('auth', () => {
  const auth = ref<StoredAuth | null>(loadAuth())

  function setFromLogin(data: LoginResponse) {
    const stored = toStoredAuth(data)
    saveAuth(stored)
    auth.value = stored
  }

  async function refreshSession() {
    if (!auth.value) return
    const data = await fetchMe()
    const stored = toStoredAuth(data, auth.value)
    saveAuth(stored)
    auth.value = stored
  }

  function logout() {
    clearAuth()
    auth.value = null
  }

  return { auth, setFromLogin, refreshSession, logout }
})
