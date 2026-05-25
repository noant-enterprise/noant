import { useEffect, useState, useCallback, useRef } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { useAPI } from '@/hooks/useAPI'
import { useToast } from '@/components/ui/Toast'
import { ChatList, ChatMessages, ChatInput, CustomerInfo } from '@/components/chat'
import { Skeleton } from '@/components/ui/Skeleton'
import { useWebSocket } from '@/hooks/useWebSocket'
import { Avatar } from '@/components/ui/Avatar'
import { cn } from '@/lib/utils'
import { ArrowLeft, Info, Sparkles } from 'lucide-react'
import type { Conversation, Message, WSMessage } from '@/types'

export default function ChatsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const activeId = searchParams.get('id')
  const isMobileChatView = !!activeId
  
  const convAPI = useAPI() as any
  const { data: convData, get: getConversations, loadMore, loading: convLoading, loadingMore } = convAPI
  
  const msgAPI = useAPI() as any
  const { data: msgData, get: getMessages } = msgAPI
  
  const postAPI = useAPI() as any
  const { post } = postAPI
  
  const putAPI = useAPI() as any
  const { put } = putAPI
  
  const { toast } = useToast()
  const { subscribe } = useWebSocket()

  const [optimisticMessages, setOptimisticMessages] = useState<Message[]>([])
  const [sending, setSending] = useState(false)
  const [typing, setTyping] = useState(false)
  const [showInfo, setShowInfo] = useState(false)
  const [page, setPage] = useState(1)
  const [allConversations, setAllConversations] = useState<Conversation[]>([])
  const [aiInitLoading, setAiInitLoading] = useState(false)
  const [pendingAI, setPendingAI] = useState<Set<string>>(new Set())
  const aiPollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const lastSentContent = useRef<string>('')

  // Safe data extraction
  const conversations = convData?.conversations || []
  const messages = msgData?.messages || []
  const hasMore = convData?.has_more || false

  useEffect(() => {
    if (!isMobileChatView) return
    window.history.pushState({ chatView: true }, '', window.location.href)
    const handlePopState = () => navigate('/chats', { replace: true })
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [isMobileChatView, navigate])

  // Poll every 800ms for AI response when pending
  useEffect(() => {
    if (!activeId || !pendingAI.has(activeId)) {
      if (aiPollRef.current) {
        clearInterval(aiPollRef.current)
        aiPollRef.current = null
      }
      return
    }

    aiPollRef.current = setInterval(() => {
      getMessages(`/chats/conversations/${activeId}`)
    }, 800)

    return () => {
      if (aiPollRef.current) {
        clearInterval(aiPollRef.current)
        aiPollRef.current = null
      }
    }
  }, [activeId, pendingAI, getMessages])

  // Stop polling once AI response arrives
  useEffect(() => {
    if (!activeId || !pendingAI.has(activeId)) return
    const hasAIResponse = messages.some((m: Message) => m.sender_type === 'ai')
    const hasUserMessage = messages.some((m: Message) => m.sender_type === 'customer')
    if (hasAIResponse && hasUserMessage) {
      setPendingAI(prev => {
        const next = new Set(prev)
        next.delete(activeId)
        return next
      })
      setTyping(false)
    }
  }, [messages, activeId, pendingAI])

  // Clear optimistic messages ONLY when real message with matching content arrives
  useEffect(() => {
    if (optimisticMessages.length === 0 || !activeId) return
    
    const lastOptimistic = optimisticMessages[optimisticMessages.length - 1]
    const foundReal = messages.some((m: Message) => 
      m.sender_type === 'customer' && m.content === lastOptimistic.content
    )
    
    if (foundReal) {
      setOptimisticMessages([])
    }
  }, [messages, optimisticMessages, activeId])

  useEffect(() => {
    getConversations('/chats/conversations?page=1&limit=20')
  }, [getConversations])

  useEffect(() => {
    if (conversations.length > 0) {
      setAllConversations(prev => 
        page === 1 ? conversations : [...prev, ...conversations]
      )
    }
  }, [conversations, page])

  useEffect(() => {
    if (activeId) {
      getMessages(`/chats/conversations/${activeId}`)
      setOptimisticMessages([])
    }
  }, [activeId, getMessages])

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
        setPendingAI(prev => {
          const next = new Set(prev)
          next.delete(activeId)
          return next
        })
      }
      if (msg.type === 'new_conversation') {
        getConversations('/chats/conversations?page=1&limit=20')
      }
    })
    return unsub
  }, [activeId, subscribe, getMessages, getConversations])

  const activeConv = allConversations.find((c) => c.id === activeId) || null

  const handleLoadMore = useCallback(() => {
    if (loadingMore || !hasMore) return
    const nextPage = page + 1
    setPage(nextPage)
    loadMore(`/chats/conversations?page=${nextPage}&limit=20`)
  }, [page, loadingMore, hasMore, loadMore])

  const handleSend = async (text: string) => {
    if (!activeId) return
    
    const tempId = `temp-${Date.now()}`
    const optimisticMsg: Message = {
      id: tempId,
      conversation_id: activeId,
      content: text,
      sender_type: 'customer',
      created_at: new Date().toISOString(),
    }
    
    lastSentContent.current = text
    
    // Show immediately — DON'T clear until real message arrives
    setOptimisticMessages(prev => [...prev, optimisticMsg])
    setSending(true)
    setTyping(true)
    
    try {
      await post(`/chats/conversations/${activeId}/messages`, { content: text })
      
      // Trigger fetch but keep optimistic visible
      getMessages(`/chats/conversations/${activeId}`)
      
      // Mark AI as pending
      setPendingAI(prev => new Set(prev).add(activeId))
    } catch (err) {
      console.error('Send failed:', err)
      toast('Failed to send message', 'error')
      setOptimisticMessages(prev => prev.map(m => 
        m.id === tempId ? { ...m, content: `${m.content} (failed)` } : m
      ))
      setTyping(false)
    } finally {
      setSending(false)
    }
  }

  const handleTakeover = async () => {
    if (!activeId) return
    try {
      await put(`/chats/conversations/${activeId}/takeover`)
      toast('You took over this conversation', 'success')
      getConversations('/chats/conversations?page=1&limit=20')
    } catch {
      toast('Failed to take over', 'error')
    }
  }

  const handleBackToList = () => {
    navigate('/chats')
  }

  const openAIChat = async () => {
    const aiConv = allConversations.find((c: Conversation) => 
      c.channel === 'web' && c.customer_name === 'Noant AI'
    )
    if (aiConv) {
      setSearchParams({ id: aiConv.id })
    } else {
      setAiInitLoading(true)
      try {
        const result = await post('/chats/direct-chat', {
          customer_name: 'Noant AI',
          message: 'Hello! I am your AI assistant. How can I help you today?',
          channel: 'web'
        })
        if (result?.conversation?.id) {
          await getConversations('/chats/conversations?page=1&limit=20')
          setSearchParams({ id: result.conversation.id })
          toast('AI assistant ready!', 'success')
        }
      } catch (err) {
        console.error('Failed to create AI conversation:', err)
        toast('Could not start AI chat', 'error')
      } finally {
        setAiInitLoading(false)
      }
    }
  }

  // Deduplicate: if real message exists with same content as optimistic, show real one only
  const dedupedMessages = [...messages]
  const realContents = new Set(messages.map((m: Message) => m.content))
  
  const visibleOptimistic = optimisticMessages.filter((om) => !realContents.has(om.content))
  
  const allMessages = [...dedupedMessages, ...visibleOptimistic]

  return (
    <div className="h-full w-full flex animate-page-in overflow-hidden relative">
      {/* Conversation List */}
      <div className={cn(
        'w-full lg:w-72 lg:shrink-0 lg:border-r lg:border-default bg-surface h-full overflow-y-auto absolute lg:relative inset-0 lg:inset-auto z-10 transition-transform duration-300 lg:transition-none',
        isMobileChatView ? '-translate-x-full lg:translate-x-0' : 'translate-x-0'
      )}>
        {/* AI Assistant Button */}
        <div className="p-3 border-b border-default">
          <button
            onClick={openAIChat}
            disabled={aiInitLoading}
            className="w-full flex items-center gap-3 p-3 rounded-xl bg-gradient-to-r from-noant-sky/10 to-noant-sky/5 border border-noant-sky/20 hover:border-noant-sky/40 transition-all active:scale-[0.98] btn-press disabled:opacity-50"
          >
            <div className="w-10 h-10 rounded-full bg-noant-sky/20 flex items-center justify-center shrink-0">
              <Sparkles className="w-5 h-5 text-noant-sky" />
            </div>
            <div className="flex-1 text-left">
              <p className="text-sm font-semibold text-primary">Noant AI</p>
              <p className="text-xs text-noant-sky">
                {aiInitLoading ? 'Starting...' : 'Chat with your assistant'}
              </p>
            </div>
          </button>
        </div>

        {convLoading && page === 1 ? (
          <div className="p-3 space-y-2">
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
          <ChatList
            conversations={allConversations}
            activeId={activeId || undefined}
            hasMore={hasMore}
            loadingMore={loadingMore}
            onLoadMore={handleLoadMore}
          />
        )}
      </div>

      {/* Chat Area */}
      <div className={cn(
        'flex-1 flex flex-col bg-base lg:border-x lg:border-default min-w-0 h-full overflow-hidden absolute lg:relative inset-0 lg:inset-auto z-20 transition-transform duration-300 lg:transition-none bg-surface',
        isMobileChatView ? 'translate-x-0' : 'translate-x-full lg:translate-x-0'
      )}>
        {activeConv ? (
          <>
            <div className="h-12 lg:h-11 border-b border-default flex items-center px-3 lg:px-4 gap-3 shrink-0 select-none bg-surface/80 backdrop-blur-sm">
              <button
                onClick={handleBackToList}
                className="lg:hidden w-8 h-8 rounded-full bg-inset flex items-center justify-center text-secondary hover:text-primary active:scale-95 transition-all mr-1 btn-press"
                aria-label="Back to conversations"
              >
                <ArrowLeft className="w-[18px] h-[18px]" strokeWidth={2} />
              </button>
              
              <Avatar name={activeConv.customer_name} size="sm" />
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold text-primary truncate">{activeConv.customer_name}</p>
                <p className="text-[10px] text-tertiary capitalize flex items-center gap-1">
                  <span className={cn(
                    "w-1.5 h-1.5 rounded-full animate-live",
                    activeConv.customer_name === 'Noant AI' ? "bg-noant-sky" : "bg-emerald-500"
                  )} />
                  {activeConv.channel} - {activeConv.status}
                  {activeConv.customer_name === 'Noant AI' && (
                    <span className="text-noant-sky ml-1">● AI</span>
                  )}
                  {pendingAI.has(activeId || '') && (
                    <span className="text-noant-sky ml-1 animate-pulse">● typing...</span>
                  )}
                </p>
              </div>
              <button
                onClick={() => setShowInfo(!showInfo)}
                className={cn(
                  'w-8 h-8 rounded-full flex items-center justify-center transition-all active:scale-95 btn-press',
                  showInfo ? 'bg-noant-sky/10 text-noant-sky' : 'bg-inset text-secondary hover:text-primary'
                )}
                aria-label="Toggle customer info"
              >
                <Info className="w-[18px] h-[18px]" strokeWidth={2} />
              </button>
            </div>

            <ChatMessages messages={allMessages} />
            <ChatInput 
              onSend={handleSend} 
              onTakeover={handleTakeover} 
              disabled={!activeId || sending} 
              typing={typing || pendingAI.has(activeId || '')} 
              typingText={pendingAI.has(activeId || '') ? 'Noant AI is thinking...' : 'Customer is typing...'}
            />
          </>
        ) : (
          <div className="hidden lg:flex h-full flex-col items-center justify-center text-center p-4">
            <div className="w-14 h-14 bg-inset rounded-2xl flex items-center justify-center mb-4 animate-float">
              <Sparkles className="w-7 h-7 text-noant-sky" />
            </div>
            <p className="text-base font-semibold text-primary">AI Assistant</p>
            <p className="text-sm text-secondary max-w-xs mt-1 mb-4">
              Click the Noant AI button in the sidebar to start chatting.
            </p>
          </div>
        )}
      </div>

      <CustomerInfo
        conversation={activeConv}
        open={showInfo}
        onClose={() => setShowInfo(false)}
      />
    </div>
  )
}