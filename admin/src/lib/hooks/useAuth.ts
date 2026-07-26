import { useState, useEffect, useCallback } from 'react'
import { adminApi } from '@/lib/api'
import type { AdminUser } from '@/types'

export function useAuth() {
  const [user, setUser] = useState<AdminUser | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    adminApi.me()
      .then(res => {
        setUser(res.user as AdminUser)
      })
      .catch(() => {
        setUser(null)
      })
      .finally(() => setLoading(false))
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const res = await adminApi.login(email, password)
    const adminUser: AdminUser = {
      id: res.user.id,
      email: res.user.email,
      role: res.user.role as AdminUser['role'],
    }
    setUser(adminUser)
    return adminUser
  }, [])

  const logout = useCallback(async () => {
    await adminApi.logout().catch(() => {})
    setUser(null)
  }, [])

  return { user, loading, login, logout, isAuthenticated: !!user }
}
