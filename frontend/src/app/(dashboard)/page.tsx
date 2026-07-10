import { useEffect } from 'react'
import { useAPI } from '@/hooks/useAPI'
import { StatCard, StatGrid } from '@/components/stats'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Link } from 'react-router-dom'
import { ArrowRight, MessageCircle, Instagram, Globe, Inbox, Radio, GraduationCap, BarChart3 } from 'lucide-react'
import { StatSkeleton } from '@/components/ui/Skeleton'
import { Avatar } from '@/components/ui/Avatar'
import { timeAgo } from '@/lib/utils'
import type { Conversation } from '@/types'

const channelIcons: Record<string, { icon: React.ElementType; color: string }> = {
  whatsapp: { icon: MessageCircle, color: '#25D366' },
  instagram: { icon: Instagram, color: '#E4405F' },
  facebook: { icon: MessageCircle, color: '#1877F2' },
  telegram: { icon: MessageCircle, color: '#0088CC' },
  discord: { icon: MessageCircle, color: '#5865F2' },
  web: { icon: Globe, color: '#0ea5e9' },
}

export default function OverviewPage() {
  const statsAPI = useAPI() as any
  const { data: stats, get: getStats, loading: statsLoading } = statsAPI
  
  const convAPI = useAPI() as any
  const { data: conversations, get: getConversations, loading: convLoading } = convAPI

  useEffect(() => {
    getStats('/analytics/overview')
    getConversations('/chats/conversations?limit=6')
  }, [])

  const recentCount = conversations?.conversations?.length || 0

  return (
    <div className="animate-page-in space-y-5 lg:space-y-6 pt-2">
      {/* Stats */}
      <div className="px-1">
        {statsLoading ? (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <StatSkeleton />
            <StatSkeleton />
            <StatSkeleton />
            <StatSkeleton />
          </div>
        ) : (
          <StatGrid>
            <StatCard label="Conversations today" value={stats?.conversations_today || 0} change={12} variant="default" />
            <StatCard label="Resolved auto" value={stats?.resolved_today || 0} change={8} variant="success" />
            <StatCard label="Avg response" value={`${stats?.avg_response_time || 0}s`} change={-15} variant="info" />
            <StatCard label="Satisfaction" value={`${stats?.satisfaction || 0}%`} change={5} variant="warning" />
          </StatGrid>
        )}
      </div>

      {/* Quick actions */}
      <div className="px-1">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          <QuickAction to="/chats" icon={Inbox} label="Inbox" color="#0ea5e9" />
          <QuickAction to="/channels" icon={Radio} label="Channels" color="#25D366" />
          <QuickAction to="/teach" icon={GraduationCap} label="Teach" color="#E4405F" />
          <QuickAction to="/insights" icon={BarChart3} label="Insights" color="#5865F2" />
        </div>
      </div>

      {/* Recent activity with count */}
      <div className="px-1 pb-4">
        <Card className="overflow-hidden animate-breathe">
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <div className="flex items-center gap-2">
              <CardTitle>Recent activity</CardTitle>
              {recentCount > 0 && (
                <Badge variant="sky" className="text-[10px] px-1.5 py-0 animate-pulse-gentle">
                  {recentCount}
                </Badge>
              )}
            </div>
            <Link to="/chats" className="hidden sm:block">
              <Button variant="ghost" size="sm">
                See all <ArrowRight className="w-4 h-4 ml-1" />
              </Button>
            </Link>
          </CardHeader>
          <CardBody className="p-0">
            {convLoading ? (
              <div className="p-3 lg:p-4 space-y-2 lg:space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="animate-shimmer-slow h-14 lg:h-12 rounded-xl lg:rounded w-full" />
                ))}
              </div>
            ) : conversations?.conversations?.length === 0 ? (
              <EmptyState />
            ) : (
              <>
                {/* Desktop table */}
                <div className="hidden lg:block">
                  <table className="w-full">
                    <thead>
                      <tr className="text-left text-[11px] font-semibold uppercase tracking-wider text-tertiary bg-inset">
                        <th className="px-4 py-3 rounded-l-lg">Customer</th>
                        <th className="px-4 py-3">Channel</th>
                        <th className="px-4 py-3">Status</th>
                        <th className="px-4 py-3 rounded-r-lg text-right">Time</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(conversations?.conversations || []).map((c: Conversation) => (
                        <tr key={c.id} className="border-b border-subtle hover:bg-surface-hover transition-colors">
                          <td className="px-4 py-3">
                            <div className="flex items-center gap-3">
                              <Avatar name={c.customer_name} size="sm" />
                              <span className="text-sm font-medium text-primary">{c.customer_name}</span>
                            </div>
                          </td>
                          <td className="px-4 py-3">
                            <Badge variant="sky">{c.channel}</Badge>
                          </td>
                          <td className="px-4 py-3">
                            <Badge variant={c.status === 'resolved' ? 'success' : c.status === 'escalated' ? 'error' : 'warning'}>
                              {c.status}
                            </Badge>
                          </td>
                          <td className="px-4 py-3 text-sm text-secondary text-right">{timeAgo(c.updated_at)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* Mobile � Instagram-style cards */}
                <div className="lg:hidden divide-y divide-subtle">
                  {(conversations?.conversations || []).map((c: Conversation) => {
                    const ch = channelIcons[c.channel] || channelIcons.web
                    const Icon = ch.icon
                    return (
                      <Link
                        key={c.id}
                        to={`/chats?id=${c.id}`}
                        className="flex items-center gap-3 p-3 hover:bg-surface-hover active:bg-surface-hover transition-colors"
                      >
                        <Avatar 
                          name={c.customer_name} 
                          size="md" 
                          showChannel 
                          channelColor={ch.color} 
                          channelIcon={<Icon className="w-2 h-2" strokeWidth={3} />} 
                        />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center justify-between">
                            <span className="font-semibold text-sm text-primary truncate">{c.customer_name}</span>
                            <span className="text-[11px] text-tertiary shrink-0 ml-2">{timeAgo(c.updated_at)}</span>
                          </div>
                          <div className="flex items-center gap-2 mt-0.5">
                            <Badge 
                              variant={c.status === 'resolved' ? 'success' : c.status === 'escalated' ? 'error' : 'warning'}
                              className="text-[9px] px-1 py-0"
                            >
                              {c.status}
                            </Badge>
                            <span className="text-[10px] text-secondary capitalize">{c.channel}</span>
                          </div>
                        </div>
                      </Link>
                    )
                  })}
                </div>
              </>
            )}
          </CardBody>
        </Card>
      </div>
    </div>
  )
}

