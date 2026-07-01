import { fetchApps } from '../api/auth'
import { buildAppLogoutUrl } from './appUrl'

/** 通知各业务应用清除本地 JWT（跨端口 localStorage 互不共享） */
export async function logoutFromApps() {
  try {
    const apps = await fetchApps()
    const iframes: HTMLIFrameElement[] = []
    for (const app of apps) {
      const iframe = document.createElement('iframe')
      iframe.style.display = 'none'
      iframe.src = buildAppLogoutUrl(app.url)
      document.body.appendChild(iframe)
      iframes.push(iframe)
    }
    if (iframes.length > 0) {
      await new Promise((resolve) => setTimeout(resolve, 400))
      for (const iframe of iframes) {
        iframe.remove()
      }
    }
  } catch {
    // 忽略：主流程仍应完成 UserCore 本地退出
  }
}
