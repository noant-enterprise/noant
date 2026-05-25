cd C:\Users\USER\Downloads\omagent\frontend

# ============================================
# 1. Setup Page — Mobile tabs, single-col form, overflow fixes
# ============================================
$setup = @'
import { useEffect, useState } from 'react'
import { useAPI } from '@/hooks/useAPI'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { useToast } from '@/components/ui/Toast'
import { cn } from '@/lib/utils'
import type { User, TeamMember, Plan, APIKey } from '@/types'

export default function SetupPage() {
  const [tab, setTab] = useState<'profile' | 'team' | 'billing' | 'api'>('profile')
  const { data: profile, get: getProfile, loading: profLoading } = useAPI<User>()
  const { data: teamData, get: getTeam, loading: teamLoading } = useAPI<{ members: TeamMember[] }>()
  const { data: plans, get: getPlans, loading: planLoading } = useAPI<{ plans: Plan[] }>()
  const { data: keys, get: getKeys, loading: keyLoading } = useAPI<{ api_keys: APIKey[] }>()
  const { put, post, del } = useAPI()
  const { toast } = useToast()

  useEffect(() => {
    if (tab === 'profile') getProfile('/settings/profile')
    if (tab === 'team') getTeam('/settings/team')
    if (tab === 'billing') getPlans('/payments/plans')
    if (tab === 'api') getKeys('/settings/api-keys')
  }, [tab])

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const form = e.target as HTMLFormElement
      await put('/settings/profile', {
        first_name: (form.elements.namedItem('first') as HTMLInputElement).value,
        last_name: (form.elements.namedItem('last') as HTMLInputElement).value,
        company_name: (form.elements.namedItem('company') as HTMLInputElement).value,
        phone: (form.elements.namedItem('phone') as HTMLInputElement).value,
      })
      toast('Profile saved', 'success')
    } catch {
      toast('Failed to save profile', 'error')
    }
  }

  const handleInvite = async () => {
    const email = prompt('Email:')
    if (!email) return
    try {
      await post('/settings/team/invite', { email, role: 'agent' })
      toast('Invitation sent', 'success')
      getTeam('/settings/team')
    } catch {
      toast('Failed to send invitation', 'error')
    }
  }

  const handleRemove = async (id: string) => {
    if (!confirm('Remove team member?')) return
    try {
      await del(`/settings/team/${id}`)
      toast('Team member removed', 'success')
      getTeam('/settings/team')
    } catch {
      toast('Failed to remove member', 'error')
    }
  }

  const handleGenerateKey = async () => {
    const name = prompt('Key name:') || 'New Key'
    try {
      await post('/settings/api-keys', { name })
      toast('API key generated', 'success')
      getKeys('/settings/api-keys')
    } catch {
      toast('Failed to generate key', 'error')
    }
  }

  const handleRevoke = async (id: string) => {
    if (!confirm('Revoke this key?')) return
    try {
      await del(`/settings/api-keys/${id}`)
      toast('API key revoked', 'success')
      getKeys('/settings/api-keys')
    } catch {
      toast('Failed to revoke key', 'error')
    }
  }

  const members = teamData?.members || []
  const planList = plans?.plans || []
  const keyList = keys?.api_keys || []

  return (
    <div className="max-w-3xl mx-auto animate-fade-in px-0 lg:px-0">
      {/* Mobile tabs — horizontal scroll */}
      <div className="flex gap-0 border-b border-default mb-6 overflow-x-auto scrollbar-hide">
        {(['profile', 'team', 'billing', 'api'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={cn(
              'px-4 lg:px-5 py-3 text-sm font-semibold border-b-2 -mb-px transition-colors capitalize shrink-0 whitespace-nowrap',
              tab === t ? 'text-noant-sky-deep border-noant-sky' : 'text-tertiary border-transparent hover:text-secondary'
            )}
          >
            {t === 'api' ? 'API Keys' : t}
          </button>
        ))}
      </div>

      {tab === 'profile' && (
        <Card>
          <CardHeader><CardTitle>Your profile</CardTitle></CardHeader>
          <CardBody>
            {profLoading ? <ProfileSkeleton /> : (
              <form onSubmit={handleSaveProfile} className="space-y-5">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
                  <div>
                    <label className="block text-sm font-semibold text-primary mb-2">First name</label>
                    <Input name="first" defaultValue={profile?.first_name} />
                  </div>
                  <div>
                    <label className="block text-sm font-semibold text-primary mb-2">Last name</label>
                    <Input name="last" defaultValue={profile?.last_name} />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-semibold text-primary mb-2">Email</label>
                  <Input defaultValue={profile?.email} disabled />
                </div>
                <div>
                  <label className="block text-sm font-semibold text-primary mb-2">Company</label>
                  <Input name="company" defaultValue={profile?.company_name} />
                </div>
                <div>
                  <label className="block text-sm font-semibold text-primary mb-2">Phone</label>
                  <Input name="phone" defaultValue={profile?.phone} />
                </div>
                <Button type="submit">Save changes</Button>
              </form>
            )}
          </CardBody>
        </Card>
      )}

      {tab === 'team' && (
        <Card>
          <CardHeader>
            <CardTitle>Team members</CardTitle>
            <Button size="sm" onClick={handleInvite}>+ Invite</Button>
          </CardHeader>
          <CardBody className="p-0">
            {teamLoading ? <TeamSkeleton /> : members.length === 0 ? (
              <EmptyTeam />
            ) : (
              members.map((m) => (
                <div key={m.id} className="flex items-center justify-between p-4 border-b border-subtle last:border-b-0 gap-3">
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="w-10 h-10 rounded-full bg-noant-black text-white flex items-center justify-center font-semibold shrink-0">
                      {m.first_name[0]}{m.last_name[0]}
                    </div>
                    <div className="min-w-0">
                      <p className="font-semibold text-sm text-primary truncate">{m.first_name} {m.last_name}</p>
                      <p className="text-xs text-secondary truncate">{m.email}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <Badge variant="sky">{m.role}</Badge>
                    <Button variant="ghost" size="sm" className="text-red-600 px-2" onClick={() => handleRemove(m.id)}>Remove</Button>
                  </div>
                </div>
              ))
            )}
          </CardBody>
        </Card>
      )}

      {tab === 'billing' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {planLoading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="rounded-xl border border-default bg-surface p-6 space-y-4">
                <Skeleton className="h-6 w-24 mx-auto rounded" />
                <Skeleton className="h-10 w-32 mx-auto rounded" />
                <div className="space-y-2">
                  {Array.from({ length: 4 }).map((_, j) => (
                    <Skeleton key={j} className="h-4 w-full rounded" />
                  ))}
                </div>
                <Skeleton className="h-10 w-full rounded-lg" />
              </div>
            ))
          ) : planList.length === 0 ? (
            <EmptyPlans />
          ) : (
            planList.map((p) => (
              <Card key={p.id} className={`p-6 text-center ${p.is_popular ? 'border-2 border-noant-sky shadow-sky' : ''}`}>
                {p.is_popular && <span className="inline-block px-3 py-1 bg-noant-sky text-white text-xs font-semibold rounded-full -mt-8 mb-4">Most popular</span>}
                <p className="text-2xl font-bold text-primary mb-1">{p.name}</p>
                <p className="text-3xl font-bold text-primary mb-4">₦{p.price_ngn.toLocaleString()}<span className="text-base font-normal text-secondary">/month</span></p>
                <ul className="text-left space-y-2 mb-6">
                  {p.features.map((f, i) => <li key={i} className="text-sm text-secondary">✓ {f}</li>)}
                </ul>
                <Button variant={p.is_popular ? 'accent' : 'primary'} className="w-full">{p.price_ngn === 0 ? 'Current' : 'Upgrade'}</Button>
              </Card>
            ))
          )}
        </div>
      )}

      {tab === 'api' && (
        <Card>
          <CardHeader>
            <CardTitle>API Keys</CardTitle>
            <Button size="sm" onClick={handleGenerateKey}>+ Generate</Button>
          </CardHeader>
          <CardBody className="p-0">
            {keyLoading ? <ApiKeySkeleton /> : keyList.length === 0 ? (
              <EmptyApiKeys />
            ) : (
              keyList.map((k) => (
                <div key={k.id} className="flex items-center justify-between p-4 border-b border-subtle last:border-b-0 gap-3">
                  <div className="min-w-0 flex-1">
                    <p className="font-semibold text-sm text-primary">{k.name}</p>
                    <p className="font-mono text-xs text-secondary truncate">{k.key || '••••••••'}</p>
                    <p className="text-xs text-tertiary">Created {new Date(k.created_at).toLocaleDateString()}</p>
                  </div>
                  <Button variant="ghost" size="sm" className="text-red-600 shrink-0" onClick={() => handleRevoke(k.id)}>Revoke</Button>
                </div>
              ))
            )}
          </CardBody>
        </Card>
      )}
    </div>
  )
}

