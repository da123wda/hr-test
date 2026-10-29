<template>
  <div class="repeater-field-container">
    <div class="repeater-header">
      <span class="repeater-title">{{ label }}</span>
      <el-button
        type="primary"
        size="small"
        plain
        :disabled="disabled"
        @click="addItem"
      >
        <el-icon><Plus /></el-icon>
        添加一项
      </el-button>
    </div>

    <div v-if="items.length === 0" class="empty-repeater">
      暂无数据，请点击上方按钮添加
    </div>

    <div v-else class="repeater-list">
      <el-card
        v-for="(item, index) in items"
        :key="index"
        class="repeater-item-card"
        shadow="hover"
      >
        <template #header>
          <div class="item-card-header">
            <span class="item-index-tag">第 {{ index + 1 }} 项</span>
            <el-button
              type="danger"
              link
              size="small"
              :disabled="disabled"
              @click="removeItem(index)"
            >
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </div>
        </template>

        <el-row :gutter="16">
          <el-col
            v-for="subField in itemFields"
            :key="subField.key"
            :span="subField.gridCol || 12"
          >
            <el-form-item
              :label="subField.label"
              :required="subField.validation?.required"
              :error="getFieldError(index, subField.key)"
            >
              <!-- 基础文本 -->
              <el-input
                v-if="subField.type === 'input' || subField.type === 'textarea'"
                v-model="item[subField.key]"
                :type="subField.type === 'textarea' ? 'textarea' : 'text'"
                :placeholder="subField.placeholder"
                :disabled="disabled"
                @input="handleItemChange"
              />

              <!-- 数字 -->
              <el-input-number
                v-else-if="subField.type === 'number'"
                v-model="item[subField.key]"
                :min="subField.validation?.min"
                :max="subField.validation?.max"
                :disabled="disabled"
                style="width: 100%"
                @change="handleItemChange"
              />

              <!-- 日期 -->
              <el-date-picker
                v-else-if="subField.type === 'date'"
                v-model="item[subField.key]"
                type="date"
                value-format="YYYY-MM-DD"
                :placeholder="subField.placeholder || '选择日期'"
                :disabled="disabled"
                style="width: 100%"
                @change="handleItemChange"
              />

              <!-- 下拉选择 -->
              <el-select
                v-else-if="subField.type === 'select'"
                v-model="item[subField.key]"
                :placeholder="subField.placeholder || '请选择'"
                :disabled="disabled"
                style="width: 100%"
                @change="handleItemChange"
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
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { FieldDefinition } from '@/types/schema'

const props = defineProps<{
  label?: string
  fieldKey: string
  itemFields: FieldDefinition[]
  modelValue?: any[]
  disabled?: boolean
  serverErrors?: Record<string, string>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', val: any[]): void
  (e: 'change', val: any[]): void
}>()

const items = computed<any[]>({
  get: () => {
    return Array.isArray(props.modelValue) ? props.modelValue : []
  },
  set: (val) => {
    emit('update:modelValue', val)
    emit('change', val)
  }
})

const addItem = () => {
  const newItem: Record<string, any> = {}
  for (const f of props.itemFields) {
    newItem[f.key] = f.defaultValue ?? ''
  }
  const updated = [...items.value, newItem]
  items.value = updated
}

const removeItem = (idx: number) => {
  const updated = items.value.filter((_, i) => i !== idx)
  items.value = updated
}

const handleItemChange = () => {
  emit('update:modelValue', [...items.value])
  emit('change', [...items.value])
}

const getFieldError = (index: number, subKey: string) => {
  if (!props.serverErrors) return ''
  const path = `${props.fieldKey}[${index}].${subKey}`
  return props.serverErrors[path] || ''
}
</script>

<style scoped>
.repeater-field-container {
  width: 100%;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
  padding: 14px;
  background-color: #fcfcfd;
}
.repeater-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.repeater-title {
  font-weight: 600;
  font-size: 14px;
  color: #334155;
}
.empty-repeater {
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
  padding: 16px 0;
}
.repeater-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.repeater-item-card {
  border-radius: 6px;
}
.item-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.item-index-tag {
  font-size: 12px;
  font-weight: 600;
  color: #2563eb;
  background: #eff6ff;
  padding: 2px 8px;
  border-radius: 4px;
}
</style>
