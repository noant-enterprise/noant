import { useState, useEffect } from 'react'
import { adminApi } from '@/lib/api'
import type { RevenueResponse } from '@/types'

export function useRevenue() {
  const [data, setData] = useState<RevenueResponse>({
    mrr: 0,
    arr: 0,
    total_revenue: 0,
    paying_users: 0,
    churn_rate: 0,
    ltv: 0,
    mrr_history: [],
    plan_breakdown: [],
    failed_payments: [],
  })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    adminApi.getRevenue()
      .then(res => setData(res))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return { data, loading }
}
