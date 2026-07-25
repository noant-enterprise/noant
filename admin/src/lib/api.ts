const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

class AdminAPI {
  private token: string | null = null

  setToken(token: string) {
    this.token = token
    localStorage.setItem('admin_token', token)
  }

  getToken(): string | null {
    if (!this.token) {
      this.token = localStorage.getItem('admin_token')
    }
    return this.token
  }

  clearToken() {
    this.token = null
    localStorage.removeItem('admin_token')
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    const token = this.getToken()
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const res = await fetch(`${API_BASE}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    })

    if (res.status === 401) {
      this.clearToken()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: 'Request failed' }))
      throw new Error(error.error || `HTTP ${res.status}`)
    }

    return res.json()
  }

  // Auth
  login(email: string, password: string) {
    return this.request<{ token: string; user: { id: string; email: string; role: string } }>('POST', '/api/v1/admin/login', { email, password })
  }

  // Dashboard
  getOverview() {
    return this.request<Record<string, unknown>>('GET', '/api/v1/admin/overview')
  }

  // Users
  getUsers(params?: { page?: number; limit?: number; search?: string; plan?: string }) {
    const query = new URLSearchParams()
    if (params?.page) query.set('page', String(params.page))
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.search) query.set('search', params.search)
    if (params?.plan) query.set('plan', params.plan)
    return this.request<{ users: unknown[]; total: number }>('GET', `/api/v1/admin/users?${query}`)
  }

  getUser(id: string) {
    return this.request<unknown>('GET', `/api/v1/admin/users/${id}`)
  }

  impersonateUser(id: string) {
    return this.request<{ token: string }>('POST', `/api/v1/admin/users/${id}/impersonate`)
  }

  // Analytics
  getAnalytics() {
    return this.request<unknown>('GET', '/api/v1/admin/analytics')
  }

  // Revenue
  getRevenue() {
    return this.request<unknown>('GET', '/api/v1/admin/revenue')
  }

  // AI Health
  getAIHealth() {
    return this.request<unknown>('GET', '/api/v1/admin/ai/health')
  }

  // System
  getSystemHealth() {
    return this.request<unknown>('GET', '/api/v1/admin/system/health')
  }

  // Alerts
  getAlerts() {
    return this.request<unknown[]>('GET', '/api/v1/admin/alerts')
  }
}

export const adminApi = new AdminAPI()
