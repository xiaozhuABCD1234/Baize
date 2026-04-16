const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:8080'
import type { PaginatedResponse, ApiResponse } from './types'

export interface Favorite {
  id: number
  user_id: number
  work_id: number
  folder_id?: number
  created_at: string
  updated_at: string
  work?: {
    id: number
    title: string
    description: string
    user?: {
      id: number
      username: string
      avatar?: string
    }
  }
}

export interface CreateFavoriteRequest {
  work_id: number
  folder_id?: number
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

export const favoriteService = {
  async listByWorkId(workId: number): Promise<Favorite[]> {
    return request(`/api/v1/favorites/work/${workId}`)
  },

  async checkExists(workId: number): Promise<{ favorited: boolean }> {
    return request(`/api/v1/favorites/check/${workId}`)
  },

  async get(id: number): Promise<Favorite> {
    return request(`/api/v1/favorites/${id}`)
  },

  async listByUserId(userId: number, page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<Favorite>> {
    return request(`/api/v1/favorites/user/${userId}?page=${page}&page_size=${pageSize}`)
  },

  async create(data: CreateFavoriteRequest): Promise<Favorite> {
    return request('/api/v1/favorites', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async delete(id: number): Promise<{ message: string }> {
    return request(`/api/v1/favorites/${id}`, {
      method: 'DELETE',
    })
  },

  async deleteByWork(workId: number): Promise<{ message: string }> {
    return request(`/api/v1/favorites/work/${workId}`, {
      method: 'DELETE',
    })
  },

  async updateFolder(id: number, folderId: number): Promise<{ message: string }> {
    return request(`/api/v1/favorites/${id}/folder`, {
      method: 'PUT',
      body: JSON.stringify({ folder_id: folderId }),
    })
  },
}
