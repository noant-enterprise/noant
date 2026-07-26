import { useState, useEffect } from 'react'
import { adminApi } from '@/lib/api'
import type { User } from '@/types'

export function useUsers(search?: string, plan?: string) {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [total, setTotal] = useState(0)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    const timer = setTimeout(() => {
      adminApi.getUsers({ search, plan })
        .then(res => {
          if (!cancelled) {
            setUsers(res.users || [])
            setTotal(res.total)
          }
        })
        .catch(() => {
          if (!cancelled) {
            setUsers([])
            setTotal(0)
            setError('Failed to load users')
          }
        })
        .finally(() => {
          if (!cancelled) setLoading(false)
        })
    }, 300)

    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [search, plan])

  return { users, loading, total, error }
}

export function useUser(id: string) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(null)
    adminApi.getUser(id)
      .then(res => setUser(res as unknown as User))
      .catch(() => { setUser(null); setError('User not found') })
      .finally(() => setLoading(false))
  }, [id])

  return { user, loading, error }
}
