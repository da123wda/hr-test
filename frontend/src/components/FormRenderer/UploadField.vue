<template>
  <div class="upload-field-wrapper">
    <!-- 已上传文件展示卡片 -->
    <div v-if="fileMeta && fileMeta.fileId" class="uploaded-card">
      <el-icon class="file-icon"><Document /></el-icon>
      <div class="file-info">
        <div class="file-name" :title="fileMeta.originalName || fileMeta.name">
          {{ fileMeta.originalName || fileMeta.name || '已上传文件' }}
        </div>
        <div class="file-size" v-if="fileMeta.fileSize || fileMeta.size">
          {{ formatSize(fileMeta.fileSize || fileMeta.size) }}
        </div>
      </div>
      <div class="file-actions">
        <el-button
          v-if="fileMeta.downloadUrl"
          type="primary"
          link
          size="small"
          @click="handleDownload"
        >
          查看 / 下载
        </el-button>
        <el-button
          type="danger"
          link
          size="small"
          @click="handleRemove"
          :disabled="disabled"
        >
          移除
        </el-button>
      </div>
    </div>

    <!-- 上传触发器 -->
    <el-upload
      v-else
      class="upload-trigger"
      :http-request="customUpload"
      :show-file-list="false"
      :disabled="disabled || uploading"
      accept=".pdf,.doc,.docx,.xls,.xlsx,.png,.jpg,.jpeg"
    >
      <el-button type="primary" plain :loading="uploading" :disabled="disabled">
        <el-icon><Upload /></el-icon>
        <span>{{ uploading ? '上传中...' : '点击上传文件' }}</span>
      </el-button>
      <template #tip>
        <div class="el-upload__tip">
          支持 PDF、Word、Excel 及常见图片格式，大小不超过 20MB（文件安全存储，仅保留元数据）
        </div>
      </template>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Document, Upload } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { uploadFile } from '@/api'

const props = defineProps<{
  modelValue?: any
  disabled?: boolean
  formId?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: any): void
  (e: 'change', value: any): void
}>()

const uploading = ref(false)

const fileMeta = computed(() => {
  if (!props.modelValue) return null
  if (typeof props.modelValue === 'object') return props.modelValue
  return null
})

const customUpload = async (options: any) => {
  const { file } = options
  uploading.value = true
  try {
    const meta = await uploadFile(file, props.formId)
    ElMessage.success('文件上传成功并已安全归档')
    emit('update:modelValue', meta)
    emit('change', meta)
  } catch (err: any) {
    ElMessage.error(err.message || '文件上传失败')
  } finally {
    uploading.value = false
  }
}

const handleRemove = () => {
  emit('update:modelValue', null)
  emit('change', null)
}

const handleDownload = () => {
  if (fileMeta.value?.downloadUrl) {
    window.open(fileMeta.value.downloadUrl, '_blank')
  }
}

const formatSize = (bytes?: number) => {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>

<style scoped>
.upload-field-wrapper {
  width: 100%;
}
.uploaded-card {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  gap: 12px;
}
.file-icon {
  font-size: 24px;
  color: #3b82f6;
}
.file-info {
  flex: 1;
  min-width: 0;
}
.file-name {
  font-size: 13px;
  color: #1e293b;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.file-size {
  font-size: 11px;
  color: #64748b;
}
.file-actions {
  display: flex;
  gap: 8px;
}
.el-upload__tip {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 4px;
}
</style>
