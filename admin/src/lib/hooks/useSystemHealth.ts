import { useState, useEffect } from 'react'
import { adminApi } from '@/lib/api'
import type { SystemHealthResponse, SystemHealth, ServiceStatus } from '@/types'

export function useSystemHealth() {
  const [data, setData] = useState<SystemHealth>({
    api: { name: 'API Server', status: 'healthy', latency_ms: 0, last_check: new Date().toISOString(), uptime: 100 },
    database: { name: 'TiDB', status: 'healthy', latency_ms: 0, last_check: new Date().toISOString(), uptime: 100 },
    redis: { name: 'Redis', status: 'healthy', latency_ms: 0, last_check: new Date().toISOString(), uptime: 100 },
    whatsapp: { name: 'WhatsApp (OpenWA)', status: 'healthy', latency_ms: 0, last_check: new Date().toISOString(), uptime: 100 },
    error_rate: 0,
    p50_latency: 0,
    p95_latency: 0,
    p99_latency: 0,
    active_websockets: 0,
    job_queue_depth: 0,
  })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    adminApi.getSystemHealth()
      .then((res: SystemHealthResponse) => {
        const mapService = (name: string, s?: { name: string; status: string; latency_ms: number }): ServiceStatus => ({
          name: s?.name || name,
          status: (s?.status || 'healthy') as ServiceStatus['status'],
          latency_ms: s?.latency_ms || 0,
          last_check: new Date().toISOString(),
          uptime: s?.status === 'down' ? 0 : 99.99,
        })

        const svcMap: Record<string, { name: string; key: keyof Pick<SystemHealth, 'api' | 'database' | 'redis' | 'whatsapp'> }> = {
          'API Server': { name: 'API Server', key: 'api' },
          'Database': { name: 'TiDB', key: 'database' },
          'Redis': { name: 'Redis', key: 'redis' },
          'WhatsApp': { name: 'WhatsApp (OpenWA)', key: 'whatsapp' },
        }

        const updated: Partial<SystemHealth> = {
          error_rate: res.error_rate,
          p50_latency: res.p50_latency,
          p95_latency: res.p95_latency,
          p99_latency: res.p99_latency,
        }

        for (const svc of res.services) {
          const mapping = svcMap[svc.name]
          if (mapping) {
            updated[mapping.key] = mapService(mapping.name, svc)
          }
        }

        setData(prev => ({ ...prev, ...updated }))
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return { data, loading }
}
