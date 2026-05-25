import { useEffect } from 'react'
import { useAPI } from '@/hooks/useAPI'
import { StatCard, StatGrid, TrendChart, PeakHoursChart, ChannelDistributionChart, MetricRow } from '@/components/stats'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Skeleton, StatSkeleton } from '@/components/ui/Skeleton'

export default function InsightsPage() {
  const overviewAPI = useAPI() as any
  const { data: overview, get: getOverview, loading: ovLoading } = overviewAPI
  
  const trendsAPI = useAPI() as any
  const { data: trends, get: getTrends, loading: trLoading } = trendsAPI
  
  const insightsAPI = useAPI() as any
  const { data: insights, get: getInsights, loading: inLoading } = insightsAPI
  
  const channelsAPI = useAPI() as any
  const { data: channels, get: getChannels, loading: chLoading } = channelsAPI

  useEffect(() => {
    getOverview('/analytics/overview')
    getTrends('/analytics/trends?days=7')
    getInsights('/analytics/insights')
    getChannels('/analytics/channels')
  }, [])

  return (
    <div className="animate-fade-in space-y-5 lg:space-y-6 pt-2">
      {/* Stats */}
      <div className="px-1">
        {ovLoading ? (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <StatSkeleton />
            <StatSkeleton />
            <StatSkeleton />
            <StatSkeleton />
          </div>
        ) : (
          <StatGrid>
            <StatCard label="Total conversations" value={overview?.total_conversations || 0} variant="default" />
            <StatCard label="Active channels" value={Object.keys(channels?.distribution || {}).length} variant="info" />
            <StatCard label="Languages used" value={500} variant="success" />
            <StatCard label="Uptime this month" value="99.9%" variant="warning" />
          </StatGrid>
        )}
      </div>

      {/* Trends + Top intents */}
      <div className="px-1 grid grid-cols-1 lg:grid-cols-3 gap-5">
        <div className="lg:col-span-2">
          <Card>
            <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
              <CardTitle>Conversation trends</CardTitle>
            </CardHeader>
            <CardBody className="p-3 lg:p-4">
              {trLoading ? (
                <div className="h-[200px] lg:h-[260px] flex items-center justify-center">
                  <Skeleton className="h-[180px] lg:h-[240px] w-full rounded-lg" />
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
            ) : (
              (insights?.top_intents || []).map((item: any, i: number) => (
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

      {/* Peak hours + Channel distribution */}
      <div className="px-1 pb-4 grid grid-cols-1 lg:grid-cols-2 gap-5">
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Peak hours</CardTitle>
          </CardHeader>
          <CardBody className="p-3 lg:p-4">
            {inLoading ? (
              <div className="h-[180px] lg:h-[220px] flex items-center justify-center">
                <Skeleton className="h-[160px] lg:h-[200px] w-full rounded-lg" />
              </div>
            ) : (
              <div className="h-[180px] lg:h-[220px]">
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
            {chLoading ? (
              <div className="h-[180px] lg:h-[220px] flex items-center justify-center">
                <Skeleton className="h-[160px] lg:h-[200px] w-[160px] lg:w-[200px] rounded-full" />
              </div>
            ) : (
              <div className="h-[180px] lg:h-[220px] flex items-center justify-center">
                <ChannelDistributionChart data={channels?.distribution || {}} />
              </div>
            )}
          </CardBody>
        </Card>
      </div>
    </div>
  )
}