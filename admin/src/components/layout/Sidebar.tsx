import { NavLink } from 'react-router-dom'
import { LayoutDashboard, Users, BarChart3, DollarSign, Brain, Server, BookOpen, ClipboardList, Settings, Target } from 'lucide-react'

function NoantLogo({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 100 100" className={className} xmlns="http://www.w3.org/2000/svg">
      <circle cx="50" cy="50" r="42" stroke="currentColor" strokeWidth="5" strokeDasharray="8 10" strokeLinecap="round" fill="none" />
      <circle cx="50" cy="50" r="28" fill="currentColor" />
      <circle cx="40" cy="50" r="3" fill="white" />
      <circle cx="50" cy="50" r="4" fill="white" />
      <circle cx="60" cy="50" r="5" fill="white" />
    </svg>
  )
}

const NAV_ITEMS = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/pipeline', icon: Target, label: 'Sales Pipeline' },
  { to: '/customers', icon: Users, label: 'Customers' },
  { to: '/analytics', icon: BarChart3, label: 'Analytics' },
  { to: '/revenue', icon: DollarSign, label: 'Revenue' },
  { to: '/ai', icon: Brain, label: 'AI Health' },
  { to: '/knowledge', icon: BookOpen, label: 'Knowledge Base' },
  { to: '/system', icon: Server, label: 'System' },
  { divider: true },
  { to: '/audit-logs', icon: ClipboardList, label: 'Audit Logs' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

export function Sidebar() {
  return (
    <aside className="fixed left-0 top-0 z-40 flex h-screen w-60 flex-col border-r border-border bg-bg-surface">
      <div className="flex h-14 items-center gap-2 border-b border-border px-4">
        <NoantLogo className="h-7 w-7 text-brand-sky shrink-0" />
        <span className="text-lg font-bold text-text-primary">NOANT</span>
        <span className="ml-auto rounded bg-brand-sky/20 px-1.5 py-0.5 text-xs font-medium text-brand-sky">Admin</span>
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        {NAV_ITEMS.map((item, i) => {
          if ('divider' in item && item.divider) {
            return <div key={i} className="my-2 border-t border-border" />
          }
          const nav = item as { to: string; icon: React.ComponentType<{ className?: string }>; label: string }
          return (
            <NavLink
              key={nav.to}
              to={nav.to}
              end={nav.to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-brand-sky/10 text-brand-sky'
                    : 'text-text-secondary hover:bg-bg-inset hover:text-text-primary'
                }`
              }
            >
              <nav.icon className="h-4 w-4" />
              {nav.label}
            </NavLink>
          )
        })}
      </nav>

      <div className="border-t border-border px-4 py-3">
        <p className="text-xs text-text-tertiary">NOANT Command Center</p>
        <p className="text-xs text-text-tertiary">v0.1.0</p>
      </div>
    </aside>
  )
}
