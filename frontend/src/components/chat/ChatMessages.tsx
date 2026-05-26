import { useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { Check, CheckCheck, Bot, UserCheck, Zap } from 'lucide-react'
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll'
import type { Message } from '@/types'

interface ChatMessagesProps {
  messages: Message[]
  deliveredIds?: Set<string>
  hasMore?: boolean
  loadingMore?: boolean
  onLoadMore?: () => void
}

function SourceBadge({ message }: { message: Message }) {
  const role = message.role || message.sender_type
  const confidence = message.metadata?.confidence ?? message.confidence

  if (role === 'customer') return null

  const badgeConfig: Record<string, { label: string; icon: typeof Bot; color: string }> = {
    ai: { label: 'AI', icon: Bot, color: 'bg-sky-100 dark:bg-sky-900/30 text-sky-700 dark:text-sky-400' },
    agent: { label: 'Agent', icon: UserCheck, color: 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400' },
    system: { label: 'System', icon: Zap, color: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400' },
  }

  const config = badgeConfig[role || ''] || badgeConfig.ai
  if (!config || role === 'customer') return null
  const Icon = config.icon

  return (
    <span className={cn('inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-[9px] font-semibold', config.color)}>
      <Icon className="w-2.5 h-2.5" />
      {config.label}
      {confidence != null && role === 'ai' && (
        <span className="ml-0.5 opacity-70">{Math.round(confidence * 100)}%</span>
      )}
    </span>
  )
}

const formatDateHeader = (dateStr: string) => {
  const date = new Date(dateStr)
  const today = new Date()
  const yesterday = new Date()
  yesterday.setDate(today.getDate() - 1)

  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  } else if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  } else {
    return date.toLocaleDateString([], { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
  }
}

export function ChatMessages({
  messages,
  deliveredIds,
  hasMore = false,
  loadingMore = false,
  onLoadMore,
}: ChatMessagesProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const lastScrollHeightRef = useRef<number>(0)
  const shouldScrollToBottomRef = useRef<boolean>(true)

  const { setSentinel } = useInfiniteScroll({
    onLoadMore: onLoadMore || (() => {}),
    hasMore,
    loading: loadingMore,
  })

  // Prevent scroll jumping when prepending historical messages
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    if (shouldScrollToBottomRef.current) {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    } else {
      const diff = container.scrollHeight - lastScrollHeightRef.current
      container.scrollTop = container.scrollTop + diff
    }
  }, [messages])

  // Track if we should auto-scroll or lock position
  const handleScroll = () => {
    const container = containerRef.current
    if (!container) return

    const isNearBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 120
    shouldScrollToBottomRef.current = isNearBottom
    lastScrollHeightRef.current = container.scrollHeight
  }

  // Force scroll to bottom when loading a new thread
  useEffect(() => {
    shouldScrollToBottomRef.current = true
    bottomRef.current?.scrollIntoView({ behavior: 'auto' })
  }, [messages[0]?.conversation_id])

  if (messages.length === 0) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-center p-6 lg:p-4 bg-base">
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
    <div
      ref={containerRef}
      onScroll={handleScroll}
      className="flex-1 overflow-y-auto overscroll-contain px-3 lg:px-4 py-4 flex flex-col gap-1.5 bg-base"
    >
      {/* Sentinel loader for older messages */}
      {hasMore && (
        <div ref={setSentinel} className="py-2.5 flex justify-center w-full shrink-0">
          {loadingMore ? (
            <span className="inline-block w-4 h-4 border-2 border-noant-sky border-t-transparent rounded-full animate-spin" />
          ) : (
            <div className="h-4 w-full" />
          )}
        </div>
      )}

      <div className="flex-1" />

      {messages.map((m, index) => {
        const isOptimistic = m.id.startsWith('temp-')
        const isFailed = m.content.includes('(failed')
        const role = m.role || m.sender_type
        const isCustomer = role === 'customer'
        const isSystem = role === 'system'
        const isDelivered = deliveredIds?.has(m.id) || (!isOptimistic && isCustomer)

        const showDateHeader =
          index === 0 ||
          new Date(messages[index - 1].created_at).toDateString() !== new Date(m.created_at).toDateString()

        const dateHeader = showDateHeader ? formatDateHeader(m.created_at) : null

        return (
          <div key={m.id} className="flex flex-col gap-1 w-full shrink-0">
            {dateHeader && (
              <div className="self-center my-3 bg-surface border border-default text-secondary text-[11px] lg:text-[10px] font-semibold px-3 py-1 rounded-full uppercase tracking-wider shadow-sm select-none">
                {dateHeader}
              </div>
            )}
            <div
              className={cn(
                'max-w-[85%] lg:max-w-[75%] px-4 py-2.5 rounded-2xl text-sm lg:text-xs leading-relaxed transition-all duration-300',
                isCustomer
                  ? 'self-end bg-noant-black dark:bg-[#262626] text-white rounded-br-md'
                  : isSystem
                    ? 'self-center bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-400 text-[11px] lg:text-[10px] font-semibold uppercase tracking-wider px-4 py-2 rounded-xl border border-amber-200/50'
                    : 'self-start bg-surface border border-default text-primary rounded-bl-md',
                isOptimistic && 'opacity-60',
                isFailed && 'border-red-300 dark:border-red-700 bg-red-50 dark:bg-red-900/20'
              )}
            >
              {!isCustomer && !isSystem && <SourceBadge message={m} />}
              <p className="whitespace-pre-wrap">{m.content}</p>
              <div
                className={cn(
                  'text-[11px] lg:text-[10px] mt-1.5 flex items-center gap-1.5',
                  isCustomer ? 'text-white/50' : 'text-tertiary'
                )}
              >
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
          </div>
        )
      })}
      <div ref={bottomRef} className="h-2 shrink-0" />
    </div>
  )
}