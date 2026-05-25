import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  description?: string
  children?: React.ReactNode
  size?: 'sm' | 'md' | 'lg'
  hideClose?: boolean
}

export function Modal({ open, onClose, title, description, children, size = 'md', hideClose }: ModalProps) {
  const overlayRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    if (open) {
      document.addEventListener('keydown', handleEscape)
      document.body.style.overflow = 'hidden'
    }
    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = ''
    }
  }, [open, onClose])

  if (!open) return null

  const sizeClasses = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
  }

  return createPortal(
    <div
      ref={overlayRef}
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-fade-in"
      onClick={(e) => {
        if (e.target === overlayRef.current) onClose()
      }}
    >
      <div className={cn(
        'w-full bg-surface border border-default rounded-2xl shadow-2xl animate-toast-in overflow-hidden',
        sizeClasses[size]
      )}>
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-default">
          <div>
            <h3 className="text-base font-semibold text-primary">{title}</h3>
            {description && (
              <p className="text-sm text-secondary mt-0.5">{description}</p>
            )}
          </div>
          {!hideClose && (
            <button
              onClick={onClose}
              className="w-8 h-8 rounded-lg bg-inset flex items-center justify-center text-secondary hover:text-primary active:scale-95 transition-all"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Body */}
        <div className="px-5 py-4">
          {children}
        </div>
      </div>
    </div>,
    document.body
  )
}