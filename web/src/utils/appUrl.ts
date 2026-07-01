/** 应用注册 URL 可能是站点根或 /auth/callback，统一解析为站点根 */
export function appBaseUrl(appUrl: string): string {
  const trimmed = appUrl.replace(/\/$/, '')
  if (trimmed.endsWith('/auth/callback')) {
    return trimmed.slice(0, -'/auth/callback'.length)
  }
  return trimmed
}

/** 带 JWT 进入子应用的完整地址 */
export function buildAppLaunchUrl(appUrl: string, token: string): string {
  const url = new URL(`${appBaseUrl(appUrl)}/auth/callback`)
  url.searchParams.set('token', token)
  return url.toString()
}

/** 子应用登出地址（iframe 清 token） */
export function buildAppLogoutUrl(appUrl: string): string {
  return `${appBaseUrl(appUrl)}/auth/logout`
}
