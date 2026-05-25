/**
 * ConfirmModal — component-level confirm modal (uses local state).
 *
 * For app-wide confirms triggered imperatively from anywhere, use:
 *   const confirm = useConfirm()  →  confirm({ title, body, variant, onConfirm })
 *
 * This component is kept for backward-compat (Sidebar, TeachPage, etc.)
 * and upgraded with full accessibility + proper brand styling.
 */
import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { Trash2, AlertTriangle, LogOut, CheckCircle2, X } from 'lucide-react'
import { cn } from '@/lib/utils'

export type ConfirmVariant = 'danger' | 'warning' | 'success' | 'neutral'

interface ConfirmModalProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void | Promise<void>
  title: string
  description?: string
  confirmText?: string
  cancelText?: string
  variant?: ConfirmVariant
  loading?: boolean
  /**
   * When set, the confirm button is disabled until the user types this phrase.
   * Use for destructive / irreversible actions.
   */
  requireTypeConfirm?: boolean
  confirmPhrase?: string
  /** Close when overlay is clicked (default true; false recommended for type-to-confirm) */
  closeOnOverlayClick?: boolean
}

const variantMeta: Record<
  ConfirmVariant,
  { Icon: React.ElementType; iconBg: string; iconColor: string; btnCls: string }
> = {
  danger: {
    Icon: Trash2,
    iconBg: 'bg-red-50 dark:bg-red-950',
    iconColor: 'text-red-500',
    btnCls:
      'bg-red-500 hover:bg-red-600 text-white disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2',
  },
  warning: {
    Icon: AlertTriangle,
    iconBg: 'bg-amber-50 dark:bg-amber-950',
    iconColor: 'text-amber-500',
    btnCls: 'bg-amber-500 hover:bg-amber-600 text-white disabled:opacity-50 disabled:cursor-not-allowed',
  },
  success: {
    Icon: CheckCircle2,
    iconBg: 'bg-emerald-50 dark:bg-emerald-950',
    iconColor: 'text-emerald-500',
    btnCls: 'bg-emerald-500 hover:bg-emerald-600 text-white disabled:opacity-50 disabled:cursor-not-allowed',
  },
  neutral: {
    Icon: LogOut,
    iconBg: 'bg-inset',
    iconColor: 'text-secondary',
    btnCls:
      'bg-[#3b82f6] hover:bg-blue-700 text-white disabled:opacity-50 disabled:cursor-not-allowed',
  },
}

export function ConfirmModal({
  open,
  onClose,
  onConfirm,
  title,
  description,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  variant = 'neutral',
  loading = false,
  requireTypeConfirm = false,
  confirmPhrase = 'RESET',
  closeOnOverlayClick,
}: ConfirmModalProps) {
  const [typeValue, setTypeValue] = useState('')
  const overlayRef = useRef<HTMLDivElement>(null)
  const cancelBtnRef = useRef<HTMLButtonElement>(null)

  const resolvedCloseOnOverlay = closeOnOverlayClick ?? !requireTypeConfirm
  const typeMatch = !requireTypeConfirm || typeValue === confirmPhrase

  const meta = variantMeta[variant]
  const Icon = meta.Icon

  // Escape key + focus trap
  useEffect(() => {
    if (!open) return
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !loading) {
        onClose()
        return
      }
      if (e.key === 'Tab') {
        const dialog = overlayRef.current?.querySelector<HTMLElement>('[role="dialog"]')
        if (!dialog) return
        const focusables = Array.from(
          dialog.querySelectorAll<HTMLElement>(
            'button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex="-1"])'
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
    document.addEventListener('keydown', handleKey)
    document.body.style.overflow = 'hidden'
    const t = setTimeout(() => cancelBtnRef.current?.focus(), 50)
    return () => {
      document.removeEventListener('keydown', handleKey)
      document.body.style.overflow = ''
      clearTimeout(t)
    }
  }, [open, loading, onClose])

  // Reset type value when modal closes
  useEffect(() => {
    if (!open) setTypeValue('')
  }, [open])

  if (!open) return null

  const secondaryCls =
    'bg-surface border border-default hover:bg-surface-hover text-primary px-4 py-2 rounded-lg text-sm font-medium transition-colors'
  const confirmCls = cn(
    'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
    meta.btnCls
  )

  return createPortal(
    <div
      ref={overlayRef}
      className="fixed inset-0 z-[10000] flex items-center justify-center p-4 bg-overlay backdrop-blur-sm"
      style={{ animation: 'noantOverlayIn 200ms ease forwards' }}
      onClick={(e) => {
        if (resolvedCloseOnOverlay && e.target === overlayRef.current) onClose()
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="confirm-modal-title"
        aria-describedby={description ? 'confirm-modal-desc' : undefined}
        className="bg-surface rounded-xl shadow-2xl max-w-md w-full mx-4 p-6 relative"
        style={{ animation: 'noantModalIn 200ms ease forwards' }}
      >
        {/* Close X */}
        <button
          onClick={onClose}
          disabled={loading}
          className="absolute top-4 right-4 w-7 h-7 rounded-lg bg-inset hover:bg-surface-hover flex items-center justify-center text-secondary hover:text-primary transition-colors disabled:opacity-40"
          aria-label="Close"
        >
          <X className="w-4 h-4" />
        </button>

        {/* Icon */}
        <div className={cn('w-11 h-11 rounded-full flex items-center justify-center mb-4', meta.iconBg)}>
          <Icon className={cn('w-5 h-5', meta.iconColor)} strokeWidth={1.5} />
        </div>

        {/* Title */}
        <h2 id="confirm-modal-title" className="text-lg font-semibold text-primary pr-8">
          {title}
        </h2>

        {/* Description */}
        {description && (
          <p id="confirm-modal-desc" className="text-sm text-secondary mt-2">
            {description}
          </p>
        )}

        {/* Type-to-confirm */}
        {requireTypeConfirm && (
          <div className="mt-4">
            <p className="text-xs text-tertiary mb-1.5">
              Type{' '}
              <span className="font-mono font-semibold text-primary bg-inset px-1 rounded">
                {confirmPhrase}
              </span>{' '}
              to confirm:
            </p>
            <input
              type="text"
              value={typeValue}
              onChange={(e) => setTypeValue(e.target.value)}
              placeholder={confirmPhrase}
              className="w-full px-3 py-2 text-sm border border-default bg-surface text-primary rounded-lg outline-none focus:border-red-400 focus:ring-1 focus:ring-red-200 transition-colors font-mono placeholder:text-tertiary"
              autoComplete="off"
              spellCheck={false}
              aria-label={`Type ${confirmPhrase} to confirm`}
            />
          </div>
        )}

        {/* Buttons */}
        <div className="flex justify-end gap-3 mt-6">
          <button
            ref={cancelBtnRef}
            onClick={onClose}
            disabled={loading}
            className={secondaryCls}
          >
            {cancelText}
          </button>
          <button
            onClick={() => { void onConfirm() }}
            disabled={loading || !typeMatch}
            className={confirmCls}
          >
            {variant === 'danger' && !loading && <Trash2 className="w-4 h-4" />}
            {loading ? 'Please wait…' : confirmText}
          </button>
        </div>
      </div>

      {/* Animation keyframes (idempotent — deduplicated by browser) */}
      <style>{`
        @keyframes noantOverlayIn { from { opacity:0 } to { opacity:1 } }
        @keyframes noantModalIn { from { opacity:0; transform:scale(.95) } to { opacity:1; transform:scale(1) } }
      `}</style>
    </div>,
    document.body
  )
}