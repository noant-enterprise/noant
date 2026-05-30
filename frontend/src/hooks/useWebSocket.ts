import { useEffect, useRef, useCallback } from 'react'
import { ws } from '@/lib/websocket'
import type { WSMessage } from '@/types'

/**
 * Connects the global WebSocket singleton and provides a per-component
 * subscribe helper that cleans up on unmount.
 */
export function useWebSocket() {
  const handlersRef = useRef<Set<(msg: WSMessage) => void>>(new Set())

  useEffect(() => {
    ws.connect()
    // Do not disconnect on unmount - other components share the same singleton.
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
