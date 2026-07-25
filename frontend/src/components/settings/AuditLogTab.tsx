import { useState, useEffect, useCallback } from 'react'
import { Search, Filter, ChevronLeft, ChevronRight, Clock, AlertTriangle, CheckCircle, User, Globe, Monitor } from 'lucide-react'
import { api } from '@/lib/api'
import { Skeleton } from '@/components/ui/Skeleton'

interface AuditLog {
  id: string
  user_id: string
  user_name: string
  user_email: string
  action: string
  resource_type: string
  resource_id: string | null
  details: Record<string, any> | null
  ip_address: string | null
  user_agent: string | null
  created_at: string
}

interface AuditLogsResponse {
  audit_logs: AuditLog[]
  total: number
  limit: number
  offset: number
}

const ACTION_LABELS: Record<string, { label: string; color: string; icon: typeof Clock }> = {
  'user.login.success': { label: 'Login', color: 'text-emerald-400', icon: CheckCircle },
  'user.login.failed': { label: 'Failed Login', color: 'text-red-400', icon: AlertTriangle },
  'user.login.failed.email_not_verified': { label: 'Login (Unverified)', color: 'text-amber-400', icon: AlertTriangle },
  'user.login.failed.account_locked': { label: 'Login (Locked)', color: 'text-red-500', icon: AlertTriangle },
  'user.registered': { label: 'Registration', color: 'text-sky-400', icon: User },
  'user.logout': { label: 'Logout', color: 'text-zinc-400', icon: User },
  'user.password_changed': { label: 'Password Changed', color: 'text-amber-400', icon: AlertTriangle },
  'user.password_reset': { label: 'Password Reset', color: 'text-amber-400', icon: AlertTriangle },
  'user.email_verified': { label: 'Email Verified', color: 'text-emerald-400', icon: CheckCircle },
}

function getActionInfo(action: string) {
  for (const [key, info] of Object.entries(ACTION_LABELS)) {
    if (action.includes(key) || action === key) return info
  }
  return { label: action, color: 'text-secondary', icon: Clock }
}

function timeAgo(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = now - then
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  return new Date(dateStr).toLocaleDateString()
}

function formatUA(ua: string | null): string {
  if (!ua) return 'Unknown'
  if (ua.includes('Chrome')) return 'Chrome'
  if (ua.includes('Firefox')) return 'Firefox'
  if (ua.includes('Safari')) return 'Safari'
  if (ua.includes('Edge')) return 'Edge'
  return ua.slice(0, 40)
}

