import { useState } from 'react'
import { cn } from '@/lib/utils'
import { NavLink, useLocation } from 'react-router-dom'
import { LayoutGrid, MessageSquare, GraduationCap, BarChart3, Link2, Settings, LogOut, X, CreditCard, Users, Code2 } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { ConfirmModal } from '@/components/ui/ConfirmModal'

const navSections = [
  {
    title: 'Your workspace',
    items: [
      { to: '/', icon: LayoutGrid, label: 'Overview' },
      { to: '/chats', icon: MessageSquare, label: 'Conversations' },
      { to: '/insights', icon: BarChart3, label: 'Insights' },
    ],
  },
  {
    title: 'Build your Noant',
    items: [
      { to: '/teach', icon: GraduationCap, label: 'Teach your Noant' },
      { to: '/channels', icon: Link2, label: 'Your channels' },
      { to: '/widget', icon: Code2, label: 'Web Widget' },
    ],
  },
  {
    title: 'Manage',
    items: [
      { to: '/team', icon: Users, label: 'Team' },
      { to: '/billing', icon: CreditCard, label: 'Billing' },
      { to: '/settings', icon: Settings, label: 'Settings' },
    ],
  },
]

interface SidebarProps {
  onClose?: () => void
  collapsed?: boolean
  onToggleCollapse?: () => void
}

export function Sidebar({ onClose, collapsed = false, onToggleCollapse }: SidebarProps) {
  const { user, signOut } = useAuth()
  const location = useLocation()
  const [isHovered, setIsHovered] = useState(false)
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false)
  
  const isExpanded = collapsed ? isHovered : true

  const initials = user
    ? `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
    : '--'

  const name = user
    ? `${user.first_name} ${user.last_name}`
    : 'Loading...'

  return (
    <div 
      className={cn(
        'flex flex-col h-full bg-surface border-r border-default transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] relative shrink-0',
        isExpanded ? 'w-[220px]' : 'w-[68px]'
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Toggle button — visible tab on right edge */}
      {!onClose && onToggleCollapse && (
        <button
          onClick={(e) => {
            e.stopPropagation()
            onToggleCollapse()
          }}
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
          <svg className="w-7 h-7" viewBox="0 0 200 200" fill="none">
            <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" className="text-primary" />
            <circle cx="100" cy="100" r="70" fill="currentColor" className="text-primary" />
            <circle cx="80" cy="100" r="6" fill="var(--bg-base)" />
            <circle cx="100" cy="100" r="8" fill="var(--bg-base)" />
            <circle cx="120" cy="100" r="10" fill="var(--bg-base)" />
          </svg>
          <span className="text-lg font-bold tracking-widest lowercase text-primary">noant</span>
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
      <nav className="flex-1 p-3 overflow-y-auto overflow-x-hidden">
        {navSections.map((section) => (
          <div key={section.title} className={cn('mb-5', !isExpanded && 'mb-2')}>
            <div className={cn(
              'px-3 mb-2 text-[10px] font-semibold uppercase tracking-widest text-tertiary transition-all duration-300 whitespace-nowrap',
              isExpanded ? 'opacity-100 h-auto' : 'opacity-0 h-0 overflow-hidden'
            )}>
              {section.title}
            </div>
            {section.items.map((item) => {
              const isActive = location.pathname === item.to || (item.to !== '/' && location.pathname.startsWith(item.to))
              return (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.to === '/'}
                  onClick={() => onClose?.()}
                  className={cn(
                    'flex items-center rounded-lg text-sm font-medium transition-all duration-200 group/nav active:scale-[0.98] relative',
                    isActive
                      ? 'bg-noant-sky/10 text-noant-sky-deep dark:text-noant-sky'
                      : 'text-secondary hover:bg-inset hover:text-primary',
                    isExpanded ? 'gap-3 px-3 py-2.5' : 'justify-center px-2 py-3 mx-auto w-12'
                  )}
                >
                  <item.icon 
                    className={cn(
                      'shrink-0 transition-colors',
                      isActive ? 'text-noant-sky' : 'group-hover/nav:text-noant-sky',
                      isExpanded ? 'w-[18px] h-[18px]' : 'w-5 h-5'
                    )} 
                    strokeWidth={2} 
                  />
                  <span className={cn(
                    'whitespace-nowrap transition-all duration-300',
                    isExpanded ? 'opacity-100 w-auto' : 'opacity-0 w-0 overflow-hidden absolute left-14'
                  )}>
                    {item.label}
                  </span>
                  {!isExpanded && (
                    <div className="absolute left-14 bg-noant-black text-white text-xs px-2 py-1 rounded-md opacity-0 group-hover/nav:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-50">
                      {item.label}
                    </div>
                  )}
                </NavLink>
              )
            })}
          </div>
        ))}
      </nav>

      {/* Footer */}
      <div className="p-3 border-t border-default shrink-0">
        <button
          onClick={() => setShowLogoutConfirm(true)}
          className={cn(
            'flex items-center rounded-lg hover:bg-inset transition-colors active:scale-[0.98] w-full',
            isExpanded ? 'gap-3 px-3 py-2.5' : 'justify-center px-2 py-3'
          )}
        >
          <div className="w-8 h-8 rounded-full bg-noant-black text-white dark:bg-white dark:text-noant-black flex items-center justify-center text-xs font-semibold shrink-0">
            {initials}
          </div>
          <div className={cn(
            'flex-1 min-w-0 text-left transition-all duration-300',
            isExpanded ? 'opacity-100 w-auto' : 'opacity-0 w-0 overflow-hidden'
          )}>
            <div className="text-sm font-semibold truncate text-primary">{name}</div>
            <div className="text-[10px] text-tertiary">{user?.plan || 'Free'} plan</div>
          </div>
          <LogOut className={cn(
            'w-4 h-4 text-tertiary hover:text-noant-sky transition-colors shrink-0',
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
