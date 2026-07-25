import { useLiveFeed } from '@/lib/hooks/useLiveFeed'
import { timeAgo } from '@/lib/utils'
import { UserPlus, CreditCard, Brain, AlertTriangle, Wifi, Server } from 'lucide-react'

const ICON_MAP = {
  signup: UserPlus,
  payment: CreditCard,
  ai_failure: Brain,
  escalation: AlertTriangle,
  whatsapp_issue: Wifi,
  system: Server,
}

const SEVERITY_MAP = {
  high: 'text-danger',
  medium: 'text-warning',
  low: 'text-text-tertiary',
}

const DOT_MAP = {
  high: 'bg-danger',
  medium: 'bg-warning',
  low: 'bg-success',
}

export function LiveFeed() {
  const { events } = useLiveFeed()

  return (
    <div className="space-y-1">
      {events.map(event => {
        const Icon = ICON_MAP[event.type]
        return (
          <div key={event.id} className="flex items-start gap-3 rounded-lg px-3 py-2.5 transition-colors hover:bg-bg-inset">
            <div className="relative mt-0.5">
              <div className={`h-2 w-2 rounded-full ${DOT_MAP[event.severity]}`} />
            </div>
            <Icon className={`mt-0.5 h-4 w-4 shrink-0 ${SEVERITY_MAP[event.severity]}`} />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-text-primary">{event.title}</p>
              <p className="truncate text-xs text-text-tertiary">{event.description}</p>
            </div>
            <span className="shrink-0 text-xs text-text-tertiary">{timeAgo(event.timestamp)}</span>
          </div>
        )
      })}
    </div>
  )
}
