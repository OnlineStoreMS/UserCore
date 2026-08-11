import { logoutOnServer } from './authSession'

/** 统一登出：吊销 refresh session 并清除共享 httpOnly Cookie */
export async function logoutFromApps() {
  await logoutOnServer()
}
