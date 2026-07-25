import { useState, useEffect } from 'react'
import type { LiveFeedEvent, Alert } from '@/types'

const MOCK_FEED: LiveFeedEvent[] = [
  { id: '1', type: 'signup', title: 'New signup', description: 'Amara Okonkwo (amara@jewelryng.com) just signed up', timestamp: new Date(Date.now() - 120000).toISOString(), severity: 'low' },
  { id: '2', type: 'payment', title: 'Payment received', description: 'Chioma Okafor renewed Pro plan — ₦15,000', timestamp: new Date(Date.now() - 300000).toISOString(), severity: 'low' },
  { id: '3', type: 'ai_failure', title: 'AI couldn\'t answer', description: '23 customers asked "Do you support Paystack?" — needs knowledge base update', timestamp: new Date(Date.now() - 600000).toISOString(), severity: 'medium' },
  { id: '4', type: 'escalation', title: 'Customer escalation', description: 'Fatima Abubakar\'s chat was escalated — negative sentiment detected', timestamp: new Date(Date.now() - 900000).toISOString(), severity: 'high' },
  { id: '5', type: 'whatsapp_issue', title: 'WhatsApp delay', description: 'Response time > 5s for session +234...890', timestamp: new Date(Date.now() - 1200000).toISOString(), severity: 'medium' },
  { id: '6', type: 'signup', title: 'New signup', description: 'Tunde Bakare (tunde@foodhub.ng) upgraded to Starter', timestamp: new Date(Date.now() - 1800000).toISOString(), severity: 'low' },
  { id: '7', type: 'system', title: 'Migration completed', description: 'DB migration 021_backfill_org_ids applied successfully', timestamp: new Date(Date.now() - 3600000).toISOString(), severity: 'low' },
]

const MOCK_ALERTS: Alert[] = [
  { id: '1', type: 'warning', title: 'Credits running low', message: 'Kemi Adewale has 45 credits remaining (free plan)', timestamp: new Date().toISOString(), acknowledged: false },
  { id: '2', type: 'critical', title: 'Customer inactive', message: 'Emeka Nwosu (Starter plan) hasn\'t logged in for 77 days', timestamp: new Date().toISOString(), acknowledged: false },
  { id: '3', type: 'info', title: 'AI accuracy improving', message: 'Accuracy up 2.1% this week — knowledge base updates working', timestamp: new Date().toISOString(), acknowledged: true },
]

export function useLiveFeed() {
  const [events, setEvents] = useState<LiveFeedEvent[]>(MOCK_FEED)
  const [alerts, setAlerts] = useState<Alert[]>(MOCK_ALERTS)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setEvents(MOCK_FEED)
      setAlerts(MOCK_ALERTS)
      setLoading(false)
    }, 300)
    return () => clearTimeout(timer)
  }, [])

  const acknowledgeAlert = (id: string) => {
    setAlerts(prev => prev.map(a => a.id === id ? { ...a, acknowledged: true } : a))
  }

  return { events, alerts, loading, acknowledgeAlert }
}
