import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useUsers } from '@/lib/hooks/useUsers'
import { formatNumber } from '@/lib/utils'
import { Search, ExternalLink, Users } from 'lucide-react'
import { SkeletonTableRows } from '@/components/ui/Skeleton'
import { ErrorBanner, EmptyState } from '@/components/ui/Feedback'

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
  const { users, loading, total, error } = useUsers(search, plan)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Customers</h1>
        <p className="text-sm text-text-tertiary">{formatNumber(total)} total users</p>
      </div>

      {error && <ErrorBanner message={error} onRetry={() => window.location.reload()} />}

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
                  <th className="px-4 py-3 text-right text-xs font-medium text-text-tertiary">Action</th>
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
                        description={search ? 'Try adjusting your search or filters' : 'There are no users yet'}
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
                        <Link
                          to={`/customers/${user.id}`}
                          className="inline-flex items-center gap-1 text-xs text-brand-sky hover:text-brand-sky-deep"
                        >
                          View <ExternalLink className="h-3 w-3" />
                        </Link>
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
