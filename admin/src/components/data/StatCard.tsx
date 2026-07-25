import { cn } from '@/lib/utils'
import { TrendingUp, TrendingDown } from 'lucide-react'

interface StatCardProps {
  label: string
  value: string | number
  change?: number
  changeLabel?: string
  icon?: React.ReactNode
  className?: string
}

export function StatCard({ label, value, change, changeLabel, icon, className }: StatCardProps) {
  return (
    <div className={cn('rounded-xl border border-border bg-bg-surface p-5 transition-colors hover:border-border/80', className)}>
      <div className="flex items-start justify-between">
        <div className="space-y-2">
          <p className="text-sm font-medium text-text-tertiary">{label}</p>
          <p className="text-2xl font-bold text-text-primary">{value}</p>
          {change !== undefined && (
            <div className="flex items-center gap-1">
              {change >= 0 ? (
                <TrendingUp className="h-3 w-3 text-success" />
              ) : (
                <TrendingDown className="h-3 w-3 text-danger" />
              )}
              <span className={`text-xs font-medium ${change >= 0 ? 'text-success' : 'text-danger'}`}>
                {change >= 0 ? '+' : ''}{change.toFixed(1)}%
              </span>
              {changeLabel && <span className="text-xs text-text-tertiary">{changeLabel}</span>}
            </div>
          )}
        </div>
        {icon && <div className="rounded-lg bg-bg-inset p-2 text-text-tertiary">{icon}</div>}
      </div>
    </div>
  )
}
