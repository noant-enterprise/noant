import { useState, useEffect, useCallback } from 'react'
import { adminApi } from '@/lib/api'
import type { LiveFeedEvent, Alert } from '@/types'

export function useLiveFeed() {
  const [events, setEvents] = useState<LiveFeedEvent[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      adminApi.getActivity().catch(() => ({ events: [] })),
      adminApi.getAlerts().catch(() => ({ alerts: [] })),
    ]).then(([activityRes, alertsRes]) => {
      setEvents(activityRes.events || [])
      setAlerts(
        (alertsRes.alerts || []).map(a => ({
          id: a.id,
          type: a.severity as Alert['type'],
          title: a.title,
          message: a.description,
          timestamp: a.created_at,
          acknowledged: false,
        }))
      )
    }).finally(() => setLoading(false))
  }, [])

  const acknowledgeAlert = useCallback((id: string) => {
    setAlerts(prev => prev.map(a => a.id === id ? { ...a, acknowledged: true } : a))
  }, [])

  return { events, alerts, loading, acknowledgeAlert }
}
