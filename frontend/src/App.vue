<template>
  <el-container class="app-layout">
    <el-header class="app-header">
      <div class="logo-area">
        <span class="logo-icon">⚡</span>
        <span class="logo-title">HR 动态表单与通用接口引擎</span>
        <el-tag type="info" size="small" effect="plain" class="tech-tag">
          Go (im-auth-go架构) + PostgreSQL JSONB + Vue 3 TS
        </el-tag>
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
          在线填表 (渲染引擎)
        </el-menu-item>
        <el-menu-item index="/builder">
          <el-icon><EditPen /></el-icon>
          表单设计器 (Builder & 乐观锁)
        </el-menu-item>
        <el-menu-item index="/submissions">
          <el-icon><DataAnalysis /></el-icon>
          数据中心 (动态 JSON 筛选)
        </el-menu-item>
        <el-menu-item index="/audit-logs">
          <el-icon><Timer /></el-icon>
          审计日志 (Diff 变更追溯)
        </el-menu-item>
        <el-menu-item index="/versions">
          <el-icon><Files /></el-icon>
          版本生命周期 (Q2 演进)
        </el-menu-item>
      </el-menu>
    </el-header>

    <el-main class="app-main">
      <router-view />
    </el-main>

    <el-footer class="app-footer">
      <div>HR 动态表单与通用接口引擎系统 · 生产级全栈方案</div>
      <div class="footer-links">
        <span>Go 1.27 + Gin + PostgreSQL (pgx/pq) JSONB + GIN 倒排索引</span>
        <span> | </span>
        <span>Vue 3 + TypeScript + Element Plus + DAG 拓扑防死循环</span>
      </div>
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
}
.logo-icon {
  font-size: 22px;
}
.logo-title {
  font-weight: 700;
  font-size: 16px;
  color: #0f172a;
}
.tech-tag {
  margin-left: 8px;
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
  font-size: 12px;
  color: #94a3b8;
  padding: 16px;
  border-top: 1px solid #e2e8f0;
  background: #ffffff;
}
.footer-links {
  margin-top: 4px;
}
</style>
