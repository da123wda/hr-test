<template>
  <div class="dynamic-form-renderer">
    <!-- 表单级错误警示条 (Form Level Error Alert) -->
    <el-alert
      v-if="formLevelError"
      :title="formLevelError"
      type="error"
      show-icon
      class="form-level-alert"
      :closable="true"
      @close="formLevelError = ''"
    />

    <!-- 循环依赖检测告警条 (Rule Engine Cycle Protection) -->
    <el-alert
      v-if="cycleErrorNotice"
      :title="cycleErrorNotice"
      type="warning"
      show-icon
      class="form-level-alert"
      description="系统已通过 DAG 拓扑检测自动阻断循环求值，避免浏览器无限死循环。"
    />

    <el-form
      ref="formRef"
      :model="formState"
      label-position="top"
      :disabled="loading || disabled"
      class="renderer-form"
    >
      <el-row :gutter="20">
        <el-col
          v-for="field in renderFields"
          :key="field.key"
          :span="field.gridCol || 24"
        >
          <FieldRenderer
            :field="field"
            :model-value="formState[field.key]"
            :runtime-state="getFieldRuntimeState(field.key)"
            :disabled="disabled"
            :form-id="schema?.formId"
            :server-error="serverValidationErrors[field.key]"
            :server-errors="serverValidationErrors"
            @update:model-value="(val) => updateFieldValue(field.key, val)"
            @change="handleFieldChange"
          />
        </el-col>
      </el-row>

      <!-- 表单底部操作槽 (提交/重置/脏状态提示) -->
      <div v-if="showFooter" class="form-footer-actions">
        <div class="dirty-status-indicator">
          <el-tag :type="isDirty ? 'warning' : 'info'" size="small" effect="plain">
            {{ isDirty ? '● 表单已修改 (Dirty)' : '○ 表单未修改 (Clean)' }}
          </el-tag>
        </div>
        <div class="actions-right">
          <el-button @click="resetForm" :disabled="loading">重置</el-button>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleSubmit"
          >
            {{ submitButtonText || '立即提交' }}
          </el-button>
        </div>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import type { FormSchema, FieldDefinition } from '@/types/schema'
import { FrontendRuleEngine, type FieldRuntimeState } from './RuleEngine'
import FieldRenderer from './FieldRenderer.vue'

const props = withDefaults(
  defineProps<{
    schema?: FormSchema
    modelValue?: Record<string, any>
    disabled?: boolean
    loading?: boolean
    showFooter?: boolean
    submitButtonText?: string
    showDeprecated?: boolean // 是否展示已废弃的历史字段 (默认表单不展示，历史详情可开启)
  }>(),
  {
    showFooter: true,
    showDeprecated: false
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: Record<string, any>): void
  (e: 'submit', val: Record<string, any>): void
  (e: 'change', key: string, val: any): void
}>()

const formRef = ref()
const ruleEngine = new FrontendRuleEngine()

// 1. formState 表单数据
const formState = reactive<Record<string, any>>({})
// 初始状态镜像，用于计算 dirtyState
const initialState = ref<Record<string, any>>({})
const dirtyKeys = ref<Set<string>>(new Set())

// 2. 错误状态
const formLevelError = ref('')
const cycleErrorNotice = ref('')
const serverValidationErrors = ref<Record<string, string>>({})

// 3. 字段运行时动态状态 (由规则引擎计算得出)
const runtimeStates = ref<Record<string, FieldRuntimeState>>({})

// 过滤后的字段 (废弃字段默认在填表时不渲染)
const renderFields = computed(() => {
  if (!props.schema?.fields) return []
  return props.schema.fields.filter((f) => {
    if (f.deprecated && !props.showDeprecated) {
      return false
    }
    return true
  })
})

// dirtyState 计算属性
const isDirty = computed(() => dirtyKeys.value.size > 0)

// 获取字段运行时状态
const getFieldRuntimeState = (key: string): FieldRuntimeState => {
  return (
    runtimeStates.value[key] || {
      visible: true,
      required: false,
      alert: ''
    }
  )
}

// 初始化表单状态
const initFormState = () => {
  if (!props.schema) return

  // 清空
  for (const k of Object.keys(formState)) {
    delete formState[k]
  }
  dirtyKeys.value.clear()
  serverValidationErrors.value = {}
  formLevelError.value = ''

  for (const f of props.schema.fields) {
    if (props.modelValue && props.modelValue[f.key] !== undefined) {
      formState[f.key] = JSON.parse(JSON.stringify(props.modelValue[f.key]))
    } else if (f.defaultValue !== undefined) {
      formState[f.key] = JSON.parse(JSON.stringify(f.defaultValue))
    } else {
      if (f.type === 'repeater') {
        formState[f.key] = []
      } else if (f.type === 'object') {
        formState[f.key] = {}
      } else if (f.type === 'checkbox') {
        formState[f.key] = []
      } else {
        formState[f.key] = ''
      }
    }
  }

  initialState.value = JSON.parse(JSON.stringify(formState))
  evaluateRules()
  emit('update:modelValue', { ...formState })
}

