import { useState, useEffect } from 'react'
import type { AnalyticsData } from '@/types'

const MOCK_ANALYTICS: AnalyticsData = {
  visitors_today: 1247,
  visitors_yesterday: 1089,
  signups_today: 23,
  conversion_rate: 1.84,
  bounce_rate: 42.3,
  avg_session_duration: 185,
  page_views: [
    { path: '/', views: 4230, unique_visitors: 2810, avg_time_on_page: 45, bounce_rate: 38 },
    { path: '/pricing', views: 1890, unique_visitors: 1420, avg_time_on_page: 120, bounce_rate: 25 },
    { path: '/features', views: 1560, unique_visitors: 1180, avg_time_on_page: 95, bounce_rate: 35 },
    { path: '/signup', views: 890, unique_visitors: 890, avg_time_on_page: 180, bounce_rate: 15 },
    { path: '/docs', views: 670, unique_visitors: 450, avg_time_on_page: 240, bounce_rate: 20 },
    { path: '/blog', views: 340, unique_visitors: 280, avg_time_on_page: 300, bounce_rate: 30 },
  ],
  traffic_sources: [
    { source: 'Google Ads', visitors: 420, signups: 12, conversion: 2.86 },
    { source: 'Organic Search', visitors: 380, signups: 8, conversion: 2.11 },
    { source: 'Twitter/X', visitors: 210, signups: 3, conversion: 1.43 },
    { source: 'Direct', visitors: 145, signups: 0, conversion: 0 },
    { source: 'Referral', visitors: 92, signups: 0, conversion: 0 },
  ],
  visitor_history: [
    { date: 'Jul 20', visitors: 890, signups: 15 },
    { date: 'Jul 21', visitors: 1020, signups: 18 },
    { date: 'Jul 22', visitors: 1150, signups: 22 },
    { date: 'Jul 23', visitors: 980, signups: 16 },
    { date: 'Jul 24', visitors: 1089, signups: 19 },
    { date: 'Jul 25', visitors: 1180, signups: 21 },
    { date: 'Jul 26', visitors: 1247, signups: 23 },
  ],
  funnel: [
    { step: 'Landing Page Visit', count: 1247, percentage: 100, dropoff: 0 },
    { step: 'Viewed Pricing', count: 680, percentage: 54.5, dropoff: 45.5 },
    { step: 'Started Signup', count: 234, percentage: 18.8, dropoff: 35.7 },
    { step: 'Completed Signup', count: 89, percentage: 7.1, dropoff: 11.7 },
    { step: 'Onboarded', count: 56, percentage: 4.5, dropoff: 2.6 },
    { step: 'First AI Message', count: 41, percentage: 3.3, dropoff: 1.2 },
  ],
}

export function useAnalytics() {
  const [data, setData] = useState<AnalyticsData>(MOCK_ANALYTICS)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setData(MOCK_ANALYTICS)
      setLoading(false)
    }, 400)
    return () => clearTimeout(timer)
  }, [])

  return { data, loading }
}
