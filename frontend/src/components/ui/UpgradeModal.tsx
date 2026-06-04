import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { Crown, Check, X, ArrowRight } from 'lucide-react'
import { Button } from './Button'

interface UpgradeModalProps {
  open: boolean
  onClose: () => void
  title?: string
  description?: string
  featureList?: string[]
}

const defaultFeatures = [
  'Unlimited AI responses',
  'All communication channels (WhatsApp, Telegram, Gmail, Web)',
  'Unlimited team members',
  'Unlimited inventory items',
  'Priority support & white-label widget',
]

export function UpgradeModal({
  open,
  onClose,
  title = 'Upgrade to unlock premium features',
  description = 'You have reached the limits of the Free plan. Upgrade to a paid plan to continue scaling your customer service.',
  featureList = defaultFeatures,
}: UpgradeModalProps) {
  const overlayRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      }
    }
    document.addEventListener('keydown', handleKey)
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', handleKey)
      document.body.style.overflow = ''
    }
  }, [open, onClose])

  if (!open) return null

  const handleGoToBilling = () => {
    window.location.href = '/billing'
  }

  return createPortal(
    <div
      ref={overlayRef}
      className="fixed inset-0 z-[15000] flex items-center justify-center overflow-y-auto p-4 sm:p-6 bg-overlay backdrop-blur-sm"
      style={{ animation: 'noantOverlayIn 200ms ease forwards' }}
      onClick={(e) => {
        if (e.target === overlayRef.current) onClose()
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        className="bg-surface border border-default rounded-2xl shadow-2xl max-w-md w-full mx-auto max-h-[calc(100dvh-2rem)] sm:max-h-[calc(100dvh-3rem)] p-6 relative overflow-hidden flex flex-col"
        style={{ animation: 'noantModalIn 200ms ease forwards' }}
      >
        {/* Glow decoration */}
        <div className="absolute -top-24 -right-24 w-48 h-48 rounded-full bg-gradient-to-br from-noant-sky/20 to-purple-500/20 blur-2xl pointer-events-none" />

        {/* Close Button */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 w-8 h-8 rounded-xl bg-inset hover:bg-surface-hover flex items-center justify-center text-secondary hover:text-primary transition-all"
          aria-label="Close"
        >
          <X className="w-4 h-4" />
        </button>

        {/* Header Icon */}
        <div className="w-12 h-12 rounded-2xl bg-gradient-to-tr from-noant-sky to-purple-500 flex items-center justify-center mb-4 shadow-lg shadow-sky/20">
          <Crown className="w-6 h-6 text-white" strokeWidth={2} />
        </div>

        {/* Title & Description */}
        <h3 className="text-lg font-bold text-primary pr-8 leading-snug">
          {title}
        </h3>
        <p className="text-xs sm:text-sm text-secondary mt-2 leading-relaxed">
          {description}
        </p>

        {/* Premium Features List */}
        <div className="mt-5 p-4 rounded-xl bg-inset/50 border border-default overflow-y-auto min-h-0">
          <p className="text-xs font-bold text-primary mb-3 uppercase tracking-wide">Included in Pro plan:</p>
          <ul className="space-y-2.5">
            {featureList.map((feature, i) => (
              <li key={i} className="flex items-start gap-2.5 text-xs text-secondary leading-normal">
                <div className="w-4 h-4 rounded-full bg-noant-sky/10 flex items-center justify-center shrink-0 mt-0.5">
                  <Check className="w-3 h-3 text-noant-sky" strokeWidth={3} />
                </div>
                <span>{feature}</span>
              </li>
            ))}
          </ul>
        </div>

        {/* Actions */}
        <div className="flex gap-3 mt-6">
          <Button
            variant="ghost"
            onClick={onClose}
            className="flex-1 text-xs"
          >
            Maybe Later
          </Button>
          <Button
            onClick={handleGoToBilling}
            className="flex-1 text-xs bg-noant-sky text-white hover:bg-noant-sky-deep shadow-lg shadow-sky/15 flex items-center justify-center gap-1.5"
          >
            Upgrade Now <ArrowRight className="w-3 h-3" />
          </Button>
        </div>
      </div>

      <style>{`
        @keyframes noantOverlayIn { from { opacity:0 } to { opacity:1 } }
        @keyframes noantModalIn { from { opacity:0; transform:scale(.95) } to { opacity:1; transform:scale(1) } }
      `}</style>
    </div>,
    document.body
  )
}
