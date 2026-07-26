import { useState, useEffect } from 'react'
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

  useEffect(() => {
    adminApi.getAnalytics()
      .then(res => setData(res))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return { data, loading }
}
