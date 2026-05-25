import { useState, useEffect } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../../../lib/api'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useToast } from '@/components/ui/Toast'
import { Eye, EyeOff, Lock, CheckCircle, XCircle } from 'lucide-react'

function PasswordStrength({ password }: { password: string }) {
  const checks = [
    { label: 'At least 8 characters', ok: password.length >= 8 },
    { label: 'Contains uppercase letter', ok: /[A-Z]/.test(password) },
    { label: 'Contains a number', ok: /\d/.test(password) },
    { label: 'Contains special character', ok: /[^A-Za-z0-9]/.test(password) },
  ]
  const score = checks.filter(c => c.ok).length

  const colors = ['bg-red-500', 'bg-orange-500', 'bg-yellow-500', 'bg-emerald-500']
  const labels = ['Weak', 'Fair', 'Good', 'Strong']

  if (!password) return null

  return (
    <div className="mt-2 space-y-2">
      <div className="flex gap-1">
        {[0, 1, 2, 3].map(i => (
          <div
            key={i}
            className={`h-1 flex-1 rounded-full transition-all duration-300 ${i < score ? colors[score - 1] : 'bg-white/10'}`}
          />
        ))}
      </div>
      <p className="text-xs text-tertiary">
        Strength: <span className={`font-medium ${score >= 3 ? 'text-emerald-400' : score >= 2 ? 'text-yellow-400' : 'text-red-400'}`}>
          {labels[score - 1] || 'Weak'}
        </span>
      </p>
      <ul className="space-y-1">
        {checks.map(check => (
          <li key={check.label} className="flex items-center gap-1.5 text-xs">
            {check.ok
              ? <CheckCircle className="w-3 h-3 text-emerald-400 shrink-0" />
              : <XCircle className="w-3 h-3 text-white/20 shrink-0" />}
            <span className={check.ok ? 'text-emerald-400' : 'text-tertiary'}>{check.label}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { toast } = useToast()

  const token = searchParams.get('token') || ''

  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showNew, setShowNew] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    if (!token) {
      toast('Invalid or missing reset token', 'error')
      navigate('/forgot-password')
    }
  }, [token])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newPassword || !confirmPassword) {
      toast('Please fill in all fields', 'error')
      return
    }
    if (newPassword.length < 8) {
      toast('Password must be at least 8 characters', 'error')
      return
    }
    if (newPassword !== confirmPassword) {
      toast('Passwords do not match', 'error')
      return
    }
    setLoading(true)
    try {
      await api.post('/auth/reset-password', { token, new_password: newPassword })
      setSuccess(true)
      setTimeout(() => navigate('/login'), 3000)
    } catch (err: any) {
      toast(err?.message || 'Reset failed. The link may have expired.', 'error')
    } finally {
      setLoading(false)
    }
  }

  if (success) {
    return (
      <div className="w-full animate-fade-in flex flex-col h-full lg:h-auto justify-center items-center text-center">
        <div className="w-16 h-16 rounded-full bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center mb-6">
          <CheckCircle className="w-8 h-8 text-emerald-400" />
        </div>
        <h1 className="text-xl lg:text-2xl font-bold text-primary mb-2">Password updated!</h1>
        <p className="text-sm text-secondary mb-6 max-w-xs">
          Your password has been reset successfully. Redirecting you to login…
        </p>
        <Link to="/login" className="text-sm text-noant-sky hover:text-noant-sky-deep transition-colors font-medium">
          Go to sign in now →
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
        <Lock className="w-6 h-6 text-noant-sky" />
      </div>

      <h1 className="text-xl lg:text-3xl font-bold text-primary mb-0.5">Set a new password</h1>
      <p className="text-xs lg:text-sm text-secondary mb-6">
        Choose a strong password you haven't used before.
      </p>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-semibold text-primary mb-1">New password</label>
          <div className="relative">
            <Input
              type={showNew ? 'text' : 'password'}
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="Min. 8 characters"
              autoComplete="new-password"
              className="h-9 text-sm pr-10"
              autoFocus
            />
            <button
              type="button"
              onClick={() => setShowNew(!showNew)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-tertiary hover:text-secondary transition-colors"
              tabIndex={-1}
            >
              {showNew ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
          <PasswordStrength password={newPassword} />
        </div>
        <div>
          <label className="block text-xs font-semibold text-primary mb-1">Confirm password</label>
          <div className="relative">
            <Input
              type={showConfirm ? 'text' : 'password'}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Repeat your password"
              autoComplete="new-password"
              className="h-9 text-sm pr-10"
            />
            <button
              type="button"
              onClick={() => setShowConfirm(!showConfirm)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-tertiary hover:text-secondary transition-colors"
              tabIndex={-1}
            >
              {showConfirm ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
          {confirmPassword && newPassword !== confirmPassword && (
            <p className="text-xs text-red-400 mt-1 flex items-center gap-1">
              <XCircle className="w-3 h-3" /> Passwords do not match
            </p>
          )}
          {confirmPassword && newPassword === confirmPassword && (
            <p className="text-xs text-emerald-400 mt-1 flex items-center gap-1">
              <CheckCircle className="w-3 h-3" /> Passwords match
            </p>
          )}
        </div>
        <Button
          type="submit"
          className="w-full h-10 text-sm"
          loading={loading}
          disabled={!newPassword || !confirmPassword || newPassword !== confirmPassword}
        >
          Reset password
        </Button>
      </form>

      <Link
        to="/login"
        className="mt-5 text-center text-xs text-secondary hover:text-primary transition-colors"
      >
        Remember your password? Sign in
      </Link>
    </div>
  )
}
