import { useState } from 'react'
import { Outlet, useLocation, useSearchParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Sidebar } from '@/components/layout/Sidebar'
import { Header } from '@/components/layout/Header'
import { BottomNav } from '@/components/layout/BottomNav'


// Pages that need to fill the entire viewport without any padding wrapper
// (they manage their own layout internally)
const FULL_BLEED_ROUTES = ['/chats']

export function DashboardLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [collapsed, setCollapsed] = useState(true)
  const location = useLocation()
  const [searchParams] = useSearchParams()

  const isFullBleed = FULL_BLEED_ROUTES.some(r => location.pathname.startsWith(r))
  // Hide header + bottom nav on mobile only when viewing a chat thread (not the chat list)
  const isChatThread = location.pathname.startsWith('/chats') && !!searchParams.get('id')
  const hideChromeOnMobile = isChatThread

  return (
    <div className="h-screen flex overflow-hidden bg-base relative">
      {/* Mobile overlay backdrop */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/40 z-40 lg:hidden backdrop-blur-sm transition-opacity"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar container */}
      <aside className={cn(
        'fixed lg:static inset-y-0 left-0 z-50 transition-transform duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] lg:transition-none lg:transform-none shrink-0',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
      )}>
        <Sidebar
          onClose={() => setSidebarOpen(false)}
          collapsed={collapsed}
          onToggleCollapse={() => setCollapsed(!collapsed)}
        />
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden relative z-0">
        <div className={cn(hideChromeOnMobile && 'hidden lg:block')}>
          <Header onMenuClick={() => setSidebarOpen(true)} />
        </div>
        <main className={cn(
          'flex-1 relative',
          isFullBleed
            ? 'overflow-hidden'                                        // chat: no scroll, no padding — fills exactly
            : 'overflow-auto pb-16 lg:pb-0'                           // all other pages: scroll + bottom nav space
        )}>
          {isFullBleed ? (
            // Full-bleed: outlet fills 100% of main, no padding at all
            <div className="h-full w-full overflow-hidden">
              <Outlet />
            </div>
          ) : (
            // Standard: responsive gutters so pages don't touch the edges
            <div className="px-4 sm:px-5 lg:px-6 py-4 lg:py-5 min-h-full">
              <Outlet />
            </div>
          )}
        </main>
        {/* Bottom nav — mobile only, hidden when viewing a chat thread */}
        <div className={cn(hideChromeOnMobile && 'hidden lg:block')}>
          <BottomNav />
        </div>
      </div>

      <style>{`
        @keyframes noantSlideDown {
          from { transform: translate(-50%, -100%); opacity: 0; }
          to { transform: translate(-50%, 0); opacity: 1; }
        }
        .animate-slide-down {
          animation: noantSlideDown 300ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
        }
      `}</style>
    </div>
  )
}
