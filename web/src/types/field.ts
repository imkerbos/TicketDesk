// 字段类型常量
export const FieldType = {
  TEXT: 'text',
  TEXTAREA: 'textarea',
  NUMBER: 'number',
  DATE: 'date',
  SELECT: 'select',
  MULTISELECT: 'multiselect',
  USER: 'user',
  VERSION: 'version',
  COMPONENT: 'component',
  LABEL: 'label',
  EPIC_LINK: 'epic_link',
  TIME_ESTIMATE: 'time_estimate',
} as const

export type FieldTypeValue = typeof FieldType[keyof typeof FieldType]

// 字段选项
export interface FieldOption {
  value: string
  label: string
}

// 字段定义
export interface FieldDefinition {
  id: number
  project_id: number | null
  field_key: string
  field_name: string
  field_type: FieldTypeValue
  description: string
  is_system: boolean
  is_active: boolean
  options: string // JSON字符串
  validation: string // JSON字符串
  default_value: string
  sort_order: number
  created_at: string
  updated_at: string
}

// 字段方案项
export interface FieldSchemeItem {
  id: number
  project_id: number
  issue_type_id: number
  field_id: number
  field?: FieldDefinition
  is_required: boolean
  is_visible_create: boolean
  is_visible_edit: boolean
  is_visible_detail: boolean
  sort_order: number
  default_value: string
  created_at: string
  updated_at: string
}

// 字段值
export interface FieldValue {
  field_id: number
  field_key: string
  field_name: string
  field_type: FieldTypeValue
  value: any
  display_value: string
}

// 自定义字段值（用于提交）
export interface CustomFieldValue {
  field_id: number
  value: any
}

// 项目版本
export interface ProjectVersion {
  id: number
  project_id: number
  name: string
  description: string
  release_date: string | null
  status: 'unreleased' | 'released' | 'archived'
  sort_order: number
  created_at: string
  updated_at: string
}

// 项目组件
export interface ProjectComponent {
  id: number
  project_id: number
  name: string
  description: string
  lead_user_id: number | null
  lead_user?: {
    id: number
    username: string
    display_name: string
    avatar_url: string
  }
  created_at: string
  updated_at: string
}

// 工单标签
export interface IssueLabel {
  id: number
  project_id: number
  name: string
  color: string
  description: string
  created_at: string
  updated_at: string
}

// 创建字段请求
export interface CreateFieldRequest {
  field_key: string
  field_name: string
  field_type: FieldTypeValue
  description?: string
  options?: string
  validation?: string
  default_value?: string
}

// 更新字段请求
export interface UpdateFieldRequest {
  field_name?: string
  description?: string
  options?: string
  validation?: string
  default_value?: string
  is_active?: boolean
  sort_order?: number
}

// 更新字段方案请求项
export interface FieldSchemeItemRequest {
  field_id: number
  is_required: boolean
  is_visible_create: boolean
  is_visible_edit: boolean
  is_visible_detail: boolean
  sort_order: number
  default_value?: string
}

// 更新字段方案请求
export interface UpdateFieldSchemeRequest {
  items: FieldSchemeItemRequest[]
}

// 创建版本请求
export interface CreateVersionRequest {
  name: string
  description?: string
  release_date?: string
}

// 更新版本请求
export interface UpdateVersionRequest {
  name?: string
  description?: string
  release_date?: string
  status?: 'unreleased' | 'released' | 'archived'
}

// 创建组件请求
export interface CreateComponentRequest {
  name: string
  description?: string
  lead_user_id?: number
}

// 更新组件请求
export interface UpdateComponentRequest {
  name?: string
  description?: string
  lead_user_id?: number
}

// 创建标签请求
export interface CreateLabelRequest {
  name: string
  color?: string
  description?: string
}

// 更新标签请求
export interface UpdateLabelRequest {
  name?: string
  color?: string
  description?: string
}

// 解析字段选项
export function parseFieldOptions(optionsJson: string): FieldOption[] {
  if (!optionsJson) return []
  try {
    return JSON.parse(optionsJson)
  } catch {
    return []
  }
}

// 获取字段类型显示名称
export function getFieldTypeLabel(type: FieldTypeValue): string {
  const labels: Record<FieldTypeValue, string> = {
    [FieldType.TEXT]: '单行文本',
    [FieldType.TEXTAREA]: '多行文本',
    [FieldType.NUMBER]: '数字',
    [FieldType.DATE]: '日期',
    [FieldType.SELECT]: '单选',
    [FieldType.MULTISELECT]: '多选',
    [FieldType.USER]: '用户',
    [FieldType.VERSION]: '版本',
    [FieldType.COMPONENT]: '组件',
    [FieldType.LABEL]: '标签',
    [FieldType.EPIC_LINK]: 'Epic链接',
    [FieldType.TIME_ESTIMATE]: '时间估算',
  }
  return labels[type] || type
}
