import { useState, useEffect, useCallback, useRef } from 'react'
import { clearAuth, getCurrentUser, logout } from '@/lib/auth'
import type { User } from '@/types'

export function useAuth() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const fetchedRef = useRef(false)

  useEffect(() => {
    if (fetchedRef.current) return
    fetchedRef.current = true
    
    getCurrentUser()
      .then(setUser)
      .catch((err: any) => {
        setUser(null)
        if (err?.status === 401) {
          clearAuth()
        }
      })
      .finally(() => setLoading(false))
  }, [])

  const signOut = useCallback(() => {
    logout()
  }, [])

  const refreshUser = useCallback(async () => {
    try {
      const updated = await getCurrentUser()
      setUser(updated)
    } catch {
      // ignore
    }
  }, [])

  return { user, loading, signOut, refreshUser }
}
