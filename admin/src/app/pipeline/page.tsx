import { useState, useEffect, useCallback } from 'react'
import { adminApi } from '@/lib/api'
import type { SalesLead } from '@/types'
import { timeAgo } from '@/lib/utils'
import { SkeletonCard } from '@/components/ui/Skeleton'
import { ErrorBanner, EmptyState } from '@/components/ui/Feedback'
import { Plus, X, ClipboardList, ChevronDown, QrCode, Link2, Copy, Check, Download } from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { useAdminWS } from '@/lib/hooks/useAdminWS'

const STATUS_META: Record<string, { label: string; color: string; bg: string }> = {
  contacted: { label: 'Contacted', color: 'text-amber-600', bg: 'bg-amber-500/10' },
  demo_sent: { label: 'Demo Sent', color: 'text-brand-sky', bg: 'bg-brand-sky/10' },
  signed_up: { label: 'Signed Up', color: 'text-success', bg: 'bg-success/10' },
  paying: { label: 'Paying', color: 'text-purple-500', bg: 'bg-purple-500/10' },
  lost: { label: 'Lost', color: 'text-danger', bg: 'bg-danger/10' },
}

const STAT_COLORS: Record<string, { border: string; text: string; bg: string }> = {
  contacted: { border: 'border-amber-500/20', text: 'text-amber-500', bg: 'bg-amber-500/5' },
  demo_sent: { border: 'border-brand-sky/20', text: 'text-brand-sky', bg: 'bg-brand-sky/5' },
  signed_up: { border: 'border-success/20', text: 'text-success', bg: 'bg-success/5' },
  paying: { border: 'border-purple-500/20', text: 'text-purple-500', bg: 'bg-purple-500/5' },
  lost: { border: 'border-danger/20', text: 'text-danger', bg: 'bg-danger/5' },
}

const BUSINESS_TYPES = ['Restaurant', 'Salon', 'Shop', 'Pharmacy', 'Tech', 'Other']

