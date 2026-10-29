import apiClient from './client'
import type { FormSchema, FormSchemaVersion, FormSubmission, AuditLog } from '@/types/schema'

// 1. Schema 相关
// 获取对外发布的正式版本定义 (面向候选人填表)
export const getSchema = async (formId: string) => {
  const res = await apiClient.get<{ code: number; data: FormSchema }>(`/schemas/${formId}`)
  return res.data.data
}

// 获取 HR 工作台草稿 (面向表单设计器)
export const getDraftSchema = async (formId: string) => {
  const res = await apiClient.get<{ code: number; data: FormSchema }>(`/schemas/${formId}/draft`)
  return res.data.data
}

export const listSchemas = async () => {
  const res = await apiClient.get<{ code: number; data: FormSchema[] }>(`/schemas`)
  return res.data.data
}

export const saveSchema = async (data: Partial<FormSchema>, expectedVersion?: number) => {
  const res = await apiClient.post<{ code: number; data: FormSchema }>(`/schemas`, {
    ...data,
    expectedVersion
  })
  return res.data.data
}

export const publishSchema = async (formId: string, expectedVersion?: number) => {
  const res = await apiClient.post<{ code: number; data: FormSchema }>(`/schemas/${formId}/publish`, {
    expectedVersion
  })
  return res.data.data
}

export const listSchemaVersions = async (formId: string) => {
  const res = await apiClient.get<{ code: number; data: FormSchemaVersion[] }>(`/schemas/${formId}/versions`)
  return res.data.data
}

export const getSchemaVersion = async (formId: string, version: number) => {
  const res = await apiClient.get<{ code: number; data: FormSchemaVersion }>(`/schemas/${formId}/versions/${version}`)
  return res.data.data
}

// 2. 动态提交与查询相关 (场景二 Q0, Q1)
export const submitDynamicData = async (formId: string, data: Record<string, any>) => {
  const res = await apiClient.post<{ code: number; data: any }>(`/dynamic-data/${formId}`, {
    data
  })
  return res.data.data
}

export const queryDynamicData = async (formId: string, params: Record<string, any>) => {
  const res = await apiClient.get<{ code: number; data: { total: number; page: number; pageSize: number; items: FormSubmission[] } }>(
    `/dynamic-data/${formId}`,
    { params }
  )
  return res.data.data
}

export const getSubmission = async (formId: string, id: string) => {
  const res = await apiClient.get<{ code: number; data: FormSubmission }>(`/dynamic-data/${formId}/${id}`)
  return res.data.data
}

export const updateSubmission = async (formId: string, id: string, data: Record<string, any>) => {
  const res = await apiClient.put<{ code: number; data: FormSubmission }>(`/dynamic-data/${formId}/${id}`, {
    data
  })
  return res.data.data
}

export const deleteSubmission = async (formId: string, id: string) => {
  const res = await apiClient.delete(`/dynamic-data/${formId}/${id}`)
  return res.data
}

// 3. 文件上传安全相关 (Q3)
export const uploadFile = async (file: File, formId?: string) => {
  const formData = new FormData()
  formData.append('file', file)
  if (formId) {
    formData.append('formId', formId)
  }
  const res = await apiClient.post<{ code: number; data: any }>('/files/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
  return res.data.data
}

// 4. 审计日志相关 (Q6)
export const queryAuditLogs = async (params: { formId?: string; targetType?: string; targetId?: string; page?: number; pageSize?: number }) => {
  const res = await apiClient.get<{ code: number; data: { total: number; page: number; pageSize: number; items: AuditLog[] } }>(
    '/audit-logs',
    { params }
  )
  return res.data.data
}
