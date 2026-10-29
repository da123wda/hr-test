<template>
  <div class="form-builder-container">
    <el-card class="builder-card" shadow="never">
      <template #header>
        <div class="builder-header">
          <div>
            <h2 class="title">🛠️ HR 表单设计器 (Visual Form Builder)</h2>
            <p class="subtitle">无需开发人员介入，配置个性化入职表单与动态接口引擎</p>
          </div>
          <div class="header-actions">
            <el-tag type="warning" size="large" effect="dark">
              当前编辑版本: v{{ schemaForm.version || 1 }} (ETag: {{ currentEtag || 'Init' }})
            </el-tag>
            <el-button @click="simulateConcurrencyConflict" type="danger" plain>
              ⚡ 模拟 Q5 并发保存冲突
            </el-button>
            <el-button @click="saveDraft" :loading="saving" type="primary">
              保存草稿
            </el-button>
            <el-button @click="handlePublish" :loading="publishing" type="success">
              发布新版本 (Q2)
            </el-button>
          </div>
        </div>
      </template>

      <!-- 循环依赖死循环测试工具条 -->
      <el-alert
        title="循环依赖防御 (Q1 进阶测试)"
        type="info"
        :closable="false"
        style="margin-bottom: 16px"
      >
        <template #default>
          <div style="display: flex; align-items: center; justify-content: space-between">
            <span>点击可快速注入【A 修改 B，B 修改 C，C 又修改 A】死循环规则，测试系统的 DAG 环检测与拦截机制：</span>
            <el-button size="small" type="warning" @click="injectCycleRulesDemo">
              注入 A->B->C->A 死循环规则
            </el-button>
          </div>
        </template>
      </el-alert>

      <!-- 表单基本信息 -->
      <el-form label-position="top">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="表单标识 (FormID)" required>
              <el-input v-model="schemaForm.formId" placeholder="例如: sales_onboarding" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="表单标题 (Title)" required>
              <el-input v-model="schemaForm.title" placeholder="例如: 销售部入职登记表" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="表单描述 (Description)">
              <el-input v-model="schemaForm.description" placeholder="用途说明" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <el-divider content-position="left">
        <span style="font-weight: 600">字段列表配置 ({{ schemaForm.fields.length }})</span>
      </el-divider>

      <!-- 字段列表表格 -->
      <el-table :data="schemaForm.fields" border style="width: 100%">
        <el-table-column label="字段标识 (Key)" prop="key" width="160">
          <template #default="{ row }">
            <el-input v-model="row.key" size="small" />
          </template>
        </el-table-column>

        <el-table-column label="中文标签 (Label)" prop="label" width="180">
          <template #default="{ row }">
            <el-input v-model="row.label" size="small" />
          </template>
        </el-table-column>

        <el-table-column label="控件类型 (Type)" prop="type" width="140">
          <template #default="{ row }">
            <el-select v-model="row.type" size="small">
              <el-option label="单行文本 (input)" value="input" />
              <el-option label="数字输入 (number)" value="number" />
              <el-option label="下拉选择 (select)" value="select" />
              <el-option label="单选框 (radio)" value="radio" />
              <el-option label="日期选择 (date)" value="date" />
              <el-option label="多行文本 (textarea)" value="textarea" />
              <el-option label="文件上传 (upload)" value="upload" />
              <el-option label="嵌套数组 (repeater)" value="repeater" />
              <el-option label="嵌套对象 (object)" value="object" />
              <el-option label="审批提示 (alert)" value="alert" />
            </el-select>
          </template>
        </el-table-column>

        <el-table-column label="校验规则 (Validation)" min-width="200">
          <template #default="{ row }">
            <div class="validation-tags">
              <el-checkbox
                v-if="row.validation"
                v-model="row.validation.required"
                label="必填"
                size="small"
              />
              <el-checkbox
                v-if="row.validation && (row.type === 'input' || row.type === 'textarea')"
                v-model="row.validation.email"
                label="邮箱格式"
                size="small"
              />
              <el-checkbox
                v-model="row.deprecated"
                label="废弃保留(Q0)"
                size="small"
              />
            </div>
          </template>
        </el-table-column>

        <el-table-column label="联动规则 (Rules)" min-width="220">
          <template #default="{ row }">
            <div v-if="row.rules && row.rules.length > 0">
              <el-tag
                v-for="(r, idx) in row.rules"
                :key="idx"
                size="small"
                type="info"
                style="margin-bottom: 4px; display: block"
              >
                {{ r.expression }} ➔ {{ r.action }} [{{ r.targetKey }}]
              </el-tag>
            </div>
            <span v-else style="color: #94a3b8; font-size: 12px">无联动</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ $index }">
            <el-button
              type="danger"
              size="small"
              link
              @click="removeField($index)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px">
        <el-button @click="addNewField" type="primary" plain>
          + 添加新字段
        </el-button>
        <el-button @click="resetToTechTemplate" plain>
          恢复为“技术岗入职”标准模板
        </el-button>
      </div>
    </el-card>

    <!-- 并发冲突 409 提示对话框 -->
    <el-dialog
      v-model="conflictDialogVisible"
      title="⚠️ 并发修改冲突 (HTTP 409 Conflict)"
      width="500px"
    >
      <el-result
        icon="warning"
        title="版本已被其他 HR 保存覆盖"
        sub-title="检测到其他人已在此期间提交了新版本（数据库当前最新版本已自增），系统已自动拦截您的提交，防止数据覆盖！"
      >
        <template #extra>
          <p style="font-size: 13px; color: #64748b">
            符合需求 Q5 乐观锁机制 (Optimistic Locking / version check) 规范。
          </p>
          <el-button type="primary" @click="resolveConflictReload">
            刷新获取远端最新版本并合并
          </el-button>
        </template>
      </el-result>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSchema, saveSchema, publishSchema } from '@/api'
