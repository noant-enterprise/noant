import { useState, useRef, useEffect, type FormEvent } from 'react'
import { Send, User, Smile } from 'lucide-react'
import EmojiPicker, { Theme } from 'emoji-picker-react'
import { cn } from '@/lib/utils'
import { useOffline } from '@/hooks/useOffline'

interface ChatInputProps {
  onSend: (message: string) => void
  onTakeover: () => void
  disabled?: boolean
}

export function ChatInput({ onSend, onTakeover, disabled }: ChatInputProps) {
  const [text, setText] = useState('')
  const [showEmoji, setShowEmoji] = useState(false)
  const emojiRef = useRef<HTMLDivElement>(null)
  const isOffline = useOffline()
  const isInputDisabled = disabled || isOffline

  const isDark = typeof document !== 'undefined' && document.documentElement.classList.contains('dark')

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (emojiRef.current && !emojiRef.current.contains(e.target as Node)) {
        setShowEmoji(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!text.trim() || isOffline) return
    onSend(text.trim())
    setText('')
  }

  return (
    <div className="border-t border-default bg-surface shrink-0">
      <form onSubmit={handleSubmit} className="pt-3 px-3 flex gap-2 items-end">
        <input
          type="text"
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={isOffline ? "You are currently offline..." : "Message..."}
          disabled={isInputDisabled}
          className={cn(
            'flex-1 px-4 py-3 lg:py-2.5 bg-inset border border-default rounded-full text-sm lg:text-xs text-primary placeholder:text-tertiary',
            'focus:outline-none focus:border-noant-sky/50 focus:ring-1 focus:ring-noant-sky/20',
            'transition-all duration-200',
            isOffline ? 'border-red-500/20 bg-red-500/5 focus:border-red-500/30 placeholder:text-red-500/40' : '',
            'disabled:opacity-50'
          )}
          style={{ fontFamily: "'Apple Color Emoji', 'Segoe UI Emoji', 'Noto Color Emoji', 'Twemoji Mozilla', sans-serif" }}
        />
        
        <div className="relative shrink-0" ref={emojiRef}>
          <button
            type="button"
            onClick={() => setShowEmoji(!showEmoji)}
            disabled={isInputDisabled}
            className={cn(
              'w-11 h-11 lg:w-9 lg:h-9 rounded-full shrink-0',
              'flex items-center justify-center transition-all active:scale-90',
              showEmoji
                ? 'bg-noant-sky/10 text-noant-sky'
                : 'text-tertiary hover:text-secondary hover:bg-inset',
              'disabled:opacity-40'
            )}
          >
            <Smile className="w-5 h-5 lg:w-4 lg:h-4" strokeWidth={1.5} />
          </button>
          {showEmoji && (
            <div className="absolute bottom-full right-0 mb-2 z-50" style={{
              '--epr-highlight-color': '#0ea5e9',
              '--epr-category-icon-active-color': '#0ea5e9',
              '--epr-search-border-color': '#0ea5e9',
              '--epr-picker-border-radius': '12px',
            } as React.CSSProperties}>
              <EmojiPicker
                onEmojiClick={(emojiData) => {
                  setText(prev => prev + emojiData.emoji)
                }}
                theme={isDark ? Theme.DARK : Theme.LIGHT}
                width={300}
                height={350}
              />
            </div>
          )}
        </div>

        <button
          type="submit"
          disabled={isInputDisabled || !text.trim()}
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
          disabled={isInputDisabled}
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
