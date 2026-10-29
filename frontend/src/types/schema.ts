export type FieldType =
  | 'input'
  | 'number'
  | 'select'
  | 'radio'
  | 'checkbox'
  | 'date'
  | 'textarea'
  | 'upload'
  | 'repeater'
  | 'object'
  | 'alert'

export interface OptionItem {
  label: string
  value: string | number
}

export interface ValidationRules {
  required?: boolean
  email?: boolean
  regex?: string
  min?: number
  max?: number
  minLength?: number
  maxLength?: number
  customMessage?: string
}

export type RuleAction = 'show' | 'hide' | 'setRequired' | 'unsetRequired' | 'setValue' | 'showAlert'

export interface FieldRule {
  targetKey: string
  action: RuleAction
  expression: string // 如 "isFreshGraduate == 'yes'" 或 "expectedSalary > 30000"
  parameters?: Record<string, any>
}

export interface AlertConfig {
  title: string
  type?: 'info' | 'warning' | 'success' | 'error'
}

export interface FieldDefinition {
  key: string
  label: string
  type: FieldType
  placeholder?: string
  defaultValue?: any
  deprecated?: boolean // Q0/Q2: 废弃字段历史保留，新表单不渲染
  validation?: ValidationRules
  options?: OptionItem[]
  children?: FieldDefinition[] // 嵌套 Object
  itemFields?: FieldDefinition[] // 嵌套 Repeater 数组
  rules?: FieldRule[] // 字段间联动规则
  gridCol?: number // 栅格占据列数 (1-24)
  alertConfig?: AlertConfig
}

export interface FormSchema {
  id?: string
  formId: string
  title: string
  description?: string
  version: number // 乐观锁版本号
  publishedVersion: number // 最新发布版本号
  status: 'draft' | 'published'
  fields: FieldDefinition[]
  migrations?: Record<string, string> // 历史字段重命名映射
  createdAt?: string
  updatedAt?: string
  createdBy?: string
}

export interface FormSchemaVersion {
  id: number
  formId: string
  version: number
  title: string
  description?: string
  fields: FieldDefinition[]
  publishedAt: string
  publishedBy: string
}

export interface FormSubmission {
  id: string
  formId: string
  schemaVersion: number
  data: Record<string, any>
  status: string
  submitterId?: string
  submitterName?: string
  createdAt: string
  updatedAt: string
}

export interface AuditLog {
  id: number
  action: string
  targetType: string
  targetId: string
  formId: string
  operatorId: string
  operatorName: string
  clientIp: string
  diff?: Record<string, { old: any; new: any }>
  createdAt: string
}
