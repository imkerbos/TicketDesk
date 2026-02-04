import request, { type ApiResponse } from '@/utils/request'

// 用户信息
export interface User {
  id: number
  username: string
  email: string
  display_name: string
  avatar_url: string
  status: number
  roles: string[]
  created_at: string
  updated_at: string
}

// 登录请求
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应
export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}

// 刷新 Token 请求
export interface RefreshTokenRequest {
  refresh_token: string
}

// 登录
export const login = (data: LoginRequest) => {
  return request.post<ApiResponse<LoginResponse>>('/auth/login', data)
}

// 刷新 Token
export const refreshToken = (data: RefreshTokenRequest) => {
  return request.post<ApiResponse<LoginResponse>>('/auth/refresh', data)
}

// 获取当前用户信息
export const getCurrentUser = () => {
  return request.get<ApiResponse<User>>('/users/me')
}

// 忘记密码请求
export interface ForgotPasswordRequest {
  email: string
}

// 重置密码请求
export interface ResetPasswordWithTokenRequest {
  token: string
  new_password: string
}

// 忘记密码
export const forgotPassword = (data: ForgotPasswordRequest) => {
  return request.post<ApiResponse<{ message: string }>>('/auth/forgot-password', data)
}

// 验证重置密码令牌
export const verifyResetToken = (token: string) => {
  return request.get<ApiResponse<{ message: string }>>('/auth/verify-reset-token', {
    params: { token }
  })
}

// 使用令牌重置密码
export const resetPasswordWithToken = (data: ResetPasswordWithTokenRequest) => {
  return request.post<ApiResponse<{ message: string }>>('/auth/reset-password', data)
}
