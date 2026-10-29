<template>
  <div class="submissions-container">
    <el-card class="box-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <h2 class="title">数据中心</h2>
            <p class="subtitle">动态表单数据查询与多维字段筛选</p>
          </div>
          <el-button type="primary" plain @click="fetchData">
            刷新列表
          </el-button>
        </div>
      </template>

      <el-form :inline="true" class="filter-form">
        <el-form-item label="工作年限">
          <el-select
            v-model="filters.experienceOp"
            placeholder="筛选条件"
            style="width: 160px"
            @change="handleFilterChange"
          >
            <el-option label="全部工作年限" value="" />
            <el-option label="大于 3 年" value="gt:3" />
            <el-option label="大于等于 5 年" value="gte:5" />
            <el-option label="应届 / 1 年以内" value="lte:1" />
          </el-select>
        </el-form-item>

        <el-form-item label="所属部门">
          <el-select
            v-model="filters.department"
            placeholder="按部门筛选"
            clearable
            style="width: 140px"
            @change="handleFilterChange"
          >
            <el-option label="基础架构部" value="infra" />
            <el-option label="后端研发部" value="tech_dept" />
            <el-option label="前端中台组" value="fe_team" />
            <el-option label="AI 智能实验室" value="ai_lab" />
          </el-select>
        </el-form-item>

        <el-form-item label="是否应届生">
          <el-select
            v-model="filters.isFreshGraduate"
            placeholder="是否应届生"
            clearable
            style="width: 130px"
            @change="handleFilterChange"
          >
            <el-option label="是 (yes)" value="yes" />
            <el-option label="否 (no)" value="no" />
          </el-select>
        </el-form-item>

        <el-form-item label="全局搜索">
          <el-input
            v-model="filters.keyword"
            placeholder="搜索姓名、邮箱、前公司等"
            clearable
            @keyup.enter="handleFilterChange"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleFilterChange">查询</el-button>
          <el-button @click="resetFilters">重置筛选</el-button>
        </el-form-item>
      </el-form>

      <!-- 提交记录表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        style="width: 100%"
      >
        <el-table-column label="提交ID" prop="id" width="130" />
        
        <el-table-column label="Schema版本" width="110" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="success">v{{ row.schemaVersion }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="姓名" min-width="110">
          <template #default="{ row }">
            <strong>{{ row.data.fullName || row.data.employeeName || '-' }}</strong>
          </template>
        </el-table-column>

        <el-table-column label="工作邮箱" min-width="170">
          <template #default="{ row }">
            {{ row.data.email || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="部门" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ formatDept(row.data.department) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="应届生" width="90" align="center">
          <template #default="{ row }">
            <el-tag
              size="small"
              :type="row.data.isFreshGraduate === 'yes' ? 'warning' : 'info'"
            >
              {{ row.data.isFreshGraduate === 'yes' ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="工作年限" width="100" align="center" sortable prop="data.experience">
          <template #default="{ row }">
            <span :style="{ fontWeight: row.data.experience > 3 ? 'bold' : 'normal', color: row.data.experience > 3 ? '#2563eb' : 'inherit' }">
              {{ row.data.experience ?? '-' }} 年
            </span>
          </template>
        </el-table-column>

        <el-table-column label="期望薪资" width="120" align="right" sortable prop="data.expectedSalary">
          <template #default="{ row }">
            <span :style="{ color: row.data.expectedSalary > 30000 ? '#ea580c' : 'inherit', fontWeight: 'bold' }">
              ¥ {{ (row.data.expectedSalary || 0).toLocaleString() }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="附件" width="80" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.data.idCardCopy?.downloadUrl"
              link
              type="primary"
              size="small"
              @click="previewFile(row.data.idCardCopy.downloadUrl)"
            >
              材料
            </el-button>
            <span v-else style="color: #cbd5e1">-</span>
          </template>
        </el-table-column>

        <el-table-column label="提交时间" width="170" prop="createdAt">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEditDialog(row)">
              编辑 (触发审计)
            </el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页组件 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[5, 10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <!-- 编辑并触发审计日志弹窗 (Q6: 修改薪资 salary 30000 -> 35000) -->
    <el-dialog
      v-model="editDialogVisible"
      title="✏️ 修改员工信息 (系统将自动记录前后 Diff 审计日志)"
      width="600px"
    >
      <el-form v-if="editingRow" label-position="top">
        <el-form-item label="姓名">
          <el-input v-model="editForm.fullName" />
        </el-form-item>
        <el-form-item label="所属部门">
          <el-select v-model="editForm.department" style="width: 100%">
            <el-option label="基础架构部" value="infra" />
            <el-option label="后端研发部" value="tech_dept" />
            <el-option label="前端中台组" value="fe_team" />
            <el-option label="AI 智能实验室" value="ai_lab" />
          </el-select>
        </el-form-item>
        <el-form-item label="期望薪资 (元/月)">
          <el-input-number v-model="editForm.expectedSalary" :min="1000" :max="200000" style="width: 100%" />
          <div style="font-size: 12px; color: #ea580c; margin-top: 4px">
            💡 提示：若将薪资由 30,000 修改为 35,000，保存后前往“审计日志”中心即可查看精确字段差异对比。
          </div>
        </el-form-item>
        <el-form-item label="工作年限">
          <el-input-number v-model="editForm.experience" :min="0" :max="50" style="width: 100%" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingEdit" @click="submitEdit">
          保存修改并记录审计
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { queryDynamicData, updateSubmission, deleteSubmission } from '@/api'
import type { FormSubmission } from '@/types/schema'

const router = useRouter()
const formId = 'tech_onboarding'
const loading = ref(false)
const tableData = ref<FormSubmission[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const filters = reactive({
  experienceOp: '',
  department: '',
  isFreshGraduate: '',
  keyword: ''
})

const editDialogVisible = ref(false)
const savingEdit = ref(false)
const editingRow = ref<FormSubmission | null>(null)
const editForm = reactive<Record<string, any>>({})

const fetchData = async () => {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      pageSize: pagination.pageSize
    }

    if (filters.keyword) {
      params.keyword = filters.keyword
    }
    if (filters.department) {
      params['filter[department]'] = filters.department
    }
    if (filters.isFreshGraduate) {
      params['filter[isFreshGraduate]'] = filters.isFreshGraduate
    }
    // experience > 3 动态过滤 (Q1)
    if (filters.experienceOp) {
      params['filter[experience]'] = filters.experienceOp
    }

    const res = await queryDynamicData(formId, params)
    tableData.value = res.items || []
    pagination.total = res.total || 0
  } catch (err: any) {
    ElMessage.error('查询数据失败: ' + err.message)
  } finally {
    loading.value = false
  }
}

const handleFilterChange = () => {
  pagination.page = 1
  fetchData()
}

const resetFilters = () => {
  filters.experienceOp = ''
  filters.department = ''
  filters.isFreshGraduate = ''
  filters.keyword = ''
  handleFilterChange()
}

const openEditDialog = (row: FormSubmission) => {
  editingRow.value = row
  Object.assign(editForm, JSON.parse(JSON.stringify(row.data)))
  editDialogVisible.value = true
}

const submitEdit = async () => {
  if (!editingRow.value) return
  savingEdit.value = true
  try {
    await updateSubmission(formId, editingRow.value.id, editForm)
    ElMessage.success('更新成功！变更明细已写入审计日志。')
    editDialogVisible.value = false
    fetchData()
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败')
  } finally {
    savingEdit.value = false
  }
}

const handleDelete = async (row: FormSubmission) => {
  try {
    await ElMessageBox.confirm('确定要删除这条提交记录吗？操作将记录到审计日志。', '删除确认', {
      type: 'warning'
    })
    await deleteSubmission(formId, row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    // 取消
  }
}

const previewFile = (url: string) => {
  window.open(url, '_blank')
}

const formatDept = (d?: string) => {
  const map: Record<string, string> = {
    infra: '基础架构部',
    tech_dept: '后端研发部',
    fe_team: '前端中台组',
    ai_lab: 'AI实验室'
  }
  return (d && map[d]) || d || '-'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN', { hour12: false })
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.submissions-container {
  max-width: 1200px;
  margin: 0 auto;
}
.box-card {
  border-radius: 10px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.title {
  margin: 0 0 6px 0;
  font-size: 20px;
  color: #0f172a;
}
.subtitle {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}
.filter-form {
  background: #f8fafc;
  padding: 14px 14px 2px 14px;
  border-radius: 8px;
  margin-bottom: 16px;
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
