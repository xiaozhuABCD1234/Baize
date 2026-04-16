const API_BASE = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:8080'
import type { ApiResponse, Craft } from './types'

export interface KnowledgeItem {
  id: string;
  title: string;
  description: string;
  icon: string;
  details: {
    origin: string;
    history: string;
    features: string;
    significance: string;
  };
  inheritors?: {
    id: string;
    name: string;
    title: string;
    bio: string;
    contributions: string;
  }[];
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

// 获取知识科普列表（使用技艺API）
export async function fetchKnowledgeItems(): Promise<KnowledgeItem[]> {
  try {
    // 使用技艺API获取数据
    const crafts = await request<Craft[]>('/api/v1/crafts');
    
    // 转换为KnowledgeItem格式
    return crafts.map(craft => ({
      id: craft.id.toString(),
      title: craft.name,
      description: craft.description,
      icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M12 6.253v13m0-13C10.832 5.223 9.246 4.5 7.5 4.5S4.168 5.223 3 6.253v13C4.168 18.777 5.754 19.5 7.5 19.5s3.332-.723 4.5-1.747zm0 0C13.168 19.777 14.754 20.5 16.5 20.5s3.332-.723 4.5-1.747v-13C20.832 5.223 19.246 4.5 17.5 4.5s-3.332.723-4.5 1.747z\" /></svg>",
      details: {
        origin: "",
        history: craft.history || "",
        features: craft.characteristics || "",
        significance: ""
      },
      inheritors: []
    }));
  } catch (error) {
    console.error('Error fetching knowledge items:', error);
    return [];
  }
}

// 获取单个科普详情（使用技艺API）
export async function fetchKnowledgeItem(id: string): Promise<KnowledgeItem | null> {
  try {
    // 使用技艺API获取单个技艺详情
    const craft = await request<Craft>(`/api/v1/crafts/${id}`);
    
    // 转换为KnowledgeItem格式
    return {
      id: craft.id.toString(),
      title: craft.name,
      description: craft.description,
      icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M12 6.253v13m0-13C10.832 5.223 9.246 4.5 7.5 4.5S4.168 5.223 3 6.253v13C4.168 18.777 5.754 19.5 7.5 19.5s3.332-.723 4.5-1.747zm0 0C13.168 19.777 14.754 20.5 16.5 20.5s3.332-.723 4.5-1.747v-13C20.832 5.223 19.246 4.5 17.5 4.5s-3.332.723-4.5 1.747z\" /></svg>",
      details: {
        origin: "",
        history: craft.history || "",
        features: craft.characteristics || "",
        significance: ""
      },
      inheritors: []
    };
  } catch (error) {
    console.error(`Error fetching knowledge item ${id}:`, error);
    return null;
  }
}

// 知识服务
export const knowledgeService = {
  async list(): Promise<KnowledgeItem[]> {
    return fetchKnowledgeItems();
  },
  
  async get(id: string): Promise<KnowledgeItem | null> {
    return fetchKnowledgeItem(id);
  }
}