import { useState } from 'react'
import { Check, Zap, Building2, Rocket, Star, ArrowRight, Crown } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { api } from '../../../lib/api'
import { useToast } from '@/components/ui/Toast'

type Currency = 'NGN' | 'USD'

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
}

const plans: Plan[] = [
  {
    id: 'free',
    name: 'Free',
    icon: Star,
    tagline: 'Try NOANT with no commitment',
    priceNGN: 0,
    priceUSD: 0,
    period: 'forever',
    cta: 'Current plan',
    features: [
      '500 AI responses/month',
      '1 channel',
      '50 Q&A pairs',
      'Basic analytics',
      'Community support',
    ],
  },
  {
    id: 'starter',
    name: 'Starter',
    icon: Zap,
    tagline: 'For small businesses getting started',
    priceNGN: 10000,
    priceUSD: 19,
    period: 'month',
    cta: 'Upgrade to Starter',
    features: [
      '5,000 AI responses/month',
      '3 channels (WhatsApp, Telegram, Web)',
      '500 Q&A pairs',
      'Unknown questions training',
      'Full analytics & insights',
      'Email notifications',
      'Email support',
    ],
  },
  {
    id: 'pro',
    name: 'Pro',
    icon: Rocket,
    tagline: 'For growing teams with heavy volume',
    priceNGN: 35000,
    priceUSD: 49,
    period: 'month',
    cta: 'Upgrade to Pro',
    highlight: true,
    features: [
      '25,000 AI responses/month',
      'Unlimited channels',
      'Unlimited Q&A pairs',
      'CSV bulk upload & re-training',
      'Team members (up to 5)',
      'API access',
      'Priority support',
      'Advanced analytics + export',
    ],
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    icon: Building2,
    tagline: 'Custom solutions for large organisations',
    priceNGN: null,
    priceUSD: null,
    period: 'month',
    cta: 'Contact Sales',
    features: [
      'Unlimited AI responses',
      'Unlimited channels',
      'Dedicated AI training',
      'Unlimited team members',
      'SLA guarantee',
      'Custom integrations',
      'Dedicated account manager',
      'On-premise option',
    ],
  },
]

function formatPrice(amount: number | null, currency: Currency) {
  if (amount === null) return 'Custom'
  if (amount === 0) return 'Free'
  if (currency === 'NGN') return `₦${amount.toLocaleString()}`
  return `$${amount}`
}

export default function BillingPage() {
  const { user } = useAuth()
  const { toast: showToast } = useToast()
  const [currency, setCurrency] = useState<Currency>('NGN')
  const [loading, setLoading] = useState<string | null>(null)
  const currentPlan = user?.plan || 'free'

  const handleUpgrade = async (planId: string) => {
    if (planId === 'free' || planId === currentPlan) return
    if (planId === 'enterprise') {
      window.open('mailto:hello@noant.ai?subject=Enterprise Enquiry', '_blank')
      return
    }
    setLoading(planId)
    try {
      const res = await api.post<{ checkout_url?: string }>('/payments/subscribe', { plan: planId, currency })
      if (res?.checkout_url) {
        window.open(res.checkout_url, '_blank')
      } else {
        showToast('Upgrade initiated! Check your email.', 'success')
      }
    } catch {
      showToast('Failed to start upgrade. Please try again.', 'error')
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

          {/* Currency toggle */}
          <div className="inline-flex items-center gap-1 mt-4 bg-inset p-1 rounded-xl border border-default">
            {(['NGN', 'USD'] as Currency[]).map(c => (
              <button
                key={c}
                onClick={() => setCurrency(c)}
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

        {/* Trial banner */}
        {currentPlan === 'free' && (
          <div className="mb-6 p-4 rounded-xl bg-gradient-to-r from-noant-sky/20 to-purple-500/10 border border-noant-sky/30 flex items-center gap-3">
            <Crown className="w-5 h-5 text-noant-sky shrink-0" />
            <div>
              <p className="text-sm font-semibold text-primary">You're on the Free plan</p>
              <p className="text-xs text-secondary">Upgrade to unlock unlimited responses, more channels, and team collaboration.</p>
            </div>
          </div>
        )}

        {/* Plan grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {plans.map(plan => {
            const Icon = plan.icon
            const isCurrent = plan.id === currentPlan
            const isDisabled = isCurrent || loading !== null

            return (
              <div
                key={plan.id}
                className={`relative flex flex-col rounded-2xl border p-5 transition-all ${
                  plan.highlight
                    ? 'border-noant-sky bg-gradient-to-b from-noant-sky/10 to-transparent shadow-sky'
                    : 'border-default bg-surface hover:border-strong'
                }`}
              >
                {plan.highlight && (
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2 px-3 py-0.5 rounded-full bg-noant-sky text-white text-xs font-bold whitespace-nowrap">
                    Most Popular
                  </div>
                )}

                <div className="mb-4">
                  <div className={`w-10 h-10 rounded-xl mb-3 flex items-center justify-center ${plan.highlight ? 'bg-noant-sky text-white' : 'bg-inset text-secondary'}`}>
                    <Icon className="w-5 h-5" />
                  </div>
                  <div className="font-bold text-primary text-lg">{plan.name}</div>
                  <div className="text-xs text-secondary mt-0.5 leading-snug">{plan.tagline}</div>
                </div>

                <div className="mb-4">
                  <div className="flex items-baseline gap-1">
                    <span className="text-3xl font-black text-primary">
                      {formatPrice(currency === 'NGN' ? plan.priceNGN : plan.priceUSD, currency)}
                    </span>
                    {plan.priceNGN !== null && plan.priceNGN > 0 && (
                      <span className="text-sm text-secondary">/{plan.period}</span>
                    )}
                  </div>
                </div>

                <ul className="space-y-2 flex-1 mb-5">
                  {plan.features.map(f => (
                    <li key={f} className="flex items-start gap-2 text-xs text-secondary">
                      <Check className="w-3.5 h-3.5 text-noant-sky shrink-0 mt-0.5" />
                      {f}
                    </li>
                  ))}
                </ul>

                <button
                  onClick={() => handleUpgrade(plan.id)}
                  disabled={isDisabled}
                  className={`w-full flex items-center justify-center gap-1.5 py-2.5 rounded-xl text-sm font-semibold transition-all active:scale-[0.98] ${
                    isCurrent
                      ? 'bg-inset text-secondary cursor-default border border-default'
                      : plan.highlight
                      ? 'bg-noant-sky text-white hover:bg-noant-sky-deep shadow-sky'
                      : 'bg-noant-black text-white dark:bg-white dark:text-noant-black hover:opacity-90'
                  } disabled:opacity-60`}
                >
                  {isCurrent ? (
                    <><Check className="w-3.5 h-3.5" /> Current plan</>
                  ) : loading === plan.id ? (
                    'Loading…'
                  ) : (
                    <>{plan.cta} <ArrowRight className="w-3.5 h-3.5" /></>
                  )}
                </button>
              </div>
            )
          })}
        </div>

        {/* Payment history placeholder */}
        <div className="mt-10">
          <h2 className="text-base font-semibold text-primary mb-3">Payment History</h2>
          <div className="rounded-xl border border-default bg-surface overflow-hidden">
            <div className="p-8 text-center">
              <div className="w-10 h-10 rounded-xl bg-inset flex items-center justify-center mx-auto mb-3">
                <Star className="w-5 h-5 text-tertiary" />
              </div>
              <p className="text-sm text-secondary">No payment history yet.</p>
              <p className="text-xs text-tertiary mt-1">Invoices will appear here after your first charge.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
