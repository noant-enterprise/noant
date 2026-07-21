import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  ReactNode,
} from 'react'
import { createPortal } from 'react-dom'
import { Trash2, AlertTriangle, X } from 'lucide-react'

// ─── Types ────────────────────────────────────────────────────────────────────

export interface ModalOptions {
  title: string
  body: string | ReactNode
  confirmText?: string
  cancelText?: string
  /** 'danger' = red+trash; 'warning' = amber; 'primary' = sky-blue */
  variant?: 'primary' | 'danger' | 'warning'
  /** When true, user must type confirmPhrase before confirm button enables */
  requireTypeConfirm?: boolean
  /** Phrase user must type — defaults to "RESET" */
  confirmPhrase?: string
  /** Async-safe confirm handler */
  onConfirm: () => void | Promise<void>
  /** Called after cancel */
  onCancel?: () => void
  /** Clicking overlay closes modal (default: true unless requireTypeConfirm) */
  closeOnOverlayClick?: boolean
}

interface ModalContextValue {
  showModal: (options: ModalOptions) => void
  closeModal: () => void
  isOpen: boolean
}

// ─── Context ──────────────────────────────────────────────────────────────────

const ModalContext = createContext<ModalContextValue>({
  showModal: () => {},
  closeModal: () => {},
  isOpen: false,
})

// ─── Provider ─────────────────────────────────────────────────────────────────

export function ModalProvider({ children }: { children: ReactNode }) {
  const [modalOptions, setModalOptions] = useState<ModalOptions | null>(null)
  const [isConfirming, setIsConfirming] = useState(false)
  const triggerRef = useRef<Element | null>(null)

  const showModal = useCallback((options: ModalOptions) => {
    // Remember which element triggered the modal so we can return focus on close
    triggerRef.current = document.activeElement
    setModalOptions(options)
    setIsConfirming(false)
  }, [])

  const closeModal = useCallback(() => {
    setModalOptions(null)
    setIsConfirming(false)
    // Return focus to trigger
    setTimeout(() => {
      if (triggerRef.current && triggerRef.current instanceof HTMLElement) {
        triggerRef.current.focus()
      }
    }, 0)
  }, [])

  const handleConfirm = useCallback(async () => {
    if (!modalOptions) return
    setIsConfirming(true)
    try {
      await modalOptions.onConfirm()
    } finally {
      setIsConfirming(false)
      setModalOptions(null)
      setTimeout(() => {
        if (triggerRef.current && triggerRef.current instanceof HTMLElement) {
          triggerRef.current.focus()
        }
      }, 0)
    }
  }, [modalOptions])

  const handleCancel = useCallback(() => {
    const onCancel = modalOptions?.onCancel
    setModalOptions(null)
    setIsConfirming(false)
    onCancel?.()
    setTimeout(() => {
      if (triggerRef.current && triggerRef.current instanceof HTMLElement) {
        triggerRef.current.focus()
      }
    }, 0)
  }, [modalOptions])

  return (
    <ModalContext.Provider value={{ showModal, closeModal, isOpen: modalOptions !== null }}>
      {children}
      {modalOptions && (
        <ConfirmModalRenderer
          options={modalOptions}
          isConfirming={isConfirming}
          onConfirm={handleConfirm}
          onCancel={handleCancel}
        />
      )}
    </ModalContext.Provider>
  )
}

export function useModalContext() {
  return useContext(ModalContext)
}

// ─── Internal Modal Renderer ──────────────────────────────────────────────────

interface ConfirmModalRendererProps {
  options: ModalOptions
  isConfirming: boolean
  onConfirm: () => void
  onCancel: () => void
}

