import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/fill'
  },
  {
    path: '/fill',
    name: 'FormFill',
    component: () => import('@/views/FormFill.vue'),
    meta: { title: '在线入职填表 (渲染引擎)' }
  },
  {
    path: '/builder',
    name: 'FormBuilder',
    component: () => import('@/views/FormBuilder.vue'),
    meta: { title: '表单设计器 (可视化配置)' }
  },
  {
    path: '/submissions',
    name: 'Submissions',
    component: () => import('@/views/Submissions.vue'),
    meta: { title: '数据中心 (通用查询与筛选)' }
  },
  {
    path: '/audit-logs',
    name: 'AuditLogs',
    component: () => import('@/views/AuditLogs.vue'),
    meta: { title: '审计日志 (敏感数据追踪)' }
  },
  {
    path: '/versions',
    name: 'SchemaVersions',
    component: () => import('@/views/SchemaVersions.vue'),
    meta: { title: '版本生命周期 (Q2演进)' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
