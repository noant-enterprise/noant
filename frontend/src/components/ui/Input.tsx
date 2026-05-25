import { cn } from '@/lib/utils'
import { type InputHTMLAttributes, forwardRef } from 'react'

const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        'w-full bg-surface border border-default rounded-lg',
        'px-4 py-3 text-base text-primary placeholder:text-tertiary',
        'focus:outline-none focus:border-noant-sky focus:ring-2 focus:ring-noant-sky/20',
        'transition-all duration-200',
        className
      )}
      {...props}
    />
  )
)

Input.displayName = 'Input'

export { Input }