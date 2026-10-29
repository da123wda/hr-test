<template>
  <div class="object-field-container">
    <div class="object-header">
      <span class="object-title">{{ label }}</span>
    </div>
    <el-card shadow="never" class="object-card">
      <el-row :gutter="16">
        <el-col
          v-for="subField in children"
          :key="subField.key"
          :span="subField.gridCol || 12"
        >
          <el-form-item
            :label="subField.label"
            :required="subField.validation?.required"
            :error="getFieldError(subField.key)"
          >
            <!-- 文本输入 -->
            <el-input
              v-if="subField.type === 'input' || subField.type === 'textarea'"
              v-model="objData[subField.key]"
              :type="subField.type === 'textarea' ? 'textarea' : 'text'"
              :placeholder="subField.placeholder"
              :disabled="disabled"
              @input="handleChange"
            />

            <!-- 数字 -->
            <el-input-number
              v-else-if="subField.type === 'number'"
              v-model="objData[subField.key]"
              :disabled="disabled"
              style="width: 100%"
              @change="handleChange"
            />

            <!-- 下拉 -->
            <el-select
              v-else-if="subField.type === 'select'"
              v-model="objData[subField.key]"
              :placeholder="subField.placeholder || '请选择'"
              :disabled="disabled"
              style="width: 100%"
              @change="handleChange"
            >
              <el-option
                v-for="opt in subField.options || []"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FieldDefinition } from '@/types/schema'

const props = defineProps<{
  label?: string
  fieldKey: string
  children: FieldDefinition[]
  modelValue?: Record<string, any>
  disabled?: boolean
  serverErrors?: Record<string, string>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', val: Record<string, any>): void
  (e: 'change', val: Record<string, any>): void
}>()

const objData = computed<Record<string, any>>({
  get: () => {
    return props.modelValue && typeof props.modelValue === 'object' ? props.modelValue : {}
  },
  set: (val) => {
    emit('update:modelValue', val)
    emit('change', val)
  }
})

const handleChange = () => {
  emit('update:modelValue', { ...objData.value })
  emit('change', { ...objData.value })
}

const getFieldError = (subKey: string) => {
  if (!props.serverErrors) return ''
  const path = `${props.fieldKey}.${subKey}`
  return props.serverErrors[path] || ''
}
</script>

<style scoped>
.object-field-container {
  width: 100%;
}
.object-header {
  margin-bottom: 8px;
}
.object-title {
  font-weight: 600;
  font-size: 13px;
  color: #475569;
}
.object-card {
  background: #f8fafc;
  border-color: #e2e8f0;
}
</style>