function ProfileSkeleton() {
  return (
    <div className="space-y-5">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <Skeleton className="h-12 rounded-lg" />
        <Skeleton className="h-12 rounded-lg" />
      </div>
      <Skeleton className="h-12 rounded-lg" />
      <Skeleton className="h-12 rounded-lg" />
      <Skeleton className="h-12 rounded-lg" />
      <Skeleton className="h-10 w-32 rounded-lg" />
    </div>
  )
}

function TeamSkeleton() {
  return (
    <div className="divide-y divide-subtle">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="flex items-center justify-between p-4 gap-3">
          <div className="flex items-center gap-3 min-w-0">
            <Skeleton className="w-10 h-10 rounded-full shrink-0" />
            <div className="space-y-1 min-w-0">
              <Skeleton className="h-4 w-32 rounded" />
              <Skeleton className="h-3 w-24 rounded" />
            </div>
          </div>
          <div className="flex gap-2 shrink-0">
            <Skeleton className="h-6 w-16 rounded" />
            <Skeleton className="h-8 w-16 rounded" />
          </div>
        </div>
      ))}
    </div>
  )
}

function ApiKeySkeleton() {
  return (
    <div className="divide-y divide-subtle">
      {Array.from({ length: 2 }).map((_, i) => (
        <div key={i} className="flex items-center justify-between p-4 gap-3">
          <div className="space-y-2 min-w-0 flex-1">
            <Skeleton className="h-4 w-24 rounded" />
            <Skeleton className="h-3 w-48 rounded" />
            <Skeleton className="h-3 w-32 rounded" />
          </div>
          <Skeleton className="h-8 w-16 rounded shrink-0" />
        </div>
      ))}
    </div>
  )
}

