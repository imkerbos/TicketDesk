import type { FieldSchemeItem } from '@/types/field'
import { BUILTIN_FIELD_MAP, isBuiltinField } from '@/types/field'
import type { Issue } from '@/types/issue'
import dayjs from 'dayjs'

/**
 * 从字段方案值中分离内置字段和扩展字段
 * @param fieldScheme 字段方案列表
 * @param fieldValues 所有字段值（field_id → value）
 * @returns builtinValues: 内置字段值（已转换为 API 请求属性名），customFields: 扩展字段数组
 */
export function extractBuiltinFields(
  fieldScheme: FieldSchemeItem[],
  fieldValues: Record<number, any>,
): {
  builtinValues: Record<string, any>
  customFields: { field_id: number; value: any }[]
} {
  const builtinValues: Record<string, any> = {}
  const customFields: { field_id: number; value: any }[] = []

  for (const item of fieldScheme) {
    const fieldKey = item.field?.field_key || ''
    const value = fieldValues[item.field_id]

    if (isBuiltinField(fieldKey)) {
      // 内置字段：转换 key 名
      const apiKey = BUILTIN_FIELD_MAP[fieldKey]
      if (apiKey && value !== undefined && value !== null && value !== '') {
        builtinValues[apiKey] = value
      }
    } else {
      // 扩展字段：过滤空值
      if (value === undefined || value === null || value === '') continue
      if (Array.isArray(value) && value.length === 0) continue
      customFields.push({ field_id: item.field_id, value })
    }
  }

  return { builtinValues, customFields }
}

/**
 * 编辑回填：将 Issue 对象的内置字段值写入 fieldValues map
 * @param issue 当前工单对象
 * @param fieldScheme 编辑用字段方案
 * @param fieldValues 字段值 map（会被修改）
 */
export function backfillBuiltinFields(
  issue: Issue,
  fieldScheme: FieldSchemeItem[],
  fieldValues: Record<number, any>,
): void {
  for (const item of fieldScheme) {
    const fieldKey = item.field?.field_key || ''
    if (!isBuiltinField(fieldKey)) continue

    switch (fieldKey) {
      case 'description':
        fieldValues[item.field_id] = issue.description || ''
        break
      case 'priority':
        fieldValues[item.field_id] = issue.priority || 'P2'
        break
      case 'assignee':
        fieldValues[item.field_id] = issue.assignee_id || undefined
        break
      case 'planned_start_date':
        fieldValues[item.field_id] = issue.planned_start_date
          ? dayjs(issue.planned_start_date).format('YYYY-MM-DD')
          : undefined
        break
      case 'planned_end_date':
        fieldValues[item.field_id] = issue.planned_end_date
          ? dayjs(issue.planned_end_date).format('YYYY-MM-DD')
          : undefined
        break
      case 'epic_link':
        fieldValues[item.field_id] = issue.epic_id || undefined
        break
    }
  }
}
