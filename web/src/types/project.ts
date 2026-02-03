// 项目相关类型定义

// 用户简要信息（项目负责人）
export interface UserBrief {
  id: number
  username: string
  display_name: string
  avatar_url?: string
}

export interface Project {
  id: number
  project_key: string
  name: string
  description: string
  lead_user_id: number
  lead_user?: UserBrief
  status: number // 0-禁用, 1-启用
  member_count?: number
  created_at: string
  updated_at: string
}

export interface ProjectListRequest {
  page?: number
  page_size?: number
  keyword?: string
  status?: number
}

export interface ProjectListResponse {
  items: Project[]
  total: number
  page: number
  page_size: number
}

export interface CreateProjectRequest {
  project_key: string
  name: string
  description?: string
  lead_user_id?: number
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
  lead_user_id?: number
  status?: number
}

export interface ProjectMember {
  id: number
  project_id: number
  user_id: number
  user?: UserBrief
  role: string // owner, admin, member
  created_at: string
}

export interface AddProjectMemberRequest {
  user_id: number
  role: string
}

export interface ProjectIssueType {
  id: number
  project_id?: number
  name: string
  display_name: string
  description: string
  icon: string
  color: string
  created_at: string
  updated_at: string
}

export interface CreateIssueTypeRequest {
  name: string
  display_name: string
  description?: string
  icon?: string
  color?: string
}
