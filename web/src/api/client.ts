import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'
import { getToken } from '../utils/token'
import { handleUnauthorized, tryRefreshAccessToken } from '../utils/authSession'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

const client: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

client.interceptors.request.use((config) => {
  // Cookie SSO：凭 httpOnly uc_access；过渡期若仍有 Bearer 可读则附带
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

async function retryAfterRefresh(config: InternalAxiosRequestConfig) {
  const ok = await tryRefreshAccessToken()
  if (!ok) {
    handleUnauthorized()
    return Promise.reject(new Error('登录已过期，请重新登录'))
  }
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  ;(config as InternalAxiosRequestConfig & { _retry?: boolean })._retry = true
  return client.request(config)
}

client.interceptors.response.use(
  async (res) => {
    const body = res.data as ApiResponse
    if (body.code === 401) {
      const cfg = res.config as InternalAxiosRequestConfig & { _retry?: boolean }
      if (!cfg._retry && !String(cfg.url || '').includes('/auth/refresh')) {
        return retryAfterRefresh(cfg)
      }
      handleUnauthorized()
      return Promise.reject(new Error(body.message || '登录已过期，请重新登录'))
    }
    if (body.code !== 200) {
      return Promise.reject(new Error(body.message || '请求失败'))
    }
    return res
  },
  async (err) => {
    const cfg = err.config as (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined
    if (err.response?.status === 401 && cfg && !cfg._retry && !String(cfg.url || '').includes('/auth/refresh')) {
      return retryAfterRefresh(cfg)
    }
    if (err.response?.status === 401) {
      handleUnauthorized()
      const body = err.response.data as ApiResponse | undefined
      return Promise.reject(new Error(body?.message || '登录已过期，请重新登录'))
    }
    return Promise.reject(err)
  },
)

export function unwrap<T>(res: { data: ApiResponse<T> }): T {
  return res.data.data as T
}

export default client
