import type { WSMessage } from '@/types'

type MessageHandler = (msg: WSMessage) => void

// Match backend port - adjust if your backend runs on a different port
const BACKEND_PORT = 8080

class WebSocketManager {
  private ws: WebSocket | null = null
  private handlers: Set<MessageHandler> = new Set()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private url: string = ''
  private currentToken: string = ''

  connect(token: string): void {
    // Don't reconnect if already connected with same token
    if (this.ws?.readyState === WebSocket.OPEN && this.currentToken === token) {
      return
    }
    
    // Close existing before opening new
    if (this.ws) {
      this.ws.onclose = null // Prevent auto-reconnect on intentional close
      this.ws.close()
      this.ws = null
    }
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    this.currentToken = token
    
    // Use same hostname but backend port for WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.hostname
    const port = host === 'localhost' ? BACKEND_PORT : window.location.port
    const portStr = port ? `:${port}` : ''
    this.url = `${protocol}//${host}${portStr}/ws?token=${token}`
    
    try {
      this.ws = new WebSocket(this.url)
      
      this.ws.onopen = () => {
        console.log('WebSocket connected to', this.url)
      }
      
      this.ws.onmessage = (event) => {
        try {
          const raw = JSON.parse(event.data)
          // Normalize backend shape: { conversation_id, type, data } -> WSMessage
          const msg: WSMessage = {
            type: raw.type || 'new_message',
            conversation_id: raw.conversation_id || raw.ConversationID,
            content: raw.content || raw.data?.content || raw.Data?.content,
            sender_type: raw.sender_type || raw.data?.sender_type || raw.Data?.sender_type,
            timestamp: raw.timestamp || raw.created_at || new Date().toISOString(),
          }
          this.handlers.forEach(h => h(msg))
        } catch (err) {
          console.error('WebSocket message parse error:', err)
        }
      }
      
      this.ws.onclose = () => {
        this.ws = null
        // Only auto-reconnect if we haven't explicitly disconnected
        if (this.currentToken) {
          // Use latest token from storage in case it was refreshed by api.ts
          const latestToken = localStorage.getItem('noant_token') || this.currentToken
          this.reconnectTimer = setTimeout(() => {
            this.connect(latestToken)
          }, 3000)
        }
      }
      
      this.ws.onerror = (err) => {
        console.error('WebSocket error:', err)
        this.ws?.close()
      }
    } catch (err) {
      this.reconnectTimer = setTimeout(() => this.connect(token), 5000)
    }
  }

  onMessage(handler: MessageHandler): () => void {
    this.handlers.add(handler)
    return () => this.handlers.delete(handler)
  }

  disconnect(): void {
    this.currentToken = '' // Stop auto-reconnect
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.onclose = null // Prevent reconnect loop
      this.ws.close()
      this.ws = null
    }
    // DON'T clear handlers � they belong to components that may persist across route changes
  }
}

export const ws = new WebSocketManager()