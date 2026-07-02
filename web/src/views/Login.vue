<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login, type TenantBrief } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

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
    await router.push('/apps')
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
    await router.push('/apps')
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
            <el-input v-model="email" placeholder="请输入邮箱" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="password" type="password" show-password placeholder="请输入密码" @keyup.enter="onLogin" />
          </el-form-item>
          <el-button type="primary" class="submit" :loading="loading" @click="onLogin">登录</el-button>
        </el-form>
      </template>

      <template v-else>
        <p class="tenant-title">选择要进入的租户</p>
        <el-radio-group v-model="selectedTenantId" class="tenant-list">
          <el-radio v-for="t in tenants" :key="t.id" :value="t.id" border>
            {{ t.name }}（{{ t.code }}）
          </el-radio>
        </el-radio-group>
        <el-button type="primary" class="submit" :loading="loading" @click="onSelectTenant">进入</el-button>
        <el-button text @click="step = 'login'">返回</el-button>
      </template>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #eef2ff 0%, #f8fafc 50%, #ecfeff 100%);
}
.login-card {
  width: 400px;
  padding: 32px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(15, 23, 42, 0.08);
}
h1 {
  margin: 0;
  font-size: 28px;
}
.subtitle {
  margin: 8px 0 24px;
  color: #64748b;
}
.submit {
  width: 100%;
  margin-top: 8px;
}
.tenant-title {
  margin-bottom: 12px;
  color: #334155;
}
.tenant-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
</style>
