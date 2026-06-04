import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useAPI } from '@/hooks/useAPI'
import { UploadZone, UnknownQuestionItem } from '@/components/training'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import { useModal } from '@/hooks/useModal'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { useToast } from '@/components/ui/Toast'
import { api } from '../../../lib/api'
import { useOffline } from '@/hooks/useOffline'
import {
  Bot, Send, X, Loader2, Sparkles, Search, Plus, Trash2, Edit3,
  FolderPlus, FileText, ChevronRight, BookOpen, GraduationCap, Upload
} from 'lucide-react'

// ─── Shared Modal Shell (Portal) ─────────────────────────────────────────────

function ModalPortal({
  open,
  onClose,
  children,
  maxWidth = 'max-w-lg',
  zIndex = 11000,
}: {
  open: boolean
  onClose: () => void
  children: React.ReactNode
  maxWidth?: string
  zIndex?: number
}) {
  if (!open || typeof document === 'undefined') return null
  return createPortal(
    <div
      className="fixed inset-0 flex items-center justify-center overflow-y-auto p-4 sm:p-6"
      style={{ zIndex }}
    >
      <div
        className="absolute inset-0 bg-overlay backdrop-blur-sm"
        onClick={onClose}
      />
      <div
        className={`relative w-full ${maxWidth} bg-surface border border-default rounded-2xl shadow-2xl overflow-hidden animate-slide-up flex flex-col max-h-[calc(100dvh-2rem)] sm:max-h-[calc(100dvh-3rem)]`}
      >
        {children}
      </div>
    </div>,
    document.body
  )
}

// ─── Shared Modal Header ──────────────────────────────────────────────────────

function ModalHeader({
  icon,
  title,
  subtitle,
  onClose,
}: {
  icon: React.ReactNode
  title: string
  subtitle?: string
  onClose: () => void
}) {
  return (
    <div className="flex items-center justify-between px-6 py-4 border-b border-default shrink-0">
      <div className="flex items-center gap-3">
        <div className="w-9 h-9 rounded-xl bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center shrink-0">
          {icon}
        </div>
        <div>
          <h3 className="text-sm font-bold text-primary">{title}</h3>
          {subtitle && <p className="text-xs text-tertiary mt-0.5">{subtitle}</p>}
        </div>
      </div>
      <button
        onClick={onClose}
        className="p-1.5 rounded-lg text-tertiary hover:text-primary hover:bg-inset transition-all active:scale-95"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  )
}

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
    } catch {
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
          className="flex items-center justify-between px-5 py-4 border-b shrink-0"
          style={{ borderColor: 'var(--border-default)' }}
        >
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center">
              <Sparkles className="w-4 h-4 text-noant-sky" />
            </div>
            <div>
              <p className="text-sm font-bold text-primary">Test Noant AI</p>
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
        <div className="p-4 border-t shrink-0" style={{ borderColor: 'var(--border-default)', background: 'var(--bg-surface)' }}>
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

// ─── CSV Upload Modal ─────────────────────────────────────────────────────────

function CSVUploadModal({
  open,
  onClose,
  onUpload,
  uploading,
  progress,
}: {
  open: boolean
  onClose: () => void
  onUpload: (file: File) => void
  uploading: boolean
  progress: number
}) {
  return (
    <ModalPortal open={open} onClose={onClose} maxWidth="max-w-xl" zIndex={11500}>
      <ModalHeader
        icon={<FileText className="w-4 h-4 text-noant-sky" />}
        title="Bulk Upload Knowledge Base"
        subtitle="Upload a CSV spreadsheet to train your AI instantly"
        onClose={onClose}
      />
      <div className="p-6 overflow-y-auto">
        <div className="mb-5 p-4 rounded-xl bg-inset border border-default space-y-2">
          <p className="text-xs font-semibold text-secondary">CSV Format Requirements</p>
          <div className="grid grid-cols-3 gap-2 text-center">
            {['Category', 'Question', 'Answer'].map((col, i) => (
              <div key={col} className="bg-surface rounded-lg py-2 px-3 border border-default">
                <p className="text-[10px] font-bold text-tertiary uppercase tracking-wider mb-0.5">Column {i + 1}</p>
                <p className="text-xs font-bold text-primary">{col}</p>
              </div>
            ))}
          </div>
          <p className="text-[11px] text-tertiary leading-relaxed">
            Each row is one Q&amp;A pair. The category column helps organize knowledge. Headers should match exactly.
          </p>
        </div>

        <UploadZone onUpload={onUpload} uploading={uploading} progress={progress} />
      </div>
      <div className="px-6 py-4 border-t border-default bg-inset/30 flex justify-end">
        <Button variant="ghost" size="sm" onClick={onClose} disabled={uploading}>
          {uploading ? 'Uploading…' : 'Close'}
        </Button>
      </div>
    </ModalPortal>
  )
}

// ─── Add Category Modal ───────────────────────────────────────────────────────

function AddCategoryModal({
  open,
  onClose,
  onSubmit,
  loading,
}: {
  open: boolean
  onClose: () => void
  onSubmit: (name: string, desc: string, color: string) => Promise<void>
  loading: boolean
}) {
  const [name, setName] = useState('')
  const [desc, setDesc] = useState('')
  const [color, setColor] = useState('#0ea5e9')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    await onSubmit(name.trim(), desc.trim(), color)
    setName('')
    setDesc('')
    setColor('#0ea5e9')
  }

  return (
    <ModalPortal open={open} onClose={onClose} maxWidth="max-w-md" zIndex={11000}>
      <ModalHeader
        icon={<FolderPlus className="w-4 h-4 text-noant-sky" />}
        title="New Category Folder"
        subtitle="Organize your AI's knowledge into folders"
        onClose={onClose}
      />
      <form onSubmit={handleSubmit} className="p-6 space-y-4 overflow-y-auto">
        <div>
          <label className="block text-xs font-semibold text-secondary mb-1.5">Category Name <span className="text-red-400">*</span></label>
          <input
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="e.g. Refunds, Shipping, Technical FAQ…"
            className="w-full text-sm px-4 py-2.5 rounded-xl border border-default bg-inset outline-none focus:border-noant-sky focus:ring-2 focus:ring-noant-sky/15 transition-all text-primary"
            required
            disabled={loading}
            autoFocus
          />
        </div>
        <div>
          <label className="block text-xs font-semibold text-secondary mb-1.5">Description <span className="text-tertiary font-normal">(optional)</span></label>
          <textarea
            value={desc}
            onChange={e => setDesc(e.target.value)}
            placeholder="Brief summary of what this category handles…"
            className="w-full text-sm px-4 py-2 rounded-xl border border-default bg-inset outline-none focus:border-noant-sky focus:ring-2 focus:ring-noant-sky/15 transition-all text-primary resize-none"
            rows={2}
            disabled={loading}
          />
        </div>
        <div>
          <label className="block text-xs font-semibold text-secondary mb-1.5">Color Tag</label>
          <div className="flex items-center gap-3">
            <input
              type="color"
              value={color}
              onChange={e => setColor(e.target.value)}
              className="w-12 h-10 rounded-xl border border-default bg-inset cursor-pointer outline-none shrink-0 p-1"
              disabled={loading}
            />
            <div className="flex-1 p-2.5 rounded-xl border border-default bg-inset">
              <p className="text-xs font-mono text-secondary uppercase">{color}</p>
            </div>
            <div
              className="w-10 h-10 rounded-xl border border-default shrink-0"
              style={{ backgroundColor: color }}
            />
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" size="sm" onClick={onClose} disabled={loading}>Cancel</Button>
          <Button type="submit" size="sm" loading={loading} disabled={!name.trim() || loading}>
            Create Folder
          </Button>
        </div>
      </form>
    </ModalPortal>
  )
}

