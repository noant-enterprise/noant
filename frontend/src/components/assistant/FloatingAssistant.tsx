import { useState, useRef, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { api } from '@/lib/api'
import { Bot, X, Send, ChevronRight, Loader2, MessageSquare, Sparkles, ThumbsUp, ThumbsDown } from 'lucide-react'
import { createPortal } from 'react-dom'

interface AssistantAction {
  type: string
  path?: string
  label?: string
}

interface AssistantStep {
  title: string
  description: string
  path?: string
  action?: string
}

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  action?: AssistantAction
  steps?: AssistantStep[]
  suggestions?: string[]
}

function renderMarkdown(text: string): string {
  let html = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  html = html
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`(.+?)`/g, '<code class="bg-surface-hover px-1 rounded text-xs">$1</code>')
    .replace(/^### (.+)$/gm, '<strong class="block text-sm mt-2 mb-1">$1</strong>')
    .replace(/^## (.+)$/gm, '<strong class="block text-base mt-3 mb-1">$1</strong>')
    .replace(/^# (.+)$/gm, '<strong class="block text-lg mt-3 mb-2">$1</strong>')
    .replace(/^- (.+)$/gm, '<span class="block pl-3 before:content-["•"] before:mr-2 before:text-noant-sky">$1</span>')
    .replace(/\n/g, '<br/>')
  return html
}

function TypewriterText({ content, onDone }: { content: string; onDone?: () => void }) {
  const [displayed, setDisplayed] = useState('')
  const [done, setDone] = useState(false)
  const idxRef = useRef(0)
  const words = content.split(' ')

  useEffect(() => {
    idxRef.current = 0
    setDisplayed('')
    setDone(false)

    if (words.length <= 3) {
      setDisplayed(content)
      setDone(true)
      onDone?.()
      return
    }

    const interval = setInterval(() => {
      idxRef.current++
      if (idxRef.current >= words.length) {
        setDisplayed(words.join(' '))
        clearInterval(interval)
        setDone(true)
        onDone?.()
        return
      }
      setDisplayed(words.slice(0, idxRef.current).join(' ') + ' ▊')
    }, Math.max(15, 80 - words.length))

    return () => clearInterval(interval)
  }, [content])

  if (done) {
    return <span dangerouslySetInnerHTML={{ __html: renderMarkdown(content) }} />
  }
  return <span>{displayed}</span>
}

export function FloatingAssistant() {
  const [open, setOpen] = useState(false)
  const [messages, setMessages] = useState<ChatMessage[]>([
    { role: 'assistant', content: 'Hi! I\'m your Noant guide. Ask me anything about using the platform.' }
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [streamingId, setStreamingId] = useState<number | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const navigate = useNavigate()

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 300)
    }
  }, [open])

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'g') {
        e.preventDefault()
        setOpen(o => !o)
      }
      if (e.key === 'Escape' && open) {
        setOpen(false)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [open])

  const handleSend = useCallback(async (suggestionText?: string) => {
    const text = suggestionText || input.trim()
    if (!text || loading) return
    if (!suggestionText) setInput('')

    setMessages(prev => [...prev, { role: 'user', content: text }])
    setLoading(true)

    try {
      const response = await api.post<{ content: string; action?: AssistantAction; steps?: AssistantStep[]; suggestions?: string[] }>(
        '/assistant/chat',
        { message: text }
      )

      const msgId = Date.now()
      setStreamingId(msgId)
      const msg: ChatMessage = {
        role: 'assistant',
        content: response.content || 'Got it! Let me help you with that.',
        action: response.action,
        steps: response.steps,
        suggestions: response.suggestions,
      }
      setMessages(prev => [...prev, msg])
    } catch {
      setMessages(prev => [...prev, {
        role: 'assistant',
        content: 'Sorry, I had trouble connecting. Please try again.'
      }])
    } finally {
      setLoading(false)
      setStreamingId(null)
    }
  }, [input, loading])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleNavigate = (path: string) => {
    navigate(path)
    setOpen(false)
  }

  return createPortal(
    <>
      {/* FAB */}
      <button
        onClick={() => setOpen(o => !o)}
        className={cn(
          'fixed bottom-6 right-6 z-[200] w-14 h-14 rounded-full flex items-center justify-center shadow-xl transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)]',
          'bg-gradient-to-br from-noant-sky to-blue-600 hover:from-blue-500 hover:to-blue-700',
          'text-white hover:scale-105 active:scale-95',
          open && 'scale-0 opacity-0 pointer-events-none'
        )}
      >
        <Bot className="w-6 h-6" />
      </button>

      {/* Panel */}
      <div
        className={cn(
          'fixed bottom-4 right-4 sm:bottom-6 sm:right-6 z-[200]',
          'w-[calc(100vw-2rem)] sm:w-[400px]',
          'h-[calc(100vh-8rem)] sm:h-[600px]',
          'max-h-[calc(100vh-4rem)] sm:max-h-[calc(100vh-8rem)]',
          'rounded-2xl border border-default bg-surface shadow-2xl flex flex-col overflow-hidden',
          'transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)]',
          open ? 'scale-100 opacity-100 translate-y-0' : 'scale-75 opacity-0 translate-y-8 pointer-events-none'
        )}
      >
        {/* Header */}
        <div className="shrink-0 flex items-center justify-between px-5 py-4 border-b border-default bg-gradient-to-r from-noant-sky/5 to-blue-500/5">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-gradient-to-br from-noant-sky to-blue-600 flex items-center justify-center">
              <Sparkles className="w-4 h-4 text-white" />
            </div>
            <div>
              <p className="text-sm font-semibold text-primary">Noant Guide</p>
              <p className="text-[11px] text-tertiary">AI onboarding assistant</p>
            </div>
          </div>
          <button
            onClick={() => setOpen(false)}
            className="w-8 h-8 rounded-lg hover:bg-surface-hover flex items-center justify-center transition-colors text-tertiary hover:text-primary"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4 scroll-smooth">
          {messages.map((msg, i) => (
            <div key={i} className={cn('flex', msg.role === 'user' ? 'justify-end' : 'justify-start')}>
              <div className={cn(
                'max-w-[85%] space-y-1.5',
                msg.role === 'user' && 'order-1'
              )}>
                {msg.role === 'assistant' && (
                  <div className="flex items-center gap-1.5 mb-1">
                    <Bot className="w-3.5 h-3.5 text-noant-sky" />
                    <span className="text-[11px] font-medium text-tertiary">Guide</span>
                  </div>
                )}
                <div className={cn(
                  'rounded-2xl px-4 py-2.5 text-sm leading-relaxed',
                  msg.role === 'user'
                    ? 'bg-noant-sky text-white rounded-br-md'
                    : 'bg-inset text-primary border border-default rounded-bl-md'
                )}>
                  {msg.role === 'assistant' && streamingId === i ? (
                    <TypewriterText content={msg.content} />
                  ) : msg.role === 'assistant' ? (
                    <span dangerouslySetInnerHTML={{ __html: renderMarkdown(msg.content) }} />
                  ) : (
                    msg.content
                  )}
                </div>

                {/* Quick action chips */}
                {msg.role === 'assistant' && msg.suggestions && msg.suggestions.length > 0 && (
                  <div className="flex flex-wrap gap-1.5 pt-1">
                    {msg.suggestions.map((suggestion, si) => (
                      <button
                        key={si}
                        onClick={() => handleSend(suggestion)}
                        disabled={loading}
                        className="px-2.5 py-1 rounded-full bg-noant-sky/10 border border-noant-sky/20 text-noant-sky text-[11px] font-medium hover:bg-noant-sky/20 transition-colors disabled:opacity-50"
                      >
                        {suggestion}
                      </button>
                    ))}
                  </div>
                )}

                {msg.steps && msg.steps.length > 0 && (
                  <div className="space-y-1.5 pt-1">
                    {msg.steps.map((step, si) => (
                      <div
                        key={si}
                        className="flex items-start gap-2 px-3 py-2 rounded-xl bg-inset border border-default cursor-pointer hover:bg-surface-hover transition-colors"
                        onClick={() => step.path && handleNavigate(step.path)}
                      >
                        <div className="w-5 h-5 rounded-full bg-noant-sky/10 text-noant-sky flex items-center justify-center shrink-0 mt-0.5">
                          <span className="text-[10px] font-bold">{si + 1}</span>
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-xs font-semibold text-primary">{step.title}</p>
                          <p className="text-[11px] text-tertiary">{step.description}</p>
                        </div>
                        {step.path && (
                          <ChevronRight className="w-3.5 h-3.5 text-tertiary shrink-0 mt-0.5" />
                        )}
                      </div>
                    ))}
                  </div>
                )}
                {(() => {
                  const act = msg.action
                  if (!act?.path) return null
                  return (
                    <button
                      onClick={() => handleNavigate(act.path!)}
                      className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-noant-sky/10 text-noant-sky text-xs font-medium hover:bg-noant-sky/20 transition-colors"
                    >
                      <MessageSquare className="w-3 h-3" />
                      {act.label || `Go to ${act.path}`}
                    </button>
                  )
                })()}

                {/* CSAT thumbs */}
                {msg.role === 'assistant' && i === messages.length - 1 && msg.content && !loading && (
                  <div className="flex items-center gap-2 pt-1.5 opacity-60 hover:opacity-100 transition-opacity">
                    <button className="p-1 rounded hover:bg-surface-hover text-tertiary hover:text-green-500 transition-colors">
                      <ThumbsUp className="w-3 h-3" />
                    </button>
                    <button className="p-1 rounded hover:bg-surface-hover text-tertiary hover:text-red-500 transition-colors">
                      <ThumbsDown className="w-3 h-3" />
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}
          {loading && !streamingId && (
            <div className="flex justify-start">
              <div className="max-w-[85%]">
                <div className="flex items-center gap-1.5 mb-1">
                  <Bot className="w-3.5 h-3.5 text-noant-sky" />
                  <span className="text-[11px] font-medium text-tertiary">Guide</span>
                </div>
                <div className="rounded-2xl rounded-bl-md px-4 py-3 bg-inset border border-default">
                  <Loader2 className="w-4 h-4 animate-spin text-tertiary" />
                </div>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <div className="shrink-0 border-t border-default p-3">
          <div className="flex items-center gap-2 bg-inset rounded-xl border border-default px-3 py-1.5 focus-within:border-noant-sky/50 focus-within:ring-1 focus-within:ring-noant-sky/20 transition-all">
            <input
              ref={inputRef}
              type="text"
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Ask me anything..."
              className="flex-1 bg-transparent text-sm text-primary placeholder:text-tertiary outline-none min-w-0 py-1"
            />
            <button
              onClick={() => handleSend()}
              disabled={!input.trim() || loading}
              className={cn(
                'w-8 h-8 rounded-lg flex items-center justify-center shrink-0 transition-all',
                input.trim() && !loading
                  ? 'bg-noant-sky text-white hover:bg-blue-600'
                  : 'bg-surface-hover text-tertiary cursor-not-allowed'
              )}
            >
              <Send className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </>,
    document.body
  )
}
