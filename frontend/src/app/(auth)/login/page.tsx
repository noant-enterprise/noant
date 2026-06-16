import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { login } from '@/lib/auth'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useToast } from '@/components/ui/Toast'
import { Eye, EyeOff } from 'lucide-react'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { toast } = useToast()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email || !password) {
      toast('Please fill in all fields', 'error')
      return
    }
    setLoading(true)
    try {
      await login({ email, password })
      toast('Welcome back!', 'success')
      navigate('/dashboard')
    } catch (err: any) {
      // If backend says email_not_verified, redirect to the verification page
      if (err?.code === 'email_not_verified' || err?.message === 'email_not_verified') {
        toast('Please verify your email before signing in.', 'error')
        navigate(`/verify-email?email=${encodeURIComponent(email)}`)
        return
      }
      toast(err?.message || 'Invalid credentials', 'error')
    } finally {
      setLoading(false)
    }
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

      <h1 className="text-xl lg:text-3xl font-bold text-primary mb-0.5">Welcome back</h1>
      <p className="text-xs lg:text-sm text-secondary mb-4">Sign in to manage your AI customer support</p>

      <form onSubmit={handleSubmit} className="space-y-3">
        <div>
          <label className="block text-xs font-semibold text-primary mb-1">Email</label>
          <Input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@company.com"
            autoComplete="email"
            className="h-9 text-sm"
          />
        </div>
        <div>
          <label className="block text-xs font-semibold text-primary mb-1">Password</label>
          <div className="relative">
            <Input
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="********"
              autoComplete="current-password"
              className="h-9 text-sm pr-10"
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-tertiary hover:text-secondary transition-colors"
              tabIndex={-1}
            >
              {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
        </div>
        <div className="flex items-center justify-between">
          <label className="flex items-center gap-2 text-xs text-secondary cursor-pointer">
            <input type="checkbox" className="w-3.5 h-3.5 rounded border-default accent-noant-sky" />
            Remember me
          </label>
          <Link to="/forgot-password" className="text-xs text-noant-sky hover:text-noant-sky-deep transition-colors">
            Forgot?
          </Link>
        </div>
        <Button type="submit" className="w-full h-10 text-sm" loading={loading}>
          Sign in
        </Button>
      </form>

      <p className="mt-4 text-center text-xs text-secondary">
        No account?{' '}
        <Link to="/signup" className="text-noant-sky hover:text-noant-sky-deep font-semibold transition-colors">
          Get started
        </Link>
      </p>
    </div>
  )
}
