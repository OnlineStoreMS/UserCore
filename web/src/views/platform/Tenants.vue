<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createTenant, fetchCompanies, fetchTenants, updateTenant, type CompanyRow, type TenantRow,
} from '../../api/admin'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()

const loading = ref(false)
const tenants = ref<TenantRow[]>([])
const companies = ref<CompanyRow[]>([])
const dialogVisible = ref(false)
const editing = ref<TenantRow | null>(null)
const form = reactive({ companyId: 0, name: '', code: '', remark: '', status: 1 })

async function load() {
  loading.value = true
  try {
    const [tenantData, companyData] = await Promise.all([
      fetchTenants({ pageSize: 100 }),
      fetchCompanies({ pageSize: 100 }),
    ])
    tenants.value = tenantData.list
    companies.value = companyData.list
  } catch (e) {
    ElMessage.error((e as Error).message)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, {
    companyId: companies.value[0]?.id || 0,
    name: '',
    code: '',
    remark: '',
    status: 1,
  })
  dialogVisible.value = true
}

function openEdit(row: TenantRow) {
  editing.value = row
  Object.assign(form, {
    companyId: row.companyId,
    name: row.name,
    code: row.code,
    remark: row.remark,
    status: row.status,
  })
  dialogVisible.value = true
}

async function submit() {
  try {
    if (editing.value) {
      await updateTenant(editing.value.id, {
        name: form.name,
        code: form.code,
        remark: form.remark,
        status: form.status,
      })
    } else {
      await createTenant(form)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await load()
    await auth.refreshSession()
  } catch (e) {
    ElMessage.error((e as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <h2>租户管理</h2>
      <el-button type="primary" @click="openCreate">新建租户</el-button>
    </div>
    <el-table v-loading="loading" :data="tenants" stripe>
      <el-table-column prop="name" label="租户" width="140" />
      <el-table-column prop="code" label="编码" width="120" />
      <el-table-column prop="companyName" label="所属公司" min-width="140" />
      <el-table-column prop="remark" label="备注" min-width="120" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">{{ row.status === 1 ? '启用' : '禁用' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑租户' : '新建租户'" width="480px">
      <el-form label-width="80px">
        <el-form-item v-if="!editing" label="公司" required>
          <el-select v-model="form.companyId" style="width: 100%">
            <el-option v-for="c in companies" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="编码" required>
          <el-input v-model="form.code" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" />
        </el-form-item>
        <el-form-item v-if="editing" label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
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
</style>
