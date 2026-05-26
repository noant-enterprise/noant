import { useState } from 'react'
import { Button } from '@/components/ui/Button'

interface UnknownQuestionProps {
  question: {
    id: string
    question: string
  }
  onTrain: (id: string, answer: string, categoryId: string) => Promise<void>
  onIgnore: (id: string) => void
  categories?: Array<{ id: string; name: string }>
}

export function UnknownQuestionItem({ question, onIgnore, onTrain, categories = [] }: UnknownQuestionProps) {
  const [isTraining, setIsTraining] = useState(false)
  const [answer, setAnswer] = useState('')
  const [categoryId, setCategoryId] = useState('default')
  const [loading, setLoading] = useState(false)

  const handleTrainSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!answer.trim()) return
    setLoading(true)
    try {
      await onTrain(question.id, answer.trim(), categoryId.trim())
    } catch (err) {
      // error handled by parent / hook
    } finally {
      setLoading(false)
    }
  }

  if (isTraining) {
    return (
      <div className="p-4 lg:p-5 bg-surface border border-noant-sky/30 rounded-xl mb-3 shadow-sm animate-slide-up">
        <div className="flex items-start justify-between mb-3.5">
          <div className="space-y-0.5">
            <span className="text-[11px] font-semibold text-noant-sky uppercase tracking-wider">Train AI Response</span>
            <h4 className="text-sm font-semibold text-primary italic">"{question.question}"</h4>
          </div>
          <button
            type="button"
            onClick={() => setIsTraining(false)}
            className="text-xs text-tertiary hover:text-primary transition-colors font-medium"
            disabled={loading}
          >
            Cancel
          </button>
        </div>

        <form onSubmit={handleTrainSubmit} className="space-y-3">
          <div>
            <label htmlFor={`answer-${question.id}`} className="block text-xs font-semibold text-secondary mb-1">
              Correct Answer
            </label>
            <textarea
              id={`answer-${question.id}`}
              value={answer}
              onChange={(e) => setAnswer(e.target.value)}
              placeholder="What is the correct answer to this question?"
              rows={3}
              className="w-full text-sm px-3 py-2 rounded-lg border border-default outline-none bg-inset focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all text-primary"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label htmlFor={`cat-${question.id}`} className="block text-xs font-semibold text-secondary mb-1">
              Select Category
            </label>
            <select
              id={`cat-${question.id}`}
              value={categoryId}
              onChange={(e) => setCategoryId(e.target.value)}
              className="w-full text-sm px-3 py-2 rounded-lg border border-default outline-none bg-inset focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all text-primary"
              disabled={loading}
            >
              <option value="default">Default</option>
              {categories.map(cat => (
                cat.id !== 'default' && (
                  <option key={cat.id} value={cat.id}>
                    {cat.name}
                  </option>
                )
              ))}
            </select>
          </div>

          <div className="flex justify-end gap-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsTraining(false)}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button type="submit" size="sm" loading={loading} disabled={!answer.trim() || loading}>
              Save Answer
            </Button>
          </div>
        </form>
      </div>
    )
  }

  return (
    <div className="flex items-center justify-between p-3 lg:p-4 bg-surface border border-default rounded-xl mb-2 gap-3 hover:border-strong transition-colors duration-200">
      <span className="text-sm text-primary italic truncate min-w-0">"{question.question}"</span>
      <div className="flex gap-2 shrink-0">
        <Button size="sm" onClick={() => setIsTraining(true)}>
          Train
        </Button>
        <Button variant="ghost" size="sm" className="text-tertiary hover:text-red-500" onClick={() => onIgnore(question.id)}>
          Ignore
        </Button>
      </div>
    </div>
  )
}
