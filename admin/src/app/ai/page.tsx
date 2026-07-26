import { useAIHealth } from '@/lib/hooks/useAIHealth'
import { useAutoRefresh } from '@/lib/hooks/useAutoRefresh'
import { StatCard } from '@/components/data/StatCard'
import { SkeletonCard, SkeletonTableRows } from '@/components/ui/Skeleton'
import { ErrorBanner, EmptyState } from '@/components/ui/Feedback'
import { formatNumber } from '@/lib/utils'
import { Brain, TrendingUp, MessageSquare, AlertTriangle, HelpCircle } from 'lucide-react'
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const SENTIMENT_COLORS = { positive: '#22c55e', neutral: '#666', negative: '#ef4444' }

export default function AIHealthPage() {
  const { data, loading, error, refetch } = useAIHealth()
  useAutoRefresh(refetch, 30000)

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">AI Health</h1>
          <p className="text-sm text-text-tertiary">AI accuracy, performance, and knowledge gaps</p>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => <SkeletonCard key={i} />)}
        </div>
        <div className="grid grid-cols-3 gap-4">
          <div className="col-span-2 rounded-xl border border-border bg-bg-surface p-5 h-[300px] animate-pulse" />
          <div className="rounded-xl border border-border bg-bg-surface p-5 h-[300px] animate-pulse" />
        </div>
        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <table className="w-full"><tbody><SkeletonTableRows rows={5} cols={3} /></tbody></table>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">AI Health</h1>
          <p className="text-sm text-text-tertiary">AI accuracy, performance, and knowledge gaps</p>
        </div>
        <ErrorBanner message={error} onRetry={refetch} />
      </div>
    )
  }

  const sentimentData = [
    { name: 'Positive', value: data.sentiment_breakdown.positive, color: SENTIMENT_COLORS.positive },
    { name: 'Neutral', value: data.sentiment_breakdown.neutral, color: SENTIMENT_COLORS.neutral },
    { name: 'Negative', value: data.sentiment_breakdown.negative, color: SENTIMENT_COLORS.negative },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">AI Health</h1>
        <p className="text-sm text-text-tertiary">AI accuracy, performance, and knowledge gaps</p>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Accuracy" value={`${data.accuracy}%`} change={data.accuracy_trend} changeLabel="vs last week" icon={<Brain className="h-4 w-4" />} />
        <StatCard label="Total Queries" value={formatNumber(data.total_queries)} icon={<MessageSquare className="h-4 w-4" />} />
        <StatCard label="Answered" value={formatNumber(data.answered_correctly)} icon={<TrendingUp className="h-4 w-4" />} />
        <StatCard label="Unanswered" value={data.unanswered_questions.length} icon={<AlertTriangle className="h-4 w-4" />} />
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div className="col-span-2 rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Accuracy Trend</h3>
          <ResponsiveContainer width="100%" height={240}>
            <AreaChart data={data.accuracy_history}>
              <defs>
                <linearGradient id="aiGreen" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#22c55e" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#22c55e" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#666' }} axisLine={false} tickLine={false} />
              <YAxis domain={[85, 100]} tick={{ fontSize: 12, fill: '#666' }} axisLine={false} tickLine={false} tickFormatter={v => `${v}%`} />
              <Tooltip contentStyle={{ backgroundColor: '#1c1c1c', border: '1px solid #2a2a2a', borderRadius: '8px', fontSize: 12 }} formatter={(v) => [`${String(v)}%`, 'Accuracy']} />
              <Area type="monotone" dataKey="accuracy" stroke="#22c55e" fill="url(#aiGreen)" strokeWidth={2} />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Sentiment</h3>
          <div className="flex items-center justify-center">
            <ResponsiveContainer width={160} height={160}>
              <PieChart>
                <Pie data={sentimentData} dataKey="value" cx="50%" cy="50%" innerRadius={45} outerRadius={65}>
                  {sentimentData.map((entry, i) => <Cell key={i} fill={entry.color} />)}
                </Pie>
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 space-y-2">
            {sentimentData.map(s => (
              <div key={s.name} className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full" style={{ backgroundColor: s.color }} />
                  <span className="text-text-secondary">{s.name}</span>
                </div>
                <span className="font-medium text-text-primary">{s.value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-bg-surface p-5">
        <h3 className="mb-4 text-sm font-medium text-text-secondary">Top Unanswered Questions (Knowledge Gaps)</h3>
        <p className="mb-3 text-xs text-text-tertiary">These are questions customers ask that the AI cannot answer. Update your knowledge base to fix these.</p>
        {data.unanswered_questions.length === 0 ? (
          <EmptyState
            icon={HelpCircle}
            title="No unanswered questions"
            description="The AI is answering all questions successfully."
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border">
                  <th className="px-3 py-2 text-left text-xs font-medium text-text-tertiary">Question</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-text-tertiary">Count</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-text-tertiary">Last Seen</th>
                </tr>
              </thead>
              <tbody>
                {data.unanswered_questions.map((q, i) => (
                  <tr key={i} className="border-b border-border-subtle">
                    <td className="px-3 py-2 text-sm text-text-primary">{q.question}</td>
                    <td className="px-3 py-2 text-right text-sm font-medium text-warning">{q.count}</td>
                    <td className="px-3 py-2 text-right text-xs text-text-tertiary">{new Date(q.last_seen).toLocaleDateString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
