<template>
  <el-container class="app-layout">
    <el-header class="app-header">
      <div class="logo-area">
        <span class="logo-icon">⚡</span>
        <span class="logo-title">HR 动态表单系统</span>
      </div>

      <el-menu
        :default-active="activeMenu"
        mode="horizontal"
        router
        class="nav-menu"
        :ellipsis="false"
      >
        <el-menu-item index="/fill">
          <el-icon><DocumentChecked /></el-icon>
          在线填表
        </el-menu-item>
        <el-menu-item index="/builder">
          <el-icon><EditPen /></el-icon>
          表单设计
        </el-menu-item>
        <el-menu-item index="/submissions">
          <el-icon><DataAnalysis /></el-icon>
          数据中心
        </el-menu-item>
        <el-menu-item index="/audit-logs">
          <el-icon><Timer /></el-icon>
          审计日志
        </el-menu-item>
        <el-menu-item index="/versions">
          <el-icon><Files /></el-icon>
          版本管理
        </el-menu-item>
      </el-menu>
    </el-header>

    <el-main class="app-main">
      <router-view />
    </el-main>

    <el-footer class="app-footer">
      <div>HR 动态表单与通用接口引擎系统</div>
    </el-footer>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  DocumentChecked,
  EditPen,
  DataAnalysis,
  Timer,
  Files
} from '@element-plus/icons-vue'

const route = useRoute()
const activeMenu = computed(() => route.path)
</script>

<style>
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f8fafc;
  color: #1e293b;
}
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.app-header {
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 60px;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.03);
}
.logo-area {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
  white-space: nowrap;
}
.logo-icon {
  font-size: 20px;
}
.logo-title {
  font-weight: 700;
  font-size: 16px;
  color: #0f172a;
  white-space: nowrap;
}
.nav-menu {
  border-bottom: none !important;
  height: 100%;
}
.app-main {
  flex: 1;
  padding: 24px;
}
.app-footer {
  text-align: center;
  font-size: 13px;
  color: #94a3b8;
  padding: 16px;
  border-top: 1px solid #e2e8f0;
  background: #ffffff;
}
</style>
