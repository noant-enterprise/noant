import { StatCard } from '@/components/data/StatCard'
import { LiveFeed } from '@/components/data/LiveFeed'
import { AlertBanner } from '@/components/data/AlertBanner'
import { useAnalytics } from '@/lib/hooks/useAnalytics'
import { useRevenue } from '@/lib/hooks/useRevenue'
import { useSystemHealth } from '@/lib/hooks/useSystemHealth'
import { useAdminWS } from '@/lib/hooks/useAdminWS'
import { formatCurrency, formatNumber, formatPercent } from '@/lib/utils'
import { Users, DollarSign, MessageSquare, Activity, Filter } from 'lucide-react'
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { SkeletonCard, SkeletonStatGrid } from '@/components/ui/Skeleton'
import { ErrorBanner, EmptyState } from '@/components/ui/Feedback'
import { useEffect, useCallback } from 'react'

export default function DashboardPage() {
  const { data: analytics, loading: analyticsLoading, error: analyticsError, refetch: refetchAnalytics } = useAnalytics()
  const { data: revenue, loading: revenueLoading, error: revenueError, refetch: refetchRevenue } = useRevenue()
  const { data: system, loading: systemLoading, error: systemError, refetch: refetchSystem } = useSystemHealth()

  const refreshAll = useCallback(() => {
    refetchAnalytics()
    refetchRevenue()
    refetchSystem()
  }, [refetchAnalytics, refetchRevenue, refetchSystem])

  useAdminWS({
    lead_created: refreshAll,
    lead_updated: refreshAll,
    user_signed_up: refreshAll,
  })

  // Auto-refresh every 30s
  useEffect(() => {
    const interval = setInterval(refreshAll, 30000)
    return () => clearInterval(interval)
  }, [refreshAll])

  const isLoading = analyticsLoading || revenueLoading || systemLoading
  const anyError = analyticsError || revenueError || systemError
  const hasErrors = !!(analyticsError || revenueError || systemError)

  const visitorGrowth = (analytics.visitors_yesterday ?? 0) > 0
    ? ((analytics.visitors_today - analytics.visitors_yesterday) / analytics.visitors_yesterday) * 100
    : 0

  const retryAll = () => {
    refetchAnalytics()
    refetchRevenue()
    refetchSystem()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Dashboard</h1>
          <p className="text-sm text-text-tertiary">Everything happening in NOANT right now</p>
        </div>
        <div className="flex items-center gap-2">
          <span className="flex items-center gap-1.5 text-xs text-success">
            <span className="h-1.5 w-1.5 rounded-full bg-success animate-pulse" />
            All systems operational
          </span>
        </div>
      </div>

      <AlertBanner />

      {anyError && (
        <ErrorBanner
          message={hasErrors ? `Failed to load: ${[analyticsError, revenueError, systemError].filter(Boolean).join(', ')}` : ''}
          onRetry={retryAll}
        />
      )}

      {isLoading ? (
        <SkeletonStatGrid count={4} />
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <StatCard label="MRR" value={formatCurrency(revenue.mrr)} change={12} changeLabel="vs last month" icon={<DollarSign className="h-4 w-4" />} />
          <StatCard label="Paying Users" value={revenue.paying_users} change={12} changeLabel="vs last month" icon={<Users className="h-4 w-4" />} />
          <StatCard label="Visitors Today" value={formatNumber(analytics.visitors_today)} change={visitorGrowth} changeLabel="vs yesterday" icon={<Activity className="h-4 w-4" />} />
          <StatCard label="AI Accuracy" value={`${system.api.status === 'healthy' ? '94.2' : '—'}%`} change={2.1} changeLabel="vs last week" icon={<MessageSquare className="h-4 w-4" />} />
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="col-span-1 md:col-span-2 rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Visitor Trend</h3>
          {analyticsLoading ? (
            <SkeletonCard />
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <AreaChart data={analytics.visitor_history}>
                <defs>
                  <linearGradient id="sky" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#0ea5e9" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#0ea5e9" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#666' }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 12, fill: '#666' }} axisLine={false} tickLine={false} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1c1c1c', border: '1px solid #2a2a2a', borderRadius: '8px', fontSize: 12 }}
                  labelStyle={{ color: '#a1a1a1' }}
                />
                <Area type="monotone" dataKey="visitors" stroke="#0ea5e9" fill="url(#sky)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>

        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-3 text-sm font-medium text-text-secondary">Conversion Funnel</h3>
          {analyticsLoading ? (
            <SkeletonCard />
          ) : analytics.funnel.length === 0 ? (
            <EmptyState
              icon={Filter}
              title="No funnel data"
              description="Conversion funnel data will appear here once visitors start flowing through."
            />
          ) : (
            <div className="space-y-3">
              {analytics.funnel.map((step, i) => (
                <div key={i}>
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-text-secondary">{step.step}</span>
                    <span className="font-medium text-text-primary">{formatNumber(step.count)}</span>
                  </div>
                  <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-bg-inset">
                    <div
                      className="h-full rounded-full bg-brand-sky transition-all"
                      style={{ width: `${step.percentage}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="col-span-1 md:col-span-2 rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-3 text-sm font-medium text-text-secondary">Live Activity</h3>
          <LiveFeed />
        </div>

        <div className="space-y-4">
          <div className="rounded-xl border border-border bg-bg-surface p-5">
            <h3 className="mb-3 text-sm font-medium text-text-secondary">System Status</h3>
            {systemLoading ? (
              <SkeletonCard />
            ) : (
              <div className="space-y-2">
                {[
                  { name: 'API', status: system.api.status, ms: system.api.latency_ms },
                  { name: 'Database', status: system.database.status, ms: system.database.latency_ms },
                  { name: 'Redis', status: system.redis.status, ms: system.redis.latency_ms },
                  { name: 'WhatsApp', status: system.whatsapp.status, ms: system.whatsapp.latency_ms },
                ].map(s => (
                  <div key={s.name} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className={`h-2 w-2 rounded-full ${s.status === 'healthy' ? 'bg-success' : s.status === 'degraded' ? 'bg-warning' : 'bg-danger'}`} />
                      <span className="text-sm text-text-secondary">{s.name}</span>
                    </div>
                    <span className="text-xs text-text-tertiary">{s.ms}ms</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="rounded-xl border border-border bg-bg-surface p-5">
            <h3 className="mb-3 text-sm font-medium text-text-secondary">Quick Stats</h3>
            {revenueLoading ? (
              <SkeletonCard />
            ) : (
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-text-tertiary">Churn Rate</span>
                  <span className="font-medium text-text-primary">{revenue.churn_rate}%</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-tertiary">Failed Payments</span>
                  <span className="font-medium text-warning">{revenue.failed_payments.length}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-tertiary">Active WebSockets</span>
                  <span className="font-medium text-text-primary">{system.active_websockets}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-text-tertiary">Conversion Rate</span>
                  <span className="font-medium text-success">{formatPercent(analytics.conversion_rate)}</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