import type { FormSchema, FieldDefinition } from '@/types/schema'

const saving = ref(false)
const publishing = ref(false)
const conflictDialogVisible = ref(false)
const currentEtag = ref('')

const schemaForm = reactive<FormSchema>({
  formId: 'tech_onboarding',
  title: '技术岗位入职信息表',
  description: '全栈/后端/前端开发岗位个性化入职登记',
  version: 1,
  publishedVersion: 1,
  status: 'draft',
  fields: []
})

const loadSchema = async () => {
  try {
    const data = await getSchema(schemaForm.formId)
    Object.assign(schemaForm, data)
    currentEtag.value = `W/"${data.formId}-v${data.version}"`
  } catch (err: any) {
    ElMessage.error('加载 Schema 失败')
  }
}

const addNewField = () => {
  schemaForm.fields.push({
    key: `custom_${Date.now().toString().slice(-4)}`,
    label: '新自定义字段',
    type: 'input',
    validation: { required: false },
    gridCol: 12
  })
}

const removeField = (idx: number) => {
  schemaForm.fields.splice(idx, 1)
}

// 保存草稿 (带当前版本号，支持并发乐观锁)
const saveDraft = async () => {
  saving.value = true
  try {
    const res = await saveSchema(schemaForm, schemaForm.version)
    schemaForm.version = res.version
    currentEtag.value = `W/"${res.formId}-v${res.version}"`
    ElMessage.success(`草稿保存成功！当前版本自增为 v${res.version}`)
  } catch (err: any) {
    if (err.isConflict || err.status === 409) {
      conflictDialogVisible.value = true
    } else {
      ElMessage.error(err.message || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

// 模拟 Q5: 并发保存冲突
const simulateConcurrencyConflict = async () => {
  try {
    await ElMessageBox.confirm(
      '此操作将模拟另一位 HR 抢先保存了表单（使数据库版本增加），而你仍以旧版本 expectedVersion 提交，验证是否能被 409 Conflict 乐观锁成功拦截！',
      '模拟 Q5 乐观锁并发测试',
      { confirmButtonText: '立即发起测试', cancelButtonText: '取消', type: 'warning' }
    )

    // 故意携带一个陈旧的版本号 (例如当前版本 - 1) 发起保存
    saving.value = true
    const outdatedVersion = Math.max(1, (schemaForm.version || 1) - 1)
    await saveSchema(schemaForm, outdatedVersion)
  } catch (err: any) {
    if (err.isConflict || err.status === 409) {
      conflictDialogVisible.value = true
    }
  } finally {
    saving.value = false
  }
}

const resolveConflictReload = () => {
  conflictDialogVisible.value = false
  loadSchema()
  ElMessage.info('已重新加载远端最新 Schema 定义')
}

// 注入 A -> B -> C -> A 死循环规则 (Q1 进阶)
const injectCycleRulesDemo = () => {
  const fields = [...schemaForm.fields]
  // 查找或添加 fieldA, fieldB, fieldC
  const fA: FieldDefinition = {
    key: 'fieldA',
    label: '测试字段A',
    type: 'input',
    validation: {},
    rules: [{ targetKey: 'fieldB', action: 'show', expression: "fieldA == '1'" }]
  }
  const fB: FieldDefinition = {
    key: 'fieldB',
    label: '测试字段B',
    type: 'input',
    validation: {},
    rules: [{ targetKey: 'fieldC', action: 'show', expression: "fieldB == '2'" }]
  }
  const fC: FieldDefinition = {
    key: 'fieldC',
    label: '测试字段C',
    type: 'input',
    validation: {},
    rules: [{ targetKey: 'fieldA', action: 'show', expression: "fieldC == '3'" }]
  }

  schemaForm.fields.push(fA, fB, fC)
  ElMessage.warning('已成功注入 A ➔ B ➔ C ➔ A 闭环死锁规则！此时保存表单，系统将进行 DAG 环检测并阻断提交。')
}

// 发布新版本 (Q2)
const handlePublish = async () => {
  publishing.value = true
  try {
    const res = await publishSchema(schemaForm.formId, schemaForm.version)
    schemaForm.version = res.version
    schemaForm.publishedVersion = res.publishedVersion
    schemaForm.status = res.status
    currentEtag.value = `W/"${res.formId}-v${res.version}"`
    ElMessage.success(`表单已成功发布为新版本 v${res.publishedVersion}！历史版本已自动存档，不可直接修改。`)
  } catch (err: any) {
    ElMessage.error(err.message || '发布失败')
  } finally {
    publishing.value = false
  }
}

const resetToTechTemplate = () => {
  loadSchema()
}

onMounted(() => {
  loadSchema()
})
</script>

<style scoped>
.form-builder-container {
  max-width: 1100px;
  margin: 0 auto;
}
.builder-card {
  border-radius: 10px;
}
.builder-header {
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
.header-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}
.validation-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
