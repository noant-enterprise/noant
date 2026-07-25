import { Search, Bell, LogOut } from 'lucide-react'
import { useAuth } from '@/lib/hooks/useAuth'
import { useNavigate } from 'react-router-dom'
import { useLiveFeed } from '@/lib/hooks/useLiveFeed'

export function Topbar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const { alerts } = useLiveFeed()
  const unacknowledged = alerts.filter(a => !a.acknowledged).length

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-border bg-bg-surface/80 px-6 backdrop-blur-md">
      <div className="flex items-center gap-3">
        <button
          className="flex items-center gap-2 rounded-lg border border-border bg-bg-inset px-3 py-1.5 text-sm text-text-tertiary transition-colors hover:text-text-secondary"
          onClick={() => {
            const input = document.getElementById('command-search') as HTMLInputElement | null
            input?.focus()
          }}
        >
          <Search className="h-4 w-4" />
          <span>Search...</span>
          <kbd className="ml-4 rounded border border-border bg-bg-surface px-1.5 py-0.5 text-xs text-text-tertiary">⌘K</kbd>
        </button>
      </div>

      <div className="flex items-center gap-4">
        <button className="relative rounded-lg p-2 text-text-secondary transition-colors hover:bg-bg-inset hover:text-text-primary">
          <Bell className="h-4 w-4" />
          {unacknowledged > 0 && (
            <span className="absolute -right-0.5 -top-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-danger text-[10px] font-bold text-white">
              {unacknowledged}
            </span>
          )}
        </button>

        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-brand-sky/20 text-xs font-bold text-brand-sky">
            {user?.email?.[0]?.toUpperCase() || 'A'}
          </div>
          <span className="text-sm text-text-secondary">{user?.email}</span>
        </div>

        <button
          onClick={handleLogout}
          className="rounded-lg p-2 text-text-tertiary transition-colors hover:bg-bg-inset hover:text-danger"
          title="Logout"
        >
          <LogOut className="h-4 w-4" />
        </button>
      </div>
    </header>
  )
}
