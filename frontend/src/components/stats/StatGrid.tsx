import { type ReactNode } from 'react'

interface StatGridProps {
  children: ReactNode
}

export function StatGrid({ children }: StatGridProps) {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4 lg:gap-5 mb-6 sm:mb-8">
      {children}
    </div>
  )
}