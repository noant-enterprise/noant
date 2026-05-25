import { cn } from '@/lib/utils'
import { type HTMLAttributes } from 'react'

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: 'sky' | 'success' | 'warning' | 'error' | 'neutral'
}

export function Badge({ className, variant = 'neutral', ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center px-3 py-1 rounded-full text-xs font-medium tracking-wide',
        {
          'bg-sky-50 text-noant-sky-deep border border-sky-100': variant === 'sky',
          'bg-emerald-50 text-emerald-700 border border-emerald-100': variant === 'success',
          'bg-amber-50 text-amber-700 border border-amber-100': variant === 'warning',
          'bg-red-50 text-red-700 border border-red-100': variant === 'error',
          'bg-slate-100 text-slate-600 border border-slate-200': variant === 'neutral',
        },
        className
      )}
      {...props}
    />
  )
}