const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:1323'
import type { PaginatedResponse, ApiResponse } from './types'

export interface Work {
  id: number
  user_id: number
  craft_id: number
  category_id: number
  region_id: number
  title: string
  description: string
  status: number
  is_master: boolean
  weight: number
  view_count: number
  like_count: number
  comment_count: number
  created_at: string
  updated_at: string
  user?: {
    id: number
    username: string
    avatar?: string
  }
  craft?: {
    id: number
    name: string
  }
  category?: {
    id: number
    name: string
  }
  region?: {
    id: number
    name: string
  }
  media?: WorkMedia[]
}

export interface WorkMedia {
  id: number
  work_id: number
  media_type: string
  media_url: string
  thumbnail_url?: string
  order: number
}

export interface CreateWorkRequest {
  craft_id: number
  category_id: number
  region_id: number
  title: string
  description: string
  status?: number
  is_master?: boolean
  media?: {
    media_type: string
    media_url: string
    thumbnail_url?: string
    order: number
  }[]
}

export interface WorkListRequest {
  user_id?: number
  craft_id?: number
  region_id?: number
  is_master?: boolean
  order_by?: 'newest' | 'hot' | 'weight'
  page?: number
  page_size?: number
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('token')
  
  const res = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` }),
      ...options?.headers,
    },
  })

  const json: ApiResponse<T> = await res.json()

  if (!res.ok || json.error) {
    throw new Error(json.error?.message || '请求失败')
  }

  return json.data
}

export const workService = {
  async list(params?: WorkListRequest): Promise<PaginatedResponse<Work>> {
    const queryParams = new URLSearchParams()
    if (params?.user_id) queryParams.append('user_id', params.user_id.toString())
    if (params?.craft_id) queryParams.append('craft_id', params.craft_id.toString())
    if (params?.region_id) queryParams.append('region_id', params.region_id.toString())
    if (params?.is_master !== undefined) queryParams.append('is_master', params.is_master.toString())
    if (params?.order_by) queryParams.append('order_by', params.order_by)
    if (params?.page) queryParams.append('page', params.page.toString())
    if (params?.page_size) queryParams.append('page_size', params.page_size.toString())

    return request(`/api/v1/works?${queryParams.toString()}`)
  },

  async listTop(limit: number = 10): Promise<Work[]> {
    return request(`/api/v1/works/top?limit=${limit}`)
  },

  async listRecommended(limit: number = 10): Promise<Work[]> {
    return request(`/api/v1/works/recommended?limit=${limit}`)
  },

  async get(id: number): Promise<Work> {
    return request(`/api/v1/works/${id}`)
  },

  async getDetailed(id: number): Promise<Work> {
    return request(`/api/v1/works/${id}/detailed`)
  },

  async create(data: CreateWorkRequest): Promise<Work> {
    return request('/api/v1/works', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async update(id: number, data: CreateWorkRequest): Promise<Work> {
    return request(`/api/v1/works/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  async delete(id: number): Promise<{ message: string }> {
    return request(`/api/v1/works/${id}`, {
      method: 'DELETE',
    })
  },

  async updateStatus(id: number, status: number): Promise<{ message: string }> {
    return request(`/api/v1/works/${id}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    })
  },

  async incrementCount(id: number, field: string, delta: number = 1): Promise<{ message: string }> {
    return request(`/api/v1/works/${id}/count`, {
      method: 'PUT',
      body: JSON.stringify({ field, delta }),
    })
  },
}
