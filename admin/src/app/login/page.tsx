import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '@/lib/hooks/useAuth'
import { Loader2 } from 'lucide-react'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<{ email?: string; password?: string }>({})
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()

  const mapError = (err: unknown): string => {
    if (err instanceof TypeError && err.message === 'Failed to fetch') {
      return 'Unable to connect to the server. Please try again.'
    }
    const msg = err instanceof Error ? err.message : ''
    if (/5\d\d/.test(msg) || /internal server/i.test(msg)) {
      return 'Something went wrong on our end. Please try again later.'
    }
    return 'Invalid email or password'
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    const next: typeof fieldErrors = {}
    if (!email.trim()) next.email = 'Email is required'
    if (!password) next.password = 'Password is required'
    setFieldErrors(next)
    if (Object.keys(next).length > 0) return

    setLoading(true)
    try {
      await login(email, password)
      navigate('/')
    } catch (err) {
      setError(mapError(err))
    } finally {
      setLoading(false)
    }
  }

  const inputBase =
    'w-full rounded-lg border bg-bg-inset px-3 py-2.5 text-sm text-text-primary outline-none transition-colors focus:border-brand-sky focus:ring-2 focus:ring-brand-sky/20'

  return (
    <div className="flex min-h-screen items-center justify-center bg-bg-base">
      <div className="w-full max-w-sm space-y-8">
        <div className="text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-brand-sky/10">
            <img src="/Logo A.png" alt="NOANT" className="h-10 w-10" />
          </div>
          <h1 className="text-2xl font-bold text-text-primary">NOANT Admin</h1>
          <p className="mt-2 text-sm text-text-tertiary">Command Center</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-text-secondary">Email</label>
            <input
              type="email"
              value={email}
              onChange={e => { setEmail(e.target.value); setFieldErrors(p => ({ ...p, email: undefined })) }}
              className={`${inputBase} ${fieldErrors.email ? 'border-danger' : 'border-border'}`}
              placeholder="admin@noant.com"
            />
            {fieldErrors.email && <p className="mt-1 text-xs text-danger">{fieldErrors.email}</p>}
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-text-secondary">Password</label>
            <input
              type="password"
              value={password}
              onChange={e => { setPassword(e.target.value); setFieldErrors(p => ({ ...p, password: undefined })) }}
              className={`${inputBase} ${fieldErrors.password ? 'border-danger' : 'border-border'}`}
              placeholder="••••••••"
            />
            {fieldErrors.password && <p className="mt-1 text-xs text-danger">{fieldErrors.password}</p>}
          </div>

          {error && (
            <p className="text-sm text-danger">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="flex w-full items-center justify-center gap-2 rounded-lg bg-brand-sky py-2.5 text-sm font-semibold text-white transition-colors hover:bg-brand-sky-deep disabled:opacity-50"
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  )
}
