import { useState, type FormEvent } from 'react'
import { Send, User } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ChatInputProps {
  onSend: (message: string) => void
  onTakeover: () => void
  disabled?: boolean
  typing?: boolean
  typingText?: string
}

export function ChatInput({ onSend, onTakeover, disabled, typing, typingText }: ChatInputProps) {
  const [text, setText] = useState('')

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!text.trim()) return
    onSend(text.trim())
    setText('')
  }

  return (
    <div className="border-t border-default bg-surface shrink-0">
      {typing && (
        <div className="px-4 py-2 text-[11px] lg:text-[10px] text-tertiary flex items-center gap-2 animate-fade-in">
          <span className="flex gap-1">
            <span className="w-1 h-1 bg-noant-sky rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
            <span className="w-1 h-1 bg-noant-sky rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
            <span className="w-1 h-1 bg-noant-sky rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
          </span>
          {typingText || 'Customer is typing...'}
        </div>
      )}
      
      <form onSubmit={handleSubmit} className="p-3 lg:p-3 flex gap-2 items-end pb-[max(0.75rem,env(safe-area-inset-bottom))]">
        <input
          type="text"
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Message..."
          disabled={disabled}
          className={cn(
            'flex-1 px-4 py-3 lg:py-2.5 bg-inset border border-default rounded-full text-sm lg:text-xs text-primary placeholder:text-tertiary',
            'focus:outline-none focus:border-noant-sky/50 focus:ring-1 focus:ring-noant-sky/20',
            'transition-all duration-200',
            'disabled:opacity-50'
          )}
        />
        
        <button
          type="submit"
          disabled={disabled || !text.trim()}
          className={cn(
            'w-11 h-11 lg:w-9 lg:h-9 rounded-full bg-noant-sky text-white shrink-0',
            'hover:bg-noant-sky-deep disabled:opacity-40 disabled:cursor-not-allowed',
            'flex items-center justify-center transition-all active:scale-90 shadow-sm'
          )}
        >
          <Send className="w-5 h-5 lg:w-4 lg:h-4 ml-0.5" strokeWidth={2.5} />
        </button>
        
        <button
          type="button"
          onClick={onTakeover}
          disabled={disabled}
          className={cn(
            'hidden lg:flex px-4 py-2.5 border border-default text-secondary rounded-full font-medium text-xs',
            'hover:border-noant-sky hover:text-noant-sky-deep disabled:opacity-50',
            'items-center gap-1.5 transition-all active:scale-95 shrink-0'
          )}
        >
          <User className="w-3.5 h-3.5" />
          Take over
        </button>
      </form>
    </div>
  )
}
