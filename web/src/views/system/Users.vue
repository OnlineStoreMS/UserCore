<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createUser, fetchRoles, fetchUsers, removeUser, updateUser, type RoleRow, type UserRow,
} from '../../api/admin'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const users = ref<UserRow[]>([])
const roles = ref<RoleRow[]>([])
const dialogVisible = ref(false)
const editing = ref<UserRow | null>(null)
const form = reactive({
  email: '',
  password: '',
  displayName: '',
  phone: '',
  status: 1,
  roleIds: [] as number[],
})

const currentTenantName = computed(() => auth.auth?.tenant.name ?? '')

async function load() {
  loading.value = true
  try {
    const [userData, roleList] = await Promise.all([fetchUsers({ pageSize: 100 }), fetchRoles()])
    users.value = userData.list
    roles.value = roleList
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { email: '', password: '', displayName: '', phone: '', status: 1, roleIds: [] })
  dialogVisible.value = true
}

function openEdit(row: UserRow) {
  editing.value = row
  Object.assign(form, {
    email: row.email,
    password: '',
    displayName: row.displayName,
    phone: row.phone,
    status: row.status,
    roleIds: row.roles.map((r) => r.id),
  })
  dialogVisible.value = true
}

async function submit() {
  try {
    if (editing.value) {
      const payload: Record<string, unknown> = {
        displayName: form.displayName,
        phone: form.phone,
        status: form.status,
        roleIds: form.roleIds,
      }
      if (form.password) payload.password = form.password
      await updateUser(editing.value.id, payload)
      ElMessage.success('已更新')
    } else {
      await createUser({
        email: form.email,
        password: form.password,
        displayName: form.displayName,
        phone: form.phone,
        roleIds: form.roleIds,
      })
      ElMessage.success('已创建')
    }
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

async function onRemove(row: UserRow) {
  if (row.id === auth.auth?.user.id) {
    ElMessage.warning('不能移除当前登录用户')
    return
  }
  if (row.isPlatform) {
    ElMessage.warning('不能移除平台管理员')
    return
  }
  await ElMessageBox.confirm(
    `确定将「${row.displayName}」从租户「${currentTenantName.value}」移除？账号仍保留，可再次添加。`,
    '确认移除',
  )
  try {
    await removeUser(row.id)
    ElMessage.success('已移除')
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
      <div>
        <h2>用户管理</h2>
        <p v-if="currentTenantName" class="subtitle">当前租户：{{ currentTenantName }}</p>
      </div>
      <el-button type="primary" @click="openCreate">添加用户</el-button>
    </div>
    <el-table v-loading="loading" :data="users" stripe>
      <el-table-column prop="displayName" label="姓名" width="120" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column prop="phone" label="手机" width="120" />
      <el-table-column label="角色" min-width="200">
        <template #default="{ row }">
          <el-tag v-if="row.isPlatform" type="warning" size="small" style="margin-right: 4px">平台账号</el-tag>
          <el-tag v-for="r in row.roles" :key="r.id" size="small" style="margin-right: 4px">{{ r.name }}</el-tag>
          <span v-if="!row.isPlatform && !row.roles.length" class="muted">未分配</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">{{ row.status === 1 ? '启用' : '禁用' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button
            link
            type="danger"
            :disabled="row.id === auth.auth?.user.id || row.isPlatform"
            @click="onRemove(row)"
          >
            移除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑用户' : '添加用户'" width="480px">
      <el-form label-width="80px">
        <el-form-item v-if="!editing" label="邮箱" required>
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item :label="editing ? '新密码' : '密码'" :required="!editing">
          <el-input v-model="form.password" type="password" show-password :placeholder="editing ? '留空则不修改' : ''" />
        </el-form-item>
        <el-form-item label="姓名" required>
          <el-input v-model="form.displayName" />
        </el-form-item>
        <el-form-item label="手机">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item v-if="editing" label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.roleIds" multiple style="width: 100%">
            <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
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
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.toolbar h2 { margin: 0; font-size: 18px; }
.subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #909399;
}
.muted {
  color: #c0c4cc;
  font-size: 12px;
}
</style>
