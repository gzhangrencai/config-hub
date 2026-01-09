import axios, { AxiosError } from 'axios'
import { message } from 'antd'

// 支持环境变量配置 API 地址
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

const client = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
client.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
client.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ message?: string; code?: string }>) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
      return Promise.reject(error)
    }

    const errorMessage = error.response?.data?.message || '请求失败，请稍后重试'
    message.error(errorMessage)
    return Promise.reject(error)
  }
)

export default client

// API 类型定义
export interface Project {
  id: number
  name: string
  description: string
  access_mode: string
  created_at: string
  updated_at: string
}

export interface Config {
  id: number
  project_id: number
  name: string
  namespace: string
  environment: string
  file_type: string
  current_version: number
  created_at: string
  updated_at: string
}

export interface ConfigVersion {
  id: number
  config_id: number
  version: number
  content: string
  commit_hash: string
  commit_message: string
  author: string
  created_at: string
}

export interface ProjectKey {
  id: number
  project_id: number
  name: string
  access_key: string
  permissions: string
  is_active: boolean
  expires_at: string | null
  created_at: string
}

export interface AuditLog {
  id: number
  project_id: number
  user_id: number | null
  access_key_id: number | null
  action: string
  resource_type: string
  resource_id: number
  resource_name: string
  ip_address: string
  user_agent: string
  created_at: string
}

export interface Release {
  id: number
  project_id: number
  config_id: number
  version: number
  environment: string
  status: string
  release_type: string
  gray_percentage: number
  released_by: string
  released_at: string
}