function StatusBadge({ status }: { status: string }) {
  const meta = STATUS_META[status] ?? STATUS_META.contacted!
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${meta.bg} ${meta.color}`}>
      {meta.label}
    </span>
  )
}

function LeadCard({ lead, onStatusUpdate }: { lead: SalesLead; onStatusUpdate: (id: string, status: string) => void }) {
  const [expanded, setExpanded] = useState(false)
  const [updating, setUpdating] = useState(false)

  const handleStatusChange = async (newStatus: string) => {
    setUpdating(true)
    try {
      await adminApi.updateSalesLead(lead.id, { status: newStatus })
      onStatusUpdate(lead.id, newStatus)
    } catch {
      // silent
    } finally {
      setUpdating(false)
    }
  }

  return (
    <div className="rounded-xl border border-border bg-bg-surface p-5 transition-colors hover:border-border/80">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 mb-1">
            <h3 className="text-sm font-semibold text-text-primary truncate">{lead.contact_name}</h3>
            <StatusBadge status={lead.status} />
          </div>
          {(lead.business_name || lead.business_type) && (
            <p className="text-xs text-text-tertiary mb-2">
              {lead.business_name}{lead.business_name && lead.business_type ? ' · ' : ''}{lead.business_type}
            </p>
          )}
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-text-secondary">
            {lead.contact_phone && <span>{lead.contact_phone}</span>}
            {lead.contact_email && <span>{lead.contact_email}</span>}
            {lead.meeting_location && <span className="text-text-tertiary">{lead.meeting_location}</span>}
          </div>
          {lead.notes && (
            <div className="mt-2">
              <p className={`text-xs text-text-tertiary ${expanded ? '' : 'line-clamp-2'}`}>{lead.notes}</p>
              {lead.notes.length > 80 && (
                <button onClick={() => setExpanded(!expanded)} className="text-xs text-brand-sky mt-1 hover:underline">
                  {expanded ? 'Show less' : 'Show more'}
                </button>
              )}
            </div>
          )}
        </div>
        <div className="flex flex-col items-end gap-2 shrink-0">
          <span className="text-xs text-text-tertiary whitespace-nowrap">{timeAgo(lead.created_at)}</span>
          <div className="relative">
            <select
              value={lead.status}
              onChange={e => handleStatusChange(e.target.value)}
              disabled={updating}
              className="appearance-none rounded-lg border border-border bg-bg-inset px-3 py-1.5 pr-7 text-xs text-text-secondary outline-none transition-colors focus:border-brand-sky cursor-pointer disabled:opacity-50"
            >
              {Object.entries(STATUS_META).map(([key, meta]) => (
                <option key={key} value={key}>{meta.label}</option>
              ))}
            </select>
            <ChevronDown className="pointer-events-none absolute right-2 top-1/2 h-3 w-3 -translate-y-1/2 text-text-tertiary" />
          </div>
        </div>
      </div>
    </div>
  )
}

export default function PipelinePage() {
  const [leads, setLeads] = useState<SalesLead[]>([])
  const [stats, setStats] = useState<{ status: string; count: number }[]>([])
  const [totalLeads, setTotalLeads] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState('all')
  const [showForm, setShowForm] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [showQR, setShowQR] = useState(false)
  const [referral, setReferral] = useState<{ code: string; url: string; clicks: number; signups: number } | null>(null)
  const [copied, setCopied] = useState(false)

  const [form, setForm] = useState({
    contact_name: '',
    contact_phone: '',
    contact_email: '',
    business_name: '',
    business_type: '',
    meeting_location: '',
    notes: '',
    status: 'contacted',
  })

  const fetchData = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [leadsRes, statsRes] = await Promise.all([
        adminApi.getSalesLeads(filter !== 'all' ? { status: filter } : undefined),
        adminApi.getPipelineStats(),
      ])
      setLeads(leadsRes.leads ?? [])
      setTotalLeads(leadsRes.total ?? 0)
      setStats(statsRes.pipeline ?? [])
    } catch (err: any) {
      setError(err.message || 'Failed to load pipeline')
    } finally {
      setLoading(false)
    }
  }, [filter])

  useEffect(() => { fetchData() }, [fetchData])

  const handleWSLeadCreated = useCallback(() => {
    fetchData()
  }, [fetchData])

  const handleWSLeadUpdated = useCallback(() => {
    fetchData()
  }, [fetchData])

  const { connected } = useAdminWS({
    lead_created: handleWSLeadCreated,
    lead_updated: handleWSLeadUpdated,
  })

  // Polling fallback every 15s
  useEffect(() => {
    if (connected) return
    const interval = setInterval(fetchData, 15000)
    return () => clearInterval(interval)
  }, [connected, fetchData])

  const handleStatusUpdate = (id: string, newStatus: string) => {
    setLeads(prev => prev.map(l => l.id === id ? { ...l, status: newStatus as SalesLead['status'] } : l))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.contact_name.trim()) return
    setSubmitting(true)
    try {
      await adminApi.createSalesLead({
        contact_name: form.contact_name.trim(),
        contact_phone: form.contact_phone.trim() || undefined,
        contact_email: form.contact_email.trim() || undefined,
        business_name: form.business_name.trim() || undefined,
        business_type: form.business_type || undefined,
        meeting_location: form.meeting_location.trim() || undefined,
        notes: form.notes.trim() || undefined,
        status: form.status,
      })
      setForm({ contact_name: '', contact_phone: '', contact_email: '', business_name: '', business_type: '', meeting_location: '', notes: '', status: 'contacted' })
      setShowForm(false)
      fetchData()
    } catch {
      // error handled by retry
    } finally {
      setSubmitting(false)
    }
  }

  const statCounts = stats.reduce((acc, s) => { acc[s.status] = s.count; return acc }, {} as Record<string, number>)

  const fetchReferral = async () => {
    try {
      const res = await adminApi.getReferral()
      setReferral(res)
    } catch { /* silent */ }
  }

  const handleShowQR = () => {
    setShowQR(true)
    if (!referral) fetchReferral()
  }

  const handleCopyLink = async () => {
    if (!referral?.url) return
    try {
      await navigator.clipboard.writeText(referral.url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* silent */ }
  }

  const handleDownloadQR = () => {
    const svg = document.getElementById('referral-qr')?.querySelector('svg')
    if (!svg) return
    const data = new XMLSerializer().serializeToString(svg)
    const blob = new Blob([data], { type: 'image/svg+xml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `noant-qr-${referral?.code ?? 'code'}.svg`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Sales Pipeline</h1>
          <p className="text-sm text-text-tertiary">{totalLeads} total leads</p>
        </div>
        <button
          onClick={() => adminApi.exportCSV('users')}
          className="ml-auto inline-flex items-center gap-1.5 rounded-lg bg-brand-sky/10 px-3 py-1.5 text-xs font-medium text-brand-sky hover:bg-brand-sky/20"
        >
          <Download className="h-3 w-3" />
          Export CSV
        </button>
        <div className={`flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ${connected ? 'bg-success/10 text-success' : 'bg-amber-500/10 text-amber-600'}`}>
          <span className={`h-1.5 w-1.5 rounded-full ${connected ? 'bg-success animate-pulse' : 'bg-amber-500'}`} />
          {connected ? 'Live' : 'Syncing...'}
        </div>
      </div>

      {error && <ErrorBanner message={error} onRetry={fetchData} />}

      {/* Stat cards */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
        {(['contacted', 'demo_sent', 'signed_up', 'paying', 'lost'] as const).map(status => {
          const meta = STAT_COLORS[status]!
          return (
            <div key={status} className={`rounded-xl border ${meta.border} ${meta.bg} p-4`}>
              <p className="text-xs font-medium text-text-tertiary">{STATUS_META[status]!.label}</p>
              <p className={`text-2xl font-bold mt-1 ${meta.text}`}>{statCounts[status] ?? 0}</p>
            </div>
          )
        })}
      </div>

      {/* Filter + Add button */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <div className="flex items-center gap-2">
          {['all', 'contacted', 'demo_sent', 'signed_up', 'paying', 'lost'].map(s => (
            <button
              key={s}
              onClick={() => setFilter(s)}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
                filter === s
                  ? 'bg-brand-sky/10 text-brand-sky'
                  : 'text-text-tertiary hover:text-text-secondary hover:bg-bg-inset'
              }`}
            >
              {s === 'all' ? 'All' : STATUS_META[s]?.label ?? s}
            </button>
          ))}
        </div>
        <div className="ml-auto flex items-center gap-2">
          <button
            onClick={handleShowQR}
            className="flex items-center gap-1.5 rounded-lg border border-border bg-bg-inset px-3.5 py-2 text-xs font-medium text-text-secondary transition-colors hover:bg-bg-surface hover:text-text-primary"
          >
            <QrCode className="h-3.5 w-3.5" />
            Share QR
          </button>
          <button
            onClick={() => { setShowForm(true); setForm({ contact_name: '', contact_phone: '', contact_email: '', business_name: '', business_type: '', meeting_location: '', notes: '', status: 'contacted' }) }}
            className="flex items-center gap-1.5 rounded-lg bg-brand-sky px-3.5 py-2 text-xs font-medium text-white transition-colors hover:bg-brand-sky-deep"
          >
            <Plus className="h-3.5 w-3.5" />
            Add Lead
          </button>
        </div>
      </div>

      {/* Add Lead Modal */}
  {showForm && (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={() => setShowForm(false)}>
      <div className="w-full max-w-lg rounded-xl border border-border bg-bg-base p-6 space-y-4" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-bold text-text-primary">Add New Lead</h2>
          <button onClick={() => setShowForm(false)} className="rounded-lg p-1 text-text-tertiary hover:text-text-primary hover:bg-bg-inset">
            <X className="h-4 w-4" />
          </button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-text-tertiary mb-1">Contact Name *</label>
              <input
                required
                value={form.contact_name}
                onChange={e => setForm(f => ({ ...f, contact_name: e.target.value }))}
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary outline-none transition-colors focus:border-brand-sky"
                placeholder="e.g. Amina Bello"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-text-tertiary mb-1">Phone</label>
              <input
                value={form.contact_phone}
                onChange={e => setForm(f => ({ ...f, contact_phone: e.target.value }))}
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary outline-none transition-colors focus:border-brand-sky"
                placeholder="0801 234 5678"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-text-tertiary mb-1">Email</label>
              <input
                type="email"
                value={form.contact_email}
                onChange={e => setForm(f => ({ ...f, contact_email: e.target.value }))}
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary outline-none transition-colors focus:border-brand-sky"
                placeholder="amina@example.com"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-text-tertiary mb-1">Business Name</label>
              <input
                value={form.business_name}
                onChange={e => setForm(f => ({ ...f, business_name: e.target.value }))}
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary outline-none transition-colors focus:border-brand-sky"
                placeholder="Amina's Restaurant"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-text-tertiary mb-1">Business Type</label>
              <select
                value={form.business_type}
                onChange={e => setForm(f => ({ ...f, business_type: e.target.value }))}
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary outline-none"
              >
                <option value="">Select type</option>
                {BUSINESS_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-text-tertiary mb-1">Meeting Location</label>
              <input
                value={form.meeting_location}
                onChange={e => setForm(f => ({ ...f, meeting_location: e.target.value }))}
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary outline-none transition-colors focus:border-brand-sky"
                placeholder="Lekki Market"
              />
            </div>
            <div className="sm:col-span-2">
              <label className="block text-xs font-medium text-text-tertiary mb-1">Status</label>
              <select
                value={form.status}
                onChange={e => setForm(f => ({ ...f, status: e.target.value }))}
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary outline-none"
              >
                {Object.entries(STATUS_META).map(([key, meta]) => (
                  <option key={key} value={key}>{meta.label}</option>
                ))}
              </select>
            </div>
            <div className="sm:col-span-2">
              <label className="block text-xs font-medium text-text-tertiary mb-1">Notes</label>
              <textarea
                value={form.notes}
                onChange={e => setForm(f => ({ ...f, notes: e.target.value }))}
                rows={3}
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary outline-none transition-colors focus:border-brand-sky resize-none"
                placeholder="Interested in WhatsApp automation..."
              />
            </div>
          </div>
          <div className="flex items-center gap-3 pt-2">
            <button
              type="submit"
              disabled={submitting || !form.contact_name.trim()}
              className="rounded-lg bg-brand-sky px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-brand-sky-deep disabled:opacity-50"
            >
              {submitting ? 'Adding...' : 'Add Lead'}
            </button>
            <button
              type="button"
              onClick={() => setShowForm(false)}
              className="rounded-lg bg-bg-inset px-4 py-2 text-sm font-medium text-text-secondary transition-colors hover:bg-bg-surface"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )}

  {/* Lead cards */}
      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <SkeletonCard key={i} className="h-28" />
          ))}
        </div>
      ) : leads.length === 0 ? (
        <div className="rounded-xl border border-border bg-bg-surface">
          <EmptyState
            icon={ClipboardList}
            title="No leads yet"
            description="Start meeting people and add them here."
          />
        </div>
      ) : (
        <div className="space-y-3">
          {leads.map(lead => (
            <LeadCard key={lead.id} lead={lead} onStatusUpdate={handleStatusUpdate} />
          ))}
        </div>
      )}

      {/* QR Code Modal */}
      {showQR && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={() => setShowQR(false)}>
          <div className="w-full max-w-sm rounded-2xl border border-border bg-bg-base p-6 space-y-5" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-bold text-text-primary">Your Referral QR</h2>
              <button onClick={() => setShowQR(false)} className="rounded-lg p-1 text-text-tertiary hover:text-text-primary hover:bg-bg-inset">
                <X className="h-4 w-4" />
              </button>
            </div>
            {referral ? (
              <>
                <div className="flex justify-center rounded-xl border border-border bg-white p-6" id="referral-qr">
                  <QRCodeSVG value={referral.url} size={180} level="H" />
                </div>
                <div className="space-y-3">
                  <div className="flex items-center gap-2 rounded-lg border border-border bg-bg-inset px-3 py-2">
                    <Link2 className="h-3.5 w-3.5 shrink-0 text-text-tertiary" />
                    <span className="flex-1 truncate text-xs text-text-secondary font-mono">{referral.url}</span>
                    <button onClick={handleCopyLink} className="shrink-0 rounded-md p-1 text-text-tertiary hover:text-brand-sky transition-colors">
                      {copied ? <Check className="h-3.5 w-3.5 text-success" /> : <Copy className="h-3.5 w-3.5" />}
                    </button>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="rounded-lg border border-border bg-bg-inset p-3 text-center">
                      <p className="text-2xl font-bold text-text-primary">{referral.clicks}</p>
                      <p className="text-xs text-text-tertiary mt-0.5">Clicks</p>
                    </div>
                    <div className="rounded-lg border border-border bg-bg-inset p-3 text-center">
                      <p className="text-2xl font-bold text-brand-sky">{referral.signups}</p>
                      <p className="text-xs text-text-tertiary mt-0.5">Signups</p>
                    </div>
                  </div>
                  <button onClick={handleDownloadQR} className="w-full rounded-lg border border-border bg-bg-inset px-4 py-2.5 text-xs font-medium text-text-secondary transition-colors hover:bg-bg-surface hover:text-text-primary">
                    Download QR
                  </button>
                  <p className="text-center text-xs text-text-tertiary">Show this QR code during meetings. People scan it to sign up with your referral.</p>
                </div>
              </>
            ) : (
              <div className="flex justify-center py-8">
                <div className="h-6 w-6 animate-spin rounded-full border-2 border-brand-sky border-t-transparent" />
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
