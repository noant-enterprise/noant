import { useState, useEffect, useCallback } from 'react'
import { adminApi } from '@/lib/api'
import type { AnalyticsResponse } from '@/types'

export function useAnalytics() {
  const [data, setData] = useState<AnalyticsResponse>({
    visitors_today: 0,
    visitors_yesterday: 0,
    signups_today: 0,
    conversion_rate: 0,
    total_signups: 0,
    bounce_rate: 0,
    avg_session_duration: 0,
    page_views: [],
    traffic_sources: [],
    visitor_history: [],
    funnel: [],
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = useCallback(() => {
    setLoading(true)
    setError(null)
    adminApi.getAnalytics()
      .then(res => setData({
        ...res,
        page_views: res.page_views ?? [],
        traffic_sources: res.traffic_sources ?? [],
        visitor_history: res.visitor_history ?? [],
        funnel: res.funnel ?? [],
      }))
      .catch(() => setError('Failed to load analytics'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}
