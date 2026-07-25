import { useAnalytics } from '@/lib/hooks/useAnalytics'
import { StatCard } from '@/components/data/StatCard'
import { formatNumber } from '@/lib/utils'
import { Users, Eye, MousePointerClick, Clock } from 'lucide-react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const COLORS = ['#0ea5e9', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6']

export default function AnalyticsPage() {
  const { data } = useAnalytics()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Analytics</h1>
        <p className="text-sm text-text-tertiary">Landing page performance and visitor insights</p>
      </div>

      <div className="grid grid-cols-4 gap-4">
        <StatCard label="Visitors Today" value={formatNumber(data.visitors_today)} change={((data.visitors_today - data.visitors_yesterday) / data.visitors_yesterday) * 100} changeLabel="vs yesterday" icon={<Eye className="h-4 w-4" />} />
        <StatCard label="Signups Today" value={data.signups_today} icon={<Users className="h-4 w-4" />} />
        <StatCard label="Conversion Rate" value={`${data.conversion_rate}%`} icon={<MousePointerClick className="h-4 w-4" />} />
        <StatCard label="Avg Session" value={`${Math.floor(data.avg_session_duration / 60)}m ${data.avg_session_duration % 60}s`} icon={<Clock className="h-4 w-4" />} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Visitor History</h3>
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={data.visitor_history}>
              <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#666' }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fontSize: 12, fill: '#666' }} axisLine={false} tickLine={false} />
              <Tooltip contentStyle={{ backgroundColor: '#1c1c1c', border: '1px solid #2a2a2a', borderRadius: '8px', fontSize: 12 }} />
              <Bar dataKey="visitors" fill="#0ea5e9" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Traffic Sources</h3>
          <div className="flex items-center gap-6">
            <ResponsiveContainer width={140} height={140}>
              <PieChart>
                <Pie data={data.traffic_sources} dataKey="visitors" nameKey="source" cx="50%" cy="50%" innerRadius={40} outerRadius={60}>
                  {data.traffic_sources.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                </Pie>
              </PieChart>
            </ResponsiveContainer>
            <div className="flex-1 space-y-2">
              {data.traffic_sources.map((source, i) => (
                <div key={source.source} className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full" style={{ backgroundColor: COLORS[i % COLORS.length] }} />
                    <span className="text-text-secondary">{source.source}</span>
                  </div>
                  <span className="font-medium text-text-primary">{formatNumber(source.visitors)}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-bg-surface p-5">
        <h3 className="mb-4 text-sm font-medium text-text-secondary">Page Performance</h3>
        <table className="w-full">
          <thead>
            <tr className="border-b border-border">
              <th className="px-3 py-2 text-left text-xs font-medium text-text-tertiary">Page</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-text-tertiary">Views</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-text-tertiary">Unique</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-text-tertiary">Avg Time</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-text-tertiary">Bounce</th>
            </tr>
          </thead>
          <tbody>
            {data.page_views.map(pv => (
              <tr key={pv.path} className="border-b border-border-subtle">
                <td className="px-3 py-2 text-sm font-medium text-text-primary">{pv.path}</td>
                <td className="px-3 py-2 text-right text-sm text-text-secondary">{formatNumber(pv.views)}</td>
                <td className="px-3 py-2 text-right text-sm text-text-secondary">{formatNumber(pv.unique_visitors)}</td>
                <td className="px-3 py-2 text-right text-sm text-text-secondary">{pv.avg_time_on_page}s</td>
                <td className="px-3 py-2 text-right text-sm text-text-secondary">{pv.bounce_rate}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
