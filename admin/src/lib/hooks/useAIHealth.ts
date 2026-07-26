import { useState, useEffect } from 'react'
import { adminApi } from '@/lib/api'
import type { AIHealthResponse } from '@/types'

export function useAIHealth() {
  const [data, setData] = useState<AIHealthResponse>({
    accuracy: 0,
    accuracy_trend: 0,
    total_queries: 0,
    answered_correctly: 0,
    unanswered_questions: [],
    accuracy_history: [],
    sentiment_breakdown: { positive: 0, neutral: 0, negative: 0 },
  })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    adminApi.getAIHealth()
      .then(res => setData(res))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return { data, loading }
}
