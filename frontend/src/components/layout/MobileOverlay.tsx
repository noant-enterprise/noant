import { cn } from '@/lib/utils'

export function MobileOverlay({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <div
      onClick={onClose}
        className={cn(
          'fixed inset-0 bg-overlay z-40 lg:hidden transition-opacity duration-300',
          open ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'
        )}
    />
  )
}