function QuickAction({ to, icon: Icon, label, color }: { to: string; icon: React.ElementType; label: string; color: string }) {
  return (
    <Link 
      to={to}
      className="flex items-center gap-3 p-3 lg:p-4 bg-surface border border-default rounded-2xl hover:border-noant-sky hover:shadow-sm transition-all active:scale-95 btn-press"
    >
      <div className="w-9 h-9 lg:w-10 lg:h-10 rounded-xl flex items-center justify-center text-white shrink-0" style={{ background: color }}>
        <Icon className="w-4 h-4 lg:w-5 lg:h-5" strokeWidth={2} />
      </div>
      <span className="text-sm font-semibold text-primary">{label}</span>
    </Link>
  )
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-12 lg:py-16 text-center px-4 col-span-full">
      <div className="w-14 h-14 lg:w-16 lg:h-16 bg-inset rounded-2xl flex items-center justify-center mb-4 animate-float">
        <svg className="w-7 h-7 lg:w-8 lg:h-8 text-tertiary" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
        </svg>
      </div>
      <p className="text-base lg:text-lg font-semibold text-primary mb-1">No conversations yet</p>
      <p className="text-sm text-secondary max-w-xs mb-6">
        Your Noant is ready to chat. Connect a channel to start receiving messages.
      </p>
      <Link to="/channels">
        <Button size="sm">Connect a channel</Button>
      </Link>
    </div>
  )
}
