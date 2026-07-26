import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useUsers } from '@/lib/hooks/useUsers'
import { useAutoRefresh } from '@/lib/hooks/useAutoRefresh'
import { formatNumber } from '@/lib/utils'
import { Search, ExternalLink, Users, Download, Ban, Check, AlertCircle, Send } from 'lucide-react'
import { SkeletonTableRows } from '@/components/ui/Skeleton'
import { ErrorBanner, EmptyState } from '@/components/ui/Feedback'
import { adminApi } from '@/lib/api'

const PLAN_COLORS: Record<string, string> = {
  free: 'bg-text-tertiary/10 text-text-tertiary',
  starter: 'bg-brand-sky/10 text-brand-sky',
  pro: 'bg-success/10 text-success',
}

const HEALTH_COLORS = (score: number) => {
  if (score >= 80) return 'text-success'
  if (score >= 50) return 'text-warning'
  return 'text-danger'
}

export default function CustomersPage() {
  const [search, setSearch] = useState('')
  const [plan, setPlan] = useState('all')
  const { users, loading, total, error, refetch } = useUsers(search, plan)
  useAutoRefresh(refetch, 20000)

  const [actionMsg, setActionMsg] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const handleSuspend = async (id: string, suspended: boolean) => {
    setActionError(null)
    setActionMsg(null)
    try {
      await adminApi.suspendUser(id, suspended)
      setActionMsg(suspended ? 'User suspended' : 'User activated')
      refetch()
    } catch (e: unknown) {
      setActionError((e as Error).message)
    }
  }

  const handleUpgrade = async (id: string, planID: string) => {
    setActionError(null)
    setActionMsg(null)
    try {
      await adminApi.upgradeUserPlan(id, planID)
      setActionMsg(`Plan upgraded to ${planID}`)
      refetch()
    } catch (e: unknown) {
      setActionError((e as Error).message)
    }
  }

  const handleResendVerify = async (id: string) => {
    setActionError(null)
    setActionMsg(null)
    try {
      await adminApi.resendVerification(id)
      setActionMsg('Verification resent')
    } catch (e: unknown) {
      setActionError((e as Error).message)
    }
  }

  const handleNotify = async (id: string) => {
    setActionError(null)
    setActionMsg(null)
    const title = prompt('Notification title:')
    if (!title) return
    const message = prompt('Notification message:')
    if (!message) return
    try {
      await adminApi.sendUserNotification(id, title, message, 'info')
      setActionMsg('Notification sent')
    } catch (e: unknown) {
      setActionError((e as Error).message)
    }
  }

  const handleExportUsers = () => {
    adminApi.exportCSV('users', { search, plan })
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Customers</h1>
          <p className="text-sm text-text-tertiary">{formatNumber(total)} total users</p>
        </div>
        <button
          onClick={handleExportUsers}
          className="inline-flex items-center gap-2 rounded-lg bg-brand-sky px-4 py-2 text-sm font-medium text-white hover:bg-brand-sky-deep"
        >
          <Download className="h-4 w-4" />
          Export CSV
        </button>
      </div>

      {error && <ErrorBanner message={error} onRetry={() => window.location.reload()} />}
      {actionError && <ErrorBanner message={actionError} onRetry={() => setActionError(null)} />}
      {actionMsg && (
        <div className="rounded-lg bg-success/10 border border-success/20 px-4 py-2 text-sm text-success">
          {actionMsg}
        </div>
      )}

      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-tertiary" />
          <input
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="Search by name or email..."
            className="w-full rounded-lg border border-border bg-bg-inset py-2 pl-9 pr-3 text-sm text-text-primary outline-none transition-colors focus:border-brand-sky"
          />
        </div>
        <select
          value={plan}
          onChange={e => setPlan(e.target.value)}
          className="rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-secondary outline-none"
        >
          <option value="all">All Plans</option>
          <option value="free">Free</option>
          <option value="starter">Starter</option>
          <option value="pro">Pro</option>
        </select>
      </div>

      <div className="rounded-xl border border-border bg-bg-surface overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border bg-bg-inset">
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-tertiary">User</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-tertiary">Plan</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-tertiary">Conversations</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-tertiary">Credits</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-tertiary">Health</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-text-tertiary">Last Active</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-text-tertiary">Actions</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <SkeletonTableRows rows={8} cols={7} />
                ) : users.length === 0 ? (
                  <tr>
                    <td colSpan={7}>
                      <EmptyState
                        icon={Users}
                        title="No users found"
                        description={search ? 'Try adjusting your search and filters' : 'There are no users yet'}
                      />
                    </td>
                  </tr>
                ) : (
                  users.map(user => (
                    <tr key={user.id} className="border-b border-border-subtle transition-colors hover:bg-bg-inset">
                      <td className="px-4 py-3">
                        <div>
                          <p className="text-sm font-medium text-text-primary">{user.first_name} {user.last_name}</p>
                          <p className="text-xs text-text-tertiary">{user.email}</p>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-flex rounded-md px-2 py-0.5 text-xs font-medium ${PLAN_COLORS[user.plan_id] || ''}`}>
                          {user.plan_id}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-text-secondary">{formatNumber(user.total_conversations ?? 0)}</td>
                      <td className="px-4 py-3 text-sm text-text-secondary">{formatNumber(user.credits_remaining ?? 0)}</td>
                      <td className="px-4 py-3">
                        <span className={`text-sm font-medium ${HEALTH_COLORS(user.health_score ?? 100)}`}>{user.health_score ?? 100}%</span>
                      </td>
                      <td className="px-4 py-3 text-xs text-text-tertiary">{user.last_login_at ? new Date(user.last_login_at).toLocaleDateString() : '—'}</td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            onClick={() => handleSuspend(user.id, user.status === 'active')}
                            className={`inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium ${
                              user.status === 'active'
                                ? 'bg-danger/10 text-danger hover:bg-danger/20'
                                : 'bg-success/10 text-success hover:bg-success/20'
                            }`}
                            title={user.status === 'active' ? 'Suspend' : 'Activate'}
                          >
                            {user.status === 'active' ? <Ban className="h-3 w-3" /> : <Check className="h-3 w-3" />}
                          </button>
                          <button
                            onClick={() => handleUpgrade(user.id, user.plan_id === 'free' ? 'starter' : user.plan_id === 'starter' ? 'pro' : 'free')}
                            className="inline-flex items-center gap-1 rounded-md bg-brand-sky/10 px-2 py-1 text-xs font-medium text-brand-sky hover:bg-brand-sky/20"
                            title="Cycle plan"
                          >
                            ⇅
                          </button>
                          <button
                            onClick={() => handleResendVerify(user.id)}
                            className="inline-flex items-center gap-1 rounded-md bg-warning/10 px-2 py-1 text-xs font-medium text-warning hover:bg-warning/20"
                            title="Resend verification"
                          >
                            <AlertCircle className="h-3 w-3" />
                          </button>
                          <button
                            onClick={() => handleNotify(user.id)}
                            className="inline-flex items-center gap-1 rounded-md bg-info/10 px-2 py-1 text-xs font-medium text-info hover:bg-info/20"
                            title="Send notification"
                          >
                            <Send className="h-3 w-3" />
                          </button>
                          <Link
                            to={`/customers/${user.id}`}
                            className="inline-flex items-center gap-1 text-xs text-brand-sky hover:text-brand-sky-deep"
                          >
                            View <ExternalLink className="h-3 w-3" />
                          </Link>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
      </div>
    </div>
  )
}