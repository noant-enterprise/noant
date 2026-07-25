import { useState, useEffect } from 'react'
import type { RevenueData } from '@/types'

const MOCK_REVENUE: RevenueData = {
  mrr: 18400000,
  arr: 220800000,
  total_revenue: 89200000,
  paying_users: 89,
  churn_rate: 3.2,
  ltv: 185000,
  mrr_history: [
    { month: 'Jan', revenue: 4200000, users: 23 },
    { month: 'Feb', revenue: 6100000, users: 34 },
    { month: 'Mar', revenue: 8400000, users: 47 },
    { month: 'Apr', revenue: 10200000, users: 56 },
    { month: 'May', revenue: 12800000, users: 68 },
    { month: 'Jun', revenue: 15600000, users: 78 },
    { month: 'Jul', revenue: 18400000, users: 89 },
  ],
  plan_breakdown: [
    { plan: 'Free', users: 1158, revenue: 0, percentage: 92.9 },
    { plan: 'Starter', users: 52, revenue: 5200000, percentage: 4.2 },
    { plan: 'Pro', users: 37, revenue: 13200000, percentage: 2.9 },
  ],
  failed_payments: [
    { id: '1', user_id: '5', user_email: 'emeka@autoshop.com', amount: 15000, reason: 'Insufficient funds', created_at: '2026-07-26T03:00:00Z' },
    { id: '2', user_id: '7', user_email: 'tunde@foodhub.ng', amount: 15000, reason: 'Card expired', created_at: '2026-07-25T14:00:00Z' },
  ],
}

export function useRevenue() {
  const [data, setData] = useState<RevenueData>(MOCK_REVENUE)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setData(MOCK_REVENUE)
      setLoading(false)
    }, 400)
    return () => clearTimeout(timer)
  }, [])

  return { data, loading }
}
