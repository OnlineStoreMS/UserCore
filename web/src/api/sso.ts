import client, { unwrap } from './client'

export interface SSOAuthorizeRequest {
  appCode?: string
  redirectUri?: string
}

export interface SSOAuthorizeResponse {
  code: string
  expiresIn: number
  redirectUri: string
}

export async function authorizeSSO(data: SSOAuthorizeRequest) {
  const res = await client.post('/auth/sso/authorize', data)
  return unwrap<SSOAuthorizeResponse>(res)
}
