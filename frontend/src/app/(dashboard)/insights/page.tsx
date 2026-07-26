import { useEffect, useMemo } from 'react'
import { useAPI } from '@/hooks/useAPI'
import {
  StatCard, StatGrid, TrendChart, PeakHoursChart,
  ChannelDistributionChart, MetricRow,
} from '@/components/stats'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Skeleton, StatSkeleton } from '@/components/ui/Skeleton'
import { BarChart3, TrendingUp, Clock, PieChart as PieChartIcon, MessageCircleQuestion } from 'lucide-react'
import type {
  AnalyticsOverview, TrendsResponse, InsightsResponse,
  ChannelAnalyticsResponse, CSATResponse, UnknownQuestionsStatsResponse,
  PopularQuestionsResponse, MessageTrendResponse,
} from '@/types/api'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  AreaChart, Area, PieChart, Pie, Cell,
} from 'recharts'

const CSAT_COLORS = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#16a34a']

function SkeletonBox({ h = 'h-[200px]' }: { h?: string }) {
  return (
    <div className={`${h} flex items-center justify-center`}>
      <Skeleton className={`${h.replace('h-', 'h-[').replace('px', 'px]')} w-full rounded-lg`} />
    </div>
  )
}

function CSATDonut({ distribution }: { distribution?: Record<string, number> }) {
  const data = useMemo(() => {
    if (!distribution) return []
    return Object.entries(distribution).map(([score, count]) => ({
      name: `${score} Star`,
      value: count,
      color: CSAT_COLORS[parseInt(score) - 1] || '#6b7280',
    })).filter(d => d.value > 0)
  }, [distribution])

  if (data.length === 0) {
    return <p className="text-xs text-tertiary text-center py-12">No ratings yet</p>
  }

  return (
    <ResponsiveContainer width="100%" height={200}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          innerRadius={50}
          outerRadius={80}
          paddingAngle={2}
          dataKey="value"
        >
          {data.map((entry, i) => (
            <Cell key={i} fill={entry.color} stroke="transparent" />
          ))}
        </Pie>
        <Tooltip
          contentStyle={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-default)',
            borderRadius: '8px',
            fontSize: '12px',
          }}
        />
      </PieChart>
    </ResponsiveContainer>
  )
}

