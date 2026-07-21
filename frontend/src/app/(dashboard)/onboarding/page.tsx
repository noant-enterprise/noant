import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { useToast } from '@/components/ui/Toast'
import { api } from '@/lib/api'
import {
  Building2, GraduationCap, Check, ArrowRight,
  Smartphone, Sparkles, Loader2, RefreshCw,
  AlertTriangle, MessageCircle, Bot,
} from 'lucide-react'

type Step = 'profile' | 'training' | 'whatsapp' | 'complete'

interface IndustryTemplate {
  id: string
  name: string
  icon: string
  categories: string[]
}

const industryIcons: Record<string, React.ElementType> = {
  'shopping-bag': Building2,
  'utensils-crossed': Sparkles,
  'stethoscope': Sparkles,
  'graduation-cap': GraduationCap,
  'sparkles': Sparkles,
  'dumbbell': Sparkles,
  'building': Building2,
  'hotel': Sparkles,
  'car': Sparkles,
  'briefcase': Building2,
}

const steps = [
  { id: 'profile', label: 'Profile', icon: Building2 },
  { id: 'training', label: 'Training', icon: GraduationCap },
  { id: 'whatsapp', label: 'WhatsApp', icon: Smartphone },
  { id: 'complete', label: 'Go Live', icon: Sparkles },
]

