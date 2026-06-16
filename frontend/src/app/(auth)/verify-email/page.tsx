import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { verifyEmail, resendVerification } from '@/lib/auth'
import { useToast } from '@/components/ui/Toast'
import { Button } from '@/components/ui/Button'
import { CheckCircle, MailOpen, RefreshCw, ArrowLeft, ShieldCheck } from 'lucide-react'

const RESEND_COOLDOWN = 60 // seconds

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const email = searchParams.get('email') || ''
  const navigate = useNavigate()
  const { toast } = useToast()

  const [code, setCode] = useState<string[]>(['', '', '', '', '', ''])
  const [loading, setLoading] = useState(false)
  const [resending, setResending] = useState(false)
  const [verified, setVerified] = useState(false)
  const [cooldown, setCooldown] = useState(0)
  const [shake, setShake] = useState(false)
  const inputRefs = useRef<(HTMLInputElement | null)[]>([])

  // Start initial cooldown so user can't instantly spam resend
  useEffect(() => {
    setCooldown(RESEND_COOLDOWN)
  }, [])

  useEffect(() => {
    if (cooldown <= 0) return
    const t = setTimeout(() => setCooldown((c) => c - 1), 1000)
    return () => clearTimeout(t)
  }, [cooldown])

  // Auto-focus first box
  useEffect(() => {
    inputRefs.current[0]?.focus()
  }, [])

  const handleChange = (index: number, value: string) => {
    const digit = value.replace(/\D/g, '').slice(-1)
    const next = [...code]
    next[index] = digit
    setCode(next)
    if (digit && index < 5) {
      inputRefs.current[index + 1]?.focus()
    }
  }

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !code[index] && index > 0) {
      inputRefs.current[index - 1]?.focus()
    }
  }

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault()
    const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6)
    if (pasted.length === 6) {
      setCode(pasted.split(''))
      inputRefs.current[5]?.focus()
    }
  }

  const handleVerify = useCallback(async () => {
    const fullCode = code.join('')
    if (fullCode.length < 6) {
      toast('Please enter the complete 6-digit code', 'error')
      return
    }
    setLoading(true)
    try {
      await verifyEmail(email, fullCode)
      setVerified(true)
      toast('Email verified! Welcome to Noant 🎉', 'success')
      setTimeout(() => navigate('/dashboard'), 1800)
    } catch (err: any) {
      setShake(true)
      setTimeout(() => setShake(false), 600)
      setCode(['', '', '', '', '', ''])
      inputRefs.current[0]?.focus()
      toast(err?.message === 'invalid_code' ? 'Incorrect code. Please try again.' : 'Verification failed. Please try again.', 'error')
    } finally {
      setLoading(false)
    }
  }, [code, email, navigate, toast])

  // Auto-submit when all 6 digits are entered
  useEffect(() => {
    if (code.every((d) => d !== '') && !loading && !verified) {
      handleVerify()
    }
  }, [code])

  const handleResend = async () => {
    if (cooldown > 0 || resending) return
    setResending(true)
    try {
      await resendVerification(email)
      setCooldown(RESEND_COOLDOWN)
      setCode(['', '', '', '', '', ''])
      inputRefs.current[0]?.focus()
      toast('Verification code resent! Check your inbox.', 'success')
    } catch (err: any) {
      toast(err?.message || 'Failed to resend code', 'error')
    } finally {
      setResending(false)
    }
  }

  if (verified) {
    return (
      <div className="w-full animate-fade-in flex flex-col items-center justify-center gap-5 py-8">
        <div className="w-20 h-20 rounded-full bg-green-500/10 border border-green-500/20 flex items-center justify-center animate-scale-in">
          <CheckCircle className="w-10 h-10 text-green-400" />
        </div>
        <div className="text-center">
          <h1 className="text-2xl font-bold text-primary mb-1">Email verified!</h1>
          <p className="text-sm text-secondary">Redirecting you to your dashboard…</p>
        </div>
        <div className="flex gap-1.5">
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              className="w-2 h-2 rounded-full bg-noant-sky animate-bounce"
              style={{ animationDelay: `${i * 0.15}s` }}
            />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="w-full animate-fade-in flex flex-col h-full lg:h-auto justify-center">
      {/* Mobile logo */}
      <div className="lg:hidden flex items-center justify-center gap-2 mb-4">
        <svg className="w-8 h-8" viewBox="0 0 200 200" fill="none">
          <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" className="text-primary" />
          <circle cx="100" cy="100" r="70" fill="currentColor" className="text-primary" />
          <circle cx="80" cy="100" r="6" fill="var(--bg-base)" />
          <circle cx="100" cy="100" r="8" fill="var(--bg-base)" />
          <circle cx="120" cy="100" r="10" fill="var(--bg-base)" />
        </svg>
        <span className="text-lg font-bold tracking-widest lowercase text-primary">noant</span>
      </div>

      {/* Header */}
      <div className="flex flex-col items-center text-center mb-6">
        <div className="w-14 h-14 rounded-2xl bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center mb-4">
          <MailOpen className="w-7 h-7 text-noant-sky" />
        </div>
        <h1 className="text-xl lg:text-3xl font-bold text-primary mb-1">Check your email</h1>
        <p className="text-xs lg:text-sm text-secondary">
          We sent a 6-digit code to{' '}
          <span className="text-primary font-semibold">{email || 'your email'}</span>
        </p>
      </div>

      {/* Code inputs */}
      <div
        className={`flex justify-center gap-2 mb-5 transition-transform ${shake ? 'animate-shake' : ''}`}
        onPaste={handlePaste}
      >
        {code.map((digit, i) => (
          <input
            key={i}
            ref={(el) => { inputRefs.current[i] = el }}
            type="text"
            inputMode="numeric"
            maxLength={1}
            value={digit}
            onChange={(e) => handleChange(i, e.target.value)}
            onKeyDown={(e) => handleKeyDown(i, e)}
            disabled={loading || verified}
            className={`
              w-11 h-14 text-center text-xl font-bold rounded-xl border-2 outline-none transition-all duration-150
              bg-surface text-primary
              ${digit ? 'border-noant-sky shadow-[0_0_0_3px_rgba(59,130,246,0.15)]' : 'border-default'}
              focus:border-noant-sky focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)]
              disabled:opacity-50 disabled:cursor-not-allowed
            `}
            style={{ caretColor: 'transparent' }}
          />
        ))}
      </div>

      {/* Verify button */}
      <Button
        type="button"
        className="w-full h-10 text-sm mb-4"
        loading={loading}
        disabled={code.join('').length < 6 || loading}
        onClick={handleVerify}
      >
        <ShieldCheck className="w-4 h-4 mr-2" />
        Verify email
      </Button>

      {/* Resend section */}
      <div className="text-center">
        <p className="text-xs text-secondary mb-1">Didn't receive a code?</p>
        <button
          type="button"
          onClick={handleResend}
          disabled={cooldown > 0 || resending}
          className="inline-flex items-center gap-1.5 text-xs font-semibold text-noant-sky hover:text-noant-sky-deep transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${resending ? 'animate-spin' : ''}`} />
          {cooldown > 0 ? `Resend in ${cooldown}s` : resending ? 'Sending…' : 'Resend code'}
        </button>
      </div>

      {/* Back to login */}
      <div className="mt-5 flex justify-center">
        <Link
          to="/login"
          className="inline-flex items-center gap-1 text-xs text-tertiary hover:text-secondary transition-colors"
        >
          <ArrowLeft className="w-3.5 h-3.5" />
          Back to login
        </Link>
      </div>
    </div>
  )
}
