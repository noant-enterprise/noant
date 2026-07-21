import { useState, useEffect, useCallback } from 'react'
import { Bell, BellOff, Check, CheckCheck, AlertTriangle, HelpCircle, CreditCard, Shield, Users } from 'lucide-react'
import { api } from '../../../lib/api'
import { useToast } from '@/components/ui/Toast'
import { useNavigate } from 'react-router-dom'

interface Notification {
  id: string
  type: string
  title: string
  body: string
  link?: string
  is_read: boolean
  created_at: string
}

const typeConfig: Record<string, { icon: typeof Bell; color: string; bg: string }> = {
  escalation:       { icon: AlertTriangle, color: 'text-orange-500', bg: 'bg-orange-500/10' },
  unknown_question: { icon: HelpCircle,    color: 'text-purple-500', bg: 'bg-purple-500/10' },
  payment:          { icon: CreditCard,    color: 'text-green-500',  bg: 'bg-green-500/10'  },
  security:         { icon: Shield,        color: 'text-red-500',    bg: 'bg-red-500/10'    },
  team:             { icon: Users,         color: 'text-blue-500',   bg: 'bg-blue-500/10'   },
  default:          { icon: Bell,          color: 'text-noant-sky',  bg: 'bg-noant-sky/10'  },
}

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  return `${d}d ago`
}

function groupByDate(notifications: Notification[]) {
  const groups: Record<string, Notification[]> = {}
  const today = new Date().toDateString()
  const yesterday = new Date(Date.now() - 86400000).toDateString()

  notifications.forEach(n => {
    const d = new Date(n.created_at).toDateString()
    const label = d === today ? 'Today' : d === yesterday ? 'Yesterday' : new Date(n.created_at).toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric' })
    if (!groups[label]) groups[label] = []
    groups[label].push(n)
  })
  return groups
}

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<'all' | 'unread'>('all')
  const { toast: showToast } = useToast()
  const navigate = useNavigate()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await api.get<{ notifications: Notification[] }>('/notifications?limit=50')
      setNotifications(res.notifications || [])
    } catch {
      showToast('Failed to load notifications', 'error')
    } finally {
      setLoading(false)
    }
  }, [showToast])

  useEffect(() => { load() }, [load])

  const markRead = async (id: string) => {
    try {
      await api.post(`/notifications/${id}/read`, {})
      setNotifications(prev => prev.map(n => n.id === id ? { ...n, is_read: true } : n))
    } catch {}
  }

  const markAllRead = async () => {
    try {
      await api.post('/notifications/read-all', {})
      setNotifications(prev => prev.map(n => ({ ...n, is_read: true })))
      showToast('All notifications marked as read', 'success')
    } catch {
      showToast('Failed to mark all as read', 'error')
    }
  }

  const handleClick = (n: Notification) => {
    if (!n.is_read) markRead(n.id)
    if (n.link) navigate(n.link)
  }

  const filtered = filter === 'unread' ? notifications.filter(n => !n.is_read) : notifications
  const unreadCount = notifications.filter(n => !n.is_read).length
  const grouped = groupByDate(filtered)

  if (loading) {
    return (
      <div className="min-h-screen p-4 lg:p-6">
        <div className="max-w-2xl mx-auto space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-20 rounded-xl animate-shimmer-slow" />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen p-4 lg:p-6 animate-page-in">
      <div className="max-w-2xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-bold text-primary">Notifications</h1>
            <p className="text-sm text-secondary mt-0.5">
              {unreadCount > 0 ? `${unreadCount} unread` : 'All caught up!'}
            </p>
          </div>
          {unreadCount > 0 && (
            <button
              onClick={markAllRead}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold text-noant-sky border border-noant-sky/30 hover:bg-noant-sky/10 active:scale-95 transition-all"
            >
              <CheckCheck className="w-3.5 h-3.5" />
              Mark all read
            </button>
          )}
        </div>

        {/* Filter */}
        <div className="flex gap-1 mb-5 bg-inset p-1 rounded-xl w-fit">
          {(['all', 'unread'] as const).map(f => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-all capitalize ${filter === f ? 'bg-surface text-primary shadow-sm border border-default' : 'text-secondary hover:text-primary'}`}
            >
              {f === 'all' ? `All (${notifications.length})` : `Unread (${unreadCount})`}
            </button>
          ))}
        </div>

        {/* Empty state */}
        {filtered.length === 0 && (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <div className="w-16 h-16 rounded-2xl bg-inset flex items-center justify-center mb-4">
              <BellOff className="w-7 h-7 text-tertiary" />
            </div>
            <h3 className="font-semibold text-primary mb-1">
              {filter === 'unread' ? 'No unread notifications' : 'No notifications yet'}
            </h3>
            <p className="text-sm text-secondary max-w-xs">
              {filter === 'unread' ? "You're all caught up!" : 'Notifications about escalations, unknown questions, and more will appear here.'}
            </p>
          </div>
        )}

        {/* Grouped list */}
        {Object.entries(grouped).map(([dateLabel, items]) => (
          <div key={dateLabel} className="mb-6">
            <div className="text-xs font-semibold text-tertiary uppercase tracking-widest mb-2 px-1">{dateLabel}</div>
            <div className="space-y-2">
              {items.map(n => {
                const cfg = typeConfig[n.type] ?? typeConfig.default
                if (!cfg) return null
                const Icon = cfg.icon
                return (
                  <div
                    key={n.id}
                    onClick={() => handleClick(n)}
                    className={`flex gap-3 p-4 rounded-xl border transition-all cursor-pointer group ${
                      n.is_read
                        ? 'border-default bg-surface hover:bg-inset/50'
                        : 'border-noant-sky/20 bg-noant-sky/5 hover:bg-noant-sky/10'
                    }`}
                  >
                    <div className={`w-9 h-9 rounded-xl ${cfg.bg} flex items-center justify-center shrink-0 mt-0.5`}>
                      <Icon className={`w-4 h-4 ${cfg.color}`} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <p className={`text-sm font-semibold leading-tight ${n.is_read ? 'text-primary' : 'text-primary'}`}>
                          {n.title}
                        </p>
                        <div className="flex items-center gap-2 shrink-0">
                          <span className="text-xs text-tertiary whitespace-nowrap">{timeAgo(n.created_at)}</span>
                          {!n.is_read && (
                            <button
                              onClick={e => { e.stopPropagation(); markRead(n.id) }}
                              className="opacity-0 group-hover:opacity-100 w-5 h-5 rounded-full bg-noant-sky/10 hover:bg-noant-sky/20 flex items-center justify-center transition-all"
                              title="Mark as read"
                            >
                              <Check className="w-3 h-3 text-noant-sky" />
                            </button>
                          )}
                        </div>
                      </div>
                      <p className="text-sm text-secondary mt-0.5 line-clamp-2">{n.body}</p>
                    </div>
                    {!n.is_read && (
                      <div className="w-2 h-2 rounded-full bg-noant-sky shrink-0 mt-2" />
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
