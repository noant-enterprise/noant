import { useRevenue } from '@/lib/hooks/useRevenue'
import { StatCard } from '@/components/data/StatCard'
import { formatCurrency } from '@/lib/utils'
import { DollarSign, Users, TrendingDown, Wallet } from 'lucide-react'
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'

export default function RevenuePage() {
  const { data } = useRevenue()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Revenue</h1>
        <p className="text-sm text-text-tertiary">Financial overview and subscription metrics</p>
      </div>

      <div className="grid grid-cols-4 gap-4">
        <StatCard label="MRR" value={formatCurrency(data.mrr)} change={12} changeLabel="vs last month" icon={<DollarSign className="h-4 w-4" />} />
        <StatCard label="Paying Users" value={data.paying_users} icon={<Users className="h-4 w-4" />} />
        <StatCard label="Churn Rate" value={`${data.churn_rate}%`} icon={<TrendingDown className="h-4 w-4" />} />
        <StatCard label="Avg LTV" value={formatCurrency(data.ltv)} icon={<Wallet className="h-4 w-4" />} />
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div className="col-span-2 rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">MRR Growth</h3>
          <ResponsiveContainer width="100%" height={280}>
            <AreaChart data={data.mrr_history}>
              <defs>
                <linearGradient id="green" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#22c55e" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#22c55e" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="month" tick={{ fontSize: 12, fill: '#666' }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fontSize: 12, fill: '#666' }} axisLine={false} tickLine={false} tickFormatter={v => `₦${(v / 1000000).toFixed(0)}M`} />
              <Tooltip contentStyle={{ backgroundColor: '#1c1c1c', border: '1px solid #2a2a2a', borderRadius: '8px', fontSize: 12 }} formatter={(v) => [formatCurrency(Number(v)), 'Revenue']} />
              <Area type="monotone" dataKey="amount" stroke="#22c55e" fill="url(#green)" strokeWidth={2} />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Plan Breakdown</h3>
          <div className="space-y-4">
            {data.plan_breakdown.map(plan => (
              <div key={plan.plan}>
                <div className="flex items-center justify-between text-sm">
                  <span className="font-medium text-text-primary">{plan.plan}</span>
                  <span className="text-text-tertiary">{plan.users} users</span>
                </div>
                <div className="mt-1 h-2 overflow-hidden rounded-full bg-bg-inset">
                  <div className="h-full rounded-full bg-brand-sky" style={{ width: `${Math.min(plan.percentage, 100)}%` }} />
                </div>
                <p className="mt-1 text-xs text-text-tertiary">{formatCurrency(plan.revenue)}/mo</p>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-bg-surface p-5">
        <h3 className="mb-4 text-sm font-medium text-text-secondary">Failed Payments ({data.failed_payments.length})</h3>
        {data.failed_payments.length === 0 ? (
          <p className="text-sm text-text-tertiary">No failed payments</p>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-border">
                <th className="px-3 py-2 text-left text-xs font-medium text-text-tertiary">User</th>
                <th className="px-3 py-2 text-right text-xs font-medium text-text-tertiary">Amount</th>
                <th className="px-3 py-2 text-left text-xs font-medium text-text-tertiary">Reason</th>
                <th className="px-3 py-2 text-right text-xs font-medium text-text-tertiary">Date</th>
              </tr>
            </thead>
            <tbody>
              {data.failed_payments.map(fp => (
                <tr key={fp.id} className="border-b border-border-subtle">
                  <td className="px-3 py-2 text-sm text-text-primary">{fp.user_id}</td>
                  <td className="px-3 py-2 text-right text-sm text-danger">{formatCurrency(fp.amount)}</td>
                  <td className="px-3 py-2 text-sm text-text-secondary">-</td>
                  <td className="px-3 py-2 text-right text-xs text-text-tertiary">{new Date(fp.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
