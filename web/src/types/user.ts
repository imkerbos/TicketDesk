// 用户相关类型定义

export interface User {
  id: number
  username: string
  email: string
  display_name: string
  avatar?: string
  role: string // admin, user
  status: number // 0-禁用, 1-启用
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface UserListRequest {
  page?: number
  page_size?: number
  keyword?: string
  role?: string
  status?: number
}

export interface UserListResponse {
  items: User[]
  total: number
  page: number
  page_size: number
}

export interface CreateUserRequest {
  username: string
  email: string
  password: string
  display_name: string
  role?: string
}

export interface UpdateUserRequest {
  email?: string
  display_name?: string
  role?: string
  status?: number
}

export interface UpdatePasswordRequest {
  old_password: string
  new_password: string
}

export interface UserProfile {
  id: number
  username: string
  email: string
  display_name: string
  avatar?: string
  role: string
}

// 用于选择器的简化用户信息
export interface UserOption {
  id: number
  username: string
  display_name: string
}
