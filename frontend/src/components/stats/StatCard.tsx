import { cn } from '@/lib/utils'

interface StatCardProps {
  label: string
  value: string | number
  change?: number
  variant?: 'default' | 'success' | 'warning' | 'info' | 'error'
}

const variantStyles = {
  default: 'text-primary',
  success: 'text-emerald-600 dark:text-emerald-400',
  warning: 'text-amber-600 dark:text-amber-400',
  info: 'text-noant-sky',
  error: 'text-red-600 dark:text-red-400',
}

export function StatCard({ label, value, change, variant = 'default' }: StatCardProps) {
  const isPositive = change && change > 0
  const isNegative = change && change < 0

  return (
    <div className="rounded-xl border border-default bg-surface p-4 lg:p-5 hover:border-noant-sky/50 transition-colors duration-300">
      <p className="text-xs lg:text-sm text-secondary mb-1">{label}</p>
      <div className="flex items-end justify-between">
        <p className={cn('text-xl lg:text-2xl font-bold', variantStyles[variant])}>{value}</p>
        {change !== undefined && (
          <span className={cn(
            'text-xs font-semibold px-1.5 py-0.5 rounded',
            isPositive ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' :
            isNegative ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' :
            'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
          )}>
            {isPositive ? '+' : ''}{change}%
          </span>
        )}
      </div>
    </div>
  )
}
