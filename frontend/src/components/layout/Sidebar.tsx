import { useState } from 'react'
import { cn } from '@/lib/utils'
import { NavLink, useLocation } from 'react-router-dom'
import {
  LayoutGrid, MessageSquare, GraduationCap, BarChart3, Link2,
  Settings, LogOut, X, CreditCard, Users, Code2
} from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { useSidebarAlerts } from '@/contexts/SidebarAlertsContext'

// ─── Alert Dot ────────────────────────────────────────────────────────────────

function AlertDot({ count, pulse = false }: { count?: number; pulse?: boolean }) {
  if (!count && !pulse) return null
  return (
    <span
      className={cn(
        'absolute -top-1 -right-1 min-w-[16px] h-4 bg-red-500 text-white text-[9px] font-bold rounded-full flex items-center justify-center px-1 border-2 border-surface leading-none shadow-sm',
        pulse && 'animate-pulse'
      )}
    >
      {count && count > 0 ? (count > 99 ? '99+' : count) : ''}
    </span>
  )
}

// ─── Nav Config ───────────────────────────────────────────────────────────────

const navSections = [
  {
    title: 'Your workspace',
    items: [
      { to: '/', icon: LayoutGrid, label: 'Overview', alertKey: null as null | string },
      { to: '/chats', icon: MessageSquare, label: 'Conversations', alertKey: 'unreadChats' },
      { to: '/insights', icon: BarChart3, label: 'Insights', alertKey: null },
    ],
  },
  {
    title: 'Build your Noant',
    items: [
      { to: '/teach', icon: GraduationCap, label: 'Teach your Noant', alertKey: 'unknownQuestions' },
      { to: '/channels', icon: Link2, label: 'Your channels', alertKey: 'channelIssues' },
      { to: '/widget', icon: Code2, label: 'Web Widget', alertKey: null },
    ],
  },
  {
    title: 'Manage',
    items: [
      { to: '/team', icon: Users, label: 'Team', alertKey: null },
      { to: '/billing', icon: CreditCard, label: 'Billing', alertKey: 'billingAlert' },
      { to: '/settings', icon: Settings, label: 'Settings', alertKey: null },
    ],
  },
]

// ─── Sidebar ──────────────────────────────────────────────────────────────────

interface SidebarProps {
  onClose?: () => void
  collapsed?: boolean
  onToggleCollapse?: () => void
}

