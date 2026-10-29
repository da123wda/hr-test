<template>
  <div class="form-fill-container">
    <el-card class="box-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <h2 class="form-title">{{ schema?.title || '加载中...' }}</h2>
            <p class="form-desc">{{ schema?.description }}</p>
          </div>
          <div class="header-badges">
            <el-tag type="success" effect="dark">
              Published v{{ schema?.publishedVersion }}
            </el-tag>
            <el-tag type="info">FormID: {{ formId }}</el-tag>
          </div>
        </div>
      </template>

      <!-- 快速演示工具条 -->
      <div class="demo-toolbar">
        <span class="demo-tip">🧪 快捷测试用例填充：</span>
        <el-button size="small" @click="fillFreshGraduateDemo">
          用例1: 应届毕业生 (联动隐藏前任公司并取消必填)
        </el-button>
        <el-button size="small" @click="fillExperiencedHighSalaryDemo">
          用例2: 社招资深员工 (期望薪资 35k 触发特别审批)
        </el-button>
        <el-button size="small" @click="fillInvalidEmailDemo">
          用例3: 故意填写非法邮箱 (测试服务端校验拦截)
        </el-button>
      </div>

      <el-divider style="margin: 16px 0" />

      <!-- 核心动态表单渲染引擎 -->
      <FormRenderer
        ref="rendererRef"
        :schema="schema"
        :loading="submitting"
        submit-button-text="提交入职信息"
        @submit="handleFormSubmit"
      />
    </el-card>

    <!-- 提交成功对话框 -->
    <el-dialog
      v-model="successDialogVisible"
      title="提交成功"
      width="480px"
      center
    >
      <el-result
        icon="success"
        title="入职信息已提交并归档"
        :sub-title="`记录编号: ${submittedResult?.submissionId}`"
      >
        <template #extra>
          <div style="margin-top: 10px">
            <el-button type="primary" @click="goToSubmissions">
              前往数据中心查看
            </el-button>
            <el-button @click="successDialogVisible = false">继续填表</el-button>
          </div>
        </template>
      </el-result>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getSchema, submitDynamicData } from '@/api'
import type { FormSchema } from '@/types/schema'
import FormRenderer from '@/components/FormRenderer/FormRenderer.vue'

const router = useRouter()
const formId = ref('tech_onboarding')
const schema = ref<FormSchema>()
const rendererRef = ref()
const submitting = ref(false)
const successDialogVisible = ref(false)
const submittedResult = ref<any>(null)

const loadSchemaData = async () => {
  try {
    const data = await getSchema(formId.value)
    schema.value = data
  } catch (err: any) {
    ElMessage.error('加载表单定义失败: ' + err.message)
  }
}

// 快捷填充用例 1: 应届毕业生
const fillFreshGraduateDemo = () => {
  if (!rendererRef.value) return
  rendererRef.value.formState.fullName = '陈应届'
  rendererRef.value.formState.email = 'chenyingjie@pku.edu.cn'
  rendererRef.value.formState.phone = '13812345678'
  rendererRef.value.formState.department = 'tech_dept'
  rendererRef.value.formState.githubUrl = 'https://github.com/chen-fresh'
  rendererRef.value.formState.isFreshGraduate = 'yes' // 触发联动
  rendererRef.value.formState.previousCompany = '' // 隐藏且不校验必填
  rendererRef.value.formState.experience = 0
  rendererRef.value.formState.expectedSalary = 18000
  rendererRef.value.formState.idCardCopy = {
    fileId: 'file_mock_id_card_01',
    originalName: 'chen_id_card.png',
    fileSize: 1024 * 240,
    downloadUrl: '/api/v1/files/download/file_mock_id_card_01'
  }
  rendererRef.value.formState.workHistory = []
  rendererRef.value.formState.emergencyContact = {
    name: '陈父',
    relation: 'parents',
    contactPhone: '13988887777'
  }
  ElMessage.info('已填充应届生测试数据，前公司名称已自动隐藏并取消必填')
}

// 快捷填充用例 2: 社招资深员工 (35k 高薪审批)
const fillExperiencedHighSalaryDemo = () => {
  if (!rendererRef.value) return
  rendererRef.value.formState.fullName = '马资深'
  rendererRef.value.formState.email = 'mazishen@techlead.com'
  rendererRef.value.formState.phone = '18611112222'
  rendererRef.value.formState.department = 'infra'
  rendererRef.value.formState.githubUrl = 'https://github.com/mazishen-golang'
  rendererRef.value.formState.isFreshGraduate = 'no'
  rendererRef.value.formState.previousCompany = '前沿科技集团'
  rendererRef.value.formState.experience = 7 // experience > 3
  rendererRef.value.formState.expectedSalary = 38000 // 触发 > 30000 审批提醒
  rendererRef.value.formState.idCardCopy = {
    fileId: 'file_mock_id_card_02',
    originalName: 'ma_id_card_scan.pdf',
    fileSize: 1024 * 512,
    downloadUrl: '/api/v1/files/download/file_mock_id_card_02'
  }
  rendererRef.value.formState.workHistory = [
    {
      company: '前沿科技集团',
      role: '分布式架构专家',
      startDate: '2019-06-01',
      endDate: '2024-08-31'
    }
  ]
  rendererRef.value.formState.emergencyContact = {
    name: '马配偶',
    relation: 'spouse',
    contactPhone: '18699990000'
  }
  ElMessage.warning('已填充资深社招数据，期望薪资 38k 已触发 HRBP 特别审批告警！')
}

// 快捷填充用例 3: 故意填写非法邮箱
const fillInvalidEmailDemo = () => {
  if (!rendererRef.value) return
  fillExperiencedHighSalaryDemo()
  rendererRef.value.formState.email = 'not-a-valid-email-address'
  ElMessage.error('已填入非法格式邮箱，提交时将触发服务端精确字段报错')
}

// 表单提交处理 (Q0: 通用提交接口 POST /api/v1/dynamic-data/:formId)
const handleFormSubmit = async (formData: Record<string, any>) => {
  submitting.value = true
  try {
    const res = await submitDynamicData(formId.value, formData)
    submittedResult.value = res
    successDialogVisible.value = true
    ElMessage.success('入职信息提交成功！')
  } catch (err: any) {
    if (err.isValidationError && err.errors) {
      // 服务端字段校验失败回显到表单渲染器中 (Q0 & Q1)
      rendererRef.value?.setServerValidationErrors(err.errors, err.message)
      ElMessage.error('提交失败：' + (err.message || '存在格式校验不合格字段'))
    } else {
      ElMessage.error(err.message || '提交异常')
    }
  } finally {
    submitting.value = false
  }
}

const goToSubmissions = () => {
  successDialogVisible.value = false
  router.push('/submissions')
}

onMounted(() => {
  loadSchemaData()
})
</script>

<style scoped>
.form-fill-container {
  max-width: 900px;
  margin: 0 auto;
}
.box-card {
  border-radius: 10px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.form-title {
  margin: 0 0 6px 0;
  font-size: 20px;
  color: #1e293b;
}
.form-desc {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}
.header-badges {
  display: flex;
  gap: 8px;
}
.demo-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  background: #f1f5f9;
  padding: 10px 14px;
  border-radius: 6px;
}
.demo-tip {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}
</style>
