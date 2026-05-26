import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useAPI } from '@/hooks/useAPI'
import { UploadZone, CategoryCard, UnknownQuestionItem } from '@/components/training'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import { useModal } from '@/hooks/useModal'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { useToast } from '@/components/ui/Toast'
import { api } from '../../../lib/api'
import { Bot, Send, X, Loader2, Sparkles } from 'lucide-react'

// ─── Test AI Drawer ───────────────────────────────────────────────────────────

interface ChatMsg {
  role: 'user' | 'assistant'
  content: string
}

function TestAIDrawer({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [messages, setMessages] = useState<ChatMsg[]>([
    { role: 'assistant', content: "Hi! I'm your AI. Ask me anything you've trained me on." }
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (open) {
      const timer = setTimeout(() => {
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
      }, 100)
      return () => clearTimeout(timer)
    }
  }, [messages, open])

  const send = async () => {
    if (!input.trim() || loading) return
    const userMsg = input.trim()
    setInput('')
    setMessages(prev => [...prev, { role: 'user', content: userMsg }])
    setLoading(true)
    try {
      const res = await api.post<{ message: { content: string } }>('/chats/direct-chat', {
        message: userMsg,
        channel: 'web',
        customer_name: 'Test Customer',
      })
      setMessages(prev => [...prev, { role: 'assistant', content: res.message?.content || 'No response.' }])
    } catch (err: any) {
      setMessages(prev => [...prev, { role: 'assistant', content: '⚠️ Could not get a response. Check your training data.' }])
    } finally {
      setLoading(false)
    }
  }

  const handleKey = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); send() }
  }

  if (typeof document === 'undefined') return null

  return createPortal(
    <div
      className="fixed inset-0 pointer-events-none overflow-hidden"
      style={{ zIndex: 9999 }}
    >
      {/* Backdrop */}
      <div
        className={`absolute inset-0 bg-overlay backdrop-blur-sm transition-opacity duration-300 pointer-events-auto ${
          open ? 'opacity-100' : 'opacity-0 invisible pointer-events-none'
        }`}
        onClick={onClose}
      />
      {/* Drawer */}
      <div
        className={`absolute right-0 top-0 h-full w-full sm:w-[440px] flex flex-col shadow-2xl transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] pointer-events-auto ${
          open ? 'translate-x-0' : 'translate-x-full'
        }`}
        style={{ background: 'var(--bg-surface)', borderLeft: '1px solid var(--border-default)' }}
      >
        {/* Header */}
        <div
          className="flex items-center justify-between px-4 py-3 lg:px-5 lg:py-4 border-b shrink-0"
          style={{ borderColor: 'var(--border-default)' }}
        >
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center">
              <Sparkles className="w-4 h-4 text-noant-sky" />
            </div>
            <div>
              <p className="text-sm font-semibold text-primary">Test Noant AI</p>
              <p className="text-xs text-tertiary">Using your training data</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-tertiary hover:text-primary transition-colors p-1.5 rounded-lg hover:bg-inset active:scale-95"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3 scrollbar-thin">
          {messages.map((msg, i) => (
            <div
              key={i}
              className={`flex gap-2.5 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              {msg.role === 'assistant' && (
                <div className="w-7 h-7 rounded-full bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center shrink-0 mt-0.5">
                  <Bot className="w-3.5 h-3.5 text-noant-sky" />
                </div>
              )}
              <div
                className={`max-w-[85%] px-3.5 py-2.5 rounded-2xl text-sm leading-relaxed ${
                  msg.role === 'user'
                    ? 'bg-noant-sky text-white rounded-br-sm shadow-sm shadow-sky/10'
                    : 'rounded-bl-sm'
                }`}
                style={msg.role === 'assistant' ? {
                  background: 'var(--bg-inset)',
                  color: 'var(--text-primary)',
                  border: '1px solid var(--border-default)',
                } : undefined}
              >
                {msg.content}
              </div>
            </div>
          ))}
          {loading && (
            <div className="flex gap-2.5 justify-start">
              <div className="w-7 h-7 rounded-full bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center shrink-0">
                <Bot className="w-3.5 h-3.5 text-noant-sky" />
              </div>
              <div
                className="px-3.5 py-3 rounded-2xl rounded-bl-sm"
                style={{ background: 'var(--bg-inset)', border: '1px solid var(--border-default)' }}
              >
                <div className="flex gap-1 items-center">
                  <div className="w-1.5 h-1.5 rounded-full bg-noant-sky animate-bounce" style={{ animationDelay: '0ms' }} />
                  <div className="w-1.5 h-1.5 rounded-full bg-noant-sky animate-bounce" style={{ animationDelay: '150ms' }} />
                  <div className="w-1.5 h-1.5 rounded-full bg-noant-sky animate-bounce" style={{ animationDelay: '300ms' }} />
                </div>
              </div>
            </div>
          )}
          <div ref={bottomRef} />
        </div>

        {/* Input */}
        <div className="p-3 lg:p-4 border-t shrink-0 bg-surface" style={{ borderColor: 'var(--border-default)' }}>
          <div className="flex gap-2">
            <input
              type="text"
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKey}
              placeholder="Ask something your AI should know…"
              className="flex-1 px-3.5 py-2.5 text-sm rounded-xl border outline-none transition-all duration-200 focus:border-noant-sky"
              style={{
                background: 'var(--bg-inset)',
                borderColor: 'var(--border-default)',
                color: 'var(--text-primary)',
              }}
              disabled={loading}
            />
            <button
              onClick={send}
              disabled={!input.trim() || loading}
              className="w-10 h-10 rounded-xl bg-noant-sky hover:bg-noant-sky-deep disabled:opacity-40 disabled:cursor-not-allowed transition-all flex items-center justify-center shrink-0 btn-press"
            >
              {loading ? (
                <Loader2 className="w-4 h-4 text-white animate-spin" />
              ) : (
                <Send className="w-4 h-4 text-white" />
              )}
            </button>
          </div>
          <p className="text-xs text-tertiary mt-2 text-center">Responses use your uploaded Q&amp;A training data</p>
        </div>
      </div>
    </div>,
    document.body
  )
}

// ─── Teach Page ───────────────────────────────────────────────────────────────

export default function TeachPage() {
  const { open: showIgnore, openModal: openIgnore, closeModal: closeIgnore } = useModal()
  const [ignoreId, setIgnoreId] = useState('')
  const [ignoreLoading, setIgnoreLoading] = useState(false)
  const [tab, setTab] = useState<'categories' | 'Upload' | 'unknown'>('categories')
  const [testAIOpen, setTestAIOpen] = useState(false)
  const { toast } = useToast()

  const catAPI = useAPI() as any
  const { data: catData, get: getCategories, post, loading: catLoading } = catAPI

  const uqAPI = useAPI() as any
  const { data: uqData, get: getUnknown, loading: uqLoading } = uqAPI

  const [progress, setProgress] = useState(0)
  const [Uploading, setUploading] = useState(false)

  useEffect(() => {
    getCategories('/training/categories')
    getUnknown('/training/unknown-questions?status=pending')
  }, [getCategories, getUnknown])

  const handleUpload = async (file: File) => {
    setUploading(true)
    setProgress(20)
    const formData = new FormData()
    formData.append('file', file)
    formData.append('category_id', 'default')
    setProgress(60)
    try {
      const res = await post('/training/csv-upload', formData, true) as any
      setProgress(100)
      const count = res?.count ?? 0
      toast(`Training complete: ${count} Q&A pair${count !== 1 ? 's' : ''} learned!`, 'success')
      setTimeout(() => {
        setUploading(false)
        setProgress(0)
        setTab('categories')
        getCategories('/training/categories')
      }, 500)
    } catch (err: any) {
      toast(err?.message || 'Upload failed', 'error')
      setUploading(false)
      setProgress(0)
    }
  }

  const handleTrain = async (id: string, answer: string, categoryId: string) => {
    await post(`/training/unknown-questions/${id}/train`, { answer, category_id: categoryId || 'default' })
    getUnknown('/training/unknown-questions?status=pending')
  }

  const handleIgnoreClick = (id: string) => {
    setIgnoreId(id)
    openIgnore()
  }

  const handleIgnoreConfirm = async () => {
    if (!ignoreId) return
    setIgnoreLoading(true)
    try {
      await post(`/training/unknown-questions/${ignoreId}/ignore`, {})
      getUnknown('/training/unknown-questions?status=pending')
    } catch {
      // error handled by API hook
    } finally {
      setIgnoreLoading(false)
      closeIgnore()
      setIgnoreId('')
    }
  }

  // Safe array defaults
  const categories = catData?.categories || []
  const questions = uqData?.questions || []

  return (
    <div className="animate-fade-in pt-2">
      {/* Header row with Test AI button */}
      <div className="px-1 flex items-center justify-between mb-1 gap-2">
        <div className="flex gap-0 border-b border-default flex-1 overflow-x-auto scrollbar-hide">
          {(['categories', 'Upload', 'unknown'] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-4 lg:px-5 py-3 text-sm font-semibold border-b-2 -mb-px transition-colors capitalize shrink-0 whitespace-nowrap ${
                tab === t ? 'text-noant-sky-deep border-noant-sky' : 'text-tertiary border-transparent hover:text-secondary'
              }`}
            >
              {t === 'unknown' ? `Unknown (${questions.length})` : t}
            </button>
          ))}
        </div>
        <button
          onClick={() => setTestAIOpen(true)}
          className="flex items-center gap-2 px-3.5 py-2 rounded-xl border text-sm font-semibold shrink-0 transition-all hover:bg-noant-sky/10 hover:border-noant-sky/30 hover:text-noant-sky"
          style={{ borderColor: 'var(--border-default)', color: 'var(--text-secondary)' }}
        >
          <Sparkles className="w-3.5 h-3.5" />
          Test AI
        </button>
      </div>

      <div className="px-1 pb-4 mt-5">
        {tab === 'categories' && (
          <>
            <div className="flex justify-between items-center mb-4 lg:mb-5">
              <h2 className="text-base lg:text-lg font-semibold text-primary">What your Noant knows</h2>
              <Button size="sm" onClick={() => getCategories('/training/categories')} loading={catLoading}>
                Refresh
              </Button>
            </div>
            {catLoading ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 lg:gap-4">
                {Array.from({ length: 6 }).map((_, i) => (
                  <div key={i} className="rounded-xl border border-default bg-surface p-4 lg:p-5 space-y-3">
                    <div className="flex items-center gap-3">
                      <Skeleton className="w-3 h-3 rounded-full" />
                      <Skeleton className="h-5 w-32 rounded" />
                    </div>
                    <Skeleton className="h-3 w-24 rounded" />
                  </div>
                ))}
              </div>
            ) : categories.length === 0 ? (
              <EmptyCategories />
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 lg:gap-4">
                {categories.map((c: any) => (
                  <CategoryCard key={c.id} category={c} />
                ))}
              </div>
            )}
          </>
        )}

        {tab === 'Upload' && (
          <Card>
            <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
              <CardTitle>Share your knowledge</CardTitle>
            </CardHeader>
            <CardBody className="p-3 lg:p-4">
              <p className="text-sm text-secondary mb-4 lg:mb-6">
                Upload a spreadsheet with your FAQs, product details, or any information your customers ask about.
                Your Noant learns instantly.
              </p>
              <UploadZone onUpload={handleUpload} uploading={Uploading} progress={progress} />
            </CardBody>
          </Card>
        )}

        {tab === 'unknown' && (
          <Card>
            <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
              <CardTitle>Questions your Noant could not answer</CardTitle>
            </CardHeader>
            <CardBody className="p-3 lg:p-4">
              {uqLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <div key={i} className="flex items-center justify-between p-3 lg:p-4 border border-default rounded-lg">
                      <Skeleton className="h-4 w-2/3 rounded" />
                      <div className="flex gap-2">
                        <Skeleton className="h-8 w-16 rounded" />
                        <Skeleton className="h-8 w-16 rounded" />
                      </div>
                    </div>
                  ))}
                </div>
              ) : questions.length === 0 ? (
                <div className="text-center py-12">
                  <p className="text-secondary">No unknown questions. Your Noant is handling everything!</p>
                </div>
              ) : (
                questions.map((q: any) => (
                  <UnknownQuestionItem key={q.id} question={q} onTrain={handleTrain} onIgnore={handleIgnoreClick} />
                ))
              )}
            </CardBody>
          </Card>
        )}
      </div>

      <ConfirmModal
        open={showIgnore}
        onClose={closeIgnore}
        onConfirm={handleIgnoreConfirm}
        title="Ignore this question?"
        description="It will be marked as ignored and won't appear in training suggestions."
        variant="neutral"
        confirmText="Ignore"
        loading={ignoreLoading}
      />

      {/* Test AI Drawer */}
      <TestAIDrawer open={testAIOpen} onClose={() => setTestAIOpen(false)} />
    </div>
  )
}

function EmptyCategories() {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center px-4">
      <div className="w-14 h-14 lg:w-16 lg:h-16 bg-inset rounded-2xl flex items-center justify-center mb-4">
        <svg className="w-7 h-7 lg:w-8 lg:h-8 text-tertiary" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
          <path d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18a8.967 8.967 0 00-6 2.292m0-14.25v14.25" />
        </svg>
      </div>
      <p className="text-base lg:text-lg font-semibold text-primary mb-1">No categories yet</p>
      <p className="text-sm text-secondary max-w-xs mb-6">
        Upload your first CSV to teach your Noant about your business.
      </p>
    </div>
  )
}
