import request, { type ApiResponse } from '@/utils/request'
import type {
  FieldDefinition,
  FieldSchemeTemplate,
  FieldSchemeTemplateDetail,
  FieldUsage,
  CreateFieldRequest,
  UpdateFieldRequest,
  CreateTemplateRequest,
  UpdateTemplateRequest,
  TemplateItemInput,
} from '@/types/field'

// ============ 全局字段管理 API ============

export function getGlobalFields() {
  return request.get<ApiResponse<FieldDefinition[]>>('/admin/fields')
}

export function createGlobalField(data: CreateFieldRequest) {
  return request.post<ApiResponse<FieldDefinition>>('/admin/fields', data)
}

export function updateGlobalField(fieldId: number, data: UpdateFieldRequest) {
  return request.put<ApiResponse<FieldDefinition>>(`/admin/fields/${fieldId}`, data)
}

export function deleteGlobalField(fieldId: number) {
  return request.delete(`/admin/fields/${fieldId}`)
}

export function getFieldUsage(fieldId: number) {
  return request.get<ApiResponse<FieldUsage>>(`/admin/fields/${fieldId}/usage`)
}

// ============ 方案模板管理 API ============

export function getTemplates() {
  return request.get<ApiResponse<FieldSchemeTemplate[]>>('/admin/field-scheme-templates')
}

export function createTemplate(data: CreateTemplateRequest) {
  return request.post<ApiResponse<FieldSchemeTemplate>>('/admin/field-scheme-templates', data)
}

export function getTemplate(templateId: number) {
  return request.get<ApiResponse<FieldSchemeTemplateDetail>>(`/admin/field-scheme-templates/${templateId}`)
}

export function updateTemplate(templateId: number, data: UpdateTemplateRequest) {
  return request.put<ApiResponse<FieldSchemeTemplate>>(`/admin/field-scheme-templates/${templateId}`, data)
}

export function deleteTemplate(templateId: number) {
  return request.delete(`/admin/field-scheme-templates/${templateId}`)
}

export function updateTemplateItems(templateId: number, items: TemplateItemInput[]) {
  return request.put(`/admin/field-scheme-templates/${templateId}/items`, { items })
}
