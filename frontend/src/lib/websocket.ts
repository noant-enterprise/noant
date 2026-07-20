import type { WSMessage } from '@/types'

type MessageHandler = (msg: WSMessage) => void

class WebSocketManager {
  private ws: WebSocket | null = null
  private handlers: Set<MessageHandler> = new Set()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private url = ''
  private shouldReconnect = false
  private retryCount = 0
  private maxRetries = 15
  private maxRetryDelay = 30000
  private initialRetryDelay = 1000

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    if (this.ws) {
      this.ws.onclose = null
      this.ws.close()
      this.ws = null
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    if (this.retryCount >= this.maxRetries) {
      return
    }

    this.shouldReconnect = true

    const configuredUrl = import.meta.env.VITE_WS_URL as string | undefined
    if (configuredUrl) {
      this.url = configuredUrl
    } else {
      const apiUrl = import.meta.env.VITE_API_URL as string | undefined
      if (apiUrl && apiUrl.startsWith('http')) {
        const wsBase = apiUrl.replace(/^http/, 'ws')
        const urlWithoutV1 = wsBase.replace(/\/api\/v1\/?$/, '').replace(/\/$/, '')
        this.url = `${urlWithoutV1}/ws`
      } else if (apiUrl && !apiUrl.startsWith('/')) {
        const host = apiUrl.replace(/\/api\/v1\/?$/, '').replace(/\/$/, '')
        this.url = `wss://${host}/ws`
      } else {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        this.url = `${protocol}//${window.location.host}/ws`
      }
    }

    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        this.retryCount = 0
      }

      this.ws.onmessage = (event) => {
        try {
          const raw = JSON.parse(event.data)
          const msg: WSMessage = {
            type: raw.type || 'new_message',
            conversation_id: raw.conversation_id || raw.ConversationID,
            content: raw.content || raw.data?.content || raw.Data?.content,
            sender_type: raw.sender_type || raw.data?.sender_type || raw.Data?.sender_type || raw.data?.role || raw.Data?.role,
            timestamp: raw.timestamp || raw.created_at || new Date().toISOString(),
            data: raw.data || raw.Data,
          }
          this.handlers.forEach((handler) => handler(msg))
        } catch {
          // ignore parse errors
        }
      }

      this.ws.onclose = () => {
        this.ws = null
        if (this.shouldReconnect && this.retryCount < this.maxRetries) {
          this.retryCount++
          const delay = Math.min(
            this.initialRetryDelay * Math.pow(2, this.retryCount - 1),
            this.maxRetryDelay
          )
          this.reconnectTimer = setTimeout(() => {
            this.connect()
          }, delay)
        }
      }

      this.ws.onerror = () => {
        this.ws?.close()
      }
    } catch {
      if (this.shouldReconnect && this.retryCount < this.maxRetries) {
        this.retryCount++
        const delay = Math.min(
          this.initialRetryDelay * Math.pow(2, this.retryCount - 1),
          this.maxRetryDelay
        )
        this.reconnectTimer = setTimeout(() => this.connect(), delay)
      }
    }
  }

  onMessage(handler: MessageHandler): () => void {
    this.handlers.add(handler)
    return () => this.handlers.delete(handler)
  }

  reset(): void {
    this.retryCount = 0
  }

  disconnect(): void {
    this.shouldReconnect = false
    this.retryCount = 0
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.onclose = null
      this.ws.close()
      this.ws = null
    }
  }
}

export const ws = new WebSocketManager()
