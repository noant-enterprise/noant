import { useEffect, useRef, useCallback } from 'react'
import { ws } from '@/lib/websocket'
import { getToken } from '@/lib/auth'
import type { WSMessage } from '@/types'

/**
 * Connects the global WebSocket singleton (if not already connected) and
 * provides a per-component subscribe helper that cleans up on unmount.
 *
 * NOTE: We intentionally do NOT call ws.disconnect() on unmount.
 * The singleton should stay alive for the lifetime of the authenticated
 * session; only clearAuth() (called on logout) tears it down.
 */
export function useWebSocket() {
  const handlersRef = useRef<Set<(msg: WSMessage) => void>>(new Set())

  useEffect(() => {
    const token = getToken()
    if (token) ws.connect(token)
    // Do not disconnect on unmount — other components share the same singleton
  }, [])

  const subscribe = useCallback((handler: (msg: WSMessage) => void) => {
    handlersRef.current.add(handler)
    const unsubscribe = ws.onMessage(handler)
    return () => {
      handlersRef.current.delete(handler)
      unsubscribe()
    }
  }, [])

  return { subscribe }
}