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
