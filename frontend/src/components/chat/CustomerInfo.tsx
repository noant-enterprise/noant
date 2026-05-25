import { cn } from '@/lib/utils'
import { X, ArrowLeft } from 'lucide-react'
import { Avatar } from '@/components/ui/Avatar'
import type { Conversation } from '@/types'

interface CustomerInfoProps {
  conversation: Conversation | null
  open: boolean
  onClose: () => void
}

export function CustomerInfo({ conversation, open, onClose }: CustomerInfoProps) {
  if (!open) return null

  return (
    <>
      {/* Desktop sidebar */}
      <div className="hidden lg:flex flex-col bg-surface border-l border-default w-[240px] shrink-0 overflow-hidden animate-fade-in">
        <div className="w-[240px] flex-1 flex flex-col">
          {!conversation ? (
            <div className="p-4 flex-1 flex items-center justify-center">
              <p className="text-xs text-tertiary text-center">Select a conversation to see details</p>
            </div>
          ) : (
            <>
              <div className="flex items-center justify-between p-4 border-b border-default">
                <span className="text-[10px] font-semibold uppercase tracking-widest text-tertiary">
                  Customer
                </span>
                <button
                  onClick={onClose}
                  className="p-1 rounded-md hover:bg-inset text-tertiary hover:text-primary transition-colors"
                  aria-label="Close customer info"
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              </div>

              <div className="p-4 flex-1 overflow-y-auto">
                <div className="flex flex-col items-center gap-3 pb-4 border-b border-default">
                  <Avatar name={conversation.customer_name} size="xl" />
                  <div className="text-center">
                    <p className="font-semibold text-sm text-primary">{conversation.customer_name}</p>
                    <p className="text-[11px] text-tertiary capitalize">{conversation.channel}</p>
                  </div>
                </div>

                <div className="space-y-1 pt-4">
                  <InfoRow label="Channel" value={conversation.channel} />
                  <InfoRow label="Status" value={conversation.status} />
                  <InfoRow label="Intent" value={conversation.intent || 'Unknown'} />
                  <InfoRow label="Priority" value={conversation.priority} />
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Mobile bottom sheet */}
      <div className="lg:hidden fixed inset-0 z-50 flex flex-col justify-end">
        <div 
          className="absolute inset-0 bg-black/40 backdrop-blur-sm animate-fade-in" 
          onClick={onClose}
        />
        <div className="relative bg-surface rounded-t-3xl animate-slide-up overflow-hidden max-h-[85vh] flex flex-col">
          <div className="flex items-center justify-center pt-3 pb-1">
            <div className="w-10 h-1 bg-border-default rounded-full bg-default" />
          </div>
          
          {!conversation ? (
            <div className="p-8 flex items-center justify-center">
              <p className="text-sm text-tertiary text-center">Select a conversation to see details</p>
            </div>
          ) : (
            <>
              <div className="flex items-center justify-between px-4 py-3 border-b border-default">
                <button
                  onClick={onClose}
                  className="w-8 h-8 rounded-full bg-inset flex items-center justify-center text-secondary active:scale-95 transition-all"
                >
                  <ArrowLeft className="w-4 h-4" />
                </button>
                <span className="text-sm font-semibold text-primary">Customer Info</span>
                <div className="w-8" />
              </div>

              <div className="p-6 overflow-y-auto">
                <div className="flex flex-col items-center gap-4 mb-6">
                  <Avatar name={conversation.customer_name} size="xl" />
                  <div className="text-center">
                    <p className="font-bold text-lg text-primary">{conversation.customer_name}</p>
                    <p className="text-sm text-tertiary capitalize">{conversation.channel}</p>
                  </div>
                </div>

                <div className="space-y-0 rounded-2xl bg-inset overflow-hidden">
                  <MobileInfoRow label="Channel" value={conversation.channel} />
                  <MobileInfoRow label="Status" value={conversation.status} />
                  <MobileInfoRow label="Intent" value={conversation.intent || 'Unknown'} />
                  <MobileInfoRow label="Priority" value={conversation.priority} isLast />
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between items-center py-2.5 border-b border-subtle px-1">
      <span className="text-[11px] font-medium text-tertiary uppercase tracking-wider">{label}</span>
      <span className={cn('text-sm font-medium capitalize', value === 'high' ? 'text-red-500' : 'text-primary')}>
        {value}
      </span>
    </div>
  )
}

function MobileInfoRow({ label, value, isLast }: { label: string; value: string; isLast?: boolean }) {
  return (
    <div className={cn(
      'flex justify-between items-center py-4 px-4',
      !isLast && 'border-b border-default'
    )}>
      <span className="text-sm font-medium text-secondary">{label}</span>
      <span className={cn('text-sm font-semibold capitalize', value === 'high' ? 'text-red-500' : 'text-primary')}>
        {value}
      </span>
    </div>
  )
}