function ConfirmModalRenderer({
  options,
  isConfirming,
  onConfirm,
  onCancel,
}: ConfirmModalRendererProps) {
  const {
    title,
    body,
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    variant = 'primary',
    requireTypeConfirm = false,
    confirmPhrase = 'RESET',
    closeOnOverlayClick,
  } = options

  const resolvedCloseOnOverlay = closeOnOverlayClick ?? !requireTypeConfirm

  const [typeValue, setTypeValue] = useState('')
  const overlayRef = useRef<HTMLDivElement>(null)
  const cancelBtnRef = useRef<HTMLButtonElement>(null)

  const typeMatch = !requireTypeConfirm || typeValue === confirmPhrase

  // Escape key + focus trap
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !isConfirming) {
        onCancel()
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
        const first = focusables[0]!
        const last = focusables[focusables.length - 1]!
        if (e.shiftKey) {
          if (document.activeElement === first) {
            e.preventDefault()
            last.focus()
          }
        } else {
          if (document.activeElement === last) {
            e.preventDefault()
            first.focus()
          }
        }
      }
    }

    document.addEventListener('keydown', handleKey)
    document.body.style.overflow = 'hidden'

    // Auto-focus cancel button on mount (safer default for destructive modals)
    const t = setTimeout(() => cancelBtnRef.current?.focus(), 50)

    return () => {
      document.removeEventListener('keydown', handleKey)
      document.body.style.overflow = ''
      clearTimeout(t)
    }
  }, [isConfirming, onCancel])

  // Button class helpers
  const secondaryBtnCls =
    'bg-surface border border-default hover:bg-surface-hover text-primary px-4 py-2 rounded-lg text-sm font-medium transition-colors'

  const confirmBtnCls =
    variant === 'danger'
      ? 'bg-red-500 hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2'
      : variant === 'warning'
      ? 'bg-amber-500 hover:bg-amber-600 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors'
      : 'bg-[#3b82f6] hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors'

  return createPortal(
    <>
      {/* Overlay */}
      <div
        ref={overlayRef}
        className="fixed inset-0 z-[15000] flex items-center justify-center overflow-y-auto p-4 sm:p-6 bg-overlay backdrop-blur-sm"
        style={{ animation: 'noantOverlayIn 200ms ease forwards' }}
        onClick={(e) => {
          if (resolvedCloseOnOverlay && e.target === overlayRef.current) {
            onCancel()
          }
        }}
      >
        {/* Modal */}
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="noant-modal-title"
          aria-describedby="noant-modal-body"
          className="bg-surface rounded-2xl shadow-2xl max-w-md w-full mx-auto max-h-[calc(100dvh-2rem)] sm:max-h-[calc(100dvh-3rem)] p-6 relative overflow-hidden flex flex-col"
          style={{ animation: 'noantModalIn 200ms ease forwards' }}
        >
          {/* Close button */}
          <button
            onClick={onCancel}
            disabled={isConfirming}
            className="absolute top-4 right-4 w-7 h-7 rounded-lg bg-inset hover:bg-surface-hover flex items-center justify-center text-secondary hover:text-primary transition-colors disabled:opacity-40"
            aria-label="Close dialog"
          >
            <X className="w-4 h-4" />
          </button>

          {/* Variant icon */}
          {variant === 'danger' && (
            <div className="w-11 h-11 rounded-full bg-red-50 dark:bg-red-950 flex items-center justify-center mb-4">
              <Trash2 className="w-5 h-5 text-red-500" strokeWidth={1.5} />
            </div>
          )}
          {variant === 'warning' && (
            <div className="w-11 h-11 rounded-full bg-amber-50 dark:bg-amber-950 flex items-center justify-center mb-4">
              <AlertTriangle className="w-5 h-5 text-amber-500" strokeWidth={1.5} />
            </div>
          )}

          {/* Title */}
          <h2
            id="noant-modal-title"
            className="text-lg font-semibold text-primary pr-8"
          >
            {title}
          </h2>

          {/* Body */}
          <div id="noant-modal-body" className="text-sm text-secondary mt-2 overflow-y-auto min-h-0">
            {typeof body === 'string' ? <p>{body}</p> : body}
          </div>

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

          {/* Action buttons */}
          <div className="flex justify-end gap-3 mt-6 shrink-0">
            <button
              ref={cancelBtnRef}
              onClick={onCancel}
              disabled={isConfirming}
              className={secondaryBtnCls}
            >
              {cancelText}
            </button>
            <button
              onClick={onConfirm}
              disabled={isConfirming || !typeMatch}
              className={confirmBtnCls}
            >
              {variant === 'danger' && !isConfirming && (
                <Trash2 className="w-4 h-4" />
              )}
              {isConfirming ? 'Please wait…' : confirmText}
            </button>
          </div>
        </div>
      </div>

      {/* Keyframe animations */}
      <style>{`
        @keyframes noantOverlayIn {
          from { opacity: 0; }
          to   { opacity: 1; }
        }
        @keyframes noantModalIn {
          from { opacity: 0; transform: scale(0.95); }
          to   { opacity: 1; transform: scale(1); }
        }
      `}</style>
    </>,
    document.body
  )
}
