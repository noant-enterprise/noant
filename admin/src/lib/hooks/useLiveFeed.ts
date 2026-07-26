import { useState, useEffect, useCallback } from 'react'
import { adminApi } from '@/lib/api'
import type { LiveFeedEvent, Alert } from '@/types'

export function useLiveFeed() {
  const [events, setEvents] = useState<LiveFeedEvent[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = useCallback(() => {
    setLoading(true)
    setError(null)
    Promise.all([
      adminApi.getActivity().catch(() => ({ events: [] as LiveFeedEvent[] })),
      adminApi.getAlerts().catch(() => ({ alerts: [] as { id: string; type: string; title: string; description: string; severity: string; created_at: string }[] })),
    ]).then(([activityRes, alertsRes]) => {
      setEvents(activityRes.events ?? [])
      setAlerts(
        (alertsRes.alerts ?? []).map(a => ({
          id: a.id,
          type: a.severity as Alert['type'],
          title: a.title,
          message: a.description,
          timestamp: a.created_at,
          acknowledged: false,
        }))
      )
    }).catch(() => setError('Failed to load live feed'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  const acknowledgeAlert = useCallback((id: string) => {
    setAlerts(prev => prev.map(a => a.id === id ? { ...a, acknowledged: true } : a))
  }, [])

  return { events, alerts, loading, error, acknowledgeAlert, refetch: fetchData }
}
