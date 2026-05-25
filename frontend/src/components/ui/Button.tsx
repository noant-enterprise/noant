import { cn } from '@/lib/utils'
import { type ButtonHTMLAttributes, forwardRef } from 'react'
import { Spinner } from './Spinner'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'accent' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', loading, children, disabled, ...props }, ref) => {
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={cn(
          'inline-flex items-center justify-center gap-2 font-semibold transition-all duration-200',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          'active:scale-[0.98]',
          {
            'bg-noant-black text-white hover:bg-noant-ink dark:bg-white dark:text-noant-black dark:hover:bg-gray-100':
              variant === 'primary',
            'bg-noant-sky text-white hover:bg-noant-sky-deep':
              variant === 'accent',
            'bg-transparent text-secondary border border-default hover:border-noant-sky hover:text-noant-sky-deep':
              variant === 'ghost',
            'px-3 py-2 text-xs': size === 'sm',
            'px-5 py-3 text-sm': size === 'md',
            'px-6 py-4 text-base': size === 'lg',
          },
          'rounded-lg',
          className
        )}
        {...props}
      >
        {loading && <Spinner size="sm" className={variant === 'primary' ? 'text-white' : 'text-current'} />}
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'

export { Button }