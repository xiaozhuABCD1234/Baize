const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:1323'
import type { PaginatedResponse, ApiResponse } from './types'

export interface User {
  id: number
  username: string
  email: string
  user_type: string
  status: string
  created_at: string
  updated_at: string
  avatar?: string
}

export interface UpdateUserRequest {
  username?: string
  email?: string
  user_type?: string
  status?: string
  avatar?: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
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

export const userService = {
  async list(page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<User>> {
    return request(`/api/v1/users?page=${page}&page_size=${pageSize}`)
  },

  async get(id: number): Promise<User> {
    return request(`/api/v1/users/${id}`)
  },

  async update(id: number, data: UpdateUserRequest): Promise<User> {
    return request(`/api/v1/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  async delete(id: number): Promise<{ message: string }> {
    return request(`/api/v1/users/${id}`, {
      method: 'DELETE',
    })
  },

  async forceDelete(id: number): Promise<{ message: string }> {
    return request(`/api/v1/users/${id}/force`, {
      method: 'DELETE',
    })
  },

  async changePassword(id: number, data: ChangePasswordRequest): Promise<{ message: string }> {
    return request(`/api/v1/users/${id}/password`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },
}
