import request, { type ApiResponse } from '@/utils/request'
import type { APIToken, CreateTokenResponse, CreateTokenRequest } from '@/types/token'

// 创建 API Token
export const createApiToken = (data: CreateTokenRequest) =>
  request.post<ApiResponse<CreateTokenResponse>>('/users/me/tokens', data)

// 列出当前用户的 API Token
export const listApiTokens = () =>
  request.get<ApiResponse<APIToken[]>>('/users/me/tokens')

// 撤销 API Token
export const deleteApiToken = (id: number) =>
  request.delete<ApiResponse<null>>(`/users/me/tokens/${id}`)
