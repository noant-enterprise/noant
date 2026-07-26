import { useState, useEffect, useCallback } from 'react'
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
  const [error, setError] = useState<string | null>(null)

  const fetchData = useCallback(() => {
    setLoading(true)
    setError(null)
    adminApi.getAIHealth()
      .then(res => setData({
        ...res,
        unanswered_questions: res.unanswered_questions ?? [],
        accuracy_history: res.accuracy_history ?? [],
        sentiment_breakdown: res.sentiment_breakdown ?? { positive: 0, neutral: 0, negative: 0 },
      }))
      .catch(() => setError('Failed to load AI health data'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => { fetchData() }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}
