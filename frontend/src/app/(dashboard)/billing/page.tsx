import { useState, useEffect } from 'react'
import { createPortal } from 'react-dom'
import { Check, X, Zap, Building2, Rocket, Star, ArrowRight, Crown, Flame, Receipt } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { api } from '../../../lib/api'
import { useToast } from '@/components/ui/Toast'

type Currency = 'NGN' | 'USD'

interface PulsePack {
  id: string
  name: string
  priceNGN: number
  priceUSD: number
  responses: number
  badge?: string
}

interface Plan {
  id: string
  name: string
  icon: typeof Zap
  tagline: string
  priceNGN: number | null
  priceUSD: number | null
  period: string
  features: string[]
  highlight?: boolean
  cta: string
  polarUrl?: string
}

interface UserCredit {
  balance: number
  expires_at: string | null
}

interface PurchaseRecord {
  id: string
  pack_type: string
  amount: number
  status: string
  purchased_at: string
  expires_at: string
}

// Response counts match backend credit.go: small=500, medium=1250, large=2500
const pulsePacks: PulsePack[] = [
  { id: 'small',  name: 'Starter', priceNGN: 2999,  priceUSD: 4,  responses: 500  },
  { id: 'medium', name: 'Growth',  priceNGN: 5999,  priceUSD: 8,  responses: 1250, badge: 'Best Value' },
  { id: 'large',  name: 'Pro',     priceNGN: 9999,  priceUSD: 13, responses: 2500, badge: 'Most Responses' },
]

const plans: Plan[] = [
  {
    id: 'free',
    name: 'Free',
    icon: Star,
    tagline: 'Start with no commitment',
    priceNGN: 0,
    priceUSD: 0,
    period: '',
    cta: 'Current plan',
    features: [
      '100 AI responses per week',
      'Web Widget + WhatsApp',
      '10 inventory items',
      '1 team member (you)',
      'Basic AI responses',
      'Handoff system enabled',
      'No instant notifications',
      'NOANT branding on widget',
    ],
  },
  {
    id: 'pulse',
    name: 'Pulse',
    icon: Flame,
    tagline: 'Pay as you go',
    priceNGN: null,
    priceUSD: null,
    period: '',
    cta: 'Buy Responses',
    polarUrl: 'https://buy.polar.sh/polar_cl_bVci27dxySJ100TaBuEMPk1uG4WQbuFyM7svO0jaMUR',
    features: [
      'All 4 channels (WhatsApp, Telegram, Gmail, Web)',
      'Unlimited inventory items',
      'Full handoff system',
      'Instant notifications',
      'AI price negotiation',
      'Unlimited negotiation rounds',
      'Basic analytics',
    ],
  },
  {
    id: 'pro',
    name: 'Pro',
    icon: Rocket,
    tagline: 'For growing teams',
    priceNGN: 21999,
    priceUSD: 30,
    period: 'month',
    cta: 'Start 14-Day Free Trial',
    highlight: true,
    polarUrl: 'https://buy.polar.sh/polar_cl_A40Sc8EW0lFIIXLhpfxOEYsi3IUZMzfxry8BM12MtID',
    features: [
      'Unlimited AI responses',
      'Unlimited team members',
      'All 4 channels',
      'Full inventory system',
      'Full handoff & lead scoring',
      'AI price negotiation',
      'Unlimited negotiation rounds',
      'Advanced analytics',
      'Priority support',
      'White-label widget (your logo, no NOANT branding)',
      'Campaign Mode',
    ],
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    icon: Building2,
    tagline: 'Custom solutions for large organisations',
    priceNGN: 99999,
    priceUSD: 130,
    period: 'month',
    cta: 'Contact Sales',
    polarUrl: 'https://buy.polar.sh/polar_cl_qobPhYYsUZEeq8gOR8Hei7rwssMZ8PfXAeUGI39TMEN',
    features: [
      'Everything in Pro',
      'Unlimited everything',
      'Custom AI training on your data',
      'API access for integrations',
      'Dedicated account manager',
      'SLA guarantee: 99.9% uptime',
      'Full white-label platform (your domain, your branding)',
      'Custom AI personality',
      'Advanced security',
      'Onboarding & training',
    ],
  },
]

function formatPrice(amount: number | null | undefined, currency: Currency) {
  if (amount == null) return 'Custom'
  if (amount === 0) return 'Free'
  if (currency === 'NGN') return `₦${amount.toLocaleString()}`
  return `$${amount}`
}

