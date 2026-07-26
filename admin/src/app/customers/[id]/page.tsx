import { useParams, Link, useNavigate } from 'react-router-dom'
import { useUser } from '@/lib/hooks/useUsers'
import { formatNumber } from '@/lib/utils'
import { ArrowLeft, Mail, MessageSquare, CreditCard, Heart, UserX } from 'lucide-react'
import { SkeletonCard } from '@/components/ui/Skeleton'
import { EmptyState, ErrorBanner } from '@/components/ui/Feedback'

const PLAN_COLORS: Record<string, string> = {
  free: 'bg-text-tertiary/10 text-text-tertiary',
  starter: 'bg-brand-sky/10 text-brand-sky',
  pro: 'bg-success/10 text-success',
}

export default function CustomerDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user, loading, error } = useUser(id || '')

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-5 w-32 rounded-md bg-bg-inset animate-pulse" />
        <div className="flex items-start gap-6">
          <div className="h-16 w-16 rounded-2xl bg-bg-inset animate-pulse" />
          <div className="flex-1 space-y-2">
            <div className="h-7 w-48 rounded-md bg-bg-inset animate-pulse" />
            <div className="h-4 w-64 rounded-md bg-bg-inset animate-pulse" />
          </div>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => <SkeletonCard key={i} />)}
        </div>
      </div>
    )
  }

  if (error) return <ErrorBanner message={error} onRetry={() => window.location.reload()} />

  if (!user) {
    return (
      <EmptyState
        icon={UserX}
        title="User not found"
        description="The customer you're looking for doesn't exist or has been removed."
        action={{ label: 'Back to Customers', onClick: () => navigate('/customers') }}
      />
    )
  }

  const healthColor = user.health_score >= 80 ? 'text-success' : user.health_score >= 50 ? 'text-warning' : 'text-danger'
  const healthBg = user.health_score >= 80 ? 'bg-success' : user.health_score >= 50 ? 'bg-warning' : 'bg-danger'

  return (
    <div className="space-y-6">
      <Link to="/customers" className="inline-flex items-center gap-1.5 text-sm text-text-tertiary hover:text-text-primary">
        <ArrowLeft className="h-4 w-4" /> Back to Customers
      </Link>

      <div className="flex items-start gap-6">
        <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-brand-sky/10 text-xl font-bold text-brand-sky">
          {user.first_name[0]}{user.last_name[0]}
        </div>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-text-primary">{user.first_name} {user.last_name}</h1>
          <div className="mt-1 flex items-center gap-4 text-sm text-text-tertiary">
            <span className="flex items-center gap-1"><Mail className="h-3.5 w-3.5" /> {user.email}</span>
            <span className={`inline-flex rounded-md px-2 py-0.5 text-xs font-medium ${PLAN_COLORS[user.plan_id]}`}>{user.plan_id}</span>
            <span className={user.status === 'active' ? 'text-success' : 'text-danger'}>{user.status}</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="rounded-xl border border-border bg-bg-surface p-4">
          <div className="flex items-center gap-2 text-text-tertiary"><MessageSquare className="h-4 w-4" /><span className="text-xs">Conversations</span></div>
          <p className="mt-1 text-xl font-bold text-text-primary">{formatNumber(user.total_conversations)}</p>
        </div>
        <div className="rounded-xl border border-border bg-bg-surface p-4">
          <div className="flex items-center gap-2 text-text-tertiary"><MessageSquare className="h-4 w-4" /><span className="text-xs">Messages</span></div>
          <p className="mt-1 text-xl font-bold text-text-primary">{formatNumber(user.total_messages)}</p>
        </div>
        <div className="rounded-xl border border-border bg-bg-surface p-4">
          <div className="flex items-center gap-2 text-text-tertiary"><CreditCard className="h-4 w-4" /><span className="text-xs">Credits</span></div>
          <p className="mt-1 text-xl font-bold text-text-primary">{formatNumber(user.credits_remaining)}</p>
        </div>
        <div className="rounded-xl border border-border bg-bg-surface p-4">
          <div className="flex items-center gap-2 text-text-tertiary"><Heart className="h-4 w-4" /><span className="text-xs">Health Score</span></div>
          <div className="mt-2 flex items-center gap-2">
            <div className="h-2 w-24 overflow-hidden rounded-full bg-bg-inset">
              <div className={`h-full rounded-full ${healthBg}`} style={{ width: `${user.health_score}%` }} />
            </div>
            <span className={`text-sm font-bold ${healthColor}`}>{user.health_score}%</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-3 text-sm font-medium text-text-secondary">Details</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between"><span className="text-text-tertiary">Joined</span><span className="text-text-primary">{new Date(user.created_at).toLocaleDateString()}</span></div>
            <div className="flex justify-between"><span className="text-text-tertiary">Last Active</span><span className="text-text-primary">{new Date(user.last_login_at).toLocaleDateString()}</span></div>
          </div>
        </div>
        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-3 text-sm font-medium text-text-secondary">Actions</h3>
          <div className="space-y-2">
            <button className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-secondary transition-colors hover:text-text-primary">Impersonate User</button>
            <button className="w-full rounded-lg border border-danger/30 bg-danger/5 px-3 py-2 text-sm text-danger transition-colors hover:bg-danger/10">Suspend Account</button>
          </div>
        </div>
      </div>
    </div>
  )
}