function CSATScoreCard({ avg_score, total_ratings }: { avg_score?: number; total_ratings?: number }) {
  const pct = avg_score ? Math.round((avg_score / 5) * 100) : 0
  return (
    <div className="flex flex-col items-center justify-center py-4">
      <div className="text-4xl font-bold text-primary">{avg_score ? avg_score.toFixed(1) : '—'}</div>
      <div className="flex gap-1 mt-1">
        {[1, 2, 3, 4, 5].map(s => (
          <span key={s} className={`text-lg ${s <= Math.round(avg_score || 0) ? 'text-amber-400' : 'text-border'}`}>★</span>
        ))}
      </div>
      <p className="text-xs text-tertiary mt-2">{total_ratings || 0} total ratings</p>
      <div className="w-full bg-border rounded-full h-2 mt-3 max-w-[160px]">
        <div className="bg-primary h-2 rounded-full transition-all" style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}

export default function InsightsPage() {
  const { data: overview, get: getOverview, loading: ovLoading } = useAPI<AnalyticsOverview>()

  const { data: trends, get: getTrends, loading: trLoading } = useAPI<TrendsResponse>()

  const { data: insights, get: getInsights, loading: inLoading } = useAPI<InsightsResponse>()

  const { data: channels, get: getChannels, loading: chLoading } = useAPI<ChannelAnalyticsResponse>()

  const { data: csat, get: getCSAT, loading: csatLoading } = useAPI<CSATResponse>()

  const { data: unknownStats, get: getUQ, loading: uqLoading } = useAPI<UnknownQuestionsStatsResponse>()

  const { data: popular, get: getPopular, loading: popLoading } = useAPI<PopularQuestionsResponse>()

  const { data: msgTrend, get: getMsgTrend, loading: msgLoading } = useAPI<MessageTrendResponse>()

  useEffect(() => {
    getOverview('/analytics/overview')
    getTrends('/analytics/trends?days=7')
    getInsights('/analytics/insights')
    getChannels('/analytics/channels')
    getCSAT('/analytics/satisfaction')
    getUQ('/analytics/unknown-questions')
    getPopular('/analytics/popular-questions')
    getMsgTrend('/analytics/messages-trend')
  }, [])

  const msgChartData = useMemo(() => {
    if (!msgTrend?.trends) return []
    return msgTrend.trends.map((d) => ({ ...d, dateShort: d.date.slice(-5) }))
  }, [msgTrend])

  const byStatus: { pending?: number; trained?: number; ignored?: number } = unknownStats?.by_status || {}

  return (
    <div className="animate-fade-in space-y-5 lg:space-y-6 pt-2">
      {/* Stats row */}
      <div className="px-1">
        {ovLoading ? (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <StatSkeleton /><StatSkeleton /><StatSkeleton /><StatSkeleton />
          </div>
        ) : (
          <StatGrid>
            <StatCard label="Total conversations" value={overview?.total_conversations || 0} variant="default" />
            <StatCard label="Active channels" value={Object.keys(channels?.distribution || {}).length} variant="info" />
            <StatCard
              label="CSAT Score"
              value={csat?.avg_score ? csat.avg_score.toFixed(1) : '—'}
              variant="success"
            />
            <StatCard
              label="Unknown questions"
              value={byStatus.pending || 0}
              variant="warning"
            />
          </StatGrid>
        )}
      </div>

      {/* CSAT section */}
      <div className="px-1 grid grid-cols-1 lg:grid-cols-3 gap-5">
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Customer Satisfaction</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4">
            {csatLoading ? <SkeletonBox h="h-[200px]" /> : <CSATScoreCard avg_score={csat?.avg_score} total_ratings={csat?.total_ratings} />}
          </CardBody>
        </Card>
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Score Distribution</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4">
            {csatLoading ? <SkeletonBox h="h-[200px]" /> : <CSATDonut distribution={csat?.distribution} />}
          </CardBody>
        </Card>
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Unknown Questions</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4 space-y-4">
            {uqLoading ? (
              <SkeletonBox h="h-[200px]" />
            ) : (
              <>
                <div className="flex items-center justify-around text-center">
                  <div>
                    <div className="text-2xl font-bold text-red-500">{byStatus.pending || 0}</div>
                    <div className="text-xs text-tertiary">Pending</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-green-500">{byStatus.trained || 0}</div>
                    <div className="text-xs text-tertiary">Trained</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-tertiary">{byStatus.ignored || 0}</div>
                    <div className="text-xs text-tertiary">Ignored</div>
                  </div>
                </div>
                <div className="w-full bg-border rounded-full h-2.5">
                  <div className="flex h-2.5 rounded-full overflow-hidden">
                    <div className="bg-red-500 transition-all" style={{ width: `${unknownStats?.total ? ((byStatus.pending || 0) / unknownStats.total) * 100 : 0}%` }} />
                    <div className="bg-green-500 transition-all" style={{ width: `${unknownStats?.total ? ((byStatus.trained || 0) / unknownStats.total) * 100 : 0}%` }} />
                    <div className="bg-tertiary/30 transition-all" style={{ width: `${unknownStats?.total ? ((byStatus.ignored || 0) / unknownStats.total) * 100 : 0}%` }} />
                  </div>
                </div>
              </>
            )}
          </CardBody>
        </Card>
      </div>

      {/* CSAT trend + Messages volume */}
      <div className="px-1 grid grid-cols-1 lg:grid-cols-2 gap-5">
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>CSAT Trend (30d)</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4">
            {csatLoading ? (
              <SkeletonBox h="h-[200px]" />
            ) : (csat?.trend || []).length === 0 ? (
              <div className="h-[200px] flex flex-col items-center justify-center text-center">
                <TrendingUp className="w-8 h-8 text-tertiary/40 mb-2" />
                <p className="text-xs text-tertiary">No data yet</p>
              </div>
            ) : (
              <div className="h-[200px]">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={csat?.trend || []} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <defs>
                      <linearGradient id="colorCsat" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#22c55e" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#22c55e" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--border-default)" vertical={false} />
                    <XAxis dataKey="date" axisLine={false} tickLine={false} tick={{ fill: 'var(--text-tertiary)', fontSize: 11 }} dy={8} tickFormatter={(v: string) => v?.slice(-5) || ''} />
                    <YAxis domain={[0, 5]} axisLine={false} tickLine={false} tick={{ fill: 'var(--text-tertiary)', fontSize: 11 }} />
                    <Tooltip contentStyle={{ background: 'var(--bg-surface)', border: '1px solid var(--border-default)', borderRadius: '8px', fontSize: '12px' }} />
                    <Area type="monotone" dataKey="avg_score" stroke="#22c55e" strokeWidth={2} fill="url(#colorCsat)" activeDot={{ r: 4 }} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardBody>
        </Card>
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Message Volume (7d)</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4">
            {msgLoading ? (
              <SkeletonBox h="h-[200px]" />
            ) : msgChartData.length === 0 ? (
              <div className="h-[200px] flex flex-col items-center justify-center text-center">
                <BarChart3 className="w-8 h-8 text-tertiary/40 mb-2" />
                <p className="text-xs text-tertiary">No data yet</p>
              </div>
            ) : (
              <div className="h-[200px]">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={msgChartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--border-default)" vertical={false} />
                    <XAxis dataKey="dateShort" axisLine={false} tickLine={false} tick={{ fill: 'var(--text-tertiary)', fontSize: 11 }} dy={8} />
                    <YAxis axisLine={false} tickLine={false} tick={{ fill: 'var(--text-tertiary)', fontSize: 11 }} />
                    <Tooltip contentStyle={{ background: 'var(--bg-surface)', border: '1px solid var(--border-default)', borderRadius: '8px', fontSize: '12px' }} />
                    <Bar dataKey="messages" fill="#0ea5e9" radius={[4, 4, 0, 0]} maxBarSize={40} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardBody>
        </Card>
      </div>

      {/* Conversation trends + Top intents */}
      <div className="px-1 grid grid-cols-1 lg:grid-cols-3 gap-5">
        <div className="lg:col-span-2">
          <Card>
            <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
              <CardTitle>Conversation trends</CardTitle>
            </CardHeader>
            <CardBody className="p-3 lg:p-4">
              {trLoading ? <SkeletonBox /> : (trends?.trends || []).length === 0 ? (
                <div className="h-[200px] lg:h-[260px] flex flex-col items-center justify-center text-center">
                  <TrendingUp className="w-8 h-8 text-tertiary/40 mb-2" />
                  <p className="text-xs text-tertiary">No data yet</p>
                </div>
              ) : (
                <div className="h-[200px] lg:h-[260px]">
                  <TrendChart data={trends?.trends || []} />
                </div>
              )}
            </CardBody>
          </Card>
        </div>
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Top intents</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4 space-y-0">
            {inLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="flex items-center justify-between py-2">
                    <div className="flex items-center gap-3">
                      <Skeleton className="w-8 h-8 rounded-md" />
                      <Skeleton className="h-4 w-24 rounded" />
                    </div>
                    <Skeleton className="h-5 w-6 rounded" />
                  </div>
                ))}
              </div>
            ) : (insights?.top_intents || []).length === 0 ? (
              <div className="py-8 text-center">
                <MessageCircleQuestion className="w-8 h-8 text-tertiary/40 mx-auto mb-2" />
                <p className="text-xs text-tertiary">No data yet</p>
              </div>
            ) : (
              (insights?.top_intents || []).map((item, i) => (
                <MetricRow
                  key={i}
                  icon={['$', '#', '?', '!', '*'][i] || '?'}
                  iconBg={['#dbeafe', '#dcfce7', '#fef3c7', '#fee2e2', '#f3e8ff'][i] || '#f3f4f6'}
                  iconColor={['#1d4ed8', '#15803d', '#b45309', '#b91c1c', '#7c3aed'][i] || '#6b7280'}
                  name={item.intent}
                  value={item.count}
                />
              ))
            )}
          </CardBody>
        </Card>
      </div>

      {/* Peak hours + Channel + Popular questions */}
      <div className="px-1 pb-4 grid grid-cols-1 lg:grid-cols-3 gap-5">
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Peak hours</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4">
            {inLoading ? <SkeletonBox h="h-[180px]" /> : (insights?.peak_hours || []).length === 0 ? (
              <div className="h-[180px] lg:h-[200px] flex flex-col items-center justify-center text-center">
                <Clock className="w-8 h-8 text-tertiary/40 mb-2" />
                <p className="text-xs text-tertiary">No data yet</p>
              </div>
            ) : (
              <div className="h-[180px] lg:h-[200px]">
                <PeakHoursChart data={insights?.peak_hours || []} />
              </div>
            )}
          </CardBody>
        </Card>
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Channel distribution</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4">
            {chLoading ? <SkeletonBox h="h-[180px]" /> : Object.keys(channels?.distribution || {}).length === 0 ? (
              <div className="h-[180px] lg:h-[200px] flex flex-col items-center justify-center text-center">
                <PieChartIcon className="w-8 h-8 text-tertiary/40 mb-2" />
                <p className="text-xs text-tertiary">No data yet</p>
              </div>
            ) : (
              <div className="h-[180px] lg:h-[200px] flex items-center justify-center">
                <ChannelDistributionChart data={channels?.distribution || {}} />
              </div>
            )}
          </CardBody>
        </Card>
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Top unanswered questions</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4 space-y-2">
            {popLoading ? (
              <SkeletonBox h="h-[180px]" />
            ) : (popular?.questions || []).length === 0 ? (
              <p className="text-xs text-tertiary text-center py-8">No unanswered questions</p>
            ) : (
              (popular?.questions || []).slice(0, 6).map((q, i) => (
                <div key={i} className="flex items-center justify-between gap-3 py-1.5 border-b border-border last:border-0">
                  <span className="text-xs truncate flex-1 text-secondary">{q.question}</span>
                  <span className="text-xs font-semibold text-tertiary shrink-0 ml-2">×{q.count}</span>
                </div>
              ))
            )}
          </CardBody>
        </Card>
      </div>
    </div>
  )
}
