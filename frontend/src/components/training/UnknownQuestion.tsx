import { Button } from '@/components/ui/Button'

interface UnknownQuestionProps {
  question: {
    id: string
    question: string
  }
  onTrain: (id: string, question: string) => void
  onIgnore: (id: string) => void
}

export function UnknownQuestionItem({ question, onIgnore, onTrain }: UnknownQuestionProps) {
  return (
    <div className="flex items-center justify-between p-3 lg:p-4 bg-surface border border-default rounded-xl mb-2 gap-3">
      <span className="text-sm text-primary italic truncate min-w-0">"{question.question}"</span>
      <div className="flex gap-2 shrink-0">
        <Button size="sm" onClick={() => onTrain(question.id, question.question)}>
          Train
        </Button>
        <Button variant="ghost" size="sm" className="text-tertiary hover:text-red-500" onClick={() => onIgnore(question.id)}>
          Ignore
        </Button>
      </div>
    </div>
  )
}
