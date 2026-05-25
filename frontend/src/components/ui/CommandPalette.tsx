import { useState, useEffect, useCallback, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Search, LayoutGrid, MessageSquare, GraduationCap, BarChart3, Link2, Settings, LogOut } from 'lucide-react'
import { cn } from '@/lib/utils'

interface CommandItem {
  id: string
  label: string
  icon: React.ReactNode
  shortcut?: string
  action: () => void
}

export function CommandPalette() {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const navigate = useNavigate()

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setOpen(prev => !prev)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  useEffect(() => {
    if (open) {
      setSearch('')
      setSelectedIndex(0)
      setTimeout(() => inputRef.current?.focus(), 50)
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [open])

  const items: CommandItem[] = [
    { id: 'overview', label: 'Go to Overview', icon: <LayoutGrid className="w-4 h-4" />, shortcut: 'G O', action: () => { navigate('/'); setOpen(false) } },
    { id: 'chats', label: 'Go to Conversations', icon: <MessageSquare className="w-4 h-4" />, shortcut: 'G C', action: () => { navigate('/chats'); setOpen(false) } },
    { id: 'teach', label: 'Go to Teach your Noant', icon: <GraduationCap className="w-4 h-4" />, shortcut: 'G T', action: () => { navigate('/teach'); setOpen(false) } },
    { id: 'insights', label: 'Go to Insights', icon: <BarChart3 className="w-4 h-4" />, shortcut: 'G I', action: () => { navigate('/insights'); setOpen(false) } },
    { id: 'channels', label: 'Go to Your channels', icon: <Link2 className="w-4 h-4" />, shortcut: 'G H', action: () => { navigate('/channels'); setOpen(false) } },
    { id: 'setup', label: 'Go to Your setup', icon: <Settings className="w-4 h-4" />, shortcut: 'G S', action: () => { navigate('/setup'); setOpen(false) } },
    { id: 'logout', label: 'Sign out', icon: <LogOut className="w-4 h-4" />, action: () => { localStorage.clear(); window.location.href = '/login' } },
  ]

  const filtered = items.filter(i => i.label.toLowerCase().includes(search.toLowerCase()))

  useEffect(() => {
    setSelectedIndex(0)
  }, [search])

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      e.preventDefault()
      setOpen(false)
      return
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex(i => (i + 1) % filtered.length)
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex(i => (i - 1 + filtered.length) % filtered.length)
    }
    if (e.key === 'Enter' && filtered[selectedIndex]) {
      e.preventDefault()
      filtered[selectedIndex].action()
    }
    if (e.key === 'Tab') {
      e.preventDefault()
      if (e.shiftKey) {
        setSelectedIndex(i => (i - 1 + filtered.length) % filtered.length)
      } else {
        setSelectedIndex(i => (i + 1) % filtered.length)
      }
    }
  }, [filtered, selectedIndex])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-[100] flex items-start justify-center pt-[20vh]"
      onClick={() => setOpen(false)}
      role="dialog"
      aria-modal="true"
      aria-label="Command palette"
    >
      <div className="absolute inset-0 bg-overlay backdrop-blur-sm" aria-hidden="true" />
      <div
        className="relative w-full max-w-xl bg-surface rounded-xl shadow-2xl border border-default overflow-hidden animate-slide-up"
        onClick={e => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        <div className="flex items-center gap-3 px-4 py-3 border-b border-default">
          <Search className="w-5 h-5 text-tertiary" aria-hidden="true" />
          <input
            ref={inputRef}
            type="text"
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="Search commands..."
            className="flex-1 bg-transparent text-primary placeholder:text-tertiary outline-none text-sm"
            aria-label="Search commands"
            role="combobox"
            aria-expanded={filtered.length > 0}
            aria-controls="cmd-list"
            aria-activedescendant={filtered[selectedIndex]?.id}
          />
          <kbd className="px-2 py-1 text-xs font-mono bg-inset rounded text-tertiary hidden sm:inline-block">ESC</kbd>
        </div>
        <div
          id="cmd-list"
          role="listbox"
          className="max-h-[60vh] overflow-y-auto py-2"
        >
          {filtered.length === 0 ? (
            <div className="px-4 py-8 text-center text-secondary text-sm">No commands found</div>
          ) : (
            filtered.map((item, i) => (
              <button
                key={item.id}
                id={item.id}
                role="option"
                aria-selected={i === selectedIndex}
                onClick={item.action}
                onMouseEnter={() => setSelectedIndex(i)}
                className={cn(
                  'w-full flex items-center gap-3 px-4 py-2.5 text-left transition-colors',
                  i === selectedIndex ? 'bg-inset text-primary' : 'text-primary hover:bg-inset'
                )}
              >
                <span className="text-secondary" aria-hidden="true">{item.icon}</span>
                <span className="flex-1 text-sm">{item.label}</span>
                {item.shortcut && (
                  <kbd className="px-1.5 py-0.5 text-[10px] font-mono bg-inset rounded text-tertiary">{item.shortcut}</kbd>
                )}
              </button>
            ))
          )}
        </div>
        <div className="px-4 py-2 border-t border-default text-xs text-tertiary flex items-center gap-4">
          <span className="flex items-center gap-1"><kbd className="px-1 bg-inset rounded">↑↓</kbd> to navigate</span>
          <span className="flex items-center gap-1"><kbd className="px-1 bg-inset rounded">↵</kbd> to select</span>
          <span className="flex items-center gap-1"><kbd className="px-1 bg-inset rounded">esc</kbd> to close</span>
        </div>
      </div>
    </div>
  )
}