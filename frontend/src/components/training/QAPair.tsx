interface QAPairProps {
  qa: {
    id: string
    question: string
    answer: string
  }
}

export function QAPair({ qa }: QAPairProps) {
  return (
    <div className="bg-surface border border-default rounded-xl p-4 lg:p-5 mb-2">
      <p className="font-semibold text-sm text-primary mb-2">{qa.question}</p>
      <p className="text-sm text-secondary pl-4 border-l-2 border-noant-sky/30">{qa.answer}</p>
    </div>
  )
}
