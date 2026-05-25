import { NavLink, useLocation } from 'react-router-dom'
import { LayoutGrid, MessageSquare, BarChart3, Link2, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'

const items = [
  { to: '/', icon: LayoutGrid, label: 'Home' },
  { to: '/chats', icon: MessageSquare, label: 'Inbox' },
  { to: '/insights', icon: BarChart3, label: 'Insights' },
  { to: '/channels', icon: Link2, label: 'Channels' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

export function BottomNav() {
  const location = useLocation()
  
  // Hide bottom nav only when actively viewing a message thread (has ?id= param)
  // Show it on chat list and all other pages
  const isChatThread = location.pathname === '/chats' && location.search.includes('id=')
  if (isChatThread) return null

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 lg:hidden bg-surface/95 backdrop-blur-lg border-t border-default pb-[env(safe-area-inset-bottom)]">
      <div className="flex h-14">
        {items.map((item) => (
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
                <div className="relative flex items-center justify-center w-7 h-7">
                  <item.icon
                    className={cn(
                      'w-[22px] h-[22px] transition-transform duration-200',
                      isActive && 'scale-110'
                    )}
                    strokeWidth={isActive ? 2.5 : 1.5}
                  />
                </div>
                <span className={cn(
                  'text-[10px] font-medium leading-none transition-all',
                  isActive && 'font-semibold'
                )}>
                  {item.label}
                </span>
                {isActive && (
                  <span className="absolute top-1 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-noant-sky" />
                )}
              </>
            )}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
