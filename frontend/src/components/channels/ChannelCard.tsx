import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { TokenDisplay } from './TokenDisplay'
import { ChannelIcon } from './ChannelIcon'
import { cn } from '@/lib/utils'

interface ChannelCardProps {
  channel: string
  name: string
  desc: string
  status: 'connected' | 'disconnected' | 'error'
  token?: string
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
  token,
  webhookUrl,
  connectedAt,
  onConnect,
  onDisconnect,
}: ChannelCardProps) {
  const isConnected = status === 'connected'

  return (
    <div className={cn(
      'bg-surface border border-default rounded-2xl p-4 lg:p-5 transition-all duration-200',
      isConnected ? 'border-noant-sky/30 shadow-sm' : 'hover:border-noant-sky/20'
    )}>
      {/* Header */}
      <div className="flex items-start gap-3 mb-4">
        <ChannelIcon channel={channel} size="md" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold text-sm text-primary truncate">{name}</h3>
            <Badge 
              variant={isConnected ? 'success' : status === 'error' ? 'error' : 'neutral'} 
              className={cn("text-[9px] px-1.5 py-0 transition-all", isConnected && "animate-pulse")}
            >
              {status}
            </Badge>
          </div>
          <p className="text-[11px] text-secondary mt-0.5">{desc}</p>
        </div>
      </div>

      {/* Token + Webhook */}
      {isConnected && (
        <div className="space-y-3 mb-4">
          <TokenDisplay token={token} label="API Token" />
          {webhookUrl && (
            <div className="space-y-1.5">
              <label className="text-[10px] font-semibold uppercase tracking-wider text-tertiary">Webhook URL</label>
              <div className="bg-inset rounded-lg px-3 py-2 font-mono text-[10px] text-secondary truncate">
                {webhookUrl}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Meta */}
      {connectedAt && (
        <p className="text-[10px] text-tertiary mb-3">
          Connected {new Date(connectedAt).toLocaleDateString()}
        </p>
      )}

      {/* Action */}
      <div className="flex justify-start">
        <Button
          variant={isConnected ? 'ghost' : 'accent'}
          className={cn(
            'inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-colors w-full sm:w-fit sm:max-w-[200px]',
            isConnected && 'border border-default'
          )}
          onClick={isConnected ? onDisconnect : onConnect}
        >
          {isConnected ? 'Disconnect' : 'Connect'}
        </Button>
      </div>
    </div>
  )
}