function EmptyTeam() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center px-4">
      <p className="text-secondary text-sm">No team members yet. Invite someone to collaborate.</p>
    </div>
  )
}

function EmptyPlans() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center col-span-full px-4">
      <p className="text-secondary text-sm">No plans available.</p>
    </div>
  )
}

function EmptyApiKeys() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center px-4">
      <p className="text-secondary text-sm">No API keys yet. Generate one to get started.</p>
    </div>
  )
}
'@
Set-Content -Path "src\app\(dashboard)\setup\page.tsx" -Value $setup -Encoding UTF8
Write-Host "✓ Setup page"

# ============================================
# 2. Channels Page — Mobile cards, horizontal scroll table
# ============================================
$channels = @'
import { useEffect } from 'react'
import { useAPI } from '@/hooks/useAPI'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { cn } from '@/lib/utils'
import type { Integration } from '@/types'

const channelConfig: Record<string, { name: string; desc: string; color: string }> = {
  telegram: { name: 'Telegram', desc: 'Bot integration', color: '#0088cc' },
  whatsapp: { name: 'WhatsApp', desc: 'Business API', color: '#25D366' },
  instagram: { name: 'Instagram', desc: 'Direct Messages', color: '#E4405F' },
  discord: { name: 'Discord', desc: 'Server bot', color: '#5865F2' },
  web: { name: 'Web Chat', desc: 'Embeddable widget', color: '#0a0a0a' },
}

