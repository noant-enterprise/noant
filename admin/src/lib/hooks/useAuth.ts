import { useState, useEffect, useCallback } from 'react'
import type { AdminUser } from '@/types'

export function useAuth() {
  const [user, setUser] = useState<AdminUser | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const stored = localStorage.getItem('admin_user')
    if (stored) {
      setUser(JSON.parse(stored))
    }
    setLoading(false)
  }, [])

  const login = useCallback(async (email: string, _password: string) => {
    // Mock auth — accept any email/password for now
    const adminUser: AdminUser = {
      id: '1',
      email,
      role: 'owner',
    }
    setUser(adminUser)
    localStorage.setItem('admin_user', JSON.stringify(adminUser))
    localStorage.setItem('admin_token', 'mock-admin-token')
    return adminUser
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('admin_user')
    localStorage.removeItem('admin_token')
    setUser(null)
  }, [])

  return { user, loading, login, logout, isAuthenticated: !!user }
}
