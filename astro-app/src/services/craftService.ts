const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:1323'
import type { ApiResponse, Craft } from './types'

export interface CreateCraftRequest {
  name: string
  description: string
  category_id: number
  difficulty: number
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

export const craftService = {
  async list(orderBy?: string): Promise<Craft[]> {
    const params = orderBy ? `?order_by=${orderBy}` : ''
    return request(`/api/v1/crafts${params}`)
  },

  async listByCategory(categoryId: number): Promise<Craft[]> {
    return request(`/api/v1/crafts/category/${categoryId}`)
  },

  async listByDifficulty(level: number): Promise<Craft[]> {
    return request(`/api/v1/crafts/difficulty/${level}`)
  },

  async get(id: number): Promise<Craft> {
    return request(`/api/v1/crafts/${id}`)
  },

  async getWithCategory(id: number): Promise<Craft> {
    return request(`/api/v1/crafts/${id}/with-category`)
  },

  async create(data: CreateCraftRequest): Promise<Craft> {
    return request('/api/v1/crafts', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async update(id: number, data: CreateCraftRequest): Promise<Craft> {
    return request(`/api/v1/crafts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  async delete(id: number): Promise<{ message: string }> {
    return request(`/api/v1/crafts/${id}`, {
      method: 'DELETE',
    })
  },
}
