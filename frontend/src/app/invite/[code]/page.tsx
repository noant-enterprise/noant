import { useState, useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowRight, MessageSquare, Bot, Zap, Check } from 'lucide-react'

export default function InvitePage() {
  const { code } = useParams<{ code: string }>()
  const [referrerName, setReferrerName] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (!code) return
    fetch(`/api/v1/referral/${code}`)
      .then(async (res) => {
        if (!res.ok) throw new Error('Invalid referral')
        const data = await res.json()
        setReferrerName(data.referrer_name || null)
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [code])

  return (
    <div className="min-h-screen bg-background text-primary font-sans antialiased flex items-center justify-center px-6 py-12">
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-7xl h-[600px] overflow-hidden pointer-events-none z-0">
        <div className="absolute top-[-20%] left-[20%] w-[500px] h-[500px] rounded-full bg-sky-500/10 blur-[120px]" />
        <div className="absolute top-[-10%] right-[20%] w-[450px] h-[450px] rounded-full bg-indigo-500/10 blur-[130px]" />
      </div>

      <div className="relative z-10 w-full max-w-lg">
        {/* Brand */}
        <div className="flex items-center justify-center gap-2.5 mb-10">
          <svg className="w-7 h-7 text-sky-400" viewBox="0 0 200 200" fill="none">
            <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" />
            <circle cx="100" cy="100" r="70" fill="currentColor" />
            <circle cx="80" cy="100" r="6" fill="#000000" />
            <circle cx="100" cy="100" r="8" fill="#000000" />
            <circle cx="120" cy="100" r="10" fill="#000000" />
          </svg>
          <span className="text-lg font-bold tracking-widest lowercase text-sky-400">noant</span>
        </div>

        {/* Card */}
        <div className="bg-surface border border-default rounded-2xl p-8 sm:p-10 text-center space-y-6">
          {loading ? (
            <div className="space-y-4 py-4">
              <div className="h-6 w-48 mx-auto bg-zinc-800 rounded-lg animate-pulse" />
              <div className="h-4 w-64 mx-auto bg-zinc-800 rounded-lg animate-pulse" />
            </div>
          ) : error ? (
            <div className="space-y-4 py-4">
              <h1 className="text-xl font-bold text-primary">Invalid or Expired Link</h1>
              <p className="text-sm text-secondary">This referral link is no longer valid.</p>
              <Link
                to="/"
                className="inline-flex items-center gap-1.5 px-5 py-2.5 rounded-xl text-sm font-semibold bg-zinc-900 hover:bg-zinc-800 text-primary border border-zinc-800 hover:border-zinc-700 transition-colors"
              >
                Go to noant.com
              </Link>
            </div>
          ) : (
            <>
              <h1 className="text-xl sm:text-2xl font-extrabold text-primary leading-tight">
                {referrerName ? (
                  <>{referrerName} invited you to try <span className="text-sky-400">NOANT</span></>
                ) : (
                  <>You've been invited to try <span className="text-sky-400">NOANT</span></>
                )}
              </h1>

              <p className="text-sm text-secondary leading-relaxed">
                AI-powered customer support for your business
              </p>

              <Link
                to={`/signup?ref=${code}`}
                className="inline-flex items-center gap-2 px-6 py-3 rounded-xl text-sm font-semibold bg-sky-500 hover:bg-sky-600 text-primary shadow-lg shadow-sky-500/20 hover:shadow-sky-500/30 active:scale-[0.98] transition-all duration-200 group"
              >
                Create Free Account
                <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
              </Link>

              {/* Feature bullets */}
              <div className="pt-6 border-t border-subtle space-y-3">
                {[
                  { icon: MessageSquare, text: 'WhatsApp support out of the box' },
                  { icon: Bot, text: 'AI auto-reply resolves queries instantly' },
                  { icon: Zap, text: 'Setup in 15 minutes — no code required' },
                ].map((f, i) => (
                  <div key={i} className="flex items-center gap-3 text-sm text-secondary">
                    <span className="w-7 h-7 rounded-lg bg-sky-500/10 text-sky-400 flex items-center justify-center shrink-0">
                      <f.icon className="w-3.5 h-3.5" />
                    </span>
                    {f.text}
                  </div>
                ))}
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        <p className="text-center text-xs text-tertiary mt-6 flex items-center justify-center gap-1.5">
          <Check className="w-3 h-3 text-sky-400" />
          Free forever. No credit card required.
        </p>
      </div>
    </div>
  )
}
