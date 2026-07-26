import { useEffect, useRef, useCallback, useState } from 'react'

type AdminEventType = 'lead_created' | 'lead_updated' | 'user_signed_up' | 'dashboard_updated'

interface AdminEvent {
  type: AdminEventType
  payload?: Record<string, unknown>
}

type EventHandler = (payload?: Record<string, unknown>) => void

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export function useAdminWS(handlers: Partial<Record<AdminEventType, EventHandler>>) {
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef(handlers)
  const [connected, setConnected] = useState(false)
  const reconnectTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)

  handlersRef.current = handlers

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = API_BASE.replace(/^https?:\/\//, '')
    const url = `${protocol}//${host}/api/v1/admin/ws`

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      setConnected(true)
      if (reconnectTimeout.current) clearTimeout(reconnectTimeout.current)
    }

    ws.onmessage = (event) => {
      try {
        const data: AdminEvent = JSON.parse(event.data)
        const handler = handlersRef.current[data.type]
        if (handler) handler(data.payload)
      } catch { /* ignore malformed */ }
    }

    ws.onclose = () => {
      setConnected(false)
      reconnectTimeout.current = setTimeout(connect, 3000)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [])

  useEffect(() => {
    connect()
    return () => {
      if (reconnectTimeout.current) clearTimeout(reconnectTimeout.current)
      wsRef.current?.close()
    }
  }, [connect])

  return { connected }
}
