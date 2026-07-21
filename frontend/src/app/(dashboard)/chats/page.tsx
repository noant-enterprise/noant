import { useEffect, useState, useCallback, useRef } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { useAPI } from '@/hooks/useAPI'
import { useToast } from '@/components/ui/Toast'
import { ChatList, ChatMessages, ChatInput, CustomerInfo } from '@/components/chat'
import { Skeleton } from '@/components/ui/Skeleton'
import { useWebSocket } from '@/hooks/useWebSocket'
import { Avatar } from '@/components/ui/Avatar'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { api } from '@/lib/api'
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

  const [showClearConfirm, setShowClearConfirm] = useState(false)
  const [clearLoading, setClearLoading] = useState(false)

  const postAPI = useAPI() as any
  const { post } = postAPI

  const putAPI = useAPI() as any
  const { put } = putAPI

  const { toast } = useToast()
  const { subscribe } = useWebSocket()

  const [optimisticMessages, setOptimisticMessages] = useState<Message[]>([])
  const [sending, setSending] = useState(false)
  const [showInfo, setShowInfo] = useState(false)
  const [page, setPage] = useState(1)
  const [allConversations, setAllConversations] = useState<Conversation[]>([])
  const [aiInitLoading, setAiInitLoading] = useState(false)
  const [pendingAI, setPendingAI] = useState<Set<string>>(new Set())
  const aiPollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const lastSentContent = useRef<string>('')

  // Paginated message states
  const [activeMessages, setActiveMessages] = useState<Message[]>([])
  const [msgPage, setMsgPage] = useState(1)
  const [msgHasMore, setMsgHasMore] = useState(false)
  const [msgLoadingMore, setMsgLoadingMore] = useState(false)
  const [messagesLoading, setMessagesLoading] = useState(false)

  // Safe data extraction
  const conversations = convData?.conversations || []
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

    aiPollRef.current = setInterval(async () => {
      try {
        const res = await api.get<any>(`/chats/conversations/${activeId}?limit=30&page=1`)
        setActiveMessages(res.messages || [])
        setMsgHasMore(res.has_more || false)
        setMsgPage(1)
      } catch (err) {
        console.error('Failed to poll latest messages:', err)
      }
    }, 800)

    return () => {
      if (aiPollRef.current) {
        clearInterval(aiPollRef.current)
        aiPollRef.current = null
      }
    }
  }, [activeId, pendingAI])

  // Stop polling once AI response arrives
  useEffect(() => {
    if (!activeId || !pendingAI.has(activeId)) return
    
    // Check if the latest message is from the AI
    if (activeMessages.length > 0) {
      const lastMessage = activeMessages[activeMessages.length - 1]
      if (lastMessage && lastMessage.sender_type === 'ai') {
        setPendingAI(prev => {
          const next = new Set(prev)
          next.delete(activeId)
          return next
        })
      }
    }
  }, [activeMessages, activeId, pendingAI])

  // Clear optimistic messages ONLY when real message with matching content arrives
  useEffect(() => {
    if (optimisticMessages.length === 0 || !activeId) return

    const lastOptimistic = optimisticMessages[optimisticMessages.length - 1]
    if (!lastOptimistic) return
    const foundReal = activeMessages.some((m: Message) =>
      m.sender_type === 'customer' && m.content === lastOptimistic.content
    )

    if (foundReal) {
      setOptimisticMessages([])
    }
  }, [activeMessages, optimisticMessages, activeId])

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

  // Initial load of latest 30 messages on selection
  useEffect(() => {
    if (activeId) {
      setActiveMessages([])
      setOptimisticMessages([])
      setMsgPage(1)
      setMsgHasMore(false)
      setMessagesLoading(true)

      const loadInitialMessages = async () => {
        try {
          const res = await api.get<any>(`/chats/conversations/${activeId}?limit=30&page=1`)
          setActiveMessages(res.messages || [])
          setMsgHasMore(res.has_more || false)
        } catch (err) {
          console.error('Failed to load initial messages:', err)
        } finally {
          setMessagesLoading(false)
        }
      }
      loadInitialMessages()
    }
  }, [activeId])

  useEffect(() => {
    setShowInfo(false)
  }, [activeId])

  // Real-time WebSocket sync (optimistic updates in place, with silent background sync)
  useEffect(() => {
    const unsub = subscribe((msg: WSMessage) => {
      if (msg.type === 'typing_indicator' && msg.conversation_id === activeId) {
        const isTyping = msg.data?.is_typing ?? false
        if (isTyping) {
          setPendingAI(prev => {
            const next = new Set(prev)
            next.add(activeId)
            return next
          })
        } else {
          setPendingAI(prev => {
            const next = new Set(prev)
            next.delete(activeId)
            return next
          })
        }
      }
      if (msg.type === 'new_message') {
        if (msg.conversation_id === activeId) {
          const newMsg: Message = {
            id: msg.data?.id || `msg-${Date.now()}`,
            conversation_id: activeId,
            content: msg.content || '',
            sender_type: (msg.sender_type || 'customer') as any,
            role: msg.sender_type || 'customer',
            created_at: msg.timestamp || new Date().toISOString(),
            confidence: msg.data?.confidence,
            source: msg.data?.source,
            metadata: msg.data?.metadata,
          }
          setActiveMessages(prev => {
            if (prev.some(m => m.id === newMsg.id)) return prev
            return [...prev, newMsg]
          })
          
          // Clear pending AI only if the incoming message is from the AI!
          if (newMsg.sender_type === 'ai') {
            setPendingAI(prev => {
              const next = new Set(prev)
              next.delete(activeId)
              return next
            })
          }
        }

        // Optimistically update conversation list in-place instantly!
        if (msg.conversation_id) {
          setAllConversations(prev =>
            prev.map(c => c.id === msg.conversation_id ? {
              ...c,
              last_message: msg.content || c.last_message,
              unread: msg.conversation_id === activeId ? 0 : (msg.sender_type === 'customer' ? c.unread + 1 : c.unread),
              updated_at: msg.timestamp || new Date().toISOString(),
            } : c)
          )
        }

        // Silently pull fresh dashboard summary in backend background to keep synced
        getConversations('/chats/conversations?page=1&limit=20')
      }
      if (msg.type === 'new_conversation') {
        getConversations('/chats/conversations?page=1&limit=20')
      }
    })
    return unsub
  }, [activeId, subscribe, getConversations])

  const activeConv = allConversations.find((c) => c.id === activeId) || null

  const handleLoadMore = useCallback(() => {
    if (loadingMore || !hasMore) return
    const nextPage = page + 1
    setPage(nextPage)
    loadMore(`/chats/conversations?page=${nextPage}&limit=20`)
  }, [page, loadingMore, hasMore, loadMore])

  // Backward message pagination infinite scroll
  const handleLoadMoreMessages = useCallback(async () => {
    if (!activeId || msgLoadingMore || !msgHasMore) return
    setMsgLoadingMore(true)
    const nextPage = msgPage + 1
    try {
      const res = await api.get<any>(`/chats/conversations/${activeId}?limit=30&page=${nextPage}`)
      setActiveMessages(prev => [...(res.messages || []), ...prev])
      setMsgHasMore(res.has_more || false)
      setMsgPage(nextPage)
    } catch (err) {
      console.error('Failed to load older messages:', err)
    } finally {
      setMsgLoadingMore(false)
    }
  }, [activeId, msgPage, msgLoadingMore, msgHasMore])

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

    // Show immediately
    setOptimisticMessages(prev => [...prev, optimisticMsg])
    setSending(true)

    // Optimistically update conversation's last message and move it to top in sidebar
    setAllConversations(prev =>
      prev.map(c => c.id === activeId ? { ...c, last_message: text, updated_at: new Date().toISOString() } : c)
    )

    // Create a streaming AI message placeholder
    const aiTempId = `ai-temp-${Date.now()}`
    const aiPlaceholder: Message = {
      id: aiTempId,
      conversation_id: activeId,
      content: '',
      sender_type: 'ai',
      role: 'ai',
      created_at: new Date().toISOString(),
    }
    setOptimisticMessages(prev => [...prev, aiPlaceholder])

    let streamedContent = ''

    try {
      api.streamPost(
        `/chats/conversations/${activeId}/stream`,
        { content: text },
        (chunk) => {
          streamedContent += chunk
          setOptimisticMessages(prev =>
            prev.map(m => m.id === aiTempId ? { ...m, content: streamedContent } : m)
          )
        },
        (_meta) => {
          // Done — sync with server to get the real message
          setSending(false)
          api.get<any>(`/chats/conversations/${activeId}?limit=30&page=1`).then(syncRes => {
            setActiveMessages(syncRes.messages || [])
            setMsgHasMore(syncRes.has_more || false)
            setMsgPage(1)
          }).catch(() => {})
        },
        (err) => {
          console.error('Stream failed:', err)
          toast('Failed to get AI response', 'error')
          setOptimisticMessages(prev => prev.filter(m => m.id !== aiTempId))
          setSending(false)
        }
      )
    } catch (err) {
      console.error('Send failed:', err)
      toast('Failed to send message', 'error')
      setOptimisticMessages(prev => prev.map(m =>
        m.id === tempId ? { ...m, content: `${m.content} (failed)` } : m
      ))
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

  const handleClearAllClick = () => {
    setShowClearConfirm(true)
  }

  const handleClearAllConfirm = async () => {
    setClearLoading(true)
    try {
      await api.delete('/chats/clear')
      toast('All conversations cleared successfully', 'success')
      setAllConversations([])
      setSearchParams({})
      setShowClearConfirm(false)
    } catch (err: any) {
      console.error('Failed to clear conversations:', err)
      toast(err?.message || 'Failed to clear conversations', 'error')
    } finally {
      setClearLoading(false)
    }
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

  // Deduplicate messages
  const dedupedMessages = [...activeMessages]
  const realContents = new Set(activeMessages.map((m: Message) => m.content))

  const visibleOptimistic = optimisticMessages.filter((om) => !realContents.has(om.content))

  const allMessages = [...dedupedMessages, ...visibleOptimistic]

  // Sort conversations by updated_at so the latest conversation is always on top
  const sortedConversations = [...allConversations].sort((a, b) =>
    new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
  )

  return (
    <div className="h-full w-full flex animate-page-in overflow-hidden relative">
      {/* Conversation List */}
      <div className={cn(
        'w-full lg:w-72 lg:shrink-0 lg:border-r lg:border-default bg-surface h-full flex flex-col overflow-hidden absolute lg:relative inset-0 lg:inset-auto z-10 transition-transform duration-300 lg:transition-none',
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

        {/* Shimmer ONLY on initial absolute blank load to avoid flashing loaders */}
        {convLoading && allConversations.length === 0 ? (
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
            conversations={sortedConversations}
            activeId={activeId || undefined}
            hasMore={hasMore}
            loadingMore={loadingMore}
            onLoadMore={handleLoadMore}
            onClearAll={handleClearAllClick}
          />
        )}
      </div>

      {/* Chat Area */}
      <div className={cn(
        'flex-1 flex flex-col bg-base min-w-0 h-full overflow-hidden absolute lg:relative inset-0 lg:inset-auto z-20 transition-transform duration-300 lg:transition-none bg-surface',
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

              <Avatar src={activeConv.customer_avatar || undefined} name={activeConv.customer_name} size="sm" />
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

            <ChatMessages
              messages={allMessages}
              hasMore={msgHasMore}
              loadingMore={msgLoadingMore}
              isLoading={messagesLoading}
              onLoadMore={handleLoadMoreMessages}
              conversationId={activeId}
            />
            <ChatInput
              onSend={handleSend}
              onTakeover={handleTakeover}
              disabled={!activeId || sending}

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

      <ConfirmModal
        open={showClearConfirm}
        onClose={() => setShowClearConfirm(false)}
        onConfirm={handleClearAllConfirm}
        title="Clear All Conversations?"
        description="Are you absolutely sure you want to delete all conversations and messages? This action is permanent and cannot be undone."
        confirmText="Clear All"
        cancelText="Cancel"
        variant="danger"
        loading={clearLoading}
        requireTypeConfirm={true}
        confirmPhrase="DELETE CHATS"
      />
    </div>
  )
}