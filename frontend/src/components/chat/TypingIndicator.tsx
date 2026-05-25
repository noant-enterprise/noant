import { useCallback, useEffect, useState } from 'react'
import { useWebSocket } from '@/hooks/useWebSocket'
import { cn } from '@/lib/utils'
import type { WSMessage } from '@/types'

interface TypingIndicatorProps {
  conversationId: string | null
}

export function TypingIndicator({ conversationId }: TypingIndicatorProps) {
  const [isTyping, setIsTyping] = useState(false)
  const { subscribe } = useWebSocket()

  const handleMessage = useCallback((msg: WSMessage) => {
    if (!conversationId) return

    const data = msg as any
    if (data.type === 'typing_indicator' && data.data?.conversation_id === conversationId) {
      setIsTyping(data.data.is_typing)
    }
  }, [conversationId])

  useEffect(() => {
    if (!conversationId) return
    const unsub = subscribe(handleMessage)
    return () => unsub()
  }, [conversationId, subscribe, handleMessage])

  // Reset typing state when conversation changes
  useEffect(() => {
    setIsTyping(false)
  }, [conversationId])

  if (!isTyping) return null

  return (
    <div
      className={cn(
        'max-w-[85%] lg:max-w-[75%] px-4 py-3 rounded-2xl rounded-bl-md',
        'self-start bg-surface border border-default text-primary',
        'flex items-center gap-1.5'
      )}
    >
      <div className="flex gap-1">
        <span className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
        <span className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
        <span className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
      </div>
      <span className="text-xs text-tertiary ml-1">AI is thinking...</span>
    </div>
  )
}