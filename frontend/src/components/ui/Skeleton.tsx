import { cn } from '@/lib/utils'

interface SkeletonProps {
  className?: string
}

export function Skeleton({ className }: SkeletonProps) {
  return (
    <div className={cn('animate-shimmer-slow rounded', className)} />
  )
}

export function StatSkeleton() {
  return (
    <div className="rounded-xl border border-default bg-surface p-4 lg:p-5 space-y-2">
      <div className="animate-shimmer-slow h-3 w-20 rounded" />
      <div className="animate-shimmer-slow h-8 w-24 rounded" />
    </div>
  )
}