// 执行联动规则求值
const evaluateRules = () => {
  if (!props.schema?.fields) return
  const { states, cycleError } = ruleEngine.evaluateRules(props.schema.fields, formState)
  runtimeStates.value = states
  cycleErrorNotice.value = cycleError || ''
}

// 字段更新
const updateFieldValue = (key: string, val: any) => {
  formState[key] = val
  // 清除该字段上的服务端错误
  if (serverValidationErrors.value[key]) {
    delete serverValidationErrors.value[key]
  }

  // 脏状态检查
  const initVal = initialState.value[key]
  if (JSON.stringify(initVal) !== JSON.stringify(val)) {
    dirtyKeys.value.add(key)
  } else {
    dirtyKeys.value.delete(key)
  }

  evaluateRules()
  emit('update:modelValue', { ...formState })
}

const handleFieldChange = (key: string, val: any) => {
  emit('change', key, val)
}

// 前端基础校验
const validate = (): boolean => {
  serverValidationErrors.value = {}
  formLevelError.value = ''
  let hasError = false
  const errMap: Record<string, string> = {}

  if (!props.schema) return true

  for (const field of renderFields.value) {
    const state = getFieldRuntimeState(field.key)
    // 隐藏的字段跳过校验
    if (!state.visible) continue

    const val = formState[field.key]
    const isEmpty = val === undefined || val === null || val === '' || (Array.isArray(val) && val.length === 0)

    // 必填
    if (state.required && isEmpty) {
      errMap[field.key] = field.validation?.customMessage || `${field.label} 为必填项`
      hasError = true
      continue
    }

    if (isEmpty) continue

    // 邮箱格式
    if (field.validation?.email && typeof val === 'string') {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      if (!emailRegex.test(val)) {
        errMap[field.key] = `字段 ${field.label} 格式不正确`
        hasError = true
        continue
      }
    }

    // 正则
    if (field.validation?.regex && typeof val === 'string') {
      try {
        const re = new RegExp(field.validation.regex)
        if (!re.test(val)) {
          errMap[field.key] = field.validation.customMessage || `${field.label} 格式不正确`
          hasError = true
          continue
        }
      } catch (e) {
        // 忽略无效正则
      }
    }

    // Repeater 嵌套子项必填校验
    if (field.type === 'repeater' && Array.isArray(val) && field.itemFields) {
      val.forEach((item, idx) => {
        for (const subF of field.itemFields || []) {
          if (subF.validation?.required) {
            const subV = item[subF.key]
            if (subV === undefined || subV === null || subV === '') {
              errMap[`${field.key}[${idx}].${subF.key}`] = `${subF.label} 为必填项`
              hasError = true
            }
          }
        }
      })
    }
  }

  if (hasError) {
    serverValidationErrors.value = errMap
    formLevelError.value = '表单包含未通过的校验项，请检查标红字段'
    return false
  }

  return true
}

const handleSubmit = () => {
  if (!validate()) return
  emit('submit', { ...formState })
}

const resetForm = () => {
  initFormState()
}

// 暴露服务端校验错误回显方法
const setServerValidationErrors = (errors: Record<string, string>, message?: string) => {
  serverValidationErrors.value = errors || {}
  formLevelError.value = message || '服务端数据校验未通过，请检查对应字段'
}

// 监听 Schema 或 modelValue 变化重置表单
watch(
  () => props.schema,
  () => {
    initFormState()
  },
  { immediate: true, deep: true }
)

watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal && Object.keys(newVal).length > 0) {
      for (const k of Object.keys(newVal)) {
        formState[k] = newVal[k]
      }
      evaluateRules()
    }
  },
  { deep: true }
)

defineExpose({
  formState,
  isDirty,
  validate,
  resetForm,
  setServerValidationErrors
})
</script>

<style scoped>
.dynamic-form-renderer {
  width: 100%;
}
.form-level-alert {
  margin-bottom: 20px;
}
.renderer-form {
  width: 100%;
}
.form-footer-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #f1f5f9;
}
.actions-right {
  display: flex;
  gap: 12px;
}
</style>
