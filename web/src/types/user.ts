// 用户相关类型定义

export interface User {
  id: number
  username: string
  email: string
  display_name: string
  avatar_url?: string
  roles: string[] // 角色数组，如 ['admin', 'user']
  status: number // 0-禁用, 1-启用
  mfa_enabled: boolean // MFA 是否启用
  last_login_at?: string // 最后登录时间
  created_at: string
  updated_at: string
}

export interface UserListRequest {
  page?: number
  page_size?: number
  keyword?: string
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
  display_name?: string
  status?: number
  roles?: string[]
}

export interface UpdateUserRequest {
  email?: string
  display_name?: string
  avatar_url?: string
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
  avatar_url?: string
  roles: string[] // 角色数组
  status: number
  created_at: string
  updated_at: string
}

// 用于选择器的简化用户信息
export interface UserOption {
  id: number
  username: string
  display_name: string
}
