const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:8080'
import type { ApiResponse } from './types'

export interface Follow {
  id: number
  follower_id: number
  following_id: number
  created_at: string
  follower?: {
    id: number
    username: string
    avatar?: string
  }
  following?: {
    id: number
    username: string
    avatar?: string
  }
}

export interface CreateFollowRequest {
  following_id: number
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

export const followService = {
  async isFollowing(userId: number): Promise<{ is_following: boolean }> {
    return request(`/api/v1/follows/check/${userId}`)
  },

  async getFollowingList(userId: number): Promise<Follow[]> {
    return request(`/api/v1/follows/following/${userId}`)
  },

  async getFollowerList(userId: number): Promise<Follow[]> {
    return request(`/api/v1/follows/followers/${userId}`)
  },

  async getFollowingCount(userId: number): Promise<{ count: number }> {
    return request(`/api/v1/follows/following/${userId}/count`)
  },

  async getFollowerCount(userId: number): Promise<{ count: number }> {
    return request(`/api/v1/follows/followers/${userId}/count`)
  },

  async create(data: CreateFollowRequest): Promise<Follow> {
    return request('/api/v1/follows', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async delete(userId: number): Promise<{ message: string }> {
    return request(`/api/v1/follows/${userId}`, {
      method: 'DELETE',
    })
  },
}
