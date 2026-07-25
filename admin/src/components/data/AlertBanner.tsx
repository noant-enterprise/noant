import { useLiveFeed } from '@/lib/hooks/useLiveFeed'
import { AlertTriangle, AlertCircle, Info, X } from 'lucide-react'

const ICON_MAP = {
  critical: AlertCircle,
  warning: AlertTriangle,
  info: Info,
}

const STYLE_MAP = {
  critical: 'border-danger/30 bg-danger/5',
  warning: 'border-warning/30 bg-warning/5',
  info: 'border-brand-sky/30 bg-brand-sky/5',
}

const ICON_STYLE_MAP = {
  critical: 'text-danger',
  warning: 'text-warning',
  info: 'text-brand-sky',
}

export function AlertBanner() {
  const { alerts, acknowledgeAlert } = useLiveFeed()
  const active = alerts.filter(a => !a.acknowledged)

  if (active.length === 0) return null

  return (
    <div className="space-y-2">
      {active.map(alert => {
        const Icon = ICON_MAP[alert.type]
        return (
          <div key={alert.id} className={`flex items-center gap-3 rounded-lg border p-3 ${STYLE_MAP[alert.type]}`}>
            <Icon className={`h-4 w-4 shrink-0 ${ICON_STYLE_MAP[alert.type]}`} />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-text-primary">{alert.title}</p>
              <p className="text-xs text-text-tertiary">{alert.message}</p>
            </div>
            <button
              onClick={() => acknowledgeAlert(alert.id)}
              className="shrink-0 rounded p-1 text-text-tertiary transition-colors hover:text-text-primary"
            >
              <X className="h-3 w-3" />
            </button>
          </div>
        )
      })}
    </div>
  )
}
