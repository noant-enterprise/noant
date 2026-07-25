import { useState, useEffect } from 'react'
import type { SystemHealth } from '@/types'

const MOCK_SYSTEM: SystemHealth = {
  api: { name: 'API Server', status: 'healthy', latency_ms: 32, last_check: new Date().toISOString(), uptime: 99.98 },
  database: { name: 'TiDB', status: 'healthy', latency_ms: 12, last_check: new Date().toISOString(), uptime: 99.99 },
  redis: { name: 'Redis', status: 'healthy', latency_ms: 1, last_check: new Date().toISOString(), uptime: 100 },
  whatsapp: { name: 'WhatsApp (OpenWA)', status: 'healthy', latency_ms: 145, last_check: new Date().toISOString(), uptime: 99.85 },
  error_rate: 0.12,
  p50_latency: 32,
  p95_latency: 89,
  p99_latency: 234,
  active_websockets: 47,
  job_queue_depth: 3,
}

export function useSystemHealth() {
  const [data, setData] = useState<SystemHealth>(MOCK_SYSTEM)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setData(MOCK_SYSTEM)
      setLoading(false)
    }, 300)
    return () => clearTimeout(timer)
  }, [])

  return { data, loading }
}
