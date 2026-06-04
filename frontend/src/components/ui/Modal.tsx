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
  const closeButtonRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (!open) return

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') { onClose(); return }

      // Focus trap
      if (e.key === 'Tab') {
        const dialog = overlayRef.current?.querySelector<HTMLElement>('[role="dialog"]')
        if (!dialog) return
        const focusables = Array.from(
          dialog.querySelectorAll<HTMLElement>(
            'button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), a[href], [tabindex]:not([tabindex="-1"])'
          )
        )
        if (focusables.length === 0) return
        const first = focusables[0]
        const last = focusables[focusables.length - 1]
        if (e.shiftKey) {
          if (document.activeElement === first) { e.preventDefault(); last.focus() }
        } else {
          if (document.activeElement === last) { e.preventDefault(); first.focus() }
        }
      }
    }

    document.addEventListener('keydown', handleEscape)
    document.body.style.overflow = 'hidden'

    // Auto-focus the first input on open
    const t = setTimeout(() => {
      if (overlayRef.current) {
        const firstInput = overlayRef.current.querySelector<HTMLElement>(
          'input:not([disabled]), textarea:not([disabled])'
        )
        if (firstInput) {
          firstInput.focus()
        } else {
          const firstFocusable = overlayRef.current.querySelector<HTMLElement>(
            'button:not([disabled]), [tabindex]:not([tabindex="-1"])'
          )
          firstFocusable?.focus()
        }
      }
    }, 50)

    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = ''
      clearTimeout(t)
    }
  }, [open, onClose, hideClose])

  if (!open) return null

  const sizeClasses = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
  }

  return createPortal(
    <div
      ref={overlayRef}
      className="fixed inset-0 z-[100] flex items-center justify-center overflow-y-auto p-4 sm:p-6 bg-overlay backdrop-blur-sm animate-fade-in"
      onClick={(e) => {
        if (e.target === overlayRef.current) onClose()
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? 'modal-primitive-title' : undefined}
        aria-describedby={description ? 'modal-primitive-desc' : undefined}
        className={cn(
          'w-full max-h-[calc(100dvh-2rem)] sm:max-h-[calc(100dvh-3rem)] bg-surface border border-default rounded-2xl shadow-2xl animate-toast-in overflow-hidden flex flex-col min-w-0',
          sizeClasses[size]
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-default shrink-0">
          <div>
            {title && (
              <h3 id="modal-primitive-title" className="text-base font-semibold text-primary">
                {title}
              </h3>
            )}
            {description && (
              <p id="modal-primitive-desc" className="text-sm text-secondary mt-0.5">
                {description}
              </p>
            )}
          </div>
          {!hideClose && (
            <button
              ref={closeButtonRef}
              onClick={onClose}
              className="w-8 h-8 rounded-lg bg-inset flex items-center justify-center text-secondary hover:text-primary active:scale-95 transition-all"
              aria-label="Close"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Body */}
        <div className="px-5 py-4 overflow-y-auto min-h-0">
          {children}
        </div>
      </div>
    </div>,
    document.body
  )
}
