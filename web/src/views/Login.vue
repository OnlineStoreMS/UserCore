<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login, type TenantBrief } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

function afterLogin() {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  if (redirect && redirect.startsWith('/') && !redirect.startsWith('//')) {
    // 跨应用路径（如 /apps/ops-m/）需整页跳转，不能走门户 Vue Router
    if (redirect.startsWith('/apps/')) {
      // replace + 末尾斜杠，避免历史栈回登录页；给 cookie 落盘留一拍
      const target = redirect.endsWith('/') ? redirect : `${redirect}/`
      window.setTimeout(() => window.location.replace(target), 50)
      return
    }
    return router.replace(redirect)
  }
  return router.replace('/apps')
}

const email = ref('')
const password = ref('')
const loading = ref(false)
const tenants = ref<TenantBrief[]>([])
const selectedTenantId = ref<number>()
const step = ref<'login' | 'tenant'>('login')

async function onLogin() {
  loading.value = true
  try {
    const data = await login({ email: email.value.trim(), password: password.value })
    if (!data.accessToken) {
      tenants.value = data.tenants || []
      if (tenants.value.length === 0) {
        ElMessage.error('账号未加入任何租户')
        return
      }
      selectedTenantId.value = tenants.value[0].id
      step.value = 'tenant'
      return
    }
    auth.setFromLogin(data)
    await afterLogin()
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

async function onSelectTenant() {
  if (!selectedTenantId.value) return
  loading.value = true
  try {
    const data = await login({
      email: email.value.trim(),
      password: password.value,
      tenantId: selectedTenantId.value,
    })
    if (!data.accessToken) {
      ElMessage.error('登录失败')
      return
    }
    auth.setFromLogin(data)
    await afterLogin()
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>UserCore</h1>
      <p class="subtitle">统一身份与应用中心</p>

      <template v-if="step === 'login'">
        <el-form label-position="top" @submit.prevent="onLogin">
          <el-form-item label="邮箱">
            <el-input
              v-model="email"
              size="large"
              type="email"
              autocomplete="username"
              inputmode="email"
              placeholder="请输入邮箱"
            />
          </el-form-item>
          <el-form-item label="密码">
            <el-input
              v-model="password"
              size="large"
              type="password"
              show-password
              autocomplete="current-password"
              placeholder="请输入密码"
              @keyup.enter="onLogin"
            />
          </el-form-item>
          <el-button type="primary" size="large" class="submit" :loading="loading" @click="onLogin">登录</el-button>
        </el-form>
      </template>

      <template v-else>
        <p class="tenant-title">选择要进入的租户</p>
        <el-radio-group v-model="selectedTenantId" class="tenant-list">
          <el-radio v-for="t in tenants" :key="t.id" :value="t.id" border>
            {{ t.name }}（{{ t.code }}）
          </el-radio>
        </el-radio-group>
        <el-button type="primary" size="large" class="submit" :loading="loading" @click="onSelectTenant">进入</el-button>
        <el-button text class="back" @click="step = 'login'">返回</el-button>
      </template>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: max(16px, env(safe-area-inset-top)) max(16px, env(safe-area-inset-right))
    max(16px, env(safe-area-inset-bottom)) max(16px, env(safe-area-inset-left));
  box-sizing: border-box;
  background: linear-gradient(160deg, #e8eefc 0%, #f8fafc 45%, #e6f7f5 100%);
}
.login-card {
  width: 100%;
  max-width: 400px;
  padding: 28px 24px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 12px 40px rgba(15, 23, 42, 0.08);
  box-sizing: border-box;
}
h1 {
  margin: 0;
  font-size: 26px;
  line-height: 1.2;
  letter-spacing: -0.02em;
}
.subtitle {
  margin: 8px 0 22px;
  color: #64748b;
  font-size: 14px;
  line-height: 1.4;
}
.submit {
  width: 100%;
  margin-top: 8px;
  height: 44px;
}
.back {
  width: 100%;
  margin-top: 4px;
  height: 40px;
}
.tenant-title {
  margin: 0 0 12px;
  color: #334155;
  font-size: 15px;
}
.tenant-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
  margin-bottom: 8px;
}
.tenant-list :deep(.el-radio) {
  width: 100%;
  margin: 0;
  height: auto;
  min-height: 44px;
  padding: 10px 14px;
  box-sizing: border-box;
}
.tenant-list :deep(.el-radio__label) {
  white-space: normal;
  line-height: 1.35;
  word-break: break-word;
}

/* 手机：输入框 16px 避免 iOS 自动放大；卡片贴边更舒适 */
:deep(.el-input__inner),
:deep(.el-input__wrapper input) {
  font-size: 16px !important;
}

@media (max-width: 480px) {
  .login-page {
    align-items: stretch;
    padding: 0;
    background: #fff;
  }
  .login-card {
    max-width: none;
    min-height: 100vh;
    min-height: 100dvh;
    border-radius: 0;
    box-shadow: none;
    padding: max(28px, env(safe-area-inset-top)) max(20px, env(safe-area-inset-right))
      max(28px, env(safe-area-inset-bottom)) max(20px, env(safe-area-inset-left));
    display: flex;
    flex-direction: column;
    justify-content: center;
  }
  h1 {
    font-size: 28px;
  }
  .subtitle {
    margin-bottom: 28px;
  }
}
</style>
