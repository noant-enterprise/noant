import { useState, useEffect, useCallback } from 'react'
import { adminApi } from '@/lib/api'
import type { AuditLogEntry } from '@/types'
import { useAutoRefresh } from '@/lib/hooks/useAutoRefresh'
import { Search, Filter, ChevronDown, ClipboardList, Download } from 'lucide-react'
import { timeAgo } from '@/lib/utils'
import { SkeletonTableRows } from '@/components/ui/Skeleton'
import { ErrorBanner, EmptyState } from '@/components/ui/Feedback'

const ACTION_COLORS: Record<string, string> = {
  'user.login': 'text-success',
  'user.registered': 'text-brand-sky',
  'user.login.failed': 'text-danger',
  'payment': 'text-success',
  'conversation': 'text-brand-sky',
  'training': 'text-warning',
}

function getActionColor(action: string): string {
  for (const [prefix, color] of Object.entries(ACTION_COLORS)) {
    if (action.startsWith(prefix)) return color
  }
  return 'text-text-tertiary'
}

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLogEntry[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [actionFilter, setActionFilter] = useState('')

  const fetchLogs = useCallback(() => {
    setLoading(true)
    setError(null)
    adminApi.getAuditLogs({ search: search || undefined, action: actionFilter || undefined, limit: 100 })
      .then(res => {
        setLogs(res.logs ?? [])
        setTotal(res.total)
      })
      .catch(() => { setError('Failed to load audit logs'); setLogs([]); setTotal(0) })
      .finally(() => setLoading(false))
  }, [search, actionFilter])

  useEffect(() => {
    const timer = setTimeout(fetchLogs, 300)
    return () => clearTimeout(timer)
  }, [fetchLogs])

  useAutoRefresh(fetchLogs, 30000)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Audit Logs</h1>
          <p className="text-sm text-text-tertiary">{total} total entries — who did what, when</p>
        </div>
        <button
          onClick={() => adminApi.exportCSV('users')}
          className="inline-flex items-center gap-1.5 rounded-lg bg-brand-sky/10 px-3 py-1.5 text-xs font-medium text-brand-sky hover:bg-brand-sky/20"
        >
          <Download className="h-3 w-3" />
          Export CSV
        </button>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-tertiary" />
          <input
            type="text"
            placeholder="Search actions, details, emails..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="w-full rounded-lg border border-border bg-bg-surface py-2 pl-10 pr-4 text-sm text-text-primary placeholder:text-text-tertiary focus:border-brand-sky focus:outline-none"
          />
        </div>
        <div className="relative">
          <Filter className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-tertiary" />
          <select
            value={actionFilter}
            onChange={e => setActionFilter(e.target.value)}
            className="appearance-none rounded-lg border border-border bg-bg-surface py-2 pl-10 pr-8 text-sm text-text-primary focus:border-brand-sky focus:outline-none"
          >
            <option value="">All actions</option>
            <option value="user.login">Login</option>
            <option value="user.logout">Logout</option>
            <option value="conversation">Conversations</option>
            <option value="training">Training</option>
            <option value="WhatsApp">WhatsApp</option>
            <option value="campaign">Campaigns</option>
            <option value="team">Team</option>
            <option value="settings">Settings</option>
          </select>
          <ChevronDown className="pointer-events-none absolute right-2 top-1/2 h-4 w-4 -translate-y-1/2 text-text-tertiary" />
        </div>
      </div>

      {error && <ErrorBanner message={error} onRetry={fetchLogs} />}

      <div className="rounded-xl border border-border bg-bg-surface">
        {loading ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <tbody>
                <SkeletonTableRows rows={8} cols={4} />
              </tbody>
            </table>
          </div>
        ) : logs.length === 0 ? (
          <EmptyState
            icon={ClipboardList}
            title="No audit logs found"
            description={search || actionFilter ? 'Try adjusting your filters' : 'Audit logs will appear here as actions are performed'}
          />
        ) : (
          <div className="overflow-x-auto">
            <div className="divide-y divide-border">
            {logs.map(log => (
              <div key={log.id} className="flex items-start gap-4 p-4">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className={`text-sm font-medium ${getActionColor(log.action)}`}>
                      {log.action}
                    </span>
                    <span className="rounded bg-bg-inset px-1.5 py-0.5 text-xs text-text-tertiary">
                      {log.resource_type}
                    </span>
                  </div>
                  <p className="mt-0.5 text-xs text-text-tertiary">
                    {log.email ? `${log.first_name} ${log.last_name} (${log.email})` : 'System'}
                  </p>
                  {log.details && log.details !== '{}' && (
                    <p className="mt-1 max-w-lg truncate text-xs text-text-tertiary">{log.details}</p>
                  )}
                </div>
                <div className="flex flex-col items-end gap-1">
                  <span className="text-xs text-text-tertiary">{timeAgo(log.created_at)}</span>
                  {log.ip_address && (
                    <span className="text-xs text-text-tertiary">{log.ip_address}</span>
                  )}
                </div>
              </div>
            ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
