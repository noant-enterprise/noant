import { useState, useCallback, useRef } from 'react'
import { api, type APIClient } from '@/lib/api'
import { APIError } from '@/types'
import { useToast } from '@/components/ui/Toast'

type APIMethod = keyof Pick<APIClient, 'get' | 'post' | 'put' | 'delete'>

interface UseAPIState<T> {
  data: T | null
  loading: boolean
  loadingMore: boolean
  error: APIError | null
}

interface LastRequest {
  method: APIMethod
  endpoint: string
  body?: unknown
  isFormData?: boolean
}

const MAX_RETRIES = 3

function getRetryDelay(attempt: number): number {
  return Math.min(1000 * Math.pow(2, attempt), 8000)
}

function isRetryableError(error: APIError): boolean {
  return error.status >= 500 || error.status === 0 || error.message.toLowerCase().includes('network')
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function mergeData(existing: any, incoming: any): any {
  if (Array.isArray(existing) && Array.isArray(incoming)) {
    return [...existing, ...incoming]
  }
  if (typeof existing === 'object' && existing !== null && typeof incoming === 'object' && incoming !== null) {
    const result = { ...existing }
    for (const key of Object.keys(incoming)) {
      const existingVal = result[key]
      const incomingVal = incoming[key]
      if (Array.isArray(existingVal) && Array.isArray(incomingVal)) {
        result[key] = [...existingVal, ...incomingVal]
      } else {
        result[key] = incomingVal
      }
    }
    return result
  }
  return incoming
}

export function useAPI<T>() {
  const { toast } = useToast()
  const [state, setState] = useState<UseAPIState<T>>({
    data: null,
    loading: false,
    loadingMore: false,
    error: null,
  })
  const lastRequest = useRef<LastRequest | null>(null)

  const execute = useCallback(async (
    method: APIMethod,
    endpoint: string,
    body?: unknown,
    isFormData?: boolean,
    attempt = 0,
    isLoadMore = false
  ): Promise<T> => {
    lastRequest.current = { method, endpoint, body, isFormData }
    setState(prev => ({
      ...prev,
      loading: isLoadMore ? false : true,
      loadingMore: isLoadMore,
      error: null,
    }))
    
    try {
      const data = await api[method]<T>(endpoint, body, isFormData)
      setState(prev => ({
        data: isLoadMore && prev.data ? mergeData(prev.data, data) as T : data,
        loading: false,
        loadingMore: false,
        error: null,
      }))
      return data
    } catch (err) {
      const error = err instanceof APIError ? err : new APIError('Unknown error', 500)
      
      if (isRetryableError(error) && attempt < MAX_RETRIES) {
        await new Promise(r => setTimeout(r, getRetryDelay(attempt)))
        return execute(method, endpoint, body, isFormData, attempt + 1, isLoadMore)
      }
      
      setState(prev => ({ ...prev, loading: false, loadingMore: false, error }))
      toast(error.message || 'Request failed. Tap Retry to try again.', 'error')
      throw error
    }
  }, [toast])

  const loadMore = useCallback((endpoint: string) => {
    return execute('get', endpoint, undefined, undefined, 0, true)
  }, [execute])

  const retry = useCallback(async (): Promise<T | undefined> => {
    if (!lastRequest.current) {
      toast('Nothing to retry', 'error')
      return
    }
    const { method, endpoint, body, isFormData } = lastRequest.current
    return execute(method, endpoint, body, isFormData, 0)
  }, [execute, toast])

  const get = useCallback((endpoint: string) => execute('get', endpoint), [execute])
  const post = useCallback((endpoint: string, body?: unknown, isFormData?: boolean) => execute('post', endpoint, body, isFormData), [execute])
  const put = useCallback((endpoint: string, body?: unknown) => execute('put', endpoint, body), [execute])
  const del = useCallback((endpoint: string) => execute('delete', endpoint), [execute])

  return { ...state, get, post, put, del, loadMore, retry }
}