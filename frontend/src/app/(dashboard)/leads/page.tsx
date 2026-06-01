import { useEffect, useState } from 'react'
import { useAPI } from '@/hooks/useAPI'
import { useToast } from '@/components/ui/Toast'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { StatCard, StatGrid } from '@/components/stats'
import { StatSkeleton } from '@/components/ui/Skeleton'
import { UserCheck, Phone, Package, Clock, CheckCircle, XCircle, AlertTriangle } from 'lucide-react'
import { api } from '@/lib/api'

interface Handoff {
  id: string
  conversation_id: string
  customer_name: string
  customer_phone: string
  customer_whatsapp: string
  customer_location: string
  product_name: string
  original_price: number
  agreed_price: number
  quantity: number
  status: 'pending' | 'sold' | 'lost' | 'expired'
  final_price: number | null
  owner_notes: string
  reminder_count: number
  created_at: string
}

const statusConfig: Record<string, { label: string; color: string; icon: React.ElementType }> = {
  pending: { label: 'HOT', color: 'warning', icon: Clock },
  sold: { label: 'SOLD', color: 'success', icon: CheckCircle },
  lost: { label: 'LOST', color: 'error', icon: XCircle },
  expired: { label: 'EXPIRED', color: 'neutral', icon: AlertTriangle },
}

