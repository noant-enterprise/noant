import { useState, useEffect } from 'react'
import type { AIHealthData } from '@/types'

const MOCK_AI: AIHealthData = {
  accuracy: 94.2,
  accuracy_trend: 2.1,
  total_queries: 12430,
  answered_correctly: 11709,
  unanswered_questions: [
    { question: 'Do you support Paystack payments?', count: 23, last_seen: '2026-07-26T08:00:00Z' },
    { question: 'Can I integrate with Shopify?', count: 18, last_seen: '2026-07-26T07:30:00Z' },
    { question: 'What is your refund policy?', count: 15, last_seen: '2026-07-25T16:00:00Z' },
    { question: 'Do you have an API?', count: 12, last_seen: '2026-07-25T14:00:00Z' },
    { question: 'Can I export my data?', count: 9, last_seen: '2026-07-24T12:00:00Z' },
    { question: 'How does pricing work for teams?', count: 8, last_seen: '2026-07-24T10:00:00Z' },
  ],
  accuracy_history: [
    { date: 'Jul 20', accuracy: 89.5, queries: 1520 },
    { date: 'Jul 21', accuracy: 91.2, queries: 1680 },
    { date: 'Jul 22', accuracy: 92.1, queries: 1890 },
    { date: 'Jul 23', accuracy: 93.0, queries: 1740 },
    { date: 'Jul 24', accuracy: 93.8, queries: 1920 },
    { date: 'Jul 25', accuracy: 94.0, queries: 2050 },
    { date: 'Jul 26', accuracy: 94.2, queries: 2140 },
  ],
  sentiment_breakdown: { positive: 62, neutral: 28, negative: 10 },
}

export function useAIHealth() {
  const [data, setData] = useState<AIHealthData>(MOCK_AI)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setData(MOCK_AI)
      setLoading(false)
    }, 400)
    return () => clearTimeout(timer)
  }, [])

  return { data, loading }
}
