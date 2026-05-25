import { api } from './api.ts'
import type { AuthResponse, LoginRequest, SignupRequest, User } from '@/types'

const TOKEN_KEY = 'noant_token'
const REFRESH_KEY = 'noant_refresh'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string, refresh?: string): void {
  localStorage.setItem(TOKEN_KEY, token)
  if (refresh) localStorage.setItem(REFRESH_KEY, refresh)
}

export function clearAuth(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_KEY)
  // Disconnect WebSocket so reconnect loop stops immediately on logout
  import('./websocket').then(({ ws }) => ws.disconnect()).catch(() => undefined)
}

export function isAuthenticated(): boolean {
  return !!getToken()
}

export async function login(credentials: LoginRequest): Promise<AuthResponse> {
  const data = await api.post<AuthResponse>('/auth/login', credentials)
  setToken(data.token, data.refresh_token)
  return data
}

// Backend /auth/register returns { user } with NO token.
// We auto-login after registration to get a token.
export async function signup(data: SignupRequest): Promise<AuthResponse> {
  await api.post<{ message: string; user: User }>('/auth/register', data)
  // Auto-login to obtain a JWT access token
  return login({ email: data.email, password: data.password })
}

export async function logout(): Promise<void> {
  try {
    // Tell backend to blacklist the current token
    await api.post<{ message: string }>('/auth/logout')
  } catch {
    // Ignore — always clear local state
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
  const refresh = localStorage.getItem(REFRESH_KEY)
  if (!refresh) return null
  try {
    const data = await api.post<{ token: string }>('/auth/refresh', { refresh_token: refresh })
    setToken(data.token)
    return data.token
  } catch {
    clearAuth()
    return null
  }
}