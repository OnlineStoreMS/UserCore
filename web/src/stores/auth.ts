import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { LoginResponse } from '../api/auth'
import { clearAuth, loadAuth, saveAuth, type StoredAuth } from '../utils/token'

export const useAuthStore = defineStore('auth', () => {
  const auth = ref<StoredAuth | null>(loadAuth())

  function setFromLogin(data: LoginResponse) {
    const stored: StoredAuth = {
      accessToken: data.accessToken,
      expiresAt: data.expiresAt,
      user: data.user,
      tenant: data.tenant,
      permissions: data.permissions,
      tenants: data.tenants || [data.tenant],
    }
    saveAuth(stored)
    auth.value = stored
  }

  function logout() {
    clearAuth()
    auth.value = null
  }

  return { auth, setFromLogin, logout }
})
