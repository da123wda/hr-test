<template>
  <div v-if="runtimeState.visible" class="field-renderer-item">
    <!-- 1. Alert 提示文本控件 -->
    <AlertField
      v-if="field.type === 'alert'"
      :title="field.alertConfig?.title || field.label"
      :description="runtimeState.alert || field.placeholder"
      :type="field.alertConfig?.type || 'warning'"
      :visible="!!runtimeState.alert || !field.rules?.length"
    />

    <!-- 2. Repeater 嵌套数组 (如工作履历) -->
    <RepeaterField
      v-else-if="field.type === 'repeater'"
      :label="field.label"
      :field-key="field.key"
      :item-fields="field.itemFields || []"
      :model-value="modelValue"
      :disabled="disabled"
      :server-errors="serverErrors"
      @update:model-value="emitValue"
    />

    <!-- 3. Object 嵌套对象 -->
    <ObjectField
      v-else-if="field.type === 'object'"
      :label="field.label"
      :field-key="field.key"
      :children="field.children || []"
      :model-value="modelValue"
      :disabled="disabled"
      :server-errors="serverErrors"
      @update:model-value="emitValue"
    />

    <!-- 4. 普通单行与标准表单项 -->
    <el-form-item
      v-else
      :label="field.label"
      :required="runtimeState.required"
      :error="errorMessage"
    >
      <template #label>
        <span class="custom-field-label">
          {{ field.label }}
          <el-tag
            v-if="field.deprecated"
            type="info"
            size="small"
            effect="plain"
            style="margin-left: 6px"
          >
            已废弃
          </el-tag>
        </span>
      </template>

      <!-- 文本输入 (input) -->
      <el-input
        v-if="field.type === 'input'"
        :model-value="modelValue"
        :placeholder="field.placeholder || `请输入${field.label}`"
        :disabled="disabled"
        clearable
        @input="emitValue"
      />

      <!-- 多行文本 (textarea) -->
      <el-input
        v-else-if="field.type === 'textarea'"
        :model-value="modelValue"
        type="textarea"
        :rows="3"
        :placeholder="field.placeholder || `请输入${field.label}`"
        :disabled="disabled"
        @input="emitValue"
      />

      <!-- 数字输入 (number) -->
      <el-input-number
        v-else-if="field.type === 'number'"
        :model-value="modelValue"
        :min="field.validation?.min"
        :max="field.validation?.max"
        :placeholder="field.placeholder"
        :disabled="disabled"
        style="width: 100%"
        @change="emitValue"
      />

      <!-- 下拉选择 (select) -->
      <el-select
        v-else-if="field.type === 'select'"
        :model-value="modelValue"
        :placeholder="field.placeholder || '请选择'"
        :disabled="disabled"
        clearable
        style="width: 100%"
        @change="emitValue"
      >
        <el-option
          v-for="opt in field.options || []"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>

      <!-- 单选框 (radio) -->
      <el-radio-group
        v-else-if="field.type === 'radio'"
        :model-value="modelValue"
        :disabled="disabled"
        @change="emitValue"
      >
        <el-radio
          v-for="opt in field.options || []"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.label }}
        </el-radio>
      </el-radio-group>

      <!-- 多选框 (checkbox) -->
      <el-checkbox-group
        v-else-if="field.type === 'checkbox'"
        :model-value="modelValue || []"
        :disabled="disabled"
        @change="emitValue"
      >
        <el-checkbox
          v-for="opt in field.options || []"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.label }}
        </el-checkbox>
      </el-checkbox-group>

      <!-- 日期选择 (date) -->
      <el-date-picker
        v-else-if="field.type === 'date'"
        :model-value="modelValue"
        type="date"
        value-format="YYYY-MM-DD"
        :placeholder="field.placeholder || '选择日期'"
        :disabled="disabled"
        style="width: 100%"
        @change="emitValue"
      />

      <!-- 文件上传 (upload) -->
      <UploadField
        v-else-if="field.type === 'upload'"
        :model-value="modelValue"
        :disabled="disabled"
        :form-id="formId"
        @update:model-value="emitValue"
      />
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FieldDefinition } from '@/types/schema'
import type { FieldRuntimeState } from './RuleEngine'
import AlertField from './AlertField.vue'
import UploadField from './UploadField.vue'
import RepeaterField from './RepeaterField.vue'
import ObjectField from './ObjectField.vue'

const props = defineProps<{
  field: FieldDefinition
  modelValue?: any
  runtimeState: FieldRuntimeState
  disabled?: boolean
  formId?: string
  serverError?: string
  serverErrors?: Record<string, string>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', val: any): void
  (e: 'change', key: string, val: any): void
}>()

const errorMessage = computed(() => {
  return props.serverError || ''
})

const emitValue = (val: any) => {
  emit('update:modelValue', val)
  emit('change', props.field.key, val)
}
</script>

<style scoped>
.field-renderer-item {
  width: 100%;
}
.custom-field-label {
  display: inline-flex;
  align-items: center;
}
</style>
