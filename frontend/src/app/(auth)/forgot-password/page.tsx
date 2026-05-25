import { useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../../lib/api'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useToast } from '@/components/ui/Toast'
import { Mail, ArrowLeft, CheckCircle } from 'lucide-react'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)
  const { toast } = useToast()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email) {
      toast('Please enter your email address', 'error')
      return
    }
    setLoading(true)
    try {
      await api.post('/auth/forgot-password', { email })
      setSent(true)
    } catch (err: any) {
      // Always show the generic message to prevent email enumeration
      setSent(true)
    } finally {
      setLoading(false)
    }
  }

  if (sent) {
    return (
      <div className="w-full animate-fade-in flex flex-col h-full lg:h-auto justify-center items-center text-center">
        <div className="w-16 h-16 rounded-full bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center mb-6">
          <CheckCircle className="w-8 h-8 text-noant-sky" />
        </div>
        <h1 className="text-xl lg:text-2xl font-bold text-primary mb-2">Check your inbox</h1>
        <p className="text-sm text-secondary mb-2 max-w-xs">
          If an account exists for <strong className="text-primary">{email}</strong>, we've sent a password reset link.
        </p>
        <p className="text-xs text-tertiary mb-8">
          The link expires in 1 hour. Check your spam folder if you don't see it.
        </p>
        <Link
          to="/login"
          className="flex items-center gap-2 text-sm text-noant-sky hover:text-noant-sky-deep transition-colors font-medium"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to sign in
        </Link>
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

      <div className="w-12 h-12 rounded-xl bg-noant-sky/10 border border-noant-sky/20 flex items-center justify-center mb-5">
        <Mail className="w-6 h-6 text-noant-sky" />
      </div>

      <h1 className="text-xl lg:text-3xl font-bold text-primary mb-0.5">Forgot your password?</h1>
      <p className="text-xs lg:text-sm text-secondary mb-6">
        No worries — enter your email and we'll send a reset link.
      </p>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-semibold text-primary mb-1">Email address</label>
          <Input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@company.com"
            autoComplete="email"
            className="h-9 text-sm"
            autoFocus
          />
        </div>
        <Button type="submit" className="w-full h-10 text-sm" loading={loading}>
          Send reset link
        </Button>
      </form>

      <Link
        to="/login"
        className="mt-5 flex items-center justify-center gap-2 text-xs text-secondary hover:text-primary transition-colors"
      >
        <ArrowLeft className="w-3.5 h-3.5" />
        Back to sign in
      </Link>
    </div>
  )
}