function packLabel(id: string) {
  const p = pulsePacks.find(x => x.id === id)
  return p ? p.name : id
}

export default function BillingPage() {
  const { user, loading: authLoading } = useAuth()
  const { toast: showToast } = useToast()
  const [currency, setCurrency] = useState<Currency>('NGN')
  const [loading, setLoading] = useState<string | null>(null)
  const [userCredit, setUserCredit] = useState<UserCredit | null>(null)
  const [purchaseHistory, setPurchaseHistory] = useState<PurchaseRecord[]>([])
  const [historyLoading, setHistoryLoading] = useState(true)
  const [showPulseModal, setShowPulseModal] = useState(false)
  // Only derive currentPlan once auth has resolved — avoids flashing 'free' banner
  const currentPlan = authLoading ? undefined : (user?.plan_id || user?.plan || 'free')

  useEffect(() => {
    if (!user?.id) return
    api.get<UserCredit>('/credits/balance').then(setUserCredit).catch(() => {})
    setHistoryLoading(true)
    api.get<PurchaseRecord[]>('/credits/history')
      .then(data => setPurchaseHistory(Array.isArray(data) ? data : []))
      .catch(() => setPurchaseHistory([]))
      .finally(() => setHistoryLoading(false))
  }, [user?.id])

  const handleUpgrade = async (planId: string) => {
    if (planId === 'free' || planId === currentPlan) return
    if (planId === 'pulse') {
      setShowPulseModal(true)
      return
    }
    const plan = plans.find(p => p.id === planId)
    if (plan?.polarUrl) {
      try {
        const url = new URL(plan.polarUrl)
        if (user?.id) {
          url.searchParams.set('metadata[user_id]', user.id)
          url.searchParams.set('metadata[plan_id]', planId)
        }
        window.open(url.toString(), '_blank')
      } catch {
        window.open(plan.polarUrl, '_blank')
      }
      return
    }
    setLoading(planId)
    try {
      const res = await api.post<{ checkout_url?: string }>('/payments/subscribe', { plan: planId, currency })
      if (res?.checkout_url) window.open(res.checkout_url, '_blank')
      else showToast('Upgrade initiated! Check your email.', 'success')
    } catch {
      showToast('Failed to start upgrade. Please try again.', 'error')
    } finally {
      setLoading(null)
    }
  }

  const handlePurchasePack = async (packType: string) => {
    if (!user?.id) return
    setLoading('purchase_' + packType)
    try {
      const res = await api.post<{ checkout_url: string }>('/credits/purchase', { pack_type: packType })
      if (res?.checkout_url) {
        window.open(res.checkout_url, '_blank')
        showToast('Redirecting to payment page…', 'success')
        setShowPulseModal(false)
      }
    } catch (err: any) {
      showToast(err.response?.data?.error || 'Failed to initiate purchase.', 'error')
    } finally {
      setLoading(null)
    }
  }

  return (
    <div className="min-h-screen p-4 lg:p-6 animate-page-in">
      <div className="max-w-5xl mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-primary mb-2">Choose your plan</h1>
          <p className="text-secondary max-w-md mx-auto">Upgrade anytime. Cancel anytime. No hidden fees.</p>
          <div className="inline-flex items-center gap-1 mt-4 bg-inset p-1 rounded-xl border border-default">
            {(['NGN', 'USD'] as Currency[]).map(c => (
              <button key={c} onClick={() => setCurrency(c)}
                className={`px-5 py-1.5 rounded-lg text-sm font-semibold transition-all ${
                  currency === c
                    ? 'bg-surface text-primary shadow-sm border border-default'
                    : 'text-secondary hover:text-primary'
                }`}
              >
                {c === 'NGN' ? '🇳🇬 NGN' : '🌍 USD'}
              </button>
            ))}
          </div>
        </div>

        {/* Trial banner — only shown after auth resolves to avoid flash */}
        {!authLoading && currentPlan === 'free' && (
          <div className="mb-6 p-4 rounded-xl bg-gradient-to-r from-noant-sky/20 to-purple-500/10 border border-noant-sky/30 flex items-center gap-3">
            <Crown className="w-5 h-5 text-noant-sky shrink-0" />
            <div>
              <p className="text-sm font-semibold text-primary">You're on the Free plan</p>
              <p className="text-xs text-secondary">Upgrade to unlock unlimited responses, more channels, and team collaboration.</p>
            </div>
          </div>
        )}

        {/* Credit balance banner for Pulse users — gated behind auth loading */}
        {!authLoading && currentPlan === 'pulse' && userCredit && (
          <div className="mb-6 p-4 rounded-xl bg-gradient-to-r from-orange-500/10 to-amber-500/10 border border-orange-500/20 flex items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <Flame className="w-5 h-5 text-orange-500 shrink-0" />
              <div>
                <p className="text-sm font-semibold text-primary">
                  {userCredit.balance.toLocaleString()} responses remaining
                </p>
                {userCredit.expires_at && (
                  <p className="text-xs text-secondary">
                    Expires {new Date(userCredit.expires_at).toLocaleDateString()}
                  </p>
                )}
              </div>
            </div>
            <button
              onClick={() => setShowPulseModal(true)}
              className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-orange-500 text-white hover:bg-orange-600 transition-all whitespace-nowrap"
            >
              Top up
            </button>
          </div>
        )}

        {/* Plan grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {plans.map(plan => {
            const Icon = plan.icon
            // While auth is loading, no card is marked current (avoids flash)
            const isCurrent = !authLoading && plan.id === currentPlan
            const isDisabled = isCurrent || loading !== null

            return (
              <div key={plan.id}
                className={`relative flex flex-col rounded-2xl border p-4 transition-all ${
                  plan.highlight
                    ? 'border-noant-sky bg-gradient-to-b from-noant-sky/10 to-transparent shadow-sky'
                    : 'border-default bg-surface hover:border-strong'
                }`}
              >
                {plan.highlight && (
                  <div className="absolute -top-2.5 left-1/2 -translate-x-1/2 px-2.5 py-0.5 rounded-full bg-noant-sky text-white text-[10px] font-bold whitespace-nowrap">
                    Most Popular
                  </div>
                )}
                {isCurrent && plan.id !== 'pro' && (
                  <div className="absolute -top-2.5 left-1/2 -translate-x-1/2 px-2.5 py-0.5 rounded-full bg-amber-500 text-white text-[10px] font-bold whitespace-nowrap">
                    Your Plan
                  </div>
                )}
                {isCurrent && plan.id === 'pro' && (
                  <div className="absolute -top-2.5 left-1/2 -translate-x-1/2 px-2.5 py-0.5 rounded-full bg-noant-sky text-white text-[10px] font-bold whitespace-nowrap">
                    Your Plan
                  </div>
                )}

                <div className="mb-3">
                  <div className={`w-9 h-9 rounded-lg mb-2 flex items-center justify-center ${
                    plan.highlight
                      ? 'bg-noant-sky text-white'
                      : plan.id === 'free'
                      ? 'bg-amber-500/10 text-amber-500'
                      : plan.id === 'pulse'
                      ? 'bg-orange-500/10 text-orange-500'
                      : plan.id === 'enterprise'
                      ? 'bg-purple-500/10 text-purple-500'
                      : 'bg-inset text-secondary'
                  }`}>
                    <Icon className="w-4.5 h-4.5" />
                  </div>
                  <div className="font-bold text-primary text-base">{plan.name}</div>
                  <div className="text-[11px] text-secondary mt-0.5 leading-snug">{plan.tagline}</div>
                </div>

                {/* Pricing */}
                <div className="mb-3">
                  {plan.id === 'pulse' ? (
                    <div>
                      <div className="text-base font-bold text-primary">Pay As You Go</div>
                      <div className="text-[11px] text-secondary mt-0.5">
                        from {formatPrice(currency === 'NGN' ? 2999 : 4, currency)}
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-baseline gap-1">
                      <span className="text-2xl font-black text-primary">
                        {formatPrice(currency === 'NGN' ? plan.priceNGN : plan.priceUSD, currency)}
                      </span>
                      {plan.priceNGN !== null && plan.priceNGN > 0 && (
                        <span className="text-xs text-secondary">/{plan.period}</span>
                      )}
                    </div>
                  )}
                </div>

                {/* Features */}
                <ul className="space-y-1.5 flex-1 mb-4">
                  {plan.features.map(f => (
                    <li key={f} className="flex items-start gap-1.5 text-[11px] text-secondary">
                      <Check className="w-3 h-3 text-noant-sky shrink-0 mt-0.5" />
                      {f}
                    </li>
                  ))}
                </ul>

                {/* Pulse packs — always visible inside Pulse card */}
                {plan.id === 'pulse' && (
                  <div className="mb-3 space-y-1.5">
                    {pulsePacks.map(pack => (
                      <div key={pack.id}
                        className={`flex items-center justify-between p-2 rounded-xl border ${
                          pack.badge ? 'border-noant-sky/40 bg-noant-sky/[0.04]' : 'border-default bg-inset/50'
                        }`}
                      >
                        <div>
                          <div className="flex items-center gap-1">
                            <span className="text-xs font-semibold text-primary">{pack.name}</span>
                            {pack.badge && (
                              <span className="text-[9px] px-1.5 py-0.5 rounded-full bg-noant-sky/10 text-noant-sky font-semibold">
                                {pack.badge}
                              </span>
                            )}
                          </div>
                          <div className="text-[11px] text-secondary">
                            {formatPrice(currency === 'NGN' ? pack.priceNGN : pack.priceUSD, currency)} · {pack.responses.toLocaleString()} responses
                          </div>
                        </div>
                        <button
                          onClick={() => handlePurchasePack(pack.id)}
                          disabled={loading !== null}
                          className="ml-2 px-2.5 py-1 rounded-lg text-[11px] font-semibold bg-noant-sky text-white hover:bg-noant-sky-deep disabled:opacity-50 transition-all"
                        >
                          {loading === 'purchase_' + pack.id ? '…' : 'Buy'}
                        </button>
                      </div>
                    ))}

                    {/* Credit balance */}
                    {userCredit && userCredit.balance > 0 && (
                      <div className="p-2 rounded-xl bg-noant-sky/[0.06] border border-noant-sky/20 text-center">
                        <div className="text-[11px] text-secondary">Your balance</div>
                        <div className="text-xs font-bold text-noant-sky">{userCredit.balance.toLocaleString()} responses</div>
                        {userCredit.expires_at && (
                          <div className="text-[9px] text-secondary">Expires {new Date(userCredit.expires_at).toLocaleDateString()}</div>
                        )}
                      </div>
                    )}
                  </div>
                )}

                <button onClick={() => {
                  if (plan.id === 'pulse') {
                    // Pulse card button just scrolls to packs / opens modal on mobile
                    setShowPulseModal(true)
                  } else {
                    handleUpgrade(plan.id)
                  }
                }}
                  disabled={isDisabled}
                  className={`w-full flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-semibold transition-all active:scale-[0.98] ${
                    isCurrent
                      ? 'bg-inset text-secondary cursor-default border border-default'
                      : plan.highlight
                      ? 'bg-noant-sky text-white hover:bg-noant-sky-deep shadow-sky'
                      : 'bg-noant-black text-white dark:bg-white dark:text-noant-black hover:opacity-90'
                  } disabled:opacity-60`}
                >
                  {isCurrent ? (
                    <><Check className="w-3 h-3" /> Current plan</>
                  ) : loading === plan.id ? (
                    'Loading…'
                  ) : (
                    <>{plan.cta} <ArrowRight className="w-3 h-3" /></>
                  )}
                </button>
              </div>
            )
          })}
        </div>

        {/* Payment / Purchase History */}
        <div className="mt-10">
          <h2 className="text-base font-semibold text-primary mb-3">Purchase History</h2>
          <div className="rounded-xl border border-default bg-surface overflow-hidden">
            {historyLoading ? (
              <div className="p-8 text-center">
                <div className="w-5 h-5 rounded-full border-2 border-noant-sky border-t-transparent animate-spin mx-auto mb-3" />
                <p className="text-sm text-secondary">Loading history…</p>
              </div>
            ) : purchaseHistory.length === 0 ? (
              <div className="p-8 text-center">
                <div className="w-10 h-10 rounded-xl bg-inset flex items-center justify-center mx-auto mb-3">
                  <Receipt className="w-5 h-5 text-tertiary" />
                </div>
                <p className="text-sm text-secondary">No purchase history yet.</p>
                <p className="text-xs text-tertiary mt-1">Your credit pack purchases will appear here.</p>
              </div>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-default text-left">
                    <th className="px-4 py-3 text-xs font-semibold text-secondary">Pack</th>
                    <th className="px-4 py-3 text-xs font-semibold text-secondary">Responses</th>
                    <th className="px-4 py-3 text-xs font-semibold text-secondary">Status</th>
                    <th className="px-4 py-3 text-xs font-semibold text-secondary">Date</th>
                    <th className="px-4 py-3 text-xs font-semibold text-secondary">Expires</th>
                  </tr>
                </thead>
                <tbody>
                  {purchaseHistory.map(record => (
                    <tr key={record.id} className="border-b border-default last:border-0 hover:bg-inset/50 transition-colors">
                      <td className="px-4 py-3 font-medium text-primary capitalize">
                        {packLabel(record.pack_type)}
                      </td>
                      <td className="px-4 py-3 text-secondary">
                        {record.amount.toLocaleString()}
                      </td>
                      <td className="px-4 py-3">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold ${
                          record.status === 'completed'
                            ? 'bg-green-500/10 text-green-500'
                            : record.status === 'pending'
                            ? 'bg-amber-500/10 text-amber-500'
                            : 'bg-red-500/10 text-red-500'
                        }`}>
                          {record.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-secondary text-xs">
                        {new Date(record.purchased_at).toLocaleDateString()}
                      </td>
                      <td className="px-4 py-3 text-secondary text-xs">
                        {new Date(record.expires_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      </div>

      {/* Pulse upgrade modal — per-pack buy buttons */}
      {showPulseModal && createPortal(
        <div className="fixed inset-0 z-[15000] flex items-center justify-center overflow-y-auto p-4 sm:p-6 bg-overlay backdrop-blur-sm" onClick={() => setShowPulseModal(false)} style={{ animation: 'noantOverlayIn 200ms ease forwards' }}>
          <div className="bg-surface rounded-2xl shadow-2xl max-w-sm w-full mx-auto max-h-[calc(100dvh-2rem)] sm:max-h-[calc(100dvh-3rem)] p-5 sm:p-6 relative overflow-hidden flex flex-col" onClick={e => e.stopPropagation()} style={{ animation: 'noantModalIn 200ms ease forwards' }}>
            <button onClick={() => setShowPulseModal(false)} className="absolute top-4 right-4 w-7 h-7 rounded-lg bg-inset hover:bg-surface-hover flex items-center justify-center text-secondary hover:text-primary transition-colors" aria-label="Close">
              <X className="w-4 h-4" />
            </button>

            <div className="w-11 h-11 rounded-full bg-orange-500/10 flex items-center justify-center mb-4">
              <Flame className="w-5 h-5 text-orange-500" strokeWidth={1.5} />
            </div>

            <h2 className="text-base sm:text-lg font-semibold text-primary pr-8">Buy Response Credits</h2>
            <p className="text-xs sm:text-sm text-secondary mt-1 mb-5">Choose a pack — credits never expire within 30 days</p>

            <div className="space-y-2 mb-2 overflow-y-auto min-h-0">
              {pulsePacks.map(pack => (
                <div key={pack.id} className={`flex items-center justify-between p-3 rounded-xl border transition-all ${
                  pack.badge ? 'border-noant-sky/40 bg-noant-sky/[0.04]' : 'border-default bg-inset/50'
                }`}>
                  <div>
                    <div className="flex items-center gap-1.5">
                      <span className="text-sm font-semibold text-primary">{pack.name}</span>
                      {pack.badge && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-noant-sky/10 text-noant-sky font-semibold">{pack.badge}</span>
                      )}
                    </div>
                    <div className="text-xs text-secondary mt-0.5">
                      {pack.responses.toLocaleString()} responses · {formatPrice(currency === 'NGN' ? pack.priceNGN : pack.priceUSD, currency)}
                    </div>
                  </div>
                  <button
                    onClick={() => handlePurchasePack(pack.id)}
                    disabled={loading !== null}
                    className="ml-3 px-3 py-1.5 rounded-lg text-xs font-semibold bg-noant-sky text-white hover:bg-noant-sky-deep disabled:opacity-50 transition-all active:scale-[0.97]"
                  >
                    {loading === 'purchase_' + pack.id ? '…' : 'Buy'}
                  </button>
                </div>
              ))}
            </div>

            {userCredit && userCredit.balance > 0 && (
              <div className="mt-3 p-3 rounded-xl bg-orange-500/[0.06] border border-orange-500/20 text-center">
                <div className="text-xs text-secondary">Current balance</div>
                <div className="text-sm font-bold text-orange-500">{userCredit.balance.toLocaleString()} responses</div>
                {userCredit.expires_at && (
                  <div className="text-[10px] text-secondary mt-0.5">Expires {new Date(userCredit.expires_at).toLocaleDateString()}</div>
                )}
              </div>
            )}
          </div>

          <style>{`
            @keyframes noantOverlayIn { from { opacity:0 } to { opacity:1 } }
            @keyframes noantModalIn { from { opacity:0; transform:scale(.95) } to { opacity:1; transform:scale(1) } }
          `}</style>
        </div>,
        document.body
      )}
    </div>
  )
}
