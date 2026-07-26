import type { OverviewResponse, UsersResponse, UserDetail, AnalyticsResponse, RevenueResponse, AIHealthResponse, SystemHealthResponse, AlertsResponse, ActivityResponse, LoginResponse, AuditLogsResponse, KnowledgeBaseResponse, TrainKnowledgeRequest, SalesLead } from '@/types'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

class AdminAPI {
  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    const res = await fetch(`${API_BASE}${path}`, {
      method,
      headers,
      credentials: 'include',
      body: body ? JSON.stringify(body) : undefined,
    })

    if (res.status === 401) {
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: 'Request failed' }))
      throw new Error(error.error || `HTTP ${res.status}`)
    }

    return res.json()
  }

  login(email: string, password: string) {
    return this.request<LoginResponse>('POST', '/api/v1/auth/login', { email, password })
  }

  me() {
    return this.request<{ user: { id: string; email: string; role: string; first_name: string; last_name: string } }>('GET', '/api/v1/auth/me')
  }

  logout() {
    return this.request<void>('POST', '/api/v1/auth/logout')
  }

  getOverview() {
    return this.request<OverviewResponse>('GET', '/api/v1/admin/overview')
  }

  getUsers(params?: { search?: string; plan?: string }) {
    const query = new URLSearchParams()
    if (params?.search) query.set('search', params.search)
    if (params?.plan) query.set('plan', params.plan)
    const qs = query.toString()
    return this.request<UsersResponse>('GET', `/api/v1/admin/users${qs ? '?' + qs : ''}`)
  }

  getUser(id: string) {
    return this.request<UserDetail>('GET', `/api/v1/admin/users/${id}`)
  }

  getAnalytics() {
    return this.request<AnalyticsResponse>('GET', '/api/v1/admin/analytics')
  }

  getRevenue() {
    return this.request<RevenueResponse>('GET', '/api/v1/admin/revenue')
  }

  getAIHealth() {
    return this.request<AIHealthResponse>('GET', '/api/v1/admin/ai/health')
  }

  getSystemHealth() {
    return this.request<SystemHealthResponse>('GET', '/api/v1/admin/system/health')
  }

  getAlerts() {
    return this.request<AlertsResponse>('GET', '/api/v1/admin/alerts')
  }

  getActivity() {
    return this.request<ActivityResponse>('GET', '/api/v1/admin/activity')
  }

  getAuditLogs(params?: { search?: string; action?: string; user_id?: string; limit?: number }) {
    const query = new URLSearchParams()
    if (params?.search) query.set('search', params.search)
    if (params?.action) query.set('action', params.action)
    if (params?.user_id) query.set('user_id', params.user_id)
    if (params?.limit) query.set('limit', String(params.limit))
    const qs = query.toString()
    return this.request<AuditLogsResponse>('GET', `/api/v1/admin/audit-logs${qs ? '?' + qs : ''}`)
  }

  getKnowledgeBase(params?: { status?: string; search?: string; limit?: number }) {
    const query = new URLSearchParams()
    if (params?.status) query.set('status', params.status)
    if (params?.search) query.set('search', params.search)
    if (params?.limit) query.set('limit', String(params.limit))
    const qs = query.toString()
    return this.request<KnowledgeBaseResponse>('GET', `/api/v1/admin/knowledge-base${qs ? '?' + qs : ''}`)
  }

  trainKnowledge(data: TrainKnowledgeRequest) {
    return this.request<{ message: string }>('POST', '/api/v1/admin/knowledge-base/train', data)
  }

  getReferral() {
    return this.request<{ code: string; url: string; clicks: number; signups: number; conversions: number }>('GET', '/api/v1/admin/referral')
  }

  getSalesLeads(params?: { status?: string }) {
    const query = new URLSearchParams()
    if (params?.status) query.set('status', params.status)
    const qs = query.toString()
    return this.request<{ leads: SalesLead[]; total: number }>('GET', `/api/v1/admin/sales-leads${qs ? '?' + qs : ''}`)
  }

  createSalesLead(data: { contact_name: string; contact_phone?: string; contact_email?: string; business_name?: string; business_type?: string; status?: string; notes?: string; meeting_location?: string }) {
    return this.request<{ id: string }>('POST', '/api/v1/admin/sales-leads', data)
  }

  updateSalesLead(id: string, data: { status?: string; notes?: string }) {
    return this.request<{ message: string }>('PUT', `/api/v1/admin/sales-leads/${id}`, data)
  }

  getPipelineStats() {
    return this.request<{ pipeline: { status: string; count: number }[]; total_leads: number }>('GET', '/api/v1/admin/pipeline-stats')
  }
}

export const adminApi = new AdminAPI()
