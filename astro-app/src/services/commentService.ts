const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:1323'
import type { PaginatedResponse, ApiResponse } from './types'

export interface Comment {
  id: number
  work_id: number
  user_id: number
  parent_id?: number
  content: string
  status: number
  like_count: number
  reply_count: number
  created_at: string
  updated_at: string
  user?: {
    id: number
    username: string
    avatar?: string
  }
  replies?: Comment[]
}

export interface CreateCommentRequest {
  work_id: number
  parent_id?: number
  content: string
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

export const commentService = {
  async listByWorkId(workId: number, page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<Comment>> {
    return request(`/api/v1/comments/work/${workId}?page=${page}&page_size=${pageSize}`)
  },

  async listRootByWorkId(workId: number): Promise<Comment[]> {
    return request(`/api/v1/comments/work/${workId}/root`)
  },

  async listByUserId(userId: number): Promise<Comment[]> {
    return request(`/api/v1/comments/user/${userId}`)
  },

  async get(id: number): Promise<Comment> {
    return request(`/api/v1/comments/${id}`)
  },

  async create(data: CreateCommentRequest): Promise<Comment> {
    return request('/api/v1/comments', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async update(id: number, content: string): Promise<Comment> {
    return request(`/api/v1/comments/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ content }),
    })
  },

  async delete(id: number): Promise<{ message: string }> {
    return request(`/api/v1/comments/${id}`, {
      method: 'DELETE',
    })
  },

  async updateStatus(id: number, status: number): Promise<{ message: string }> {
    return request(`/api/v1/comments/${id}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    })
  },

  async incrementLikeCount(id: number, delta: number = 1): Promise<{ message: string }> {
    return request(`/api/v1/comments/${id}/like`, {
      method: 'PUT',
      body: JSON.stringify({ delta }),
    })
  },
}
