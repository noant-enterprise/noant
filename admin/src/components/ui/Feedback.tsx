import { AlertCircle, RefreshCw } from 'lucide-react'

export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
}: {
  icon?: React.ComponentType<{ className?: string }>
  title: string
  description?: string
  action?: { label: string; onClick: () => void }
}) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      {Icon && <Icon className="mb-3 h-10 w-10 text-text-tertiary" />}
      <p className="text-sm font-medium text-text-secondary">{title}</p>
      {description && <p className="mt-1 text-xs text-text-tertiary">{description}</p>}
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 rounded-lg bg-brand-sky/10 px-4 py-2 text-sm font-medium text-brand-sky hover:bg-brand-sky/20 transition-colors"
        >
          {action.label}
        </button>
      )}
    </div>
  )
}

export function ErrorBanner({
  message,
  onRetry,
}: {
  message: string
  onRetry?: () => void
}) {
  return (
    <div className="flex items-center gap-3 rounded-lg border border-red-500/20 bg-red-500/5 px-4 py-3 text-sm">
      <AlertCircle className="h-4 w-4 shrink-0 text-red-400" />
      <span className="flex-1 text-text-secondary">{message}</span>
      {onRetry && (
        <button
          onClick={onRetry}
          className="flex items-center gap-1.5 rounded-md bg-red-500/10 px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-500/20 transition-colors"
        >
          <RefreshCw className="h-3 w-3" />
          Retry
        </button>
      )}
    </div>
  )
}
