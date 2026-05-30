import { api } from './api.ts'
import type { AuthResponse, LoginRequest, SignupRequest, User } from '@/types'

export function getToken(): string | null {
  return null
}

export function setToken(_token: string, _refresh?: string): void {
  // Sessions are now handled with httpOnly cookies.
}

export function clearAuth(): void {
  localStorage.removeItem('noant_token')
  localStorage.removeItem('noant_refresh')
  import('./websocket').then(({ ws }) => ws.disconnect()).catch(() => undefined)
}

export function isAuthenticated(): boolean {
  return false
}

export async function login(credentials: LoginRequest): Promise<AuthResponse> {
  return api.post<AuthResponse>('/auth/login', credentials)
}

// Backend /auth/register returns { user }; we follow up with a cookie-based login.
export async function signup(data: SignupRequest): Promise<AuthResponse> {
  await api.post<{ message: string; user: User }>('/auth/register', data)
  return login({ email: data.email, password: data.password })
}

export async function logout(): Promise<void> {
  try {
    await api.post<{ message: string }>('/auth/logout')
  } catch {
    // Ignore - always clear local state
  } finally {
    clearAuth()
    window.location.href = '/login'
  }
}

export async function getCurrentUser(): Promise<User> {
  const data = await api.get<{ user: User }>('/auth/me')
  return data.user
}

export async function refreshToken(): Promise<string | null> {
  try {
    await api.post<{ message: string }>('/auth/refresh', {})
    return 'refreshed'
  } catch {
    clearAuth()
    return null
  }
}
