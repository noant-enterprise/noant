import { AlertTriangle, Trash2, LogOut, CheckCircle2 } from 'lucide-react'
import { Modal } from './Modal'
import { Button } from './Button'
import { cn } from '@/lib/utils'

type ConfirmVariant = 'danger' | 'warning' | 'success' | 'neutral'

const variantConfig: Record<ConfirmVariant, { icon: React.ElementType; color: string; btnVariant: 'accent' | 'primary' | 'ghost' }> = {
  danger: { icon: Trash2, color: 'text-red-500', btnVariant: 'accent' },
  warning: { icon: AlertTriangle, color: 'text-amber-500', btnVariant: 'primary' },
  success: { icon: CheckCircle2, color: 'text-emerald-500', btnVariant: 'accent' },
  neutral: { icon: LogOut, color: 'text-secondary', btnVariant: 'primary' },
}

interface ConfirmModalProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  description?: string
  confirmText?: string
  cancelText?: string
  variant?: ConfirmVariant
  loading?: boolean
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
  loading,
}: ConfirmModalProps) {
  const config = variantConfig[variant]
  const Icon = config.icon

  return (
    <Modal open={open} onClose={onClose} title="" description="" size="sm" hideClose>
      <div className="flex flex-col items-center text-center py-6 px-2">
        {/* Icon at top */}
        <div className={cn('w-12 h-12 rounded-full bg-inset flex items-center justify-center mb-4', config.color)}>
          <Icon className="w-6 h-6" strokeWidth={1.5} />
        </div>

        {/* Title */}
        <h3 className="text-base font-semibold text-primary mb-1">{title}</h3>

        {/* Description */}
        {description && (
          <p className="text-sm text-secondary mb-6 max-w-[260px]">{description}</p>
        )}

        {/* Buttons */}
        <div className="flex gap-3 w-full">
          <Button variant="ghost" className="flex-1" onClick={onClose} disabled={loading}>
            {cancelText}
          </Button>
          <Button 
            variant={config.btnVariant} 
            className="flex-1" 
            onClick={onConfirm}
            loading={loading}
          >
            {confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  )
}