// ─── Add / Edit Q&A Modal ─────────────────────────────────────────────────────

function QAModal({
  open,
  onClose,
  onSubmit,
  loading,
  categories,
  initialCatId,
  mode,
  initialQuestion,
  initialAnswer,
  initialCatID,
}: {
  open: boolean
  onClose: () => void
  onSubmit: (question: string, answer: string, catId: string) => Promise<void>
  loading: boolean
  categories: any[]
  initialCatId?: string
  mode: 'add' | 'edit'
  initialQuestion?: string
  initialAnswer?: string
  initialCatID?: string
}) {
  const [question, setQuestion] = useState(initialQuestion || '')
  const [answer, setAnswer] = useState(initialAnswer || '')
  const [catId, setCatId] = useState(initialCatID || initialCatId || 'default')

  useEffect(() => {
    if (open) {
      setQuestion(initialQuestion || '')
      setAnswer(initialAnswer || '')
      setCatId(initialCatID || initialCatId || 'default')
    }
  }, [open, initialQuestion, initialAnswer, initialCatID, initialCatId])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!question.trim() || !answer.trim()) return
    await onSubmit(question.trim(), answer.trim(), catId)
  }

  const isEdit = mode === 'edit'

  return (
    <ModalPortal open={open} onClose={onClose} maxWidth="max-w-lg" zIndex={isEdit ? 12000 : 11000}>
      <ModalHeader
        icon={isEdit ? <Edit3 className="w-4 h-4 text-noant-sky" /> : <Sparkles className="w-4 h-4 text-noant-sky" />}
        title={isEdit ? 'Edit Trained Q&A' : 'Add Trained Q&A Pair'}
        subtitle={isEdit ? 'Update this knowledge entry' : 'Teach your AI a new question and answer'}
        onClose={onClose}
      />
      <form onSubmit={handleSubmit} className="p-6 space-y-4 overflow-y-auto">
        <div>
          <label className="block text-xs font-semibold text-secondary mb-1.5">Category Folder</label>
          <select
            value={catId}
            onChange={e => setCatId(e.target.value)}
            className="w-full text-sm px-4 py-2.5 rounded-xl border border-default bg-inset outline-none focus:border-noant-sky focus:ring-2 focus:ring-noant-sky/15 transition-all text-primary cursor-pointer"
            disabled={loading}
          >
            <option value="default">Default</option>
            {categories.map((c: any) =>
              c.id !== 'default' && <option key={c.id} value={c.id}>{c.name}</option>
            )}
          </select>
        </div>

        <div>
          <label className="block text-xs font-semibold text-secondary mb-1.5">
            Trained Question <span className="text-red-400">*</span>
          </label>
          <input
            type="text"
            value={question}
            onChange={e => setQuestion(e.target.value)}
            placeholder="What question does a customer ask? (e.g. Do you ship to Canada?)"
            className="w-full text-sm px-4 py-2.5 rounded-xl border border-default bg-inset outline-none focus:border-noant-sky focus:ring-2 focus:ring-noant-sky/15 transition-all text-primary"
            required
            disabled={loading}
            autoFocus={!isEdit}
          />
        </div>

        <div>
          <label className="block text-xs font-semibold text-secondary mb-1.5">
            AI Trained Answer <span className="text-red-400">*</span>
          </label>
          <textarea
            value={answer}
            onChange={e => setAnswer(e.target.value)}
            placeholder="What answer should Noant AI automatically respond with?"
            className="w-full text-sm px-4 py-2.5 rounded-xl border border-default bg-inset outline-none focus:border-noant-sky focus:ring-2 focus:ring-noant-sky/15 transition-all text-primary resize-none"
            rows={4}
            required
            disabled={loading}
          />
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="ghost" size="sm" onClick={onClose} disabled={loading}>Cancel</Button>
          <Button type="submit" size="sm" loading={loading} disabled={!question.trim() || !answer.trim() || loading}>
            {isEdit ? 'Save Changes' : 'Train AI'}
          </Button>
        </div>
      </form>
    </ModalPortal>
  )
}

