import request, { type ApiResponse } from '@/utils/request'
import type {
  User,
  UserListRequest,
  UserListResponse,
  CreateUserRequest,
  UpdateUserRequest,
  UpdatePasswordRequest,
  UserProfile,
  UserOption,
} from '@/types/user'

// 用户列表
export const getUserList = (params?: UserListRequest) => {
  return request.get<ApiResponse<UserListResponse>>('/users', { params })
}

// 用户详情
export const getUserDetail = (id: number) => {
  return request.get<ApiResponse<User>>(`/users/${id}`)
}

// 创建用户
export const createUser = (data: CreateUserRequest) => {
  return request.post<ApiResponse<User>>('/users', data)
}

// 更新用户
export const updateUser = (id: number, data: UpdateUserRequest) => {
  return request.put<ApiResponse<User>>(`/users/${id}`, data)
}

// 删除用户
export const deleteUser = (id: number) => {
  return request.delete(`/users/${id}`)
}

// 启用用户
export const enableUser = (id: number) => {
  return request.post(`/users/${id}/enable`)
}

// 禁用用户
export const disableUser = (id: number) => {
  return request.post(`/users/${id}/disable`)
}

// 重置密码
export const resetUserPassword = (id: number, newPassword: string) => {
  return request.post(`/users/${id}/reset-password`, { new_password: newPassword })
}

// 获取当前用户信息
export const getCurrentUser = () => {
  return request.get<ApiResponse<UserProfile>>('/users/me')
}

// 更新当前用户信息
export const updateCurrentUser = (data: Partial<UpdateUserRequest>) => {
  return request.put<ApiResponse<UserProfile>>('/users/me', data)
}

// 修改当前用户密码
export const updatePassword = (data: UpdatePasswordRequest) => {
  return request.post('/users/me/password', data)
}

// 获取所有用户（用于选择器，不分页）
export const getAllUsers = () => {
  return request.get<ApiResponse<UserOption[]>>('/users/all')
}

// 搜索用户（用于自动完成）
export const searchUsers = (keyword: string) => {
  return request.get<ApiResponse<UserOption[]>>('/users/search', { params: { keyword } })
}

// ============ MFA 相关 ============

// MFA 状态响应
export interface MFAStatusResponse {
  enabled: boolean
  verified_at?: string
}

// MFA 设置响应
export interface MFASetupResponse {
  secret: string
  otp_auth_url: string
  issuer: string
  account: string
}

// 获取 MFA 状态
export const getMFAStatus = () => {
  return request.get<ApiResponse<MFAStatusResponse>>('/users/me/mfa')
}

// 开始 MFA 设置
export const setupMFA = () => {
  return request.post<ApiResponse<MFASetupResponse>>('/users/me/mfa/setup')
}

// 验证并启用 MFA
export const enableMFA = (code: string) => {
  return request.post('/users/me/mfa/enable', { code })
}

// 禁用 MFA
export const disableMFA = (code: string) => {
  return request.post('/users/me/mfa/disable', { code })
}
