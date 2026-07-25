import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { Command } from 'cmdk'
import { LayoutDashboard, Users, BarChart3, DollarSign, Brain, Server, X } from 'lucide-react'

const NAVIGATION = [
  { path: '/', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/customers', label: 'Customers', icon: Users },
  { path: '/analytics', label: 'Analytics', icon: BarChart3 },
  { path: '/revenue', label: 'Revenue', icon: DollarSign },
  { path: '/ai', label: 'AI Health', icon: Brain },
  { path: '/system', label: 'System', icon: Server },
]

export function CommandPalette() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen(prev => !prev)
      }
    }
    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [])

  const handleSelect = useCallback((path: string) => {
    navigate(path)
    setOpen(false)
  }, [navigate])

  return (
    <>
      <Command.Dialog
        open={open}
        onOpenChange={setOpen}
        className="fixed left-1/2 top-1/2 z-[100] w-full max-w-lg -translate-x-1/2 -translate-y-1/2 overflow-hidden rounded-xl border border-border bg-bg-surface shadow-2xl"
      >
        <div className="flex items-center border-b border-border px-4">
          <Command.Input
            placeholder="Search pages, actions..."
            className="h-12 flex-1 bg-transparent text-sm text-text-primary outline-none placeholder:text-text-tertiary"
          />
          <button onClick={() => setOpen(false)} className="ml-2 rounded p-1 text-text-tertiary hover:text-text-primary">
            <X className="h-4 w-4" />
          </button>
        </div>
        <Command.List className="max-h-64 overflow-auto p-2">
          <Command.Empty className="py-6 text-center text-sm text-text-tertiary">No results found.</Command.Empty>
          <Command.Group heading="Navigation" className="text-xs text-text-tertiary">
            {NAVIGATION.map(item => (
              <Command.Item
                key={item.path}
                value={item.label}
                onSelect={() => handleSelect(item.path)}
                className="flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 text-sm text-text-secondary transition-colors hover:bg-bg-inset data-[selected=true]:bg-bg-inset data-[selected=true]:text-text-primary"
              >
                <item.icon className="h-4 w-4" />
                {item.label}
              </Command.Item>
            ))}
          </Command.Group>
          <Command.Separator className="my-2 h-px bg-border" />
          <Command.Group heading="Quick Actions" className="text-xs text-text-tertiary">
            <Command.Item
              value="add api key"
              onSelect={() => { navigate('/system'); setOpen(false) }}
              className="flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 text-sm text-text-secondary transition-colors hover:bg-bg-inset data-[selected=true]:bg-bg-inset data-[selected=true]:text-text-primary"
            >
              Manage API Keys
            </Command.Item>
            <Command.Item
              value="view failed payments"
              onSelect={() => { navigate('/revenue'); setOpen(false) }}
              className="flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 text-sm text-text-secondary transition-colors hover:bg-bg-inset data-[selected=true]:bg-bg-inset data-[selected=true]:text-text-primary"
            >
              View Failed Payments
            </Command.Item>
            <Command.Item
              value="knowledge gaps"
              onSelect={() => { navigate('/ai'); setOpen(false) }}
              className="flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 text-sm text-text-secondary transition-colors hover:bg-bg-inset data-[selected=true]:bg-bg-inset data-[selected=true]:text-text-primary"
            >
              AI Knowledge Gaps
            </Command.Item>
          </Command.Group>
        </Command.List>
      </Command.Dialog>
    </>
  )
}
