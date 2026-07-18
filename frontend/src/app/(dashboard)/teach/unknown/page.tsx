import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useToast } from '@/components/ui/Toast'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import {
  ArrowLeft, Check, X, BookOpen, AlertTriangle,
} from 'lucide-react'
import { api } from '@/lib/api'
import type { UnknownQuestion, UnknownQuestionListResponse } from '@/types'

type Tab = 'pending' | 'trained' | 'ignored'

export default function UnknownQuestionsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [tab, setTab] = useState<Tab>('pending')
  const [questions, setQuestions] = useState<UnknownQuestion[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [loadMoreLoading, setLoadMoreLoading] = useState(false)
  const [offset, setOffset] = useState(0)
  const limit = 25

  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [batchAnswer, setBatchAnswer] = useState('')
  const [batchCategory, setBatchCategory] = useState('default')
  const [showBatchForm, setShowBatchForm] = useState(false)
  const [categories, setCategories] = useState<Array<{ id: string; name: string }>>([])
  const [batching, setBatching] = useState(false)

  const fetchQuestions = useCallback(async (reset = false) => {
    if (reset) { setLoading(true); setOffset(0) }
    const currentOffset = reset ? 0 : offset
    try {
      const res = await api.get<UnknownQuestionListResponse>(
        `/training/unknown-questions?status=${tab}&limit=${limit}&offset=${currentOffset}`
      )
      if (reset) {
        setQuestions(res.questions)
      } else {
        setQuestions(prev => [...prev, ...res.questions])
      }
      setTotal(res.total)
      setOffset(currentOffset + limit)
    } catch {
      toast('Failed to load questions', 'error')
    } finally {
      setLoading(false)
      setLoadMoreLoading(false)
    }
  }, [tab, offset, limit, toast])

  useEffect(() => {
    setQuestions([])
    setOffset(0)
    setSelected(new Set())
    fetchQuestions(true)
  }, [tab])

  useEffect(() => {
    api.get<{ categories: Array<{ id: string; name: string }> }>('/training/categories')
      .then(res => setCategories(res.categories || []))
      .catch(() => {})
  }, [])

  const handleLoadMore = async () => {
    setLoadMoreLoading(true)
    await fetchQuestions(false)
  }

  const handleTrain = async (id: string, answer: string, categoryId: string) => {
    try {
      await api.post(`/training/unknown-questions/${id}/train`, { answer, category_id: categoryId })
      setQuestions(prev => prev.filter(q => q.id !== id))
      setTotal(prev => prev - 1)
      toast('Question trained', 'success')
    } catch {
      toast('Failed to train question', 'error')
    }
  }

  const handleIgnore = async (id: string) => {
    try {
      await api.post(`/training/unknown-questions/${id}/ignore`, {})
      setQuestions(prev => prev.filter(q => q.id !== id))
      setTotal(prev => prev - 1)
    } catch {
      toast('Failed to ignore question', 'error')
    }
  }

  const toggleSelect = (id: string) => {
    setSelected(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const selectAll = () => {
    if (selected.size === questions.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(questions.map(q => q.id)))
    }
  }

  const handleBatchTrain = async () => {
    if (!batchAnswer.trim() || selected.size === 0) return
    setBatching(true)
    try {
      await api.post('/training/unknown-questions/batch-train', {
        ids: Array.from(selected),
        answer: batchAnswer.trim(),
        category_id: batchCategory,
      })
      setQuestions(prev => prev.filter(q => !selected.has(q.id)))
      setTotal(prev => prev - selected.size)
      setSelected(new Set())
      setBatchAnswer('')
      setShowBatchForm(false)
      toast(`${selected.size} questions trained`, 'success')
    } catch {
      toast('Batch train failed', 'error')
    } finally {
      setBatching(false)
    }
  }

  const handleBatchIgnore = async () => {
    if (selected.size === 0) return
    setBatching(true)
    try {
      await api.post('/training/unknown-questions/batch-ignore', {
        ids: Array.from(selected),
      })
      setQuestions(prev => prev.filter(q => !selected.has(q.id)))
      setTotal(prev => prev - selected.size)
      setSelected(new Set())
      toast(`${selected.size} questions ignored`, 'success')
    } catch {
      toast('Batch ignore failed', 'error')
    } finally {
      setBatching(false)
    }
  }

  return (
    <div className="animate-fade-in space-y-5 max-w-4xl mx-auto pt-2">
      {/* Header */}
      <div className="flex items-center gap-3">
        <button onClick={() => navigate('/teach')} className="w-8 h-8 rounded-lg bg-surface border border-border flex items-center justify-center text-tertiary hover:text-primary transition-colors">
          <ArrowLeft size={16} />
        </button>
        <div>
          <h1 className="text-lg font-semibold">Unknown Questions Review</h1>
          <p className="text-xs text-tertiary">{total} total {tab === 'pending' ? 'need attention' : 'processed'}</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-surface rounded-xl p-1 border border-border w-fit">
        {(['pending', 'trained', 'ignored'] as Tab[]).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              tab === t ? 'bg-primary text-white shadow-sm' : 'text-secondary hover:text-primary'
            }`}
          >
            {t === 'pending' && <AlertTriangle size={14} className="inline mr-1.5 -mt-0.5" />}
            {t === 'trained' && <Check size={14} className="inline mr-1.5 -mt-0.5" />}
            {t === 'ignored' && <X size={14} className="inline mr-1.5 -mt-0.5" />}
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {/* Batch actions bar */}
      {selected.size > 0 && (
        <div className="bg-surface rounded-xl p-3 border border-primary/30 flex items-center justify-between gap-3 animate-slide-up">
          <span className="text-sm font-medium">{selected.size} selected</span>
          <div className="flex gap-2">
            <Button size="sm" variant="ghost" onClick={() => setSelected(new Set())}>Deselect</Button>
            {tab === 'pending' && (
              <>
                <Button size="sm" onClick={() => setShowBatchForm(true)}>Train All</Button>
                <Button size="sm" variant="ghost" className="text-red-500" onClick={handleBatchIgnore} loading={batching}>Ignore All</Button>
              </>
            )}
          </div>
        </div>
      )}

      {/* Batch train form */}
      {showBatchForm && (
        <Card>
          <CardHeader className="px-4 py-3">
            <CardTitle>Batch train {selected.size} questions</CardTitle>
          </CardHeader>
          <CardBody className="p-4 space-y-3">
            <div>
              <label className="text-xs font-medium text-secondary mb-1 block">Answer for all selected</label>
              <textarea
                value={batchAnswer}
                onChange={e => setBatchAnswer(e.target.value)}
                placeholder="Enter the answer that works for all selected questions..."
                rows={3}
                className="w-full text-sm px-3 py-2 rounded-lg border border-default outline-none bg-inset focus:border-primary focus:ring-1 focus:ring-primary/20 transition-all"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-secondary mb-1 block">Category</label>
              <select
                value={batchCategory}
                onChange={e => setBatchCategory(e.target.value)}
                className="w-full text-sm px-3 py-2 rounded-lg border border-default outline-none bg-inset focus:border-primary"
              >
                <option value="default">Default</option>
                {categories.map(c => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
            <div className="flex gap-2 justify-end">
              <Button size="sm" variant="ghost" onClick={() => setShowBatchForm(false)}>Cancel</Button>
              <Button size="sm" onClick={handleBatchTrain} loading={batching} disabled={!batchAnswer.trim()}>Save All</Button>
            </div>
          </CardBody>
        </Card>
      )}

      {/* Question list */}
      <div className="space-y-2">
        {loading ? (
          Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="bg-surface rounded-xl p-4 border border-border">
              <Skeleton className="h-4 w-3/4 rounded mb-2" />
              <Skeleton className="h-3 w-1/4 rounded" />
            </div>
          ))
        ) : questions.length === 0 ? (
          <div className="text-center py-16">
            <BookOpen size={40} className="mx-auto text-tertiary mb-3" />
            <p className="text-sm text-secondary font-medium">No {tab} questions</p>
            <p className="text-xs text-tertiary mt-1">
              {tab === 'pending' ? 'Your AI is answering everything correctly!' : 'No processed questions yet'}
            </p>
          </div>
        ) : (
          <>
            {/* Select all checkbox */}
            {tab === 'pending' && questions.length > 1 && (
              <div className="flex items-center gap-2 px-1 py-1">
                <input
                  type="checkbox"
                  checked={selected.size === questions.length}
                  onChange={selectAll}
                  className="rounded border-border"
                />
                <span className="text-xs text-tertiary">Select all ({questions.length})</span>
              </div>
            )}

            {questions.map(q => (
              <QuestionCard
                key={q.id}
                question={q}
                tab={tab}
                categories={categories}
                selected={selected.has(q.id)}
                onToggleSelect={() => toggleSelect(q.id)}
                onTrain={handleTrain}
                onIgnore={handleIgnore}
              />
            ))}

            {/* Load more */}
            {questions.length < total && (
              <div className="text-center pt-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleLoadMore}
                  loading={loadMoreLoading}
                >
                  Load more ({total - questions.length} remaining)
                </Button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

function QuestionCard({
  question, tab, categories, selected, onToggleSelect, onTrain, onIgnore,
}: {
  question: UnknownQuestion
  tab: Tab
  categories: Array<{ id: string; name: string }>
  selected: boolean
  onToggleSelect: () => void
  onTrain: (id: string, answer: string, categoryId: string) => Promise<void>
  onIgnore: (id: string) => void
}) {
  const [training, setTraining] = useState(false)
  const [answer, setAnswer] = useState(question.suggested_answer || '')
  const [categoryId, setCategoryId] = useState(question.category_id || 'default')
  const [saving, setSaving] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!answer.trim()) return
    setSaving(true)
    await onTrain(question.id, answer.trim(), categoryId)
    setSaving(false)
    setTraining(false)
  }

  return (
    <div className={`bg-surface rounded-xl border transition-all ${
      selected ? 'border-primary/50 ring-1 ring-primary/20' : 'border-border'
    }`}>
      <div className="flex items-start gap-3 p-4">
        {tab === 'pending' && (
          <input
            type="checkbox"
            checked={selected}
            onChange={onToggleSelect}
            className="mt-1 rounded border-border shrink-0"
          />
        )}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="text-sm font-medium text-primary">"{question.question}"</p>
              <div className="flex items-center gap-2 mt-1.5">
                {question.channel && (
                  <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-primary/10 text-primary font-medium">{question.channel}</span>
                )}
                <span className="text-[10px] text-tertiary">{new Date(question.created_at).toLocaleDateString()}</span>
              </div>
            </div>
            <div className="flex gap-1.5 shrink-0">
              {tab === 'pending' && !training && (
                <>
                  <Button size="sm" onClick={() => setTraining(true)}>Train</Button>
                  <Button size="sm" variant="ghost" className="text-tertiary hover:text-red-500" onClick={() => onIgnore(question.id)}>Ignore</Button>
                </>
              )}
              {tab !== 'pending' && (
                <span className={`text-[10px] font-medium px-2 py-1 rounded-full ${
                  tab === 'trained' ? 'bg-green-500/10 text-green-600' : 'bg-tertiary/10 text-tertiary'
                }`}>
                  {tab}
                </span>
              )}
            </div>
          </div>

          {/* Training form inline */}
          {training && (
            <form onSubmit={handleSubmit} className="mt-4 pt-4 border-t border-border space-y-3">
              <div>
                <label className="text-xs font-medium text-secondary mb-1 block">Answer</label>
                <textarea
                  value={answer}
                  onChange={e => setAnswer(e.target.value)}
                  placeholder="Write the correct answer..."
                  rows={3}
                  className="w-full text-sm px-3 py-2 rounded-lg border border-default outline-none bg-inset focus:border-primary focus:ring-1 focus:ring-primary/20 transition-all"
                  required
                />
              </div>
              <div>
                <label className="text-xs font-medium text-secondary mb-1 block">Category</label>
                <select
                  value={categoryId}
                  onChange={e => setCategoryId(e.target.value)}
                  className="w-full text-sm px-3 py-2 rounded-lg border border-default outline-none bg-inset focus:border-primary"
                >
                  <option value="default">Default</option>
                  {categories.map(c => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>
              <div className="flex justify-end gap-2">
                <Button size="sm" variant="ghost" type="button" onClick={() => setTraining(false)}>Cancel</Button>
                <Button size="sm" type="submit" loading={saving} disabled={!answer.trim()}>Save</Button>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  )
}