function timeAgo(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = Math.floor((now - then) / 1000)
  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

export default function LeadsPage() {
  const { toast } = useToast()
  const apiHook = useAPI() as any
  const { data, get: getHandoffs, loading } = apiHook
  const [statusFilter, setStatusFilter] = useState<string>('')

  const handoffs: Handoff[] = data?.handoffs || []

  useEffect(() => {
    loadHandoffs()
  }, [statusFilter])

  const loadHandoffs = () => {
    const query = statusFilter ? `?status=${statusFilter}` : ''
    getHandoffs(`/handoffs${query}`)
  }

  const handleUpdateStatus = async (id: string, status: string, notes: string = '') => {
    try {
      await api.put('/handoffs/status', { id, status, notes })
      toast(`Lead marked as ${status}`, 'success')
      loadHandoffs()
    } catch (err) {
      toast('Failed to update lead', 'error')
    }
  }

  const pending = handoffs.filter(h => h.status === 'pending').length
  const sold = handoffs.filter(h => h.status === 'sold').length
  const lost = handoffs.filter(h => h.status === 'lost').length
  const totalValue = handoffs.filter(h => h.status === 'sold').reduce((sum, h) => sum + (h.final_price || h.agreed_price), 0)

  return (
    <div className="animate-page-in space-y-5 lg:space-y-6 pt-2">
      {/* Stats */}
      <div className="px-1">
        {loading ? (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <StatSkeleton /><StatSkeleton /><StatSkeleton /><StatSkeleton />
          </div>
        ) : (
          <StatGrid>
            <StatCard label="Hot Leads" value={pending} variant="warning" />
            <StatCard label="Sold" value={sold} variant="success" />
            <StatCard label="Lost" value={lost} variant="error" />
            <StatCard label="Revenue" value={`₦${totalValue.toLocaleString()}`} variant="default" />
          </StatGrid>
        )}
      </div>

      {/* Filter tabs */}
      <div className="px-1">
        <div className="flex gap-2 flex-wrap">
          {['', 'pending', 'sold', 'lost', 'expired'].map(s => (
            <button
              key={s}
              onClick={() => setStatusFilter(s)}
              className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                statusFilter === s
                  ? 'bg-noant-sky text-white'
                  : 'bg-inset text-secondary hover:text-primary'
              }`}
            >
              {s === '' ? 'All' : statusConfig[s]?.label || s}
            </button>
          ))}
        </div>
      </div>

      {/* Leads table */}
      <div className="px-1">
        <Card>
          <CardHeader>
            <CardTitle>Lead Pipeline</CardTitle>
          </CardHeader>
          <CardBody className="p-0">
            {loading ? (
              <div className="p-4 space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="animate-shimmer-slow h-16 rounded-xl w-full" />
                ))}
              </div>
            ) : handoffs.length === 0 ? (
              <div className="p-8 text-center">
                <UserCheck className="w-12 h-12 mx-auto text-tertiary mb-3" />
                <p className="text-secondary text-sm">No leads yet</p>
                <p className="text-tertiary text-xs mt-1">Leads appear when the AI hands off a customer sale</p>
              </div>
            ) : (
              <>
                {/* Desktop table */}
                <div className="hidden lg:block overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b border-default">
                        <th className="text-left text-[10px] font-bold uppercase tracking-wider text-tertiary px-4 py-2.5">Customer</th>
                        <th className="text-left text-[10px] font-bold uppercase tracking-wider text-tertiary px-4 py-2.5">Product</th>
                        <th className="text-left text-[10px] font-bold uppercase tracking-wider text-tertiary px-4 py-2.5">Price</th>
                        <th className="text-left text-[10px] font-bold uppercase tracking-wider text-tertiary px-4 py-2.5">Status</th>
                        <th className="text-left text-[10px] font-bold uppercase tracking-wider text-tertiary px-4 py-2.5">Time</th>
                        <th className="text-right text-[10px] font-bold uppercase tracking-wider text-tertiary px-4 py-2.5">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {handoffs.map(h => {
                        const cfg = statusConfig[h.status]
                        const StatusIcon = cfg?.icon || Clock
                        return (
                          <tr key={h.id} className="border-b border-default last:border-0 hover:bg-inset/50 transition-colors">
                            <td className="px-4 py-3">
                              <div className="flex items-center gap-2">
                                <div className="w-8 h-8 rounded-full bg-noant-sky/10 flex items-center justify-center text-noant-sky text-xs font-bold">
                                  {h.customer_name?.[0] || '?'}
                                </div>
                                <div>
                                  <div className="text-sm font-medium text-primary">{h.customer_name || 'Unknown'}</div>
                                  <div className="text-[10px] text-tertiary flex items-center gap-1">
                                    <Phone className="w-3 h-3" />
                                    {h.customer_phone || h.customer_whatsapp || 'No phone'}
                                  </div>
                                </div>
                              </div>
                            </td>
                            <td className="px-4 py-3">
                              <div className="flex items-center gap-1.5 text-sm text-primary">
                                <Package className="w-3.5 h-3.5 text-tertiary" />
                                {h.product_name || 'General inquiry'}
                              </div>
                            </td>
                            <td className="px-4 py-3">
                              <div className="text-sm font-medium text-primary">₦{h.agreed_price?.toLocaleString()}</div>
                              {h.original_price !== h.agreed_price && (
                                <div className="text-[10px] text-tertiary line-through">₦{h.original_price?.toLocaleString()}</div>
                              )}
                            </td>
                            <td className="px-4 py-3">
                              <Badge variant={cfg?.color as any || 'neutral'}>
                                <StatusIcon className="w-3 h-3 mr-1" />
                                {cfg?.label || h.status}
                              </Badge>
                            </td>
                            <td className="px-4 py-3 text-xs text-tertiary">{timeAgo(h.created_at)}</td>
                            <td className="px-4 py-3 text-right">
                              {h.status === 'pending' && (
                                <div className="flex gap-1 justify-end">
                                  <Button variant="ghost" size="sm" onClick={() => handleUpdateStatus(h.id, 'sold')}>Sold</Button>
                                  <Button variant="ghost" size="sm" onClick={() => handleUpdateStatus(h.id, 'lost')}>Lost</Button>
                                </div>
                              )}
                            </td>
                          </tr>
                        )
                      })}
                    </tbody>
                  </table>
                </div>

                {/* Mobile cards */}
                <div className="lg:hidden divide-y divide-default">
                  {handoffs.map(h => {
                    const cfg = statusConfig[h.status]
                    const StatusIcon = cfg?.icon || Clock
                    return (
                      <div key={h.id} className="p-4 space-y-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <div className="w-8 h-8 rounded-full bg-noant-sky/10 flex items-center justify-center text-noant-sky text-xs font-bold">
                              {h.customer_name?.[0] || '?'}
                            </div>
                            <div>
                              <div className="text-sm font-medium text-primary">{h.customer_name || 'Unknown'}</div>
                              <div className="text-[10px] text-tertiary">{h.customer_phone || h.customer_whatsapp || 'No phone'}</div>
                            </div>
                          </div>
                          <Badge variant={cfg?.color as any || 'neutral'}>
                            <StatusIcon className="w-3 h-3 mr-1" />
                            {cfg?.label || h.status}
                          </Badge>
                        </div>
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-secondary flex items-center gap-1"><Package className="w-3.5 h-3.5" />{h.product_name || 'General'}</span>
                          <span className="font-medium text-primary">₦{h.agreed_price?.toLocaleString()}</span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-[10px] text-tertiary">{timeAgo(h.created_at)}</span>
                          {h.status === 'pending' && (
                            <div className="flex gap-1">
                              <Button variant="ghost" size="sm" onClick={() => handleUpdateStatus(h.id, 'sold')}>Sold</Button>
                              <Button variant="ghost" size="sm" onClick={() => handleUpdateStatus(h.id, 'lost')}>Lost</Button>
                            </div>
                          )}
                        </div>
                      </div>
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
