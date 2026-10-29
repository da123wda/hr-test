<template>
  <div class="schema-versions-container">
    <el-card class="box-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <h2 class="title">版本管理</h2>
            <p class="subtitle">表单历史归档快照与版本演进记录</p>
          </div>
          <el-button type="primary" plain @click="fetchVersions">
            刷新版本列表
          </el-button>
        </div>
      </template>

      <el-alert
        title="版本演进与历史数据兼容说明"
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template #default>
          <div style="font-size: 13px; line-height: 1.8">
            <p><strong>【背景】</strong>某表单版本为 v3，已有 10,000 条历史记录。HR 将 <code>employeeName</code> 改为 <code>fullName</code>、删除 <code>nickname</code>、新增必填 <code>department</code>。</p>
            <p><strong>【核心策略与实现】</strong></p>
            <ol style="margin: 0; padding-left: 20px">
              <li><strong>不可变快照 (Immutable Snapshots)：</strong>已发布的 Schema 严禁直接更新，全部版本写入 <code>form_schema_versions</code> 快照表，保留历史原貌；</li>
              <li><strong>版本强绑定 (Version Binding)：</strong>每条提交数据 <code>form_submissions</code> 严格锁定当时的 <code>schema_version</code>，老数据永远基于老 Schema 渲染；</li>
              <li><strong>废弃字段软隐藏 (Field Deprecation)：</strong>删除字段（如 <code>nickname</code>）时标记为 <code>deprecated: true</code>，新表单不渲染，但历史数据详情仍能正常解析展示；</li>
              <li><strong>迁移映射兼容 (Migrations Mapping)：</strong>通过 <code>migrations: { "employeeName": "fullName" }</code> 进行字段重命名平滑兼容。</li>
            </ol>
          </div>
        </template>
      </el-alert>

      <!-- 历史版本列表 -->
      <el-table
        v-loading="loading"
        :data="versionList"
        border
        style="width: 100%"
      >
        <el-table-column label="版本号" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="large" type="success" effect="dark">
              v{{ row.version }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="表单标题" prop="title" min-width="200" />

        <el-table-column label="包含字段数" width="130" align="center">
          <template #default="{ row }">
            {{ row.fields?.length || 0 }} 个字段
          </template>
        </el-table-column>

        <el-table-column label="发布人" prop="publishedBy" width="150" />

        <el-table-column label="发布归档时间" width="200" prop="publishedAt">
          <template #default="{ row }">
            {{ formatDate(row.publishedAt) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="viewVersionDetails(row)">
              查看此版本 Schema
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 版本详情弹窗 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="`版本 v${selectedVersion?.version} 历史归档快照`"
      width="750px"
    >
      <div v-if="selectedVersion">
        <el-descriptions :column="2" border style="margin-bottom: 16px">
          <el-descriptions-item label="表单标识">{{ selectedVersion.formId }}</el-descriptions-item>
          <el-descriptions-item label="版本号">v{{ selectedVersion.version }}</el-descriptions-item>
          <el-descriptions-item label="表单标题" :span="2">{{ selectedVersion.title }}</el-descriptions-item>
          <el-descriptions-item label="发布归档时间" :span="2">{{ formatDate(selectedVersion.publishedAt) }}</el-descriptions-item>
        </el-descriptions>

        <h4>本版本字段定义 (只读不可篡改)：</h4>
        <el-table :data="selectedVersion.fields" border size="small" max-height="350">
          <el-table-column label="Key" prop="key" width="140" />
          <el-table-column label="Label" prop="label" width="150" />
          <el-table-column label="Type" prop="type" width="100" />
          <el-table-column label="属性" min-width="120">
            <template #default="{ row }">
              <el-tag v-if="row.validation?.required" size="small" type="danger" style="margin-right: 4px">必填</el-tag>
              <el-tag v-if="row.deprecated" size="small" type="info">已废弃</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listSchemaVersions } from '@/api'
import type { FormSchemaVersion } from '@/types/schema'

const formId = 'tech_onboarding'
const loading = ref(false)
const versionList = ref<FormSchemaVersion[]>([])
const selectedVersion = ref<FormSchemaVersion | null>(null)
const detailDialogVisible = ref(false)

const fetchVersions = async () => {
  loading.value = true
  try {
    const list = await listSchemaVersions(formId)
    versionList.value = list
  } catch (err: any) {
    ElMessage.error('获取历史版本失败: ' + err.message)
  } finally {
    loading.value = false
  }
}

const viewVersionDetails = (row: FormSchemaVersion) => {
  selectedVersion.value = row
  detailDialogVisible.value = true
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN', { hour12: false })
}

onMounted(() => {
  fetchVersions()
})
</script>

<style scoped>
.schema-versions-container {
  max-width: 1100px;
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
</style>