export function Sidebar({ onClose, collapsed = false, onToggleCollapse }: SidebarProps) {
  const { user, signOut } = useAuth()
  const location = useLocation()
  const alerts = useSidebarAlerts()
  const [isHovered, setIsHovered] = useState(false)
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false)

  const isExpanded = collapsed ? isHovered : true

  const initials = user
    ? `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
    : '--'

  const name = user ? `${user.first_name} ${user.last_name}` : 'Loading...'

  // Helper: get badge count for a route
  const getAlertCount = (alertKey: string | null): number => {
    if (!alertKey) return 0
    if (alertKey === 'billingAlert') return alerts.billingAlert ? 1 : 0
    return (alerts as any)[alertKey] ?? 0
  }

  return (
    <div
      className={cn(
        'flex flex-col h-full bg-surface border-r border-default transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] relative shrink-0',
        isExpanded ? 'w-[220px]' : 'w-[68px]'
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Toggle button */}
      {!onClose && onToggleCollapse && (
        <button
          onClick={(e) => { e.stopPropagation(); onToggleCollapse() }}
          className={cn(
            'absolute right-0 top-[60px] w-5 h-12 bg-noant-sky text-white flex items-center justify-center rounded-r-lg shadow-lg hover:bg-noant-sky-deep transition-all duration-200 z-[100]',
            isExpanded ? 'translate-x-full opacity-100' : 'translate-x-0 opacity-0 hover:opacity-100'
          )}
          title={collapsed ? 'Pin open' : 'Collapse'}
        >
          <span className="text-[10px] font-bold">{collapsed ? '›' : '‹'}</span>
        </button>
      )}

      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-default h-[57px] shrink-0">
        <div className={cn(
          'flex items-center gap-2.5 transition-all duration-300',
          isExpanded ? 'opacity-100' : 'opacity-0 w-0 overflow-hidden'
        )}>
          <svg className="w-7 h-7 shrink-0" viewBox="0 0 200 200" fill="none">
            <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" className="text-primary" />
            <circle cx="100" cy="100" r="70" fill="currentColor" className="text-primary" />
            <circle cx="80" cy="100" r="6" fill="var(--bg-base)" />
            <circle cx="100" cy="100" r="8" fill="var(--bg-base)" />
            <circle cx="120" cy="100" r="10" fill="var(--bg-base)" />
          </svg>
          <span className="text-base font-bold tracking-widest lowercase text-primary">noant</span>
        </div>

        <div className={cn(
          'absolute left-4 transition-all duration-300',
          isExpanded ? 'opacity-0 scale-75' : 'opacity-100 scale-100'
        )}>
          <svg className="w-8 h-8" viewBox="0 0 200 200" fill="none">
            <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" className="text-primary" />
            <circle cx="100" cy="100" r="70" fill="currentColor" className="text-primary" />
            <circle cx="80" cy="100" r="6" fill="var(--bg-base)" />
            <circle cx="100" cy="100" r="8" fill="var(--bg-base)" />
            <circle cx="120" cy="100" r="10" fill="var(--bg-base)" />
          </svg>
        </div>

        {onClose && (
          <button
            onClick={onClose}
            className="lg:hidden w-8 h-8 rounded-full bg-inset flex items-center justify-center text-secondary hover:text-primary active:scale-95 transition-all"
            aria-label="Close menu"
          >
            <X className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Nav */}
      <nav className="flex-1 p-3 overflow-y-auto overflow-x-hidden scrollbar-thin">
        {navSections.map((section) => (
          <div key={section.title} className={cn('mb-4', !isExpanded && 'mb-2')}>
            <div className={cn(
              'px-3 mb-1.5 text-[9px] font-bold uppercase tracking-widest text-tertiary transition-all duration-300 whitespace-nowrap',
              isExpanded ? 'opacity-100 h-auto' : 'opacity-0 h-0 overflow-hidden'
            )}>
              {section.title}
            </div>

            {section.items.map((item) => {
              const isActive = location.pathname === item.to || (item.to !== '/' && location.pathname.startsWith(item.to))
              const badgeCount = getAlertCount(item.alertKey)
              const isBillingAlert = item.alertKey === 'billingAlert' && alerts.billingAlert

              return (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.to === '/'}
                  onClick={() => onClose?.()}
                  className={cn(
                    'relative flex items-center rounded-xl text-sm font-medium transition-all duration-200 group/nav active:scale-[0.98] mb-0.5',
                    isActive
                      ? 'bg-noant-sky/10 text-noant-sky-deep dark:text-noant-sky'
                      : 'text-secondary hover:bg-inset hover:text-primary',
                    isExpanded ? 'gap-3 px-3 py-2' : 'justify-center px-2 py-2.5 mx-auto w-12'
                  )}
                >
                  {/* Icon wrapper — positioned relative so badge can anchor to it */}
                  <div className="relative shrink-0">
                    <item.icon
                      className={cn(
                        'transition-colors',
                        isActive ? 'text-noant-sky' : 'group-hover/nav:text-noant-sky',
                        isExpanded ? 'w-[17px] h-[17px]' : 'w-[19px] h-[19px]'
                      )}
                      strokeWidth={isActive ? 2.5 : 2}
                    />
                    {/* Badge on collapsed icon */}
                    {!isExpanded && badgeCount > 0 && (
                      <AlertDot count={isBillingAlert ? undefined : badgeCount} pulse={isBillingAlert} />
                    )}
                    {!isExpanded && isBillingAlert && badgeCount > 0 && (
                      <AlertDot pulse />
                    )}
                  </div>

                  {/* Label */}
                  <span className={cn(
                    'whitespace-nowrap transition-all duration-300 flex-1 min-w-0 truncate text-[13px]',
                    isExpanded ? 'opacity-100 w-auto' : 'opacity-0 w-0 overflow-hidden absolute left-14'
                  )}>
                    {item.label}
                  </span>

                  {/* Expanded badge */}
                  {isExpanded && badgeCount > 0 && (
                    <span className={cn(
                      'ml-auto text-[10px] font-bold px-1.5 py-0.5 rounded-full leading-none shrink-0',
                      isBillingAlert
                        ? 'bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400 animate-pulse'
                        : 'bg-red-100 text-red-600 dark:bg-red-950/40 dark:text-red-400'
                    )}>
                      {isBillingAlert ? '!' : badgeCount > 99 ? '99+' : badgeCount}
                    </span>
                  )}

                  {/* Tooltip when collapsed */}
                  {!isExpanded && (
                    <div className="absolute left-14 bg-popover text-white text-xs px-2.5 py-1.5 rounded-lg opacity-0 group-hover/nav:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-50 shadow-lg flex items-center gap-2"
                      style={{ background: 'var(--bg-inset)', color: 'var(--text-primary)', border: '1px solid var(--border-default)' }}>
                      {item.label}
                      {badgeCount > 0 && (
                        <span className={cn(
                          'text-[9px] font-bold px-1 py-0.5 rounded-full leading-none',
                          isBillingAlert ? 'bg-amber-500/20 text-amber-500' : 'bg-red-500/20 text-red-500'
                        )}>
                          {isBillingAlert ? '!' : badgeCount}
                        </span>
                      )}
                    </div>
                  )}
                </NavLink>
              )
            })}
          </div>
        ))}
      </nav>

      {/* Footer — User profile */}
      <div className="p-3 border-t border-default shrink-0">
        <button
          onClick={() => setShowLogoutConfirm(true)}
          className={cn(
            'flex items-center rounded-xl hover:bg-inset transition-colors active:scale-[0.98] w-full group',
            isExpanded ? 'gap-3 px-3 py-2.5' : 'justify-center px-2 py-2.5'
          )}
        >
          <div className="w-8 h-8 rounded-full bg-noant-black text-white dark:bg-white dark:text-noant-black flex items-center justify-center text-xs font-bold shrink-0">
            {initials}
          </div>
          <div className={cn(
            'flex-1 min-w-0 text-left transition-all duration-300',
            isExpanded ? 'opacity-100 w-auto' : 'opacity-0 w-0 overflow-hidden'
          )}>
            <div className="text-xs font-semibold truncate text-primary">{name}</div>
            <div className="text-[10px] text-tertiary flex items-center gap-1">
              {user?.plan || 'Free'} plan
              {alerts.billingAlert && (
                <span className="inline-flex items-center justify-center w-3.5 h-3.5 bg-amber-500 text-white text-[8px] font-bold rounded-full animate-pulse">!</span>
              )}
            </div>
          </div>
          <LogOut className={cn(
            'w-3.5 h-3.5 text-tertiary group-hover:text-noant-sky transition-colors shrink-0',
            isExpanded ? 'opacity-100' : 'opacity-0 w-0'
          )} />
        </button>
      </div>

      <ConfirmModal
        open={showLogoutConfirm}
        onClose={() => setShowLogoutConfirm(false)}
        onConfirm={signOut}
        title="Sign out of Noant?"
        description="You will need to enter your email and password to log back in."
        confirmText="Sign out"
        variant="neutral"
      />
    </div>
  )
}
