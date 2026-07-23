import { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react'
import { api } from '@/lib/api'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useAuth } from '@/hooks/useAuth'

export interface SidebarAlerts {
  /** Conversations with no agent/AI reply yet */
  unreadChats: number
  /** Unknown questions pending training */
  unknownQuestions: number
  /** Channels with connection/auth issues */
  channelIssues: number
  /** Billing warnings (trial ending, payment failed, over-limit) */
  billingAlert: boolean
  /** Unread notification count (for bell icon) */
  unreadNotifications: number
  /** Total badge across all alerts */
  total: number
}

const defaultAlerts: SidebarAlerts = {
  unreadChats: 0,
  unknownQuestions: 0,
  channelIssues: 0,
  billingAlert: false,
  unreadNotifications: 0,
  total: 0,
}

interface SidebarAlertsContextValue extends SidebarAlerts {
  refreshAlerts: () => void
}

const SidebarAlertsContext = createContext<SidebarAlertsContextValue>({ ...defaultAlerts, refreshAlerts: () => {} })

export function SidebarAlertsProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth()
  const { subscribe } = useWebSocket()
  const [alerts, setAlerts] = useState<SidebarAlerts>(defaultAlerts)
  const lastFetchRef = useRef<number>(0)

  const fetchAlerts = useCallback(async () => {
    if (!user) return
    // Debounce — don't refetch more than once per 6s
    if (Date.now() - lastFetchRef.current < 6000) return
    lastFetchRef.current = Date.now()

    try {
      // Fetch in parallel — all are existing endpoints
      const [overviewRes, uqRes, notifRes, intRes] = await Promise.allSettled([
        // Analytics overview gives us total conversations and can surface billing
        api.get<{
          total_conversations?: number
          unread_conversations?: number
          billing_alert?: boolean
          plan?: string
        }>('/analytics/overview'),

        // Unknown questions list with status=pending (count from array length)
        api.get<{ questions?: any[] }>(
          '/training/unknown-questions?status=pending'
        ),

        // Unread notification count
        api.get<{ count: number }>('/notifications/unread-count'),

        // Channel integrations — count those with error status
        api.get<{
          integrations?: Array<{ status: string; channel: string }>;
          channels?: Array<{ status: string }>
        }>('/integrations/list'),
      ])

      // Unread chats: from overview, fallback 0
      const unreadChats =
        overviewRes.status === 'fulfilled'
          ? (overviewRes.value?.unread_conversations ?? 0)
          : 0

      // Billing alert: from overview
      const billingAlert =
        overviewRes.status === 'fulfilled'
          ? (overviewRes.value?.billing_alert ?? false)
          : false

      // Unknown questions: count from array
      const unknownQuestions =
        uqRes.status === 'fulfilled'
          ? (uqRes.value?.questions?.length ?? 0)
          : 0

      // Unread notifications
      const unreadNotifications =
        notifRes.status === 'fulfilled'
          ? (notifRes.value?.count ?? 0)
          : 0

      // Channel issues: integrations with error status
      let channelIssues = 0
      if (intRes.status === 'fulfilled') {
        const items = intRes.value?.integrations || intRes.value?.channels || []
        channelIssues = items.filter(
          (ch: any) => ch.status === 'error'
        ).length
      }

      const total =
        unreadChats +
        unknownQuestions +
        channelIssues +
        unreadNotifications +
        (billingAlert ? 1 : 0)

      setAlerts({
        unreadChats,
        unknownQuestions,
        channelIssues,
        billingAlert,
        unreadNotifications,
        total,
      })
    } catch {
      // Silently swallow — sidebar badges are non-critical
    }
  }, [user])

  const refreshAlerts = useCallback(() => {
    lastFetchRef.current = 0
    fetchAlerts()
  }, [fetchAlerts])

  // Initial fetch + poll every 30s
  useEffect(() => {
    fetchAlerts()
    const interval = setInterval(fetchAlerts, 30000)
    return () => clearInterval(interval)
  }, [fetchAlerts])

  // Realtime: refresh on new message or unknown question events
  useEffect(() => {
    const unsub = subscribe((msg) => {
      if (
        msg.type === 'new_message' ||
        msg.type === 'unknown_question' ||
        msg.type === 'notification'
      ) {
        // Reset debounce so the live event always triggers a refresh
        lastFetchRef.current = 0
        fetchAlerts()
      }
    })
    return unsub
  }, [subscribe, fetchAlerts])

  return (
    <SidebarAlertsContext.Provider value={{ ...alerts, refreshAlerts }}>
      {children}
    </SidebarAlertsContext.Provider>
  )
}

export function useSidebarAlerts(): SidebarAlertsContextValue {
  return useContext(SidebarAlertsContext)
}
