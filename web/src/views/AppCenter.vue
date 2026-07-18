<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { fetchApps, saveAppOrder, type AppItem } from '../api/auth'
import { getToken } from '../utils/token'
import { buildAppLaunchUrl } from '../utils/appUrl'

const router = useRouter()
const apps = ref<AppItem[]>([])
const loading = ref(false)
const saving = ref(false)
const dragFrom = ref<number | null>(null)
const dragOver = ref<number | null>(null)
const dragged = ref(false)

const iconMap = ElementPlusIconsVue as Record<string, Component>
const FallbackIcon = ElementPlusIconsVue.Grid

function resolveIcon(name?: string): Component {
  if (!name) return FallbackIcon
  return iconMap[name] || FallbackIcon
}

const iconTone = computed(() => {
  const tones = [
    { bg: '#ecf5ff', color: '#409eff' },
    { bg: '#f0f9eb', color: '#67c23a' },
    { bg: '#fdf6ec', color: '#e6a23c' },
    { bg: '#fef0f0', color: '#f56c6c' },
    { bg: '#f4f4f5', color: '#909399' },
    { bg: '#f3e8ff', color: '#8b5cf6' },
    { bg: '#e0f2fe', color: '#0284c7' },
    { bg: '#fff7ed', color: '#ea580c' },
  ]
  return (code: string) => {
    let h = 0
    for (let i = 0; i < code.length; i++) h = (h * 31 + code.charCodeAt(i)) >>> 0
    return tones[h % tones.length]
  }
})

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

async function persistOrder() {
  saving.value = true
  try {
    await saveAppOrder(apps.value.map((a) => a.id))
    ElMessage.success('排序已保存')
  } catch (e) {
    ElMessage.error((e as Error).message || '保存排序失败')
    await loadApps()
  } finally {
    saving.value = false
  }
}

function openApp(app: AppItem) {
  if (dragged.value) {
    dragged.value = false
    return
  }
  const token = getToken()
  if (!token) {
    router.push('/login')
    return
  }
  window.location.href = buildAppLaunchUrl(app.url, token)
}

function onDragStart(index: number, e: DragEvent) {
  dragFrom.value = index
  dragged.value = false
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(index))
  }
}

function onDragOver(index: number, e: DragEvent) {
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  dragOver.value = index
}

function onDragLeave(index: number) {
  if (dragOver.value === index) dragOver.value = null
}

async function onDrop(index: number, e: DragEvent) {
  e.preventDefault()
  const from = dragFrom.value
  dragFrom.value = null
  dragOver.value = null
  if (from === null || from === index) return
  dragged.value = true
  const list = [...apps.value]
  const [item] = list.splice(from, 1)
  list.splice(index, 0, item)
  apps.value = list
  await persistOrder()
}

function onDragEnd() {
  dragFrom.value = null
  dragOver.value = null
}

onMounted(loadApps)
</script>

<template>
  <div v-loading="loading || saving">
    <div class="page-head">
      <div>
        <h2 class="page-title">应用中心</h2>
        <p class="page-desc">选择要进入的业务应用；拖拽卡片可自定义排序（按账号保存）。</p>
      </div>
    </div>
    <div class="app-grid">
      <div
        v-for="(app, index) in apps"
        :key="app.code"
        class="app-card"
        :class="{ 'is-drag-over': dragOver === index, 'is-dragging': dragFrom === index }"
        draggable="true"
        @dragstart="onDragStart(index, $event)"
        @dragover="onDragOver(index, $event)"
        @dragleave="onDragLeave(index)"
        @drop="onDrop(index, $event)"
        @dragend="onDragEnd"
        @click="openApp(app)"
      >
        <div
          class="app-icon"
          :style="{ background: iconTone(app.code).bg, color: iconTone(app.code).color }"
        >
          <el-icon :size="32">
            <component :is="resolveIcon(app.icon)" />
          </el-icon>
        </div>
        <h3>{{ app.name }}</h3>
        <p>{{ app.description }}</p>
      </div>
      <el-empty v-if="!loading && !apps.length" description="暂无可用应用" />
    </div>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}
.page-title { margin: 0 0 8px; }
.page-desc { margin: 0; color: #909399; font-size: 14px; }
.app-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 20px;
}
.app-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  cursor: grab;
  border: 1px solid #ebeef5;
  transition: box-shadow 0.2s, transform 0.2s, border-color 0.2s, opacity 0.2s;
  user-select: none;
}
.app-card:active { cursor: grabbing; }
.app-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}
.app-card.is-dragging {
  opacity: 0.45;
}
.app-card.is-drag-over {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}
.app-icon {
  width: 56px; height: 56px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
.app-card h3 { margin: 0 0 8px; }
.app-card p { margin: 0; color: #909399; font-size: 13px; line-height: 1.5; }
</style>