// ─── Q&A Table Sheet (Category Detail Panel) ─────────────────────────────────

function QATableSheet({
  category,
  qaList,
  loading,
  onClose,
  onAddQA,
  onEditQA,
  onDeleteQA,
  onDeleteCategory,
}: {
  category: any | null
  qaList: any[]
  loading: boolean
  onClose: () => void
  onAddQA: () => void
  onEditQA: (qa: any) => void
  onDeleteQA: (id: string) => void
  onDeleteCategory: () => void
}) {
  const [filter, setFilter] = useState('')

  useEffect(() => {
    if (!category) setFilter('')
  }, [category])

  if (!category || typeof document === 'undefined') return null

  const filtered = qaList.filter(qa =>
    qa.question.toLowerCase().includes(filter.toLowerCase()) ||
    qa.answer.toLowerCase().includes(filter.toLowerCase())
  )

  return createPortal(
    <div
      className="fixed inset-0 flex flex-col bg-surface overflow-hidden"
      style={{
        zIndex: 10500,
        animation: 'noantPageIn 300ms cubic-bezier(0.16, 1, 0.3, 1) forwards'
      }}
    >
      <style>{`
        @keyframes noantPageIn {
          from { opacity: 0; transform: translateY(16px); }
          to { opacity: 1; transform: translateY(0); }
        }
      `}</style>

      {/* Panel Header */}
      <div
        className="border-b border-default bg-surface shrink-0 z-10"
        style={{ borderTop: `4px solid ${category.color}` }}
      >
        <div className="max-w-7xl mx-auto px-6 py-5 flex items-center justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-3">
              <span
                className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold border"
                style={{
                  backgroundColor: `${category.color}0D`,
                  borderColor: `${category.color}33`,
                  color: category.color,
                }}
              >
                <span className="w-2 h-2 rounded-full" style={{ backgroundColor: category.color }} />
                {category.name}
              </span>
              <span className="text-xs font-bold bg-inset px-2.5 py-1 rounded-full text-secondary border border-default">
                {qaList.length} Q&amp;A Pairs
              </span>
            </div>
            <p className="text-xs text-tertiary leading-relaxed">
              {category.description || 'Manage Q&A records inside this folder.'}
            </p>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            {category.id !== 'default' && (
              <button
                onClick={onDeleteCategory}
                className="p-2.5 rounded-xl border border-red-500/20 text-red-500 hover:bg-red-500/10 active:scale-95 transition-all cursor-pointer flex items-center gap-2 text-xs font-semibold"
                title="Delete Category Folder"
              >
                <Trash2 className="w-4 h-4" />
                Delete Folder
              </button>
            )}
            <button
              onClick={onClose}
              className="p-2.5 rounded-xl hover:bg-inset text-tertiary hover:text-primary active:scale-95 transition-colors cursor-pointer border border-default bg-surface shadow-sm"
              title="Close Panel"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Panel Controls */}
      <div className="border-b border-default bg-inset/50 shrink-0 z-10">
        <div className="max-w-7xl mx-auto px-6 py-4 flex flex-col sm:flex-row sm:items-center gap-4">
          <div className="relative flex-1 flex items-center bg-surface border border-default rounded-xl px-4 focus-within:ring-2 focus-within:ring-noant-sky/20 focus-within:border-noant-sky transition-all shadow-sm">
            <Search className="w-4 h-4 text-tertiary shrink-0" />
            <input
              type="text"
              value={filter}
              onChange={e => setFilter(e.target.value)}
              placeholder="Search questions or answers in this folder…"
              className="w-full pl-3 py-3 text-sm bg-transparent outline-none text-primary placeholder-tertiary"
            />
            {filter && (
              <button onClick={() => setFilter('')} className="text-tertiary hover:text-primary shrink-0 cursor-pointer">
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
          <button
            onClick={onAddQA}
            className="flex items-center justify-center gap-2 px-5 py-3 bg-noant-sky text-white rounded-xl text-sm font-semibold hover:bg-noant-sky-deep active:scale-95 transition-all shrink-0 shadow-sm shadow-sky/10 cursor-pointer"
          >
            <Plus className="w-4 h-4 text-white" />
            Add Q&amp;A to Folder
          </button>
        </div>
      </div>

      {/* Table Body */}
      <div className="flex-1 overflow-auto bg-surface scrollbar-thin">
        <div className="max-w-7xl mx-auto px-6 py-6">
          {loading ? (
            <div className="space-y-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex gap-4 p-4 border border-default rounded-2xl bg-surface shadow-sm animate-pulse">
                  <Skeleton className="h-5 w-1/3 rounded" />
                  <Skeleton className="h-5 w-2/3 rounded" />
                </div>
              ))}
            </div>
          ) : filtered.length === 0 ? (
            <div className="text-center py-24 border border-dashed border-default rounded-2xl bg-inset/10">
              <div className="w-14 h-14 rounded-2xl bg-inset flex items-center justify-center text-tertiary mx-auto mb-4 border border-default">
                <BookOpen className="w-7 h-7 text-noant-sky" />
              </div>
              <h4 className="text-base font-bold text-primary">
                {filter ? 'No matching pairs found' : 'No trained Q&As yet'}
              </h4>
              <p className="text-sm text-secondary mt-1 max-w-sm mx-auto leading-relaxed">
                {filter ? 'Try a different search term.' : 'Add a Q&A pair to get started.'}
              </p>
            </div>
          ) : (
            <div className="border border-default rounded-2xl overflow-hidden shadow-sm bg-surface">
              <table className="w-full text-left border-collapse table-fixed">
                <thead>
                  <tr className="bg-inset border-b border-default text-[11px] font-bold text-secondary uppercase tracking-wider select-none">
                    <th className="px-6 py-4 w-[40%]">Question</th>
                    <th className="px-6 py-4 w-[45%]">Answer</th>
                    <th className="px-6 py-4 w-[15%] text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-default">
                  {filtered.map((qa: any) => (
                    <tr key={qa.id} className="hover:bg-inset/30 transition-colors group">
                      <td className="px-6 py-4 text-sm font-semibold text-primary leading-relaxed align-top break-words">
                        {qa.question}
                      </td>
                      <td className="px-6 py-4 text-sm text-secondary leading-relaxed align-top break-words">
                        {qa.answer}
                      </td>
                      <td className="px-6 py-4 align-top text-right">
                        <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                          <button
                            onClick={() => onEditQA(qa)}
                            className="p-2 rounded-xl text-tertiary hover:text-noant-sky hover:bg-inset active:scale-95 transition-all cursor-pointer border border-transparent hover:border-default bg-surface shadow-sm"
                            title="Edit Q&A"
                          >
                            <Edit3 className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => onDeleteQA(qa.id)}
                            className="p-2 rounded-xl text-tertiary hover:text-red-500 hover:bg-inset active:scale-95 transition-all cursor-pointer border border-transparent hover:border-default bg-surface shadow-sm"
                            title="Delete Q&A"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Panel Footer */}
      <div className="border-t border-default bg-inset/30 shrink-0">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <span className="text-xs text-tertiary">
            Showing <strong className="text-secondary">{filtered.length}</strong> of <strong className="text-secondary">{qaList.length}</strong> pairs
          </span>
          <button
            onClick={onClose}
            className="px-5 py-2.5 border border-default bg-surface hover:bg-inset rounded-xl text-xs font-semibold active:scale-95 transition-all cursor-pointer shadow-sm text-primary"
          >
            Close Page
          </button>
        </div>
      </div>
    </div>,
    document.body
  )
}

// ─── Teach Page ───────────────────────────────────────────────────────────────

export default function TeachPage() {
  const isOffline = useOffline()
  const { open: showIgnore, openModal: openIgnore, closeModal: closeIgnore } = useModal()
  const [ignoreId, setIgnoreId] = useState('')
  const [ignoreLoading, setIgnoreLoading] = useState(false)
  const [testAIOpen, setTestAIOpen] = useState(false)
  const { toast } = useToast()

  const [clearGapsConfirmOpen, setClearGapsConfirmOpen] = useState(false)
  const [clearGapsLoading, setClearGapsLoading] = useState(false)

  // API hooks
  const catAPI = useAPI() as any
  const { data: catData, get: getCategories, loading: catLoading } = catAPI
  const uqAPI = useAPI() as any
  const { data: uqData, get: getUnknown, loading: uqLoading } = uqAPI

  // Upload state
  const [progress, setProgress] = useState(0)
  const [Uploading, setUploading] = useState(false)
  const [csvModalOpen, setCSVModalOpen] = useState(false)

  // Search
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<any[]>([])
  const [searchLoading, setSearchLoading] = useState(false)

  // Category detail panel
  const [selectedCategory, setSelectedCategory] = useState<any | null>(null)
  const [categoryQAs, setCategoryQAs] = useState<any[]>([])
  const [categoryQALoading, setCategoryQALoading] = useState(false)

  // Modals
  const [addCatOpen, setAddCatOpen] = useState(false)
  const [addCatLoading, setAddCatLoading] = useState(false)

  const [addQAOpen, setAddQAOpen] = useState(false)
  const [addQALoading, setAddQALoading] = useState(false)
  const [addQAInitialCatId, setAddQAInitialCatId] = useState('default')

  const [editQAOpen, setEditQAOpen] = useState(false)
  const [editQAPair, setEditQAPair] = useState<any | null>(null)
  const [editQALoading, setEditQALoading] = useState(false)

  const [confirmDeleteCatOpen, setConfirmDeleteCatOpen] = useState(false)
  const [deleteCatLoading, setDeleteCatLoading] = useState(false)

  // Initial data
  useEffect(() => {
    getCategories('/training/categories')
    getUnknown('/training/unknown-questions?status=pending')
  }, [getCategories, getUnknown])

  // Debounced search
  useEffect(() => {
    if (!searchQuery.trim()) { setSearchResults([]); return }
    const t = setTimeout(async () => {
      setSearchLoading(true)
      try {
        const res = await api.get<{ qa_pairs: any[] }>(`/training/search?q=${encodeURIComponent(searchQuery)}`)
        setSearchResults(res.qa_pairs || [])
      } catch { /* no-op */ } finally { setSearchLoading(false) }
    }, 300)
    return () => clearTimeout(t)
  }, [searchQuery])

  // Load category QAs
  const loadCategoryQAs = async (catId: string) => {
    setCategoryQALoading(true)
    try {
      const res = await api.get<{ qa_pairs: any[] }>(`/training/categories/${catId}/qa`)
      setCategoryQAs(res.qa_pairs || [])
    } catch { toast('Failed to load category questions', 'error') }
    finally { setCategoryQALoading(false) }
  }

  // CSV Upload
  const handleUpload = async (file: File) => {
    setUploading(true)
    setProgress(20)
    const formData = new FormData()
    formData.append('file', file)
    formData.append('category_id', selectedCategory ? selectedCategory.id : 'default')
    setProgress(60)
    try {
      const res = await api.post<{ count: number }>('/training/csv-upload', formData, true)
      setProgress(100)
      const count = res?.count ?? 0
      toast(`Training complete: ${count} Q&A pair${count !== 1 ? 's' : ''} learned!`, 'success')
      setTimeout(() => {
        setUploading(false)
        setProgress(0)
        setCSVModalOpen(false)
        getCategories('/training/categories')
        if (selectedCategory) loadCategoryQAs(selectedCategory.id)
      }, 800)
    } catch (err: any) {
      toast(err?.message || 'Upload failed', 'error')
      setUploading(false)
      setProgress(0)
    }
  }

  // Train unknown question
  const handleTrain = async (id: string, answer: string, categoryId: string) => {
    try {
      await api.post(`/training/unknown-questions/${id}/train`, { answer, category_id: categoryId || 'default' })
      toast('AI trained successfully!', 'success')
      getUnknown('/training/unknown-questions?status=pending')
      getCategories('/training/categories')
      if (selectedCategory && selectedCategory.id === categoryId) loadCategoryQAs(categoryId)
    } catch (err: any) { toast(err?.message || 'Training failed', 'error') }
  }

  // Ignore unknown question
  const handleIgnoreClick = (id: string) => { setIgnoreId(id); openIgnore() }
  const handleIgnoreConfirm = async () => {
    if (!ignoreId) return
    setIgnoreLoading(true)
    try {
      await api.post(`/training/unknown-questions/${ignoreId}/ignore`, {})
      toast('Question dismissed', 'success')
      getUnknown('/training/unknown-questions?status=pending')
    } catch (err: any) { toast(err?.message || 'Action failed', 'error') }
    finally { setIgnoreLoading(false); closeIgnore(); setIgnoreId('') }
  }

  const handleClearGapsClick = () => {
    setClearGapsConfirmOpen(true)
  }

  const handleClearGapsConfirm = async () => {
    setClearGapsLoading(true)
    try {
      await api.delete('/training/unknown-questions/clear')
      toast('All AI gaps cleared successfully', 'success')
      getUnknown('/training/unknown-questions?status=pending')
      setClearGapsConfirmOpen(false)
    } catch (err: any) {
      console.error('Failed to clear unknown questions:', err)
      toast(err?.message || 'Failed to clear AI gaps', 'error')
    } finally {
      setClearGapsLoading(false)
    }
  }

  // Category select
  const handleCategoryClick = (cat: any) => {
    setSelectedCategory(cat)
    loadCategoryQAs(cat.id)
  }

  // Add category
  const handleCreateCategory = async (name: string, desc: string, color: string) => {
    setAddCatLoading(true)
    try {
      await api.post('/training/categories', { name, description: desc, color })
      toast('Category created!', 'success')
      setAddCatOpen(false)
      getCategories('/training/categories')
    } catch (err: any) { toast(err?.message || 'Failed to create category', 'error') }
    finally { setAddCatLoading(false) }
  }

  // Add Q&A pair
  const handleCreateQAPair = async (question: string, answer: string, catId: string) => {
    setAddQALoading(true)
    try {
      await api.post('/training/qa', { category_id: catId, question, answer })
      toast('Trained Q&A added!', 'success')
      setAddQAOpen(false)
      getCategories('/training/categories')
      if (selectedCategory && selectedCategory.id === catId) loadCategoryQAs(catId)
    } catch (err: any) { toast(err?.message || 'Failed to add Q&A pair', 'error') }
    finally { setAddQALoading(false) }
  }

  // Edit Q&A pair
  const handleEditQAClick = (qa: any) => {
    setEditQAPair(qa)
    setEditQAOpen(true)
  }
  const handleUpdateQAPair = async (question: string, answer: string, catId: string) => {
    if (!editQAPair) return
    setEditQALoading(true)
    try {
      await api.put(`/training/qa/${editQAPair.id}`, { category_id: catId, question, answer })
      toast('Q&A pair updated!', 'success')
      setEditQAOpen(false)
      setEditQAPair(null)
      getCategories('/training/categories')
      if (selectedCategory) loadCategoryQAs(selectedCategory.id)
    } catch (err: any) { toast(err?.message || 'Failed to update Q&A pair', 'error') }
    finally { setEditQALoading(false) }
  }

  // Delete Q&A pair
  const handleDeleteQAPair = async (qaId: string) => {
    if (!confirm('Are you sure you want to delete this Q&A pair?')) return
    try {
      await api.delete(`/training/qa/${qaId}`)
      toast('Q&A pair deleted!', 'success')
      getCategories('/training/categories')
      if (selectedCategory) loadCategoryQAs(selectedCategory.id)
    } catch (err: any) { toast(err?.message || 'Failed to delete Q&A pair', 'error') }
  }

  // Delete category
  const handleDeleteCategory = async () => {
    if (!selectedCategory) return
    setDeleteCatLoading(true)
    try {
      await api.delete(`/training/categories/${selectedCategory.id}`)
      toast('Category and all Q&As deleted!', 'success')
      setConfirmDeleteCatOpen(false)
      setSelectedCategory(null)
      getCategories('/training/categories')
    } catch (err: any) { toast(err?.message || 'Failed to delete category', 'error') }
    finally { setDeleteCatLoading(false) }
  }

  const categories = catData?.categories || []
  const questions = uqData?.questions || []

  return (
    <div className="animate-page-in space-y-6 pt-2 pb-12">

      {/* ── Page Header ── */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-surface border border-default p-5 rounded-2xl shadow-sm relative overflow-hidden">
        <div className="absolute right-0 top-0 w-32 h-32 bg-noant-sky/5 rounded-full blur-3xl pointer-events-none" />
        <div className="flex items-start gap-4 min-w-0">
          <div className="w-12 h-12 rounded-2xl bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center shrink-0">
            <GraduationCap className="w-6 h-6 text-noant-sky animate-breathe" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-primary tracking-tight">Teach your Noant AI</h1>
            <p className="text-sm text-secondary mt-0.5 leading-relaxed">
              Train your custom AI on questions, products, and answers. It learns instantly.
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2 flex-wrap shrink-0">
          <button
            onClick={() => setCSVModalOpen(true)}
            disabled={isOffline}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl border border-default text-sm font-semibold transition-all hover:bg-inset hover:border-strong active:scale-95 shadow-sm cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            style={{ color: 'var(--text-secondary)' }}
          >
            <Upload className="w-4 h-4 text-noant-sky" />
            Upload CSV
          </button>
          <button
            onClick={() => setTestAIOpen(true)}
            disabled={isOffline}
            className="flex items-center gap-2 px-5 py-2.5 rounded-xl border text-sm font-semibold transition-all hover:bg-noant-sky hover:text-white hover:border-noant-sky active:scale-95 shadow-sm shadow-sky/5 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            style={{ borderColor: 'var(--border-default)', color: 'var(--text-secondary)' }}
          >
            <Sparkles className="w-4 h-4 text-noant-sky" />
            Test AI
          </button>
        </div>
      </div>

      {/* ── Global Search ── */}
      <div className="relative">
        <div className="relative flex items-center bg-surface border border-default rounded-2xl shadow-sm hover:border-strong transition-all focus-within:ring-2 focus-within:ring-noant-sky/20 focus-within:border-noant-sky">
          <Search className="w-5 h-5 text-tertiary ml-4 shrink-0" />
          <input
            type="text"
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            placeholder={isOffline ? "Search is disabled while offline..." : "Search trained questions, keywords, or answers…"}
            disabled={isOffline}
            className="w-full pl-3 pr-4 py-3.5 text-sm bg-transparent outline-none text-primary placeholder-tertiary rounded-2xl disabled:opacity-50"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="mr-3 p-1 rounded-full text-tertiary hover:bg-inset hover:text-primary transition-all active:scale-95 shrink-0"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>

        {/* Search Results Dropdown */}
        {searchQuery.trim() !== '' && (
          <div className="absolute left-0 right-0 mt-2 bg-surface border border-default rounded-2xl shadow-2xl z-50 overflow-hidden animate-slide-up max-h-[380px] overflow-y-auto">
            <div className="px-4 py-2.5 border-b border-default bg-inset flex justify-between items-center sticky top-0">
              <span className="text-[10px] font-bold text-secondary uppercase tracking-wider">Search Results</span>
              {searchLoading && <Loader2 className="w-3.5 h-3.5 text-noant-sky animate-spin" />}
            </div>
            {searchLoading && searchResults.length === 0 ? (
              <div className="p-8 text-center text-secondary text-sm">Searching knowledge base…</div>
            ) : searchResults.length === 0 ? (
              <div className="p-8 text-center text-secondary text-sm">No matching trained Q&amp;As found.</div>
            ) : (
              <div className="divide-y divide-default">
                {searchResults.map((qa: any) => {
                  const cat = categories.find((c: any) => c.id === qa.category_id)
                  return (
                    <div key={qa.id} className="p-4 hover:bg-inset/50 transition-colors flex items-start gap-4 justify-between">
                      <div className="space-y-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span
                            className="text-[10px] font-bold px-2 py-0.5 rounded-full shrink-0 border"
                            style={{
                              background: cat ? `${cat.color}0D` : 'var(--bg-inset)',
                              borderColor: cat ? `${cat.color}33` : 'var(--border-default)',
                              color: cat ? cat.color : 'var(--text-secondary)'
                            }}
                          >
                            {cat ? cat.name : 'Unassigned'}
                          </span>
                          <span className="text-[10px] text-tertiary">
                            Trained {new Date(qa.created_at).toLocaleDateString()}
                          </span>
                        </div>
                        <p className="text-sm font-semibold text-primary">Q: {qa.question}</p>
                        <p className="text-sm text-secondary">A: {qa.answer}</p>
                      </div>
                      <button
                        onClick={() => handleEditQAClick(qa)}
                        className="text-tertiary hover:text-noant-sky p-1.5 rounded-lg hover:bg-inset active:scale-95 shrink-0"
                        title="Edit QA"
                      >
                        <Edit3 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        )}
      </div>

      {/* ── Main Dual Pane ── */}
      <div className="grid grid-cols-1 xl:grid-cols-5 gap-6">

        {/* LEFT: AI Gaps */}
        <div className="xl:col-span-2 space-y-4">
          <div className="bg-surface border border-default rounded-2xl shadow-sm p-5 relative overflow-hidden">
            <div className="absolute top-0 left-0 w-1.5 h-full bg-amber-500 rounded-tl-2xl rounded-bl-2xl" />
            <div className="flex items-center justify-between mb-1">
              <h2 className="text-base font-bold text-primary tracking-tight">AI Gaps</h2>
              <div className="flex items-center gap-2">
                <span className="bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400 text-xs font-bold px-2.5 py-0.5 rounded-full">
                  {questions.length} pending
                </span>
                {questions.length > 0 && (
                  <button
                    onClick={handleClearGapsClick}
                    disabled={isOffline}
                    title="Clear all pending gaps"
                    className="p-1 rounded-lg hover:bg-red-500/10 text-tertiary hover:text-red-500 transition-colors active:scale-95 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                )}
              </div>
            </div>
            <p className="text-xs text-secondary leading-relaxed mb-4">
              Questions your AI didn't know how to answer. Train it below.
            </p>

            {uqLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="p-4 border border-default rounded-xl space-y-2">
                    <Skeleton className="h-4 w-2/3 rounded" />
                    <div className="flex gap-2">
                      <Skeleton className="h-8 w-16 rounded" />
                      <Skeleton className="h-8 w-16 rounded" />
                    </div>
                  </div>
                ))}
              </div>
            ) : questions.length === 0 ? (
              <div className="border border-dashed border-default rounded-xl py-10 px-4 text-center bg-inset/30 flex flex-col items-center">
                <div className="w-10 h-10 rounded-full bg-emerald-100 dark:bg-emerald-950/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center mb-2">
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2.5">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <h3 className="text-sm font-bold text-primary">No AI gaps detected!</h3>
                <p className="text-xs text-secondary max-w-xs mt-1">
                  Your AI is handling all questions flawlessly.
                </p>
              </div>
            ) : (
              <div className="max-h-[520px] overflow-y-auto pr-1 space-y-3 scrollbar-thin">
                {questions.map((q: any) => (
                  <UnknownQuestionItem
                    key={q.id}
                    question={q}
                    onTrain={handleTrain}
                    onIgnore={handleIgnoreClick}
                    categories={categories}
                  />
                ))}
              </div>
            )}
          </div>
        </div>

        {/* RIGHT: Knowledge Base */}
        <div className="xl:col-span-3 space-y-5">

          {/* Action Header Row */}
          <div className="flex flex-wrap items-center justify-between gap-3 bg-surface border border-default p-4 rounded-2xl shadow-sm">
            <div>
              <h2 className="text-base font-bold text-primary tracking-tight">Knowledge Base</h2>
              <p className="text-xs text-secondary mt-0.5">Organized folders of trained Q&amp;A pairs</p>
            </div>
            <div className="flex items-center gap-2 flex-wrap">
              <button
                onClick={() => setAddCatOpen(true)}
                disabled={isOffline}
                className="flex items-center gap-1.5 px-3.5 py-2 border border-default rounded-xl text-xs font-semibold hover:bg-inset hover:border-strong transition-all active:scale-95 cursor-pointer text-secondary disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <FolderPlus className="w-3.5 h-3.5" />
                New Category
              </button>
              <button
                onClick={() => {
                  setAddQAInitialCatId(selectedCategory?.id || 'default')
                  setAddQAOpen(true)
                }}
                disabled={isOffline}
                className="flex items-center gap-1.5 px-3.5 py-2 bg-noant-sky text-white rounded-xl text-xs font-semibold hover:bg-noant-sky-deep transition-all active:scale-95 shadow-sm shadow-sky/10 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Plus className="w-3.5 h-3.5 text-white" />
                Add Q&amp;A Pair
              </button>
            </div>
          </div>

          {/* Categories Grid */}
          {catLoading ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="rounded-2xl border border-default bg-surface p-4 space-y-3">
                  <div className="flex items-center gap-3">
                    <Skeleton className="w-3 h-3 rounded-full" />
                    <Skeleton className="h-5 w-24 rounded" />
                  </div>
                  <Skeleton className="h-3 w-16 rounded" />
                </div>
              ))}
            </div>
          ) : categories.length === 0 ? (
            <div className="flex flex-col items-center justify-center border border-dashed border-default py-16 rounded-2xl bg-surface/50 text-center">
              <div className="w-12 h-12 rounded-xl bg-inset flex items-center justify-center text-tertiary mb-3">
                <BookOpen className="w-6 h-6" />
              </div>
              <h3 className="text-sm font-semibold text-primary">No knowledge folders yet</h3>
              <p className="text-xs text-secondary mt-1 max-w-xs mb-4">
                Create a category folder or upload a CSV spreadsheet to get started.
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setAddCatOpen(true)}
                  className="flex items-center gap-1.5 px-4 py-2 border border-default rounded-xl text-xs font-semibold hover:bg-inset transition-all cursor-pointer text-secondary"
                >
                  <FolderPlus className="w-3.5 h-3.5" />
                  New Category
                </button>
                <button
                  onClick={() => setCSVModalOpen(true)}
                  className="flex items-center gap-1.5 px-4 py-2 bg-noant-sky text-white rounded-xl text-xs font-semibold hover:bg-noant-sky-deep transition-all cursor-pointer"
                >
                  <Upload className="w-3.5 h-3.5" />
                  Upload CSV
                </button>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {categories.map((c: any) => (
                <button
                  key={c.id}
                  onClick={() => handleCategoryClick(c)}
                  className="group relative flex flex-col justify-between text-left p-4 rounded-2xl bg-surface border border-default shadow-sm hover:shadow-md hover:border-strong transition-all duration-300 active:scale-[0.98] outline-none overflow-hidden cursor-pointer"
                  style={{ borderLeft: `4px solid ${c.color}` }}
                >
                  <div className="space-y-1">
                    <h3 className="text-sm font-bold text-primary truncate group-hover:text-noant-sky transition-colors">{c.name}</h3>
                    <p className="text-xs text-secondary line-clamp-2 leading-relaxed">{c.description || 'No description'}</p>
                  </div>
                  <div className="flex items-center justify-between mt-4 pt-2 border-t border-subtle w-full text-[10px] font-semibold text-tertiary">
                    <span className="bg-inset px-2 py-0.5 rounded-full text-secondary">{c.qa_count} trained pairs</span>
                    <span className="flex items-center gap-0.5 text-noant-sky opacity-0 group-hover:opacity-100 transition-opacity">
                      Open <ChevronRight className="w-3 h-3" />
                    </span>
                  </div>
                </button>
              ))}
            </div>
          )}

          {/* Quick Upload Nudge Banner (when categories exist) */}
          {!catLoading && categories.length > 0 && (
            <button
              onClick={() => setCSVModalOpen(true)}
              className="w-full flex items-center gap-3 p-4 bg-surface border border-dashed border-default rounded-2xl hover:border-noant-sky/40 hover:bg-inset/30 transition-all text-left group cursor-pointer"
            >
              <div className="w-9 h-9 rounded-xl bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center shrink-0 group-hover:bg-noant-sky/15 transition-all">
                <FileText className="w-4 h-4 text-noant-sky" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold text-primary">Bulk upload via CSV spreadsheet</p>
                <p className="text-xs text-tertiary mt-0.5">Upload hundreds of Q&amp;A pairs at once — columns: Category, Question, Answer</p>
              </div>
              <Upload className="w-4 h-4 text-tertiary group-hover:text-noant-sky transition-colors shrink-0" />
            </button>
          )}

        </div>
      </div>

      {/* ── Modals ── */}

      {/* CSV Upload Modal */}
      <CSVUploadModal
        open={csvModalOpen}
        onClose={() => !Uploading && setCSVModalOpen(false)}
        onUpload={handleUpload}
        uploading={Uploading}
        progress={progress}
      />

      {/* Add Category Modal */}
      <AddCategoryModal
        open={addCatOpen}
        onClose={() => setAddCatOpen(false)}
        onSubmit={handleCreateCategory}
        loading={addCatLoading}
      />

      {/* Add Q&A Modal */}
      <QAModal
        open={addQAOpen}
        onClose={() => setAddQAOpen(false)}
        onSubmit={handleCreateQAPair}
        loading={addQALoading}
        categories={categories}
        initialCatId={addQAInitialCatId}
        mode="add"
      />

      {/* Edit Q&A Modal */}
      <QAModal
        open={editQAOpen}
        onClose={() => { setEditQAOpen(false); setEditQAPair(null) }}
        onSubmit={handleUpdateQAPair}
        loading={editQALoading}
        categories={categories}
        mode="edit"
        initialQuestion={editQAPair?.question}
        initialAnswer={editQAPair?.answer}
        initialCatID={editQAPair?.category_id || selectedCategory?.id || 'default'}
      />

      {/* Confirm Delete Category */}
      <ConfirmModal
        open={confirmDeleteCatOpen}
        onClose={() => setConfirmDeleteCatOpen(false)}
        onConfirm={handleDeleteCategory}
        title="Delete this entire category?"
        description={`This will permanently delete "${selectedCategory?.name}" and ALL its associated trained Q&A pairs. This action cannot be undone.`}
        variant="danger"
        confirmText="Permanently Delete Everything"
        loading={deleteCatLoading}
      />

      {/* Confirm Ignore */}
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

      {/* Confirm Clear Gaps */}
      <ConfirmModal
        open={clearGapsConfirmOpen}
        onClose={() => setClearGapsConfirmOpen(false)}
        onConfirm={handleClearGapsConfirm}
        title="Clear All Pending AI Gaps?"
        description="Are you absolutely sure you want to dismiss and clear all pending questions? This action is permanent."
        confirmText="Clear All"
        cancelText="Cancel"
        variant="danger"
        loading={clearGapsLoading}
        requireTypeConfirm={true}
        confirmPhrase="DELETE GAPS"
      />

      {/* Q&A Table Sheet */}
      <QATableSheet
        category={selectedCategory}
        qaList={categoryQAs}
        loading={categoryQALoading}
        onClose={() => setSelectedCategory(null)}
        onAddQA={() => {
          setAddQAInitialCatId(selectedCategory?.id || 'default')
          setAddQAOpen(true)
        }}
        onEditQA={handleEditQAClick}
        onDeleteQA={handleDeleteQAPair}
        onDeleteCategory={() => setConfirmDeleteCatOpen(true)}
      />

      {/* Test AI Drawer */}
      <TestAIDrawer open={testAIOpen} onClose={() => setTestAIOpen(false)} />
    </div>
  )
}