export default function OnboardingPage() {
  const navigate = useNavigate()
  const { user, refreshUser } = useAuth()
  const { toast } = useToast()
  const [currentStep, setCurrentStep] = useState<Step>('profile')
  const [loading, setLoading] = useState(false)
  const [industries, setIndustries] = useState<IndustryTemplate[]>([])
  const [completed, setCompleted] = useState<Set<string>>(new Set())

  useEffect(() => {
    api.get<{
      status: string
      steps: { id: string; completed: boolean }[]
    }>('/onboarding/status').then(res => {
      const doneSteps = new Set(res.steps.filter(s => s.completed).map(s => s.id))
      setCompleted(doneSteps)
      if (doneSteps.has('complete')) {
        navigate('/dashboard', { replace: true })
        return
      }
      if (doneSteps.has('whatsapp')) { setCurrentStep('complete'); return }
      if (doneSteps.has('training')) { setCurrentStep('whatsapp'); return }
      if (doneSteps.has('profile')) { setCurrentStep('training'); return }
    }).catch(() => {})

    api.get<{ industries: IndustryTemplate[] }>('/onboarding/industry-templates')
      .then(res => setIndustries(res.industries))
      .catch(() => {})
  }, [])

  const completeStep = async (step: Step, extra?: Record<string, unknown>) => {
    setLoading(true)
    try {
      await api.post('/onboarding/step', { step, ...extra })
      setCompleted(prev => new Set(prev).add(step))
      await refreshUser()
      return true
    } catch {
      toast('Failed to save progress', 'error')
      return false
    } finally {
      setLoading(false)
    }
  }

  const advanceTo = (step: Step) => {
    const order: Step[] = ['profile', 'training', 'whatsapp', 'complete']
    const nextIdx = order.indexOf(step)
    const nextStep = order[nextIdx]
    if (!nextStep) return
    setCurrentStep(nextStep)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  return (
    <div className="max-w-3xl mx-auto px-4 pt-8 pb-20 animate-page-in">
      {/* Stepper */}
      <div className="mb-10">
        <div className="flex items-center justify-between">
          {steps.map((s, i) => {
            const isActive = s.id === currentStep
            const isDone = completed.has(s.id)
            const Icon = s.icon
            return (
              <div key={s.id} className="flex items-center flex-1">
                <div className="flex flex-col items-center">
                  <div className={`w-10 h-10 rounded-full flex items-center justify-center transition-all ${
                    isDone ? 'bg-primary text-white' :
                    isActive ? 'bg-primary/20 text-primary ring-2 ring-primary' :
                    'bg-surface text-tertiary'
                  }`}>
                    {isDone ? <Check size={18} /> : <Icon size={18} />}
                  </div>
                  <span className={`text-xs mt-1.5 font-medium ${
                    isActive || isDone ? 'text-primary' : 'text-tertiary'
                  }`}>{s.label}</span>
                </div>
                {i < steps.length - 1 && (
                  <div className={`flex-1 h-0.5 mx-3 rounded ${
                    isDone ? 'bg-primary' : 'bg-border'
                  }`} />
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* Step Content */}
      {currentStep === 'profile' && (
        <ProfileStep
          user={user}
          industries={industries}
          onComplete={async (industry) => {
            const ok = await completeStep('profile', { industry })
            if (ok) advanceTo('training')
          }}
        />
      )}

      {currentStep === 'training' && (
        <TrainingStep
          industries={industries}
          userIndustry={user?.industry}
          onSkip={async () => {
            const ok = await completeStep('training')
            if (ok) advanceTo('whatsapp')
          }}
          loading={loading}
        />
      )}

      {currentStep === 'whatsapp' && (
        <WhatsAppStep
          onComplete={async () => {
            const ok = await completeStep('whatsapp')
            if (ok) advanceTo('complete')
          }}
          onSkip={async () => {
            const ok = await completeStep('whatsapp')
            if (ok) advanceTo('complete')
          }}
        />
      )}

      {currentStep === 'complete' && (
        <CompleteStep onFinish={() => navigate('/dashboard', { replace: true })} />
      )}
    </div>
  )
}

// ===== Step 1: Profile =====
function ProfileStep({
  user, industries, onComplete,
}: {
  user: any
  industries: IndustryTemplate[]
  onComplete: (industry: string) => void
}) {
  const [selectedIndustry, setSelectedIndustry] = useState(user?.industry || '')
  const [saving, setSaving] = useState(false)

  const handleSubmit = async () => {
    if (!selectedIndustry) { return }
    setSaving(true)
    onComplete(selectedIndustry)
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold">Welcome to Noant AI</h2>
        <p className="text-secondary text-sm mt-1">
          Let's get your AI assistant set up. First, tell us about your business.
        </p>
      </div>

      {/* Company Info Summary */}
      <div className="bg-surface rounded-xl p-5 border border-border space-y-3">
        <h3 className="text-sm font-medium text-secondary">Company Details</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-tertiary">Name</span>
            <p className="font-medium">{user?.company_name || '—'}</p>
          </div>
          <div>
            <span className="text-tertiary">Email</span>
            <p className="font-medium">{user?.email || '—'}</p>
          </div>
        </div>
        <p className="text-xs text-tertiary">
          You can update these in Settings later.
        </p>
      </div>

      {/* Industry Picker */}
      <div>
        <label className="text-sm font-medium mb-3 block">What industry are you in?</label>
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
          {industries.map(ind => {
            const Icon = industryIcons[ind.icon] || Building2
            const isSelected = selectedIndustry === ind.id
            return (
              <button
                key={ind.id}
                onClick={() => setSelectedIndustry(ind.id)}
                className={`flex flex-col items-center gap-2 p-4 rounded-xl border text-center transition-all ${
                  isSelected
                    ? 'border-primary bg-primary/5 ring-1 ring-primary'
                    : 'border-border hover:border-primary/40 bg-surface'
                }`}
              >
                <Icon size={24} className={isSelected ? 'text-primary' : 'text-tertiary'} />
                <span className="text-xs font-medium leading-tight">{ind.name}</span>
              </button>
            )
          })}
        </div>
      </div>

      <button
        onClick={handleSubmit}
        disabled={!selectedIndustry || saving}
        className="w-full btn-primary flex items-center justify-center gap-2 py-3 rounded-xl disabled:opacity-50"
      >
        {saving ? <Loader2 size={18} className="animate-spin" /> : <ArrowRight size={18} />}
        Continue
      </button>
    </div>
  )
}

// ===== Step 2: Training =====
function TrainingStep({
  industries, userIndustry, onSkip, loading,
}: {
  industries: IndustryTemplate[]
  userIndustry?: string
  onSkip: () => void
  loading: boolean
}) {
  const [creating, setCreating] = useState(false)
  const [created, setCreated] = useState(false)
  const template = industries.find(i => i.id === userIndustry)

  const handleAutoCreate = async () => {
    if (!userIndustry) return
    setCreating(true)
    try {
      await api.post('/onboarding/categories/auto-create', { industry_id: userIndustry })
      setCreated(true)
    } catch {
      // ignore
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold">Train Your AI</h2>
        <p className="text-secondary text-sm mt-1">
          Give your AI assistant knowledge about your business so it can answer customer questions.
        </p>
      </div>

      {/* Auto-create from industry */}
      {template && (
        <div className="bg-surface rounded-xl p-5 border border-border space-y-4">
          <div className="flex items-start gap-3">
            <Bot size={24} className="text-primary shrink-0 mt-0.5" />
            <div>
              <h3 className="font-medium text-sm">Auto-generate from "{template.name}" template</h3>
              <p className="text-xs text-tertiary mt-1">
                We'll create {template.categories.length} categories with common questions for your industry.
                You can always add more later.
              </p>
            </div>
          </div>
          {template.categories.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {template.categories.slice(0, 5).map(c => (
                <span key={c} className="text-xs px-2.5 py-1 bg-primary/10 text-primary rounded-full">{c}</span>
              ))}
              {template.categories.length > 5 && (
                <span className="text-xs px-2.5 py-1 bg-surface text-tertiary rounded-full">
                  +{template.categories.length - 5} more
                </span>
              )}
            </div>
          )}
          <div className="flex gap-3">
            <button
              onClick={handleAutoCreate}
              disabled={creating || created}
              className="flex-1 btn-primary flex items-center justify-center gap-2 py-2.5 rounded-xl text-sm disabled:opacity-50"
            >
              {creating ? <Loader2 size={16} className="animate-spin" /> :
               created ? <Check size={16} /> : <Sparkles size={16} />}
              {created ? 'Created!' : 'Auto-generate Categories'}
            </button>
          </div>
        </div>
      )}

      {/* Manual option */}
      <div className="bg-surface rounded-xl p-5 border border-border space-y-3">
        <h3 className="font-medium text-sm">Or add your own knowledge</h3>
        <p className="text-xs text-tertiary">
          Go to the <strong>Teach</strong> page later to add Q&A pairs, upload documents, or write your own training data.
        </p>
      </div>

      <div className="flex gap-3">
        <button
          onClick={onSkip}
          disabled={loading}
          className="flex-1 py-2.5 rounded-xl border border-border text-sm font-medium hover:bg-surface transition-colors disabled:opacity-50"
        >
          {created ? 'Continue' : 'Skip for now'}
        </button>
      </div>
    </div>
  )
}

// ===== Step 3: WhatsApp =====
function WhatsAppStep({
  onComplete, onSkip,
}: {
  onComplete: () => void
  onSkip: () => void
}) {
  const [step, setStep] = useState<'form' | 'qr' | 'connecting' | 'success' | 'error'>('form')
  const [phone, setPhone] = useState('')
  const [sessionId, setSessionId] = useState('')
  const [qrCode, setQrCode] = useState('')
  const [errorMsg, setErrorMsg] = useState('')
  const pollRef = useRef<ReturnType<typeof setInterval>>()
  const { toast } = useToast()

  const startPolling = useCallback((sid: string) => {
    const poll = async () => {
      try {
        const res = await api.get<{
          status: string
          connected: boolean
          qr_code?: string
        }>(`/channels/whatsapp/status/${sid}?poll=false&t=${Date.now()}`)
        if (res.connected) {
          setStep('success')
          clearInterval(pollRef.current)
          return
        }
        if (res.qr_code && res.qr_code !== qrCode) {
          setQrCode(res.qr_code)
        }
        if (res.status === 'expired' || res.status === 'failed') {
          setErrorMsg(res.status === 'expired' ? 'QR code expired' : 'Connection failed')
          setStep('error')
          clearInterval(pollRef.current)
        }
      } catch {
        // retry
      }
    }
    poll()
    pollRef.current = setInterval(poll, 3000)
  }, [qrCode])

  useEffect(() => {
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [])

  const handleConnect = async () => {
    const cleaned = phone.startsWith('+') ? phone : `+${phone.replace(/[^0-9]/g, '')}`
    if (cleaned.length < 10) {
      toast('Enter a valid phone number with country code (e.g. +447700900123)', 'error')
      return
    }
    setStep('connecting')
    try {
      const res = await api.post<{
        session_id: string
        qr_code?: string
        status: string
      }>('/channels/whatsapp/connect', { phone: cleaned })
      setSessionId(res.session_id)
      if (res.status === 'connected') {
        setStep('success')
        return
      }
      setStep('qr')
      startPolling(res.session_id)
    } catch {
      setErrorMsg('Failed to connect. Check that OpenWA is running.')
      setStep('error')
    }
  }

  const handleRefresh = async () => {
    if (!sessionId) return
    try {
      const res = await api.post<{ session_id: string }>(`/channels/whatsapp/refresh/${sessionId}`)
      setSessionId(res.session_id)
      setQrCode('')
      setStep('qr')
      startPolling(res.session_id)
    } catch {
      toast('Failed to refresh QR', 'error')
    }
  }

  const handleVerify = async () => {
    if (!sessionId) return
    try {
      const res = await api.get<{ connected: boolean }>(
        `/channels/whatsapp/status/${sessionId}?poll=false&t=${Date.now()}`
      )
      if (res.connected) {
        setStep('success')
        clearInterval(pollRef.current)
      } else {
        toast('Not connected yet. Scan the QR code with your phone.', 'error')
      }
    } catch {
      toast('Verification failed', 'error')
    }
  }

  if (step === 'success') {
    return (
      <div className="space-y-6 text-center py-12">
        <div className="w-16 h-16 rounded-full bg-green-500/20 flex items-center justify-center mx-auto">
          <Check size={32} className="text-green-500" />
        </div>
        <div>
          <h2 className="text-xl font-semibold">WhatsApp Connected!</h2>
          <p className="text-secondary text-sm mt-1">Your AI can now send and receive WhatsApp messages.</p>
        </div>
        <button onClick={onComplete} className="btn-primary px-8 py-2.5 rounded-xl text-sm">
          Continue
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold">Connect WhatsApp</h2>
        <p className="text-secondary text-sm mt-1">
          Link your WhatsApp Business number so your AI assistant can chat with customers.
        </p>
      </div>

      {step === 'form' && (
        <div className="bg-surface rounded-xl p-5 border border-border space-y-4">
          <div>
            <label className="text-sm font-medium mb-1.5 block">WhatsApp Business Number</label>
            <input
              type="tel"
              value={phone}
              onChange={e => setPhone(e.target.value)}
              placeholder="+447700900123"
              className="w-full px-4 py-2.5 rounded-lg border border-border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/40"
            />
          </div>
          <button
            onClick={handleConnect}
            disabled={!phone}
            className="w-full btn-primary flex items-center justify-center gap-2 py-2.5 rounded-xl text-sm disabled:opacity-50"
          >
            <MessageCircle size={16} /> Connect WhatsApp
          </button>
          <button onClick={onSkip} className="w-full text-sm text-tertiary hover:text-secondary transition-colors">
            Skip — I'll do this later
          </button>
        </div>
      )}

      {step === 'connecting' && (
        <div className="flex flex-col items-center py-12 gap-4">
          <Loader2 size={32} className="animate-spin text-primary" />
          <p className="text-sm text-secondary">Initializing session...</p>
        </div>
      )}

      {(step === 'qr' || step === 'error') && (
        <div className="bg-surface rounded-xl p-5 border border-border space-y-4">
          {step === 'error' && (
            <div className="flex items-center gap-2 text-red-500 text-sm mb-2">
              <AlertTriangle size={16} /> {errorMsg}
            </div>
          )}

          {qrCode && step !== 'error' && (
            <div className="flex justify-center">
              <div className="w-56 h-56 rounded-xl border border-border p-3 bg-white">
                <img src={qrCode} alt="WhatsApp QR" className="w-full h-full object-contain" />
              </div>
            </div>
          )}

          {!qrCode && step !== 'error' && (
            <div className="flex flex-col items-center py-8 gap-3">
              <Loader2 size={24} className="animate-spin text-primary" />
              <p className="text-xs text-tertiary">Generating QR code...</p>
            </div>
          )}

          {qrCode && (
            <div className="text-center">
              <p className="text-sm font-medium">Scan this QR code with your phone</p>
              <p className="text-xs text-tertiary mt-1">
                Open WhatsApp {'>'} Linked Devices {'>'} Link a Device
              </p>
            </div>
          )}

          <div className="flex gap-3">
            <button onClick={handleRefresh} className="flex-1 py-2.5 rounded-xl border border-border text-sm flex items-center justify-center gap-2 hover:bg-surface transition-colors">
              <RefreshCw size={14} /> Refresh QR
            </button>
            <button onClick={handleVerify} className="flex-1 btn-primary py-2.5 rounded-xl text-sm flex items-center justify-center gap-2">
              <Check size={14} /> I've Scanned
            </button>
          </div>

          <button onClick={onSkip} className="w-full text-xs text-tertiary hover:text-secondary transition-colors">
            Skip this step
          </button>
        </div>
      )}
    </div>
  )
}

// ===== Step 4: Complete =====
function CompleteStep({ onFinish }: { onFinish: () => void }) {
  return (
    <div className="space-y-6 text-center py-12">
      <div className="w-20 h-20 rounded-full bg-primary/20 flex items-center justify-center mx-auto animate-breathe">
        <Sparkles size={40} className="text-primary" />
      </div>

      <div>
        <h2 className="text-2xl font-semibold">You're all set!</h2>
        <p className="text-secondary text-sm mt-2 max-w-md mx-auto">
          Your AI assistant is ready to start answering customer questions.
          Head to the dashboard to see it in action.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 max-w-lg mx-auto">
        {[
          { label: 'Add Q&A', desc: 'Teach your AI more', icon: GraduationCap },
          { label: 'View Chats', desc: 'See conversations', icon: MessageCircle },
          { label: 'Analytics', desc: 'Track performance', icon: Sparkles },
        ].map(item => (
          <div key={item.label} className="bg-surface rounded-xl p-4 border border-border text-center">
            <item.icon size={20} className="mx-auto text-primary mb-2" />
            <p className="text-sm font-medium">{item.label}</p>
            <p className="text-xs text-tertiary">{item.desc}</p>
          </div>
        ))}
      </div>

      <button onClick={onFinish} className="btn-primary px-10 py-3 rounded-xl text-sm font-medium">
        Go to Dashboard
      </button>
    </div>
  )
}
