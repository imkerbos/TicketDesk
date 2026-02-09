/**
 * 附件信息
 */
export interface Attachment {
  id: number
  issue_id: number
  file_name: string
  file_path: string
  file_size: number
  file_type: string
  is_image: boolean
  uploaded_by: number
  uploader?: UserBrief
  created_at: string
}

/**
 * 用户简要信息
 */
export interface UserBrief {
  id: number
  username: string
  display_name: string
  avatar_url: string
}
