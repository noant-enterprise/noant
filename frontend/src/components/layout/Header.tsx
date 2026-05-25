import { useState, useEffect, useRef, useCallback } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Sun, Moon, Bell, AlertTriangle, HelpCircle, CreditCard, Shield, Users, ChevronRight } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { useToast } from '@/components/ui/Toast'
import { useWebSocket } from '@/hooks/useWebSocket'
import { api } from '../../lib/api'

const titles: Record<string, string> = {
  '/': 'Overview',
  '/chats': 'Conversations',
  '/teach': 'Teach your Noant',
  '/insights': 'Insights',
  '/channels': 'Your channels',
  '/setup': 'Your setup',
  '/settings': 'Settings',
  '/notifications': 'Notifications',
  '/billing': 'Billing',
  '/team': 'Team',
  '/widget': 'Web Widget',
}

interface Notification {
  id: string
  type: string
  title: string
  body: string
  link?: string
  is_read: boolean
  created_at: string
}

const typeIcons: Record<string, typeof Bell> = {
  escalation: AlertTriangle,
  unknown_question: HelpCircle,
  payment: CreditCard,
  security: Shield,
  team: Users,
}

export function Header({ onMenuClick }: { onMenuClick: () => void }) {
  const location = useLocation()
  const navigate = useNavigate()
  const { user } = useAuth()
  const title = titles[location.pathname] || 'noant'
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [showDropdown, setShowDropdown] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  const initials = user
    ? `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
    : '--'

  const toggleTheme = () => {
    const html = document.documentElement
    const isDark = html.classList.contains('dark')
    
    if (isDark) {
      html.classList.remove('dark')
      html.classList.add('light')
      localStorage.setItem('noant_theme', 'light')
    } else {
      html.classList.remove('light')
      html.classList.add('dark')
      localStorage.setItem('noant_theme', 'dark')
    }
  }

  const { subscribe } = useWebSocket()
  const { toast: showToast } = useToast()

  const loadNotifications = useCallback(() => {
    api.get<{ notifications: Notification[] }>('/notifications?limit=5').then(res => {
      setNotifications(res.notifications || [])
    }).catch(() => {})
  }, [])

  // Load notifications (poll every 30s)
  useEffect(() => {
    loadNotifications()
    const interval = setInterval(loadNotifications, 30000)
    return () => clearInterval(interval)
  }, [loadNotifications])

  // Listen for real-time notification events
  useEffect(() => {
    const unsub = subscribe((msg) => {
      if (msg.type === 'unknown_question' || msg.type === 'notification' || msg.type === 'new_message') {
        loadNotifications()
        if (msg.type === 'unknown_question') {
          const qText = msg.content || (msg.data as any)?.question || 'Someone asked a question'
          showToast(`New unknown question: "${qText}"`, 'warning')
        }
      }
    })
    return unsub
  }, [subscribe, loadNotifications, showToast])

  // Close dropdown on outside click
  useEffect(() => {
    const handle = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setShowDropdown(false)
      }
    }
    document.addEventListener('mousedown', handle)
    return () => document.removeEventListener('mousedown', handle)
  }, [])

  const unreadCount = notifications.filter(n => !n.is_read).length

  const markRead = async (id: string) => {
    try {
      await api.post(`/notifications/${id}/read`, {})
      setNotifications(prev => prev.map(n => n.id === id ? { ...n, is_read: true } : n))
    } catch {}
  }

  const handleNotifClick = (n: Notification) => {
    markRead(n.id)
    setShowDropdown(false)
    if (n.link) navigate(n.link)
  }

  return (
    <header className="sticky top-0 h-12 bg-surface border-b border-default flex items-center justify-between px-4 z-40 transition-colors duration-300">
      <div className="flex items-center gap-3">
        {/* Mobile: Logo + name opens sidebar */}
        <button
          onClick={onMenuClick}
          className="lg:hidden flex items-center gap-2 active:scale-95 transition-transform"
          aria-label="Open menu"
        >
          {/* Static logo — no animation */}
          <svg className="w-7 h-7" viewBox="0 0 200 200" fill="none">
            <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" className="text-primary" />
            <circle cx="100" cy="100" r="70" fill="currentColor" className="text-primary" />
            <circle cx="80" cy="100" r="6" fill="var(--bg-base)" />
            <circle cx="100" cy="100" r="8" fill="var(--bg-base)" />
            <circle cx="120" cy="100" r="10" fill="var(--bg-base)" />
          </svg>
          <span className="text-lg font-bold tracking-widest lowercase text-primary">noant</span>
        </button>

        {/* Desktop: Page title only */}
        <h1 className="hidden lg:block text-base font-semibold text-primary transition-colors duration-300">
          {title}
        </h1>
      </div>

      <div className="flex items-center gap-2">
        {/* Theme toggle — desktop only */}
        <button
          onClick={toggleTheme}
          className="hidden lg:flex w-8 h-8 rounded-md border border-default items-center justify-center text-secondary hover:bg-inset hover:border-noant-sky hover:text-noant-sky transition-all duration-200 active:scale-95"
          aria-label="Toggle theme"
        >
          <Sun className="w-4 h-4 block dark:hidden" />
          <Moon className="w-4 h-4 hidden dark:block" />
        </button>

        {/* Notification bell + dropdown */}
        <div className="relative" ref={dropdownRef}>
          <button
            onClick={() => setShowDropdown(v => !v)}
            className="w-8 h-8 rounded-md border border-default flex items-center justify-center text-secondary hover:bg-inset hover:border-noant-sky hover:text-noant-sky transition-all duration-200 active:scale-95 relative"
            aria-label="Notifications"
          >
            <Bell className="w-4 h-4" />
            {unreadCount > 0 && (
              <span className="absolute -top-0.5 -right-0.5 min-w-[14px] h-3.5 bg-red-500 text-white text-[8px] font-bold rounded-full flex items-center justify-center px-0.5 border border-surface">
                {unreadCount > 9 ? '9+' : unreadCount}
              </span>
            )}
          </button>

          {showDropdown && (
            <div className="absolute right-0 top-10 w-80 bg-surface border border-default rounded-xl shadow-lg overflow-hidden animate-fade-in z-50">
              <div className="flex items-center justify-between px-4 py-2.5 border-b border-default">
                <span className="text-xs font-semibold text-primary">Notifications</span>
                {unreadCount > 0 && (
                  <span className="text-xs text-noant-sky">{unreadCount} unread</span>
                )}
              </div>

              {notifications.length === 0 ? (
                <div className="py-8 text-center text-sm text-secondary">
                  No notifications yet
                </div>
              ) : (
                <div className="max-h-72 overflow-y-auto">
                  {notifications.map(n => {
                    const Icon = typeIcons[n.type] || Bell
                    return (
                      <button
                        key={n.id}
                        onClick={() => handleNotifClick(n)}
                        className={`w-full flex items-start gap-3 px-4 py-3 text-left hover:bg-inset transition-colors ${!n.is_read ? 'bg-noant-sky/5' : ''}`}
                      >
                        <div className={`w-7 h-7 rounded-lg flex items-center justify-center shrink-0 mt-0.5 ${!n.is_read ? 'bg-noant-sky/10' : 'bg-inset'}`}>
                          <Icon className={`w-3.5 h-3.5 ${!n.is_read ? 'text-noant-sky' : 'text-secondary'}`} />
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-xs font-semibold text-primary truncate">{n.title}</p>
                          <p className="text-xs text-secondary line-clamp-1 mt-0.5">{n.body}</p>
                        </div>
                        {!n.is_read && <div className="w-1.5 h-1.5 rounded-full bg-noant-sky shrink-0 mt-2" />}
                      </button>
                    )
                  })}
                </div>
              )}

              <button
                onClick={() => { navigate('/notifications'); setShowDropdown(false) }}
                className="w-full flex items-center justify-center gap-1.5 px-4 py-2.5 border-t border-default text-xs font-semibold text-noant-sky hover:bg-noant-sky/5 transition-colors"
              >
                View all notifications
                <ChevronRight className="w-3 h-3" />
              </button>
            </div>
          )}
        </div>

        {/* Profile avatar — mobile only */}
        <div className="lg:hidden w-8 h-8 rounded-full bg-noant-black text-white dark:bg-white dark:text-noant-black flex items-center justify-center text-xs font-semibold select-none">
          {initials}
        </div>
      </div>
    </header>
  )
}
