import { cn } from '@/lib/utils'

interface CategoryCardProps {
  category: {
    id: string
    name: string
    color: string
    qa_count: number
    created_at: string
  }
  onClick?: () => void
}

export function CategoryCard({ category, onClick }: CategoryCardProps) {
  return (
    <div
      onClick={onClick}
      className={cn(
        'bg-surface border border-default rounded-xl p-4 lg:p-5 cursor-pointer transition-all duration-200',
        'hover:border-noant-sky/40 hover:shadow-sm active:scale-[0.98]'
      )}
    >
      <div className="flex items-center gap-3 mb-3">
        <div className="w-3 h-3 rounded-full shrink-0" style={{ background: category.color }} />
        <span className="font-semibold text-sm text-primary truncate">{category.name}</span>
      </div>
      <div className="flex gap-4 text-xs text-secondary">
        <span>{category.qa_count} Q&A pairs</span>
        <span className="text-tertiary">{new Date(category.created_at).toLocaleDateString()}</span>
      </div>
    </div>
  )
}
