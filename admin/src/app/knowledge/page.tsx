import { useState, useEffect } from 'react'
import { adminApi } from '@/lib/api'
import type { KnowledgeEntry } from '@/types'
import { Search, Brain, CheckCircle, XCircle, Clock } from 'lucide-react'
import { timeAgo } from '@/lib/utils'
import { SkeletonTableRows } from '@/components/ui/Skeleton'
import { ErrorBanner, EmptyState } from '@/components/ui/Feedback'

const STATUS_CONFIG = {
  pending: { icon: Clock, color: 'text-warning', bg: 'bg-warning/10', label: 'Pending' },
  trained: { icon: CheckCircle, color: 'text-success', bg: 'bg-success/10', label: 'Trained' },
  ignored: { icon: XCircle, color: 'text-text-tertiary', bg: 'bg-bg-inset', label: 'Ignored' },
}

export default function KnowledgeBasePage() {
  const [questions, setQuestions] = useState<KnowledgeEntry[]>([])
  const [summary, setSummary] = useState({ pending: 0, trained: 0, ignored: 0 })
  const [loading, setLoading] = useState(true)
  const [statusFilter, setStatusFilter] = useState('pending')
  const [search, setSearch] = useState('')
  const [trainingId, setTrainingId] = useState<string | null>(null)
  const [trainingAnswer, setTrainingAnswer] = useState('')
  const [trainingMessage, setTrainingMessage] = useState('')
  const [error, setError] = useState<string | null>(null)

  const fetchQuestions = () => {
    setLoading(true)
    setError(null)
    adminApi.getKnowledgeBase({ status: statusFilter || undefined, search: search || undefined })
      .then(res => {
        setQuestions(res.questions ?? [])
        setSummary(res.summary ?? { pending: 0, trained: 0, ignored: 0 })
      })
      .catch(() => { setError('Failed to load questions. Please try again.') })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    const timer = setTimeout(fetchQuestions, 300)
    return () => clearTimeout(timer)
  }, [statusFilter, search])

  const handleTrain = async (questionId: string) => {
    if (!trainingAnswer.trim()) return
    setTrainingMessage('')
    try {
      await adminApi.trainKnowledge({ question_id: questionId, answer: trainingAnswer })
      setTrainingMessage('Trained successfully!')
      setTrainingId(null)
      setTrainingAnswer('')
      fetchQuestions()
    } catch {
      setTrainingMessage('Failed to train — try again')
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Knowledge Base</h1>
        <p className="text-sm text-text-tertiary">Questions the AI can't answer — train it to help customers better</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {(['pending', 'trained', 'ignored'] as const).map(status => {
          const cfg = STATUS_CONFIG[status]
          return (
            <button
              key={status}
              onClick={() => setStatusFilter(status)}
              className={`rounded-xl border p-4 text-left transition-colors ${
                statusFilter === status ? 'border-brand-sky bg-brand-sky/5' : 'border-border bg-bg-surface hover:border-border'
              }`}
            >
              <div className="flex items-center gap-2">
                <cfg.icon className={`h-4 w-4 ${cfg.color}`} />
                <span className="text-sm font-medium text-text-secondary">{cfg.label}</span>
              </div>
              <p className="mt-1 text-2xl font-bold text-text-primary">{summary[status]}</p>
            </button>
          )
        })}
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-tertiary" />
        <input
          type="text"
          placeholder="Search questions..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          className="w-full rounded-lg border border-border bg-bg-surface py-2 pl-10 pr-4 text-sm text-text-primary placeholder:text-text-tertiary focus:border-brand-sky focus:outline-none"
        />
      </div>

      {trainingMessage && (
        <div className={`rounded-lg px-4 py-2 text-sm ${trainingMessage.includes('success') ? 'bg-success/10 text-success' : 'bg-danger/10 text-danger'}`}>
          {trainingMessage}
        </div>
      )}

      {error && <ErrorBanner message={error} onRetry={fetchQuestions} />}

      <div className="rounded-xl border border-border bg-bg-surface">
        {loading ? (
          <div className="overflow-x-auto">
            <table className="w-full"><tbody><SkeletonTableRows rows={5} cols={4} /></tbody></table>
          </div>
        ) : questions.length === 0 ? (
          <EmptyState
            icon={Brain}
            title={`No ${statusFilter} questions found`}
            description="Try adjusting your filters or check back later"
          />
        ) : (
          <div className="overflow-x-auto">
              <div className="divide-y divide-border">
                {questions.map(q => (
                  <div key={q.id} className="p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium text-text-primary">{q.question}</p>
                        <div className="mt-1 flex items-center gap-3 text-xs text-text-tertiary">
                          <span>{q.user_email || 'Unknown user'}</span>
                          <span>{q.channel || 'web'}</span>
                          <span>{timeAgo(q.created_at)}</span>
                        </div>
                        {q.suggested_answer && (
                          <p className="mt-2 rounded bg-bg-inset px-3 py-2 text-xs text-text-secondary">{q.suggested_answer}</p>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        {STATUS_CONFIG[q.status as keyof typeof STATUS_CONFIG] && (
                          <span className={`rounded-full px-2 py-0.5 text-xs ${STATUS_CONFIG[q.status as keyof typeof STATUS_CONFIG].bg} ${STATUS_CONFIG[q.status as keyof typeof STATUS_CONFIG].color}`}>
                            {STATUS_CONFIG[q.status as keyof typeof STATUS_CONFIG].label}
                          </span>
                        )}
                        {q.status === 'pending' && (
                          <button
                            onClick={() => { setTrainingId(trainingId === q.id ? null : q.id); setTrainingAnswer(q.suggested_answer || '') }}
                            className="rounded-lg bg-brand-sky px-3 py-1.5 text-xs font-medium text-white hover:bg-brand-sky/80"
                          >
                            {trainingId === q.id ? 'Cancel' : 'Train'}
                          </button>
                        )}
                      </div>
                    </div>
                    {trainingId === q.id && (
                      <div className="mt-3 rounded-lg border border-border bg-bg-inset p-3">
                        <label className="mb-1 block text-xs text-text-tertiary">Answer this question for the AI:</label>
                        <textarea
                          value={trainingAnswer}
                          onChange={e => setTrainingAnswer(e.target.value)}
                          rows={3}
                          className="w-full rounded border border-border bg-bg-surface px-3 py-2 text-sm text-text-primary placeholder:text-text-tertiary focus:border-brand-sky focus:outline-none"
                          placeholder="Type the answer the AI should give..."
                        />
                        <div className="mt-2 flex justify-end">
                          <button
                            onClick={() => handleTrain(q.id)}
                            disabled={!trainingAnswer.trim()}
                            className="rounded-lg bg-success px-4 py-1.5 text-xs font-medium text-white hover:bg-success/80 disabled:opacity-50"
                          >
                            Train AI
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
          </div>
        )}
      </div>
    </div>
  )
}
