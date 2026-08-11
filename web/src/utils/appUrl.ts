/** 应用注册 URL 可能是站点根或 /auth/callback，统一解析为站点根 */
export function appBaseUrl(appUrl: string): string {
  const trimmed = appUrl.replace(/\/$/, '')
  if (trimmed.endsWith('/auth/callback')) {
    return trimmed.slice(0, -'/auth/callback'.length)
  }
  return trimmed
}

/** 用一次性 SSO code 进入子应用（禁止在 URL 中携带 JWT） */
export function buildAppLaunchUrl(
  redirectUri: string,
  code: string,
  extraQuery?: Record<string, string>,
): string {
  const url = new URL(redirectUri)
  url.searchParams.set('code', code)
  if (extraQuery) {
    for (const [k, v] of Object.entries(extraQuery)) {
      if (v) url.searchParams.set(k, v)
    }
  }
  return url.toString()
}

/** 子应用登出地址（iframe 清 token） */
export function buildAppLogoutUrl(appUrl: string): string {
  return `${appBaseUrl(appUrl)}/auth/logout`
}
