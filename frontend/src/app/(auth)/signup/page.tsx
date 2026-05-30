import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { signup } from '@/lib/auth'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useToast } from '@/components/ui/Toast'
import { Eye, EyeOff } from 'lucide-react'

export default function SignupPage() {
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [company, setCompany] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { toast } = useToast()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!firstName || !lastName || !email || !password) {
      toast('Please fill in all required fields', 'error')
      return
    }
    setLoading(true)
    try {
      await signup({
        first_name: firstName,
        last_name: lastName,
        email,
        password,
        company_name: company,
      })
      toast('Account created!', 'success')
      navigate('/')
    } catch (err: any) {
      toast(err?.message || 'Failed to create account', 'error')
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

      <h1 className="text-xl lg:text-3xl font-bold text-primary mb-0.5">Get started</h1>
      <p className="text-xs lg:text-sm text-secondary mb-4">Create your AI customer support agent</p>

      <form onSubmit={handleSubmit} className="space-y-3">
        <div className="grid grid-cols-2 gap-2">
          <div>
            <label className="block text-xs font-semibold text-primary mb-1">First name</label>
            <Input
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              placeholder="John"
              autoComplete="given-name"
              className="h-9 text-sm"
            />
          </div>
          <div>
            <label className="block text-xs font-semibold text-primary mb-1">Last name</label>
            <Input
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              placeholder="Doe"
              autoComplete="family-name"
              className="h-9 text-sm"
            />
          </div>
        </div>
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
              placeholder="Min 8 chars"
              autoComplete="new-password"
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
        <div>
          <label className="block text-xs font-semibold text-primary mb-1">Company <span className="text-tertiary font-normal">(opt)</span></label>
          <Input
            value={company}
            onChange={(e) => setCompany(e.target.value)}
            placeholder="Acme Inc"
            autoComplete="organization"
            className="h-9 text-sm"
          />
        </div>
        <Button type="submit" className="w-full h-10 text-sm" loading={loading}>
          Create account
        </Button>
      </form>

      <p className="mt-4 text-center text-xs text-secondary">
        Have an account?{' '}
        <Link to="/login" className="text-noant-sky hover:text-noant-sky-deep font-semibold transition-colors">
          Sign in
        </Link>
      </p>
    </div>
  )
}
