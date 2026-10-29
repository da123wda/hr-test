<template>
  <div class="audit-logs-container">
    <el-card class="box-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <h2 class="title">审计日志</h2>
            <p class="subtitle">敏感字段变更明细与操作追溯</p>
          </div>
          <el-button type="primary" plain @click="fetchLogs">
            刷新日志
          </el-button>
        </div>
      </template>

      <!-- 审计日志表格 -->
      <el-table
        v-loading="loading"
        :data="logList"
        border
        style="width: 100%"
      >
        <el-table-column label="日志ID" prop="id" width="90" align="center" />

        <el-table-column label="操作类型" width="160">
          <template #default="{ row }">
            <el-tag :type="getActionTagType(row.action)">
              {{ formatAction(row.action) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="目标对象" width="180">
          <template #default="{ row }">
            <div>
              <span style="font-weight: 600">{{ row.targetType }}</span>: 
              <span style="font-family: monospace">{{ row.targetId }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作人与客户端 IP" width="220">
          <template #default="{ row }">
            <div>
              <div><strong>{{ row.operatorName || row.operatorId || '系统' }}</strong></div>
              <div style="font-size: 12px; color: #64748b">IP: {{ row.clientIp || '127.0.0.1' }}</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="字段级修改明细 (Diff 比对)" min-width="320">
          <template #default="{ row }">
            <div v-if="row.diff && Object.keys(row.diff).length > 0" class="diff-container">
              <div
                v-for="(diffItem, fieldKey) in row.diff"
                :key="fieldKey"
                class="diff-row"
              >
                <span class="diff-field-name">{{ fieldKey }}:</span>
                <span class="diff-old-val">[{{ formatVal(diffItem.old) }}]</span>
                <span class="diff-arrow">➔</span>
                <span class="diff-new-val">[{{ formatVal(diffItem.new) }}]</span>
              </div>
            </div>
            <span v-else style="color: #94a3b8; font-size: 12px">
              {{ row.action === 'CREATE_SUBMISSION' ? '创建全新提交记录' : '无字段变更记录' }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="操作时间" width="180" prop="createdAt">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          layout="total, prev, pager, next"
          @current-change="fetchLogs"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { queryAuditLogs } from '@/api'
import type { AuditLog } from '@/types/schema'

const loading = ref(false)
const logList = ref<AuditLog[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 15,
  total: 0
})

const fetchLogs = async () => {
  loading.value = true
  try {
    const res = await queryAuditLogs({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    logList.value = res.items || []
    pagination.total = res.total || 0
  } catch (err: any) {
    ElMessage.error('获取审计日志失败: ' + err.message)
  } finally {
    loading.value = false
  }
}

const formatAction = (action: string) => {
  const map: Record<string, string> = {
    CREATE_SCHEMA: '创建表单定义',
    UPDATE_SCHEMA: '更新表单草稿',
    PUBLISH_SCHEMA: '发布新版本',
    CREATE_SUBMISSION: '新建入职提交',
    UPDATE_SUBMISSION: '修改入职信息',
    DELETE_SUBMISSION: '删除入职提交'
  }
  return map[action] || action
}

const getActionTagType = (action: string) => {
  if (action.includes('CREATE')) return 'success'
  if (action.includes('UPDATE')) return 'warning'
  if (action.includes('DELETE')) return 'danger'
  if (action.includes('PUBLISH')) return 'primary'
  return 'info'
}

const formatVal = (v: any) => {
  if (v === undefined || v === null || v === '') return '空'
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN', { hour12: false })
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.audit-logs-container {
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
.diff-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.diff-row {
  font-size: 13px;
  font-family: monospace;
  background: #f8fafc;
  padding: 4px 8px;
  border-radius: 4px;
  border-left: 3px solid #3b82f6;
}
.diff-field-name {
  font-weight: bold;
  color: #1e293b;
  margin-right: 6px;
}
.diff-old-val {
  color: #dc2626;
  text-decoration: line-through;
  background: #fee2e2;
  padding: 1px 4px;
  border-radius: 2px;
}
.diff-arrow {
  margin: 0 6px;
  color: #64748b;
}
.diff-new-val {
  color: #16a34a;
  font-weight: bold;
  background: #dcfce7;
  padding: 1px 4px;
  border-radius: 2px;
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
