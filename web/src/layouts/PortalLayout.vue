<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Grid, OfficeBuilding, HomeFilled, Setting, SwitchButton, User, UserFilled,
} from '@element-plus/icons-vue'
import { switchTenant } from '../api/auth'
import { useAuthStore } from '../stores/auth'
import { hasPerm } from '../utils/token'
import { logoutFromApps } from '../utils/logout'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const isPlatform = computed(() => auth.auth?.user.isPlatform)

async function onSwitchTenant(tenantId: number) {
  try {
    const data = await switchTenant(tenantId)
    auth.setFromLogin(data)
    ElMessage.success(`已切换到 ${data.tenant.name}`)
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function logout() {
  await logoutFromApps()
  auth.logout()
  router.push('/login')
}

function navTo(path: string) {
  router.push(path)
}

onMounted(() => {
  void auth.refreshSession().catch(() => {})
})
</script>

<template>
  <div class="portal-layout">
    <aside class="sidebar">
      <div class="brand">UserCore</div>
      <el-menu :default-active="route.path" router>
        <el-menu-item index="/apps"><el-icon><Grid /></el-icon>应用中心</el-menu-item>
        <el-menu-item v-if="hasPerm(auth.auth?.permissions, 'tenant:admin')" index="/system/users">
          <el-icon><User /></el-icon>用户管理
        </el-menu-item>
        <el-menu-item v-if="hasPerm(auth.auth?.permissions, 'tenant:admin')" index="/system/roles">
          <el-icon><UserFilled /></el-icon>角色权限
        </el-menu-item>
        <el-menu-item v-if="isPlatform" index="/platform/companies">
          <el-icon><OfficeBuilding /></el-icon>公司管理
        </el-menu-item>
        <el-menu-item v-if="isPlatform" index="/platform/tenants">
          <el-icon><HomeFilled /></el-icon>租户管理
        </el-menu-item>
      </el-menu>
    </aside>
    <div class="main">
      <header class="topbar">
        <div class="topbar-left">
          <span v-if="auth.auth">{{ auth.auth.user.displayName }} · {{ auth.auth.tenant.name }}</span>
        </div>
        <div class="topbar-right">
          <el-select
            v-if="auth.auth && auth.auth.tenants.length > 1"
            :model-value="auth.auth.tenant.id"
            style="width: 160px"
            @change="onSwitchTenant"
          >
            <el-option v-for="t in auth.auth.tenants" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
          <el-button text :icon="Setting" @click="navTo('/apps')">应用中心</el-button>
          <el-button text :icon="SwitchButton" @click="logout">退出</el-button>
        </div>
      </header>
      <main class="content">
        <router-view :key="auth.auth?.tenant.id ?? 0" />
      </main>
    </div>
  </div>
</template>

<style scoped>
.portal-layout {
  display: flex;
  min-height: 100vh;
  background: #f5f7fa;
}
.sidebar {
  width: 220px;
  background: #fff;
  border-right: 1px solid #ebeef5;
  flex-shrink: 0;
}
.brand {
  padding: 20px 16px;
  font-size: 18px;
  font-weight: 600;
  border-bottom: 1px solid #f0f2f5;
}
.main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.topbar {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}
.topbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
}
.content {
  flex: 1;
  padding: 20px;
  overflow: auto;
}
</style>
