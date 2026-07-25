import { useSystemHealth } from '@/lib/hooks/useSystemHealth'
import { Activity, Database, Wifi, Server, Clock, Zap } from 'lucide-react'

const STATUS_STYLE = {
  healthy: { dot: 'bg-success', text: 'text-success', label: 'Healthy' },
  degraded: { dot: 'bg-warning', text: 'text-warning', label: 'Degraded' },
  down: { dot: 'bg-danger', text: 'text-danger', label: 'Down' },
}

const SERVICE_ICONS: Record<string, typeof Server> = {
  'API Server': Zap,
  TiDB: Database,
  Redis: Activity,
  'WhatsApp (OpenWA)': Wifi,
}

export default function SystemPage() {
  const { data } = useSystemHealth()

  const services = [data.api, data.database, data.redis, data.whatsapp]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">System Health</h1>
        <p className="text-sm text-text-tertiary">All services operational</p>
      </div>

      <div className="grid grid-cols-4 gap-4">
        {services.map(s => {
          const style = STATUS_STYLE[s.status]
          const Icon = SERVICE_ICONS[s.name] || Server
          return (
            <div key={s.name} className="rounded-xl border border-border bg-bg-surface p-5">
              <div className="flex items-center justify-between">
                <Icon className="h-5 w-5 text-text-tertiary" />
                <div className="flex items-center gap-1.5">
                  <span className={`h-2 w-2 rounded-full ${style.dot}`} />
                  <span className={`text-xs font-medium ${style.text}`}>{style.label}</span>
                </div>
              </div>
              <p className="mt-3 text-sm font-medium text-text-primary">{s.name}</p>
              <div className="mt-2 space-y-1 text-xs text-text-tertiary">
                <p>Latency: <span className="text-text-secondary">{s.latency_ms}ms</span></p>
                <p>Uptime: <span className="text-text-secondary">{s.uptime}%</span></p>
              </div>
            </div>
          )
        })}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Performance Metrics</h3>
          <div className="space-y-3">
            {[
              { label: 'p50 Latency', value: `${data.p50_latency}ms`, icon: Clock },
              { label: 'p95 Latency', value: `${data.p95_latency}ms`, icon: Clock },
              { label: 'p99 Latency', value: `${data.p99_latency}ms`, icon: Clock },
              { label: 'Error Rate', value: `${data.error_rate}%`, icon: Activity },
            ].map(m => (
              <div key={m.label} className="flex items-center justify-between rounded-lg bg-bg-inset px-3 py-2">
                <div className="flex items-center gap-2 text-sm text-text-secondary">
                  <m.icon className="h-3.5 w-3.5 text-text-tertiary" />
                  {m.label}
                </div>
                <span className="text-sm font-medium text-text-primary">{m.value}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Connections</h3>
          <div className="space-y-3">
            {[
              { label: 'Active WebSockets', value: data.active_websockets },
              { label: 'Job Queue Depth', value: data.job_queue_depth },
            ].map(m => (
              <div key={m.label} className="flex items-center justify-between rounded-lg bg-bg-inset px-3 py-2">
                <span className="text-sm text-text-secondary">{m.label}</span>
                <span className="text-sm font-medium text-text-primary">{m.value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
