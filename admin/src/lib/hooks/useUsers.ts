import { useState, useEffect } from 'react'
import type { User } from '@/types'

const MOCK_USERS: User[] = [
  { id: '1', email: 'chioma@hairbychi.com', first_name: 'Chioma', last_name: 'Okafor', plan_id: 'pro', status: 'active', created_at: '2026-03-15T10:00:00Z', last_login_at: '2026-07-26T08:30:00Z', total_conversations: 342, total_messages: 4210, credits_remaining: 1850, health_score: 95 },
  { id: '2', email: 'adedeji@techsolutions.ng', first_name: 'Adedeji', last_name: 'Adebayo', plan_id: 'starter', status: 'active', created_at: '2026-04-20T14:00:00Z', last_login_at: '2026-07-25T16:00:00Z', total_conversations: 128, total_messages: 1540, credits_remaining: 320, health_score: 82 },
  { id: '3', email: 'fatima@fashionhouse.com', first_name: 'Fatima', last_name: 'Abubakar', plan_id: 'pro', status: 'active', created_at: '2026-02-10T09:00:00Z', last_login_at: '2026-07-26T07:15:00Z', total_conversations: 891, total_messages: 12400, credits_remaining: 4200, health_score: 98 },
  { id: '4', email: 'kemi@bakeryplace.com', first_name: 'Kemi', last_name: 'Adewale', plan_id: 'free', status: 'active', created_at: '2026-06-01T11:00:00Z', last_login_at: '2026-07-20T12:00:00Z', total_conversations: 23, total_messages: 180, credits_remaining: 45, health_score: 45 },
  { id: '5', email: 'emeka@autoshop.com', first_name: 'Emeka', last_name: 'Nwosu', plan_id: 'starter', status: 'inactive', created_at: '2026-01-05T08:00:00Z', last_login_at: '2026-05-10T09:00:00Z', total_conversations: 67, total_messages: 890, credits_remaining: 0, health_score: 20 },
  { id: '6', email: 'blessing@beautyspot.com', first_name: 'Blessing', last_name: 'Eze', plan_id: 'pro', status: 'active', created_at: '2026-05-12T13:00:00Z', last_login_at: '2026-07-26T06:00:00Z', total_conversations: 567, total_messages: 7800, credits_remaining: 3100, health_score: 91 },
  { id: '7', email: 'tunde@foodhub.ng', first_name: 'Tunde', last_name: 'Bakare', plan_id: 'starter', status: 'active', created_at: '2026-04-01T10:00:00Z', last_login_at: '2026-07-25T18:00:00Z', total_conversations: 234, total_messages: 3100, credits_remaining: 680, health_score: 78 },
  { id: '8', email: 'amara@jewelryng.com', first_name: 'Amara', last_name: 'Okonkwo', plan_id: 'free', status: 'active', created_at: '2026-07-01T09:00:00Z', last_login_at: '2026-07-26T09:00:00Z', total_conversations: 12, total_messages: 95, credits_remaining: 88, health_score: 70 },
]

export function useUsers(search?: string, plan?: string) {
  const [users, setUsers] = useState<User[]>(MOCK_USERS)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    const timer = setTimeout(() => {
      let filtered = MOCK_USERS
      if (search) {
        const q = search.toLowerCase()
        filtered = filtered.filter(u => u.email.toLowerCase().includes(q) || u.first_name.toLowerCase().includes(q) || u.last_name.toLowerCase().includes(q))
      }
      if (plan && plan !== 'all') {
        filtered = filtered.filter(u => u.plan_id === plan)
      }
      setUsers(filtered)
      setLoading(false)
    }, 300)
    return () => clearTimeout(timer)
  }, [search, plan])

  return { users, loading, total: users.length }
}

export function useUser(id: string) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    const timer = setTimeout(() => {
      setUser(MOCK_USERS.find(u => u.id === id) || null)
      setLoading(false)
    }, 200)
    return () => clearTimeout(timer)
  }, [id])

  return { user, loading }
}
