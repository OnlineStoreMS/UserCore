import { createRouter, createWebHistory } from 'vue-router'
import PortalLayout from '../layouts/PortalLayout.vue'
import { clearAuth, getToken, loadAuth } from '../utils/token'
import { isAuthExpired } from '../utils/authSession'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'Login', component: () => import('../views/Login.vue'), meta: { public: true } },
    {
      path: '/',
      component: PortalLayout,
      redirect: '/apps',
      children: [
        { path: 'apps', name: 'Apps', component: () => import('../views/AppCenter.vue') },
        { path: 'system/users', name: 'SystemUsers', component: () => import('../views/system/Users.vue') },
        { path: 'system/roles', name: 'SystemRoles', component: () => import('../views/system/Roles.vue') },
        { path: 'platform/tenants', name: 'PlatformTenants', component: () => import('../views/platform/Tenants.vue') },
        { path: 'platform/companies', name: 'PlatformCompanies', component: () => import('../views/platform/Companies.vue') },
      ],
    },
  ],
})

router.beforeEach((to) => {
  if (to.meta.public) return true
  const token = getToken()
  if (!token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  const auth = loadAuth()
  if (isAuthExpired(auth?.expiresAt)) {
    clearAuth()
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
