import { useCallback, useEffect, useState } from 'react'
import { useWebSocket } from '@/hooks/useWebSocket'
import { cn } from '@/lib/utils'
import type { WSMessage } from '@/types'
import { ConversationLoading } from './ConversationLoading'

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
        'max-w-[85%] lg:max-w-[75%] px-4 py-2.5 rounded-2xl rounded-bl-md',
        'self-start bg-surface border border-default text-primary',
        'flex items-center gap-3 shadow-sm animate-fade-in'
      )}
    >
      <ConversationLoading size="sm" className="shrink-0" />
      <span className="text-xs text-tertiary">AI is thinking...</span>
    </div>
  )
}
