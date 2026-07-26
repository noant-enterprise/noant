import { useEffect, useRef } from 'react'
import { useAdminWS } from './useAdminWS'

type AdminEventType = 'lead_created' | 'lead_updated' | 'user_signed_up' | 'dashboard_updated'

export function useAutoRefresh(
  refetchFn: () => void,
  intervalMs: number,
  listenEvents?: AdminEventType[],
) {
  const fnRef = useRef(refetchFn)
  fnRef.current = refetchFn

  const handlers = listenEvents
    ? Object.fromEntries(listenEvents.map(e => [e, () => fnRef.current()])) as Partial<Record<AdminEventType, () => void>>
    : {}

  useAdminWS(handlers)

  useEffect(() => {
    const interval = setInterval(() => fnRef.current(), intervalMs)
    return () => clearInterval(interval)
  }, [intervalMs])
}
