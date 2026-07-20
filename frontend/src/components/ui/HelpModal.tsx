import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useAPI } from '@/hooks/useAPI'
import { useToast } from '@/components/ui/Toast'
import { ChatList, ChatMessages, ChatInput, CustomerInfo } from '@/components/chat'
import { Skeleton } from '@/components/ui/Skeleton'
import { useWebSocket } from '@/hooks/useWebSocket'
import { cn } from '@/lib/utils'
import type { Conversation, Message, WSMessage } from '@/types'

export default function ChatsPage() {
  const [searchParams] = useSearchParams()
  const activeId = searchParams.get('id')
  const { data: convData, get: getConversations, loading: convLoading } = useAPI<{ conversations: Conversation[] }>()
  const { data: msgData, get: getMessages } = useAPI<{ messages: Message[] }>()
  const { post } = useAPI<Message>()
  const { put } = useAPI()
  const { toast } = useToast()
  const { subscribe } = useWebSocket()

  const [optimisticMessages, setOptimisticMessages] = useState<Message[]>([])
  const [sending, setSending] = useState(false)
  const [typing, setTyping] = useState(false)
  const [showInfo, setShowInfo] = useState(false)

  useEffect(() => {
    getConversations('/chats/conversations')
  }, [])

  useEffect(() => {
    if (activeId) {
      getMessages(`/chats/conversations/${activeId}`)
      setOptimisticMessages([])
    }
  }, [activeId])

  useEffect(() => {
    setShowInfo(false)
  }, [activeId])

  useEffect(() => {
    const unsub = subscribe((msg: WSMessage) => {
      if (msg.type === 'typing' && msg.conversation_id === activeId) {
        setTyping(true)
        setTimeout(() => setTyping(false), 3000)
      }
      if (msg.type === 'new_message' && msg.conversation_id === activeId) {
        getMessages(`/chats/conversations/${activeId}`)
      }
      if (msg.type === 'new_conversation') {
        getConversations('/chats/conversations')
      }
    })
    return unsub
  }, [activeId, subscribe])

  const activeConv = convData?.conversations.find((c) => c.id === activeId) || null
  const allMessages = [...(msgData?.messages || []), ...optimisticMessages]

  const handleSend = async (text: string) => {
    if (!activeId) return
    const tempId = `temp-${Date.now()}`
    const optimisticMsg: Message = {
      id: tempId,
      conversation_id: activeId,
      content: text,
      sender_type: 'agent',
      created_at: new Date().toISOString(),
    }
    setOptimisticMessages(prev => [...prev, optimisticMsg])
    setSending(true)
    try {
      await post(`/chats/conversations/${activeId}/messages`, { content: text, role: 'agent' })
      await getMessages(`/chats/conversations/${activeId}`)
      setOptimisticMessages(prev => prev.filter(m => m.id !== tempId))
    } catch {
      toast('Failed to send message', 'error')
      setOptimisticMessages(prev => prev.map(m => 
        m.id === tempId ? { ...m, content: `${m.content} (failed — tap to retry)` } : m
      ))
    } finally {
      setSending(false)
    }
  }

  const handleTakeover = async () => {
    if (!activeId) return
    try {
      await put(`/chats/conversations/${activeId}/takeover`)
      toast('You took over this conversation', 'success')
    } catch {
      toast('Failed to take over', 'error')
    }
  }

  return (
    <div className="flex-1 flex overflow-hidden animate-fade-in">
      {/* Chat list sidebar */}
      <div className="w-72 shrink-0 border-r border-default bg-surface flex flex-col overflow-hidden">
        {convLoading ? (
          <div className="p-3 space-y-2 overflow-y-auto">
            <Skeleton className="h-8 w-full rounded-md" />
            <div className="flex gap-1">
              <Skeleton className="h-6 w-14 rounded-full" />
              <Skeleton className="h-6 w-14 rounded-full" />
              <Skeleton className="h-6 w-14 rounded-full" />
            </div>
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="flex gap-2 p-2">
                <Skeleton className="w-8 h-8 rounded-full shrink-0" />
                <div className="flex-1 space-y-1.5">
                  <Skeleton className="h-3 w-3/4 rounded" />
                  <Skeleton className="h-2.5 w-1/2 rounded" />
                </div>
              </div>
            ))}
          </div>
        ) : (
          <ChatList conversations={convData?.conversations || []} activeId={activeId || undefined} />
        )}
      </div>

      {/* Main chat area */}
      <div className="flex-1 flex flex-col bg-base border-x border-default min-w-0 overflow-hidden">
        {/* Chat header */}
        {activeConv && (
          <div
            onClick={() => setShowInfo(!showInfo)}
            className="h-11 border-b border-default flex items-center px-4 gap-3 cursor-pointer hover:bg-surface-hover transition-colors shrink-0 select-none"
          >
            <div className="w-7 h-7 rounded-full bg-noant-black text-white dark:bg-white dark:text-noant-black flex items-center justify-center text-[10px] font-bold">
              {activeConv.customer_name.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2)}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-primary truncate">{activeConv.customer_name}</p>
              <p className="text-[10px] text-tertiary capitalize">
                {activeConv.channel} · {activeConv.status}
              </p>
            </div>
            <button
              onClick={(e) => {
                e.stopPropagation()
                setShowInfo(!showInfo)
              }}
              className={cn(
                'p-1.5 rounded-md transition-colors',
                showInfo ? 'bg-noant-sky/10 text-noant-sky' : 'text-tertiary hover:text-primary hover:bg-inset'
              )}
              aria-label="Toggle customer info"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </button>
          </div>
        )}

        {/* Messages area — constrained so it scrolls within bounds */}
        <div className="flex-1 min-h-0 overflow-hidden">
          <ChatMessages messages={allMessages} />
        </div>

        <ChatInput onSend={handleSend} onTakeover={handleTakeover} disabled={!activeId || sending} />
      </div>

      <CustomerInfo
        conversation={activeConv}
        open={showInfo}
        onClose={() => setShowInfo(false)}
      />
    </div>
  )
}