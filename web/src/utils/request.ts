import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
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

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
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
  (error) => {
    // 支持静默错误：请求配置中设置 _silent 则不弹出错误提示
    const silent = (error.config as any)?._silent
    if (error.response) {
      const { status, data } = error.response
      if (!silent) {
        switch (status) {
          case 401:
            // 登录页面的 401 只显示错误信息，不跳转
            if (window.location.pathname === '/login') {
              ElMessage.error(data?.message || '用户名或密码错误')
            } else {
              ElMessage.error('未登录或登录已过期')
              localStorage.removeItem('token')
              window.location.href = '/login'
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
