import { useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { Check, CheckCheck } from 'lucide-react'
import type { Message } from '@/types'

interface ChatMessagesProps {
  messages: Message[]
  deliveredIds?: Set<string>
}

export function ChatMessages({ messages, deliveredIds }: ChatMessagesProps) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  if (messages.length === 0) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-center p-6 lg:p-4">
        <div className="w-16 h-16 lg:w-12 lg:h-12 bg-inset rounded-2xl flex items-center justify-center mb-4">
          <svg className="w-8 h-8 lg:w-6 lg:h-6 text-tertiary" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
          </svg>
        </div>
        <p className="text-base lg:text-sm font-semibold text-primary">Start the conversation</p>
        <p className="text-sm lg:text-xs text-secondary max-w-xs mt-1">
          Select a customer to view messages and respond in real-time.
        </p>
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-y-auto overscroll-contain px-3 lg:px-4 py-4 flex flex-col gap-1 bg-base">
      <div className="flex-1" />
      {messages.map((m) => {
        const isOptimistic = m.id.startsWith('temp-')
        const isFailed = m.content.includes('(failed')
        const isCustomer = m.sender_type === 'customer'
        const isSystem = m.sender_type === 'system'
        const isDelivered = deliveredIds?.has(m.id) || (!isOptimistic && isCustomer)
        
        return (
          <div
            key={m.id}
            className={cn(
              'max-w-[85%] lg:max-w-[75%] px-4 py-2.5 rounded-2xl text-sm lg:text-xs leading-relaxed',
              'transition-all duration-300',
              isCustomer
                ? 'self-end bg-noant-black dark:bg-[#262626] text-white rounded-br-md'
                : isSystem
                  ? 'self-center bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-400 text-[11px] lg:text-[10px] font-semibold uppercase tracking-wider px-4 py-2 rounded-xl'
                  : 'self-start bg-surface border border-default text-primary rounded-bl-md',
              isOptimistic && 'opacity-60',
              isFailed && 'border-red-300 dark:border-red-700 bg-red-50 dark:bg-red-900/20'
            )}
          >
            <p className="whitespace-pre-wrap">{m.content}</p>
            <div className={cn(
              'text-[11px] lg:text-[10px] mt-1 flex items-center gap-1.5',
              isCustomer ? 'text-white/50' : 'text-tertiary'
            )}>
              {new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              
              {isCustomer && (
                <>
                  {isOptimistic && !isFailed ? (
                    <span className="inline-block w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin" />
                  ) : isFailed ? (
                    <span className="text-red-500 font-bold text-xs">!</span>
                  ) : isDelivered ? (
                    <CheckCheck className="w-3 h-3" strokeWidth={2.5} />
                  ) : (
                    <Check className="w-3 h-3" strokeWidth={2.5} />
                  )}
                </>
              )}
            </div>
          </div>
        )
      })}
      <div ref={bottomRef} className="h-2" />
    </div>
  )
}