const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:8080'
import type { ApiResponse } from './types'

export interface AuthUser {
  id: number
  username: string
  email: string
  user_type: string
  status?: string
}

export interface LoginResponse {
  id: number
  username: string
  email: string
  user_type: string
  status: string
  token: string
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  const json: ApiResponse<T> = await res.json()

  if (!res.ok || json.error) {
    throw new Error(json.error?.message || '请求失败')
  }

  return json.data
}

export const authService = {
  async register(data: { username: string; email: string; password: string }) {
    return request('/api/v1/users/register', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async login(email: string, password: string): Promise<LoginResponse> {
    return request<LoginResponse>('/api/v1/users/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  },

  async refreshToken(refreshToken: string) {
    return request('/api/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  },

  getToken: () => localStorage.getItem('token'),
  setToken: (token: string) => localStorage.setItem('token', token),
  removeToken: () => localStorage.removeItem('token'),

  getUser: (): AuthUser | null => {
    const user = localStorage.getItem('user')
    return user ? JSON.parse(user) : null
  },
  setUser: (user: AuthUser) => localStorage.setItem('user', JSON.stringify(user)),
  removeUser: () => localStorage.removeItem('user'),

  isAuthenticated: () => !!localStorage.getItem('token'),

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  },
}