export default function ChannelsPage() {
  const { data, get: getIntegrations, loading } = useAPI<{ integrations: Integration[] }>()
  const { post } = useAPI()

  useEffect(() => {
    getIntegrations('/integrations/list')
  }, [])

  const handleConnect = async (channel: string) => {
    await post('/integrations/connect', { channel })
    getIntegrations('/integrations/list')
  }

  const handleDisconnect = async (channel: string) => {
    if (!confirm(`Disconnect ${channel}?`)) return
    await post(`/integrations/disconnect/${channel}`)
    getIntegrations('/integrations/list')
  }

  const connected = new Set(data?.integrations?.map((i) => i.channel) || [])

  return (
    <div className="animate-fade-in space-y-6">
      {/* Channel cards — horizontal scroll on mobile, grid on desktop */}
      {loading ? (
        <div className="flex gap-4 overflow-x-auto pb-2 scrollbar-hide lg:grid lg:grid-cols-4 lg:overflow-visible">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="rounded-xl border border-default bg-surface p-5 space-y-4 min-w-[280px] lg:min-w-0">
              <div className="flex items-center gap-4">
                <Skeleton className="w-12 h-12 rounded-md" />
                <div className="space-y-2">
                  <Skeleton className="h-5 w-24 rounded" />
                  <Skeleton className="h-3 w-20 rounded" />
                </div>
              </div>
              <Skeleton className="h-10 w-full rounded-lg" />
            </div>
          ))}
        </div>
      ) : (
        <div className="flex gap-4 overflow-x-auto pb-2 scrollbar-hide lg:grid lg:grid-cols-4 lg:overflow-visible snap-x snap-mandatory">
          {Object.entries(channelConfig).map(([key, cfg]) => {
            const isConnected = connected.has(key)
            return (
              <Card key={key} className="p-5 min-w-[280px] lg:min-w-0 snap-start">
                <div className="flex items-center gap-4 mb-4">
                  <div className="w-12 h-12 rounded-xl flex items-center justify-center text-white text-xl font-bold shrink-0" style={{ background: cfg.color }}>
                    {cfg.name[0]}
                  </div>
                  <div className="min-w-0">
                    <p className="font-semibold text-primary truncate">{cfg.name}</p>
                    <p className="text-xs text-secondary">{cfg.desc}</p>
                  </div>
                </div>
                <Button 
                  variant={isConnected ? 'ghost' : 'accent'} 
                  className={cn("w-full", isConnected && "border border-default")}
                  onClick={() => (isConnected ? handleDisconnect(key) : handleConnect(key))}
                >
                  {isConnected ? 'Disconnect' : 'Connect'}
                </Button>
              </Card>
            )
          })}
        </div>
      )}

      {/* Connected channels table — card list on mobile, table on desktop */}
      <Card>
        <div className="p-4 lg:p-6">
          <h3 className="text-base lg:text-lg font-semibold text-primary mb-4">Connected channels</h3>
          {loading ? (
            <div className="space-y-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="flex items-center justify-between py-3">
                  <Skeleton className="h-4 w-24 rounded" />
                  <div className="flex gap-4">
                    <Skeleton className="h-6 w-20 rounded" />
                    <Skeleton className="h-6 w-24 rounded" />
                  </div>
                </div>
              ))}
            </div>
          ) : data?.integrations?.length === 0 ? (
            <EmptyChannels />
          ) : (
            <>
              {/* Desktop table */}
              <div className="hidden lg:block overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-[11px] font-semibold uppercase tracking-wider text-tertiary bg-inset">
                      <th className="px-4 py-3 rounded-l-lg">Channel</th>
                      <th className="px-4 py-3">Status</th>
                      <th className="px-4 py-3">Connected</th>
                      <th className="px-4 py-3 rounded-r-lg text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data?.integrations.map((i) => (
                      <tr key={i.channel} className="border-b border-subtle hover:bg-surface-hover">
                        <td className="px-4 py-3 text-sm font-medium capitalize">{i.channel}</td>
                        <td className="px-4 py-3">
                          <Badge variant={i.status === 'connected' ? 'success' : 'error'}>{i.status}</Badge>
                        </td>
                        <td className="px-4 py-3 text-sm text-secondary">
                          {i.connected_at ? new Date(i.connected_at).toLocaleDateString() : '-'}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <Button variant="ghost" size="sm" onClick={() => handleDisconnect(i.channel)}>
                            Disconnect
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Mobile card list */}
              <div className="lg:hidden space-y-3">
                {data?.integrations.map((i) => (
                  <div key={i.channel} className="flex items-center justify-between p-3 bg-inset rounded-xl">
                    <div className="flex items-center gap-3">
                      <div 
                        className="w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold shrink-0"
                        style={{ background: channelConfig[i.channel]?.color || '#64748b' }}
                      >
                        {(channelConfig[i.channel]?.name || i.channel)[0]}
                      </div>
                      <div>
                        <p className="font-semibold text-sm text-primary capitalize">{i.channel}</p>
                        <p className="text-xs text-secondary">
                          {i.connected_at ? new Date(i.connected_at).toLocaleDateString() : '-'}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant={i.status === 'connected' ? 'success' : 'error'}>{i.status}</Badge>
                      <Button variant="ghost" size="sm" className="px-2 text-red-600" onClick={() => handleDisconnect(i.channel)}>
                        <span className="sr-only">Disconnect</span>
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                          <path d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </Card>
    </div>
  )
}

function EmptyChannels() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center px-4">
      <div className="w-16 h-16 bg-inset rounded-2xl flex items-center justify-center mb-4">
        <svg className="w-8 h-8 text-tertiary" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
          <path d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" />
        </svg>
      </div>
      <p className="text-lg font-semibold text-primary mb-1">No channels connected</p>
      <p className="text-sm text-secondary max-w-xs">
        Connect your first channel to start receiving customer messages.
      </p>
    </div>
  )
}
'@
Set-Content -Path "src\app\(dashboard)\channels\page.tsx" -Value $channels -Encoding UTF8
Write-Host "✓ Channels page"