export function AuditLogTab() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [offset, setOffset] = useState(0)
  const [actionFilter, setActionFilter] = useState('')
  const [showFilters, setShowFilters] = useState(false)
  const limit = 20

  const fetchLogs = useCallback(async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams({ limit: String(limit), offset: String(offset) })
      if (actionFilter) params.set('action', actionFilter)
      const res = await api.get<AuditLogsResponse>(`/settings/audit-logs/search?${params}`)
      setLogs(res.audit_logs || [])
      setTotal(res.total || 0)
    } catch {
      setLogs([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [offset, actionFilter])

  useEffect(() => { fetchLogs() }, [fetchLogs])

  const totalPages = Math.ceil(total / limit)

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-semibold text-primary">Audit Log</h3>
        <p className="text-xs text-secondary mt-1">Track account activity, logins, and security events.</p>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-tertiary" />
          <input
            type="text"
            placeholder="Filter by action..."
            value={actionFilter}
            onChange={e => { setActionFilter(e.target.value); setOffset(0) }}
            className="w-full pl-9 pr-3 py-2 bg-inset border border-default rounded-xl text-sm text-primary placeholder:text-tertiary focus:outline-none focus:border-noant-sky/50 transition-colors"
          />
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`p-2 rounded-xl border transition-colors ${showFilters ? 'bg-noant-sky/10 border-noant-sky/30 text-noant-sky' : 'bg-inset border-default text-secondary hover:text-primary'}`}
          aria-label="Toggle filters"
        >
          <Filter className="w-4 h-4" />
        </button>
      </div>

      {/* Quick filter chips */}
      {showFilters && (
        <div className="flex flex-wrap gap-1.5">
          {['login', 'logout', 'register', 'password', 'verified'].map(chip => (
            <button
              key={chip}
              onClick={() => { setActionFilter(actionFilter === chip ? '' : chip); setOffset(0) }}
              className={`px-2.5 py-1 rounded-full text-xs font-medium transition-colors ${
                actionFilter === chip
                  ? 'bg-noant-sky/15 text-noant-sky border border-noant-sky/30'
                  : 'bg-inset text-secondary border border-default hover:text-primary'
              }`}
            >
              {chip}
            </button>
          ))}
          {actionFilter && (
            <button
              onClick={() => { setActionFilter(''); setOffset(0) }}
              className="px-2.5 py-1 rounded-full text-xs text-red-400 hover:text-red-300 transition-colors"
            >
              Clear
            </button>
          )}
        </div>
      )}

      {/* Log entries */}
      {loading ? (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center gap-3 p-3 rounded-xl bg-inset">
              <Skeleton className="w-8 h-8 rounded-lg shrink-0" />
              <div className="flex-1 space-y-1.5">
                <Skeleton className="h-3.5 w-32" />
                <Skeleton className="h-3 w-48" />
              </div>
              <Skeleton className="h-3 w-16" />
            </div>
          ))}
        </div>
      ) : logs.length === 0 ? (
        <div className="text-center py-12">
          <Clock className="w-8 h-8 text-tertiary mx-auto mb-3" />
          <p className="text-sm text-secondary">No audit logs found</p>
          <p className="text-xs text-tertiary mt-1">Activity will appear here as you use the platform</p>
        </div>
      ) : (
        <div className="space-y-1">
          {logs.map(log => {
            const info = getActionInfo(log.action)
            const Icon = info.icon
            const success = log.details?.success !== false
            return (
              <div key={log.id} className="flex items-start gap-3 p-3 rounded-xl hover:bg-inset/50 transition-colors group">
                <div className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5 ${
                  success ? 'bg-emerald-500/10' : 'bg-red-500/10'
                }`}>
                  <Icon className={`w-4 h-4 ${info.color}`} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-primary">{info.label}</span>
                    {log.details?.status && (
                      <span className={`text-[10px] px-1.5 py-0.5 rounded-full font-mono ${
                        success ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'
                      }`}>
                        {log.details.status}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-3 mt-1 text-xs text-tertiary">
                    {log.user_name && (
                      <span className="flex items-center gap-1">
                        <User className="w-3 h-3" />
                        {log.user_name}
                      </span>
                    )}
                    {!log.user_name && log.user_email && (
                      <span className="flex items-center gap-1">
                        <User className="w-3 h-3" />
                        {log.user_email}
                      </span>
                    )}
                    {log.ip_address && (
                      <span className="flex items-center gap-1">
                        <Globe className="w-3 h-3" />
                        {log.ip_address}
                      </span>
                    )}
                    {log.user_agent && (
                      <span className="flex items-center gap-1">
                        <Monitor className="w-3 h-3" />
                        {formatUA(log.user_agent)}
                      </span>
                    )}
                  </div>
                </div>
                <span className="text-xs text-tertiary whitespace-nowrap" title={new Date(log.created_at).toLocaleString()}>
                  {timeAgo(log.created_at)}
                </span>
              </div>
            )
          })}
        </div>
      )}

      {/* Pagination */}
      {total > limit && (
        <div className="flex items-center justify-between pt-2 border-t border-default">
          <span className="text-xs text-tertiary">
            {offset + 1}–{Math.min(offset + limit, total)} of {total}
          </span>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setOffset(Math.max(0, offset - limit))}
              disabled={offset === 0}
              className="p-1.5 rounded-lg text-secondary hover:text-primary hover:bg-inset disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              aria-label="Previous page"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-xs text-tertiary px-2">
              {Math.floor(offset / limit) + 1}/{totalPages}
            </span>
            <button
              onClick={() => setOffset(offset + limit)}
              disabled={offset + limit >= total}
              className="p-1.5 rounded-lg text-secondary hover:text-primary hover:bg-inset disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              aria-label="Next page"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
