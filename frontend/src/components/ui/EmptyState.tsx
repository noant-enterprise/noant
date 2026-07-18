import type { ReactNode } from 'react'
import { Inbox } from 'lucide-react'
import { cn } from '@/lib/utils'

interface EmptyStateProps {
  icon?: ReactNode
  title: string
  description?: string
  action?: ReactNode
  className?: string
}

export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-16 px-4 text-center', className)}>
      <div className="w-14 h-14 rounded-2xl bg-inset flex items-center justify-center mb-4">
        {icon || <Inbox className="w-6 h-6 text-tertiary" />}
      </div>
      <h3 className="font-semibold text-primary mb-1">{title}</h3>
      {description && <p className="text-sm text-secondary max-w-sm mb-4">{description}</p>}
      {action && <div>{action}</div>}
    </div>
  )
}
