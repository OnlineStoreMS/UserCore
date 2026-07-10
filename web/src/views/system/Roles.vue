<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createRole, deleteRole, fetchPermissions, fetchRoles, updateRole,
  type PermissionRow, type RoleRow,
} from '../../api/admin'

const loading = ref(false)
const roles = ref<RoleRow[]>([])
const permissions = ref<PermissionRow[]>([])
const dialogVisible = ref(false)
const editing = ref<RoleRow | null>(null)
const form = reactive({ code: '', name: '', description: '', permissions: [] as string[] })

const permGroups = computed(() => {
  const groups: Record<string, PermissionRow[]> = {}
  for (const p of permissions.value) {
    const key = p.appCode || 'other'
    if (!groups[key]) groups[key] = []
    groups[key].push(p)
  }
  return groups
})

async function load() {
  loading.value = true
  try {
    const [roleList, permList] = await Promise.all([fetchRoles(), fetchPermissions()])
    roles.value = roleList
    permissions.value = permList
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { code: '', name: '', description: '', permissions: [] })
  dialogVisible.value = true
}

function openEdit(row: RoleRow) {
  if (row.isBuiltin) {
    ElMessage.warning('内置角色不可编辑')
    return
  }
  editing.value = row
  Object.assign(form, {
    code: row.code,
    name: row.name,
    description: row.description,
    permissions: [...row.permissions],
  })
  dialogVisible.value = true
}

async function submit() {
  try {
    if (editing.value) {
      await updateRole(editing.value.id, {
        name: form.name,
        description: form.description,
        permissions: form.permissions,
      })
    } else {
      await createRole(form)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function onDelete(row: RoleRow) {
  if (row.isBuiltin) return
  await ElMessageBox.confirm(`确定删除角色「${row.name}」？`, '确认')
  try {
    await deleteRole(row.id)
    ElMessage.success('已删除')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>角色权限</h2>
      <el-button type="primary" @click="openCreate">新建角色</el-button>
    </div>
    <el-table v-loading="loading" :data="roles" stripe>
      <el-table-column prop="name" label="角色" width="140" />
      <el-table-column prop="code" label="编码" width="140" />
      <el-table-column prop="description" label="说明" min-width="120" />
      <el-table-column label="权限" min-width="200">
        <template #default="{ row }">
          <el-tag v-for="p in row.permissions.slice(0, 3)" :key="p" size="small" style="margin: 2px">{{ p }}</el-tag>
          <span v-if="row.permissions.length > 3">+{{ row.permissions.length - 3 }}</span>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.code === 'platform_admin'" type="warning" size="small">平台</el-tag>
          <span v-else>{{ row.isBuiltin ? '内置' : '自定义' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :disabled="row.isBuiltin" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" :disabled="row.isBuiltin" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑角色' : '新建角色'" width="560px">
      <el-form label-width="80px">
        <el-form-item label="编码" required>
          <el-input v-model="form.code" :disabled="!!editing" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" />
        </el-form-item>
        <el-form-item label="权限">
          <div v-for="(list, app) in permGroups" :key="app" class="perm-group">
            <div class="perm-app">{{ app }}</div>
            <el-checkbox-group v-model="form.permissions">
              <el-checkbox v-for="p in list" :key="p.code" :value="p.code" :label="p.code">
                {{ p.name }}
              </el-checkbox>
            </el-checkbox-group>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar h2 { margin: 0; font-size: 18px; }
.perm-group { margin-bottom: 12px; }
.perm-app { font-size: 12px; color: #909399; margin-bottom: 4px; }
</style>
