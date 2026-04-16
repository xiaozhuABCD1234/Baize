const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:8080'
import type { ApiResponse } from './types'

export interface Region {
  id: number
  name: string
  code: string
  level: number
  parent_id?: number
  is_heritage_center: boolean
  created_at: string
  updated_at: string
  children?: Region[]
}

export interface CreateRegionRequest {
  name: string
  code: string
  level: number
  parent_id?: number
  is_heritage_center?: boolean
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

export const regionService = {
  async list(orderBy?: string): Promise<Region[]> {
    const params = orderBy ? `?order_by=${orderBy}` : ''
    return request(`/api/v1/regions${params}`)
  },

  async listRoot(): Promise<Region[]> {
    return request('/api/v1/regions/root')
  },

  async listByParentId(parentId: number): Promise<Region[]> {
    return request(`/api/v1/regions/parent/${parentId}`)
  },

  async listByLevel(level: number): Promise<Region[]> {
    return request(`/api/v1/regions/level/${level}`)
  },

  async listHeritageCenters(): Promise<Region[]> {
    return request('/api/v1/regions/heritage-centers')
  },

  async get(id: number): Promise<Region> {
    return request(`/api/v1/regions/${id}`)
  },

  async getWithChildren(id: number): Promise<Region> {
    return request(`/api/v1/regions/${id}/with-children`)
  },

  async getByCode(code: string): Promise<Region> {
    return request(`/api/v1/regions/code/${code}`)
  },

  async create(data: CreateRegionRequest): Promise<Region> {
    return request('/api/v1/regions', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async update(id: number, data: CreateRegionRequest): Promise<Region> {
    return request(`/api/v1/regions/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  async delete(id: number): Promise<{ message: string }> {
    return request(`/api/v1/regions/${id}`, {
      method: 'DELETE',
    })
  },
}