# ============================================
# 3. Insights Page — Better chart containers, mobile stacking
# ============================================
$insights = @'
import { useEffect } from 'react'
import { useAPI } from '@/hooks/useAPI'
import { StatCard, StatGrid, TrendChart, PeakHoursChart, ChannelDistributionChart, MetricRow } from '@/components/stats'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Skeleton, StatSkeleton } from '@/components/ui/Skeleton'
import type { OverviewStats, TrendData, IntentData, PeakHourData, ChannelDistribution } from '@/types'

export default function InsightsPage() {
  const { data: overview, get: getOverview, loading: ovLoading } = useAPI<OverviewStats>()
  const { data: trends, get: getTrends, loading: trLoading } = useAPI<{ trends: TrendData[] }>()
  const { data: insights, get: getInsights, loading: inLoading } = useAPI<{ top_intents: IntentData[]; peak_hours: PeakHourData[] }>()
  const { data: channels, get: getChannels, loading: chLoading } = useAPI<{ distribution: ChannelDistribution }>()

  useEffect(() => {
    getOverview('/analytics/overview')
    getTrends('/analytics/trends?days=7')
    getInsights('/analytics/insights')
    getChannels('/analytics/channels')
  }, [])

  return (
    <div className="animate-fade-in space-y-6">
      {/* Stats row — horizontal scroll on mobile */}
      {ovLoading ? (
        <div className="flex gap-3 overflow-x-auto pb-2 scrollbar-hide lg:grid lg:grid-cols-4">
          <StatSkeleton /><StatSkeleton /><StatSkeleton /><StatSkeleton />
        </div>
      ) : (
        <div className="flex gap-3 overflow-x-auto pb-2 scrollbar-hide lg:grid lg:grid-cols-4 snap-x snap-mandatory">
          <StatCard label="Total conversations" value={overview?.total_conversations || 0} variant="default" />
          <StatCard label="Active channels" value={Object.keys(channels?.distribution || {}).length} variant="info" />
          <StatCard label="Languages used" value={500} variant="success" />
          <StatCard label="Uptime this month" value="99.9%" variant="warning" />
        </div>
      )}

      {/* Trends + Top intents */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>Conversation trends</CardTitle>
            </CardHeader>
            <CardBody>
              {trLoading ? (
                <div className="h-[240px] flex items-center justify-center">
                  <Skeleton className="h-[200px] w-full rounded-lg" />
                </div>
              ) : (
                <div className="h-[240px] lg:h-[280px]">
                  <TrendChart data={trends?.trends || []} />
                </div>
              )}
            </CardBody>
          </Card>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Top intents</CardTitle>
          </CardHeader>
          <CardBody className="space-y-1">
            {inLoading ? (
              <div className="space-y-4">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <Skeleton className="w-9 h-9 rounded-md" />
                      <Skeleton className="h-4 w-24 rounded" />
                    </div>
                    <Skeleton className="h-6 w-8 rounded" />
                  </div>
                ))}
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

      {/* Peak hours + Channel distribution */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <Card>
          <CardHeader>
            <CardTitle>Peak hours</CardTitle>
          </CardHeader>
          <CardBody>
            {inLoading ? (
              <div className="h-[200px] flex items-center justify-center">
                <Skeleton className="h-[160px] w-full rounded-lg" />
              </div>
            ) : (
              <div className="h-[200px] lg:h-[240px]">
                <PeakHoursChart data={insights?.peak_hours || []} />
              </div>
            )}
          </CardBody>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Channel distribution</CardTitle>
          </CardHeader>
          <CardBody>
            {chLoading ? (
              <div className="h-[200px] flex items-center justify-center">
                <Skeleton className="h-[180px] w-[180px] rounded-full" />
              </div>
            ) : (
              <div className="h-[200px] lg:h-[240px] flex items-center justify-center">
                <ChannelDistributionChart data={channels?.distribution || {}} />
              </div>
            )}
          </CardBody>
        </Card>
      </div>
    </div>
  )
}
'@
Set-Content -Path "src\app\(dashboard)\insights\page.tsx" -Value $insights -Encoding UTF8
Write-Host "✓ Insights page"

# ============================================
# 4. Also fix the StatCard component if it exists — make it shrink-0 for mobile scroll
# ============================================
$statCardPath = "src\components\stats\StatCard.tsx"
if (Test-Path $statCardPath) {
    $statCard = Get-Content $statCardPath -Raw
    if ($statCard -notmatch 'shrink-0') {
        $statCard = $statCard -replace 'className="', 'className="shrink-0 '
        Set-Content $statCardPath $statCard -Encoding UTF8
        Write-Host "✓ StatCard shrink-0 added"
    }
}

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "All pages mobile-optimized!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "`nChanges:" -ForegroundColor Cyan
Write-Host "• Setup: Scrollable tabs, single-col form on mobile, truncated text with gaps"
Write-Host "• Channels: Horizontal scroll cards, mobile card list instead of table"
Write-Host "• Insights: Explicit chart heights, horizontal stat scroll on mobile"