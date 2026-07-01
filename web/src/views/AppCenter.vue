<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Goods } from '@element-plus/icons-vue'
import { fetchApps, type AppItem } from '../api/auth'
import { getToken } from '../utils/token'
import { buildAppLaunchUrl } from '../utils/appUrl'

const router = useRouter()
const apps = ref<AppItem[]>([])
const loading = ref(false)

async function loadApps() {
  loading.value = true
  try {
    apps.value = await fetchApps()
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

function openApp(app: AppItem) {
  const token = getToken()
  if (!token) {
    router.push('/login')
    return
  }
  window.location.href = buildAppLaunchUrl(app.url, token)
}

onMounted(loadApps)
</script>

<template>
  <div v-loading="loading">
    <h2 class="page-title">应用中心</h2>
    <p class="page-desc">选择要进入的业务应用，数据按当前租户隔离。</p>
    <div class="app-grid">
      <div v-for="app in apps" :key="app.code" class="app-card" @click="openApp(app)">
        <div class="app-icon"><el-icon :size="32"><Goods /></el-icon></div>
        <h3>{{ app.name }}</h3>
        <p>{{ app.description }}</p>
      </div>
      <el-empty v-if="!loading && !apps.length" description="暂无可用应用" />
    </div>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 8px; }
.page-desc { margin: 0 0 20px; color: #909399; font-size: 14px; }
.app-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 20px;
}
.app-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  cursor: pointer;
  border: 1px solid #ebeef5;
  transition: box-shadow 0.2s, transform 0.2s;
}
.app-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}
.app-icon {
  width: 56px; height: 56px; border-radius: 12px;
  background: #ecf5ff; color: #409eff;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
.app-card h3 { margin: 0 0 8px; }
.app-card p { margin: 0; color: #909399; font-size: 13px; line-height: 1.5; }
</style>
