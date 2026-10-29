<template>
  <div v-if="visible" class="alert-field-container">
    <el-alert
      :title="title"
      :description="description"
      :type="alertType"
      show-icon
      :closable="false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title?: string
  description?: string
  type?: 'info' | 'warning' | 'success' | 'error'
  visible?: boolean
}>()

const alertType = computed(() => props.type || 'warning')
</script>

<style scoped>
.alert-field-container {
  margin: 8px 0 16px 0;
  width: 100%;
}
</style>
