// 共享类型定义
export interface PaginatedResponse<T> {
  data: T[]
  page: number
  page_size: number
  total: number
}

export interface ApiResponse<T = any> {
  data: T
  error?: {
    code: string
    message: string
  }
}

// 技艺接口
export interface Craft {
  id: number
  name: string
  description: string
  category_id: number
  difficulty: number
  created_at: string
  updated_at: string
  category?: {
    id: number
    name: string
  }
  history?: string
  characteristics?: string
}
