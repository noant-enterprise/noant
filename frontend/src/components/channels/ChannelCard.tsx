import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { ChannelIcon } from './ChannelIcon'
import { cn } from '@/lib/utils'
import { Link2, Clock3 } from 'lucide-react'

interface ChannelCardProps {
  channel: string
  name: string
  desc: string
  status: 'connected' | 'active' | 'disconnected' | 'error'
  details?: Array<{ label: string; value: string }>
  webhookUrl?: string
  connectedAt?: string
  onConnect: () => void
  onDisconnect: () => void
}

export function ChannelCard({
  channel,
  name,
  desc,
  status,
  details,
  webhookUrl,
  connectedAt,
  onConnect,
  onDisconnect,
}: ChannelCardProps) {
  const isConnected = status === 'connected' || status === 'active'
  const statusLabel = isConnected ? 'connected' : status

  return (
    <div className={cn(
      'bg-surface border rounded-2xl p-4 lg:p-5 transition-all duration-200',
      isConnected
        ? 'border-noant-sky/30 shadow-sm'
        : 'border-default'
    )}>
      <div className="flex items-start gap-3 mb-4">
        <ChannelIcon channel={channel} size="md" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold text-sm text-primary truncate">{name}</h3>
            {isConnected && (
              <Badge variant="success" className="text-[9px] px-1.5 py-0">
                {statusLabel}
              </Badge>
            )}
          </div>
          <p className="text-[11px] text-secondary mt-0.5">{desc}</p>
        </div>
      </div>

      {isConnected && details && details.length > 0 && (
        <div className="space-y-2 mb-4">
          {details.slice(0, 3).map((detail) => (
            <div
              key={detail.label}
              className="flex items-center gap-2 rounded-lg border border-default bg-inset px-3 py-2 text-xs"
            >
              <span className="font-medium text-tertiary shrink-0">{detail.label}:</span>
              <span className="text-secondary truncate font-mono">{detail.value}</span>
            </div>
          ))}
          {webhookUrl && (
            <div className="flex items-center gap-2 rounded-lg border border-default bg-inset px-3 py-2 text-xs text-tertiary truncate">
              <Link2 className="w-3 h-3 shrink-0" />
              <span className="truncate">{webhookUrl}</span>
            </div>
          )}
          {connectedAt && (
            <div className="flex items-center gap-2 text-[10px] text-tertiary">
              <Clock3 className="w-3 h-3" />
              Connected {new Date(connectedAt).toLocaleDateString()}
            </div>
          )}
        </div>
      )}

      <div className="flex justify-start">
        {isConnected ? (
          <Button
            variant="ghost"
            className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold border border-default w-full sm:w-fit sm:max-w-[200px] text-red-600 hover:border-red-200 hover:text-red-700"
            onClick={onDisconnect}
          >
            Disconnect
          </Button>
        ) : (
          <Button
            variant="accent"
            className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold w-full sm:w-fit sm:max-w-[200px]"
            onClick={onConnect}
          >
            Connect
          </Button>
        )}
      </div>
    </div>
  )
}