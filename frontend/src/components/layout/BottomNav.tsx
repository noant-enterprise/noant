import { NavLink, useLocation } from 'react-router-dom'
import { LayoutGrid, MessageSquare, GraduationCap, BarChart3, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useSidebarAlerts } from '@/contexts/SidebarAlertsContext'

const items = [
  { to: '/dashboard', icon: LayoutGrid, label: 'Home', alertKey: null as string | null },
  { to: '/chats', icon: MessageSquare, label: 'Inbox', alertKey: 'unreadChats' },
  { to: '/teach', icon: GraduationCap, label: 'Teach', alertKey: 'unknownQuestions' },
  { to: '/insights', icon: BarChart3, label: 'Insights', alertKey: null },
  { to: '/settings', icon: Settings, label: 'Settings', alertKey: null },
]

export function BottomNav() {
  const location = useLocation()
  const alerts = useSidebarAlerts()

  // Hide bottom nav only when actively viewing a message thread (has ?id= param)
  const isChatThread = location.pathname === '/chats' && location.search.includes('id=')
  if (isChatThread) return null

  const getCount = (alertKey: string | null): number => {
    if (!alertKey) return 0
    return (alerts as any)[alertKey] ?? 0
  }

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 lg:hidden bg-surface/95 backdrop-blur-lg border-t border-default pb-[env(safe-area-inset-bottom)]">
      <div className="flex h-[60px]">
        {items.map((item) => {
          const badgeCount = getCount(item.alertKey)

          return (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                cn(
                  'flex-1 flex flex-col items-center justify-center gap-0.5 relative select-none active:opacity-70 transition-colors duration-200',
                  isActive ? 'text-noant-sky' : 'text-tertiary'
                )
              }
            >
              {({ isActive }) => (
                <>
                  <div className="relative flex items-center justify-center w-8 h-8">
                    <item.icon
                      className={cn(
                        'w-[20px] h-[20px] transition-transform duration-200',
                        isActive && 'scale-110'
                      )}
                      strokeWidth={isActive ? 2.5 : 1.75}
                    />
                    {/* Notification badge */}
                    {badgeCount > 0 && (
                      <span className="absolute top-0 right-0 min-w-[14px] h-3.5 bg-red-500 text-white text-[8px] font-bold rounded-full flex items-center justify-center px-0.5 border border-surface leading-none">
                        {badgeCount > 9 ? '9+' : badgeCount}
                      </span>
                    )}
                    {/* Active indicator dot */}
                    {isActive && (
                      <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-noant-sky" />
                    )}
                  </div>
                  <span className={cn(
                    'text-[9px] font-medium leading-none transition-all',
                    isActive && 'font-semibold text-noant-sky'
                  )}>
                    {item.label}
                  </span>
                </>
              )}
            </NavLink>
          )
        })}
      </div>
    </nav>
  )
}
