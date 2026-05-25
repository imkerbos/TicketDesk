// API Token 类型定义
export interface APIToken {
  id: number
  name: string
  prefix: string
  expires_at: string | null
  last_used_at: string | null
  created_at: string
  is_expired: boolean
}

// 创建 token 返回 (含一次性明文)
export interface CreateTokenResponse extends APIToken {
  token: string
}

export interface CreateTokenRequest {
  name: string
  expires_in: number // 秒数, 0 = 永久
}
