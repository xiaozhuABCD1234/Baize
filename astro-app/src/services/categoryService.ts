const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:8080'
import type { ApiResponse } from './types'

export interface Category {
  id: number
  name: string
  code: string
  level: number
  parent_id?: number
  region_code?: string
  is_active: boolean
  created_at: string
  updated_at: string
  children?: Category[]
}

export interface CreateCategoryRequest {
  name: string
  code: string
  level: number
  parent_id?: number
  region_code?: string
  is_active?: boolean
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

export const categoryService = {
  async list(orderBy?: string): Promise<Category[]> {
    const params = orderBy ? `?order_by=${orderBy}` : ''
    return request(`/api/v1/categories${params}`)
  },

  async listRoot(): Promise<Category[]> {
    return request('/api/v1/categories/root')
  },

  async listByParentId(parentId: number): Promise<Category[]> {
    return request(`/api/v1/categories/parent/${parentId}`)
  },

  async listByRegionCode(regionCode: string): Promise<Category[]> {
    return request(`/api/v1/categories/region/${regionCode}`)
  },

  async listActive(): Promise<Category[]> {
    return request('/api/v1/categories/active')
  },

  async get(id: number): Promise<Category> {
    return request(`/api/v1/categories/${id}`)
  },

  async getWithChildren(id: number): Promise<Category> {
    return request(`/api/v1/categories/${id}/with-children`)
  },

  async getByName(name: string): Promise<Category> {
    return request(`/api/v1/categories/name/${name}`)
  },

  async create(data: CreateCategoryRequest): Promise<Category> {
    return request('/api/v1/categories', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async update(id: number, data: CreateCategoryRequest): Promise<Category> {
    return request(`/api/v1/categories/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  async delete(id: number): Promise<{ message: string }> {
    return request(`/api/v1/categories/${id}`, {
      method: 'DELETE',
    })
  },
}
