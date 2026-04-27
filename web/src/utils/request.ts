import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// API 响应结构
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 分页响应结构
export interface PageResponse<T = unknown> {
  items: T[]
  total: number
  page: number
  page_size: number
}

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Token 刷新状态管理
let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

const processQueue = (token: string) => {
  pendingRequests.forEach(cb => cb(token))
  pendingRequests = []
}

const clearAuthAndRedirect = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('refresh_token')
  localStorage.removeItem('user')
  if (window.location.pathname !== '/login') {
    ElMessage.error('登录已过期，请重新登录')
    window.location.href = '/login'
  }
}

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    // FormData 时删除默认 Content-Type，让浏览器自动设置 multipart/form-data 及 boundary
    if (config.data instanceof FormData) {
      delete config.headers['Content-Type']
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    if (data.code !== 0) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message))
    }
    return response
  },
  async (error) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }
    // 支持静默错误：请求配置中设置 _silent 则不弹出错误提示
    const silent = (error.config as any)?._silent

    if (error.response) {
      const { status, data } = error.response

      // 401 且不是刷新请求本身 → 尝试用 Refresh Token 刷新
      if (status === 401 && !originalRequest._retry && window.location.pathname !== '/login') {
        const refreshTokenStr = localStorage.getItem('refresh_token')
        if (!refreshTokenStr) {
          clearAuthAndRedirect()
          return Promise.reject(error)
        }

        if (isRefreshing) {
          // 已经在刷新中，排队等待
          return new Promise((resolve) => {
            pendingRequests.push((newToken: string) => {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
              resolve(request(originalRequest))
            })
          })
        }

        originalRequest._retry = true
        isRefreshing = true

        try {
          const res = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshTokenStr })
          const { access_token, refresh_token: newRefreshToken } = res.data.data

          localStorage.setItem('token', access_token)
          if (newRefreshToken) {
            localStorage.setItem('refresh_token', newRefreshToken)
          }

          // 处理排队的请求
          processQueue(access_token)

          // 重试原请求
          originalRequest.headers.Authorization = `Bearer ${access_token}`
          return request(originalRequest)
        } catch {
          // Refresh Token 也过期了，跳登录
          processQueue('')
          clearAuthAndRedirect()
          return Promise.reject(error)
        } finally {
          isRefreshing = false
        }
      }

      if (!silent) {
        switch (status) {
          case 401:
            // 登录页面的 401 只显示错误信息，不跳转
            if (window.location.pathname === '/login') {
              ElMessage.error(data?.message || '用户名或密码错误')
            }
            break
          case 403:
            ElMessage.error(data?.message || '没有权限访问')
            break
          case 404:
            // 标记了 _redirectOn404 的请求（详情页主资源加载）自动跳转 404 页面
            if ((error.config as any)?._redirectOn404) {
              router.replace('/404')
            }
            break
          case 500:
            ElMessage.error(data?.message || '服务器错误')
            break
          default:
            ElMessage.error(data?.message || '请求失败')
        }
      }
    } else if (!silent) {
      ElMessage.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export default request
