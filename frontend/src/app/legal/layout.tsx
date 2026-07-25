import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'

interface LegalLayoutProps {
  title: string
  effectiveDate: string
  children: React.ReactNode
}

export default function LegalLayout({ title, effectiveDate, children }: LegalLayoutProps) {
  useEffect(() => {
    window.scrollTo(0, 0)
  }, [])

  return (
    <div className="min-h-screen" style={{ background: 'var(--bg-base)' }}>
      <div className="max-w-3xl mx-auto px-6 py-12">
        <Link
          to="/"
          className="inline-flex items-center gap-2 text-sm text-secondary hover:text-primary transition-colors mb-8"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to home
        </Link>

        <h1 className="text-3xl font-bold mb-2" style={{ color: 'var(--text-primary)' }}>
          {title}
        </h1>
        <p className="text-sm mb-8" style={{ color: 'var(--text-tertiary)' }}>
          Effective Date: {effectiveDate}
        </p>

        <div
          className="prose prose-sm max-w-none"
          style={{ color: 'var(--text-secondary)' }}
        >
          {children}
        </div>

        <div className="mt-12 pt-8 border-t" style={{ borderColor: 'var(--border-default)' }}>
          <p className="text-xs" style={{ color: 'var(--text-tertiary)' }}>
            Questions? Contact us at{' '}
            <a href="mailto:legal@noant.com" className="underline hover:text-primary transition-colors">
              legal@noant.com
            </a>
          </p>
        </div>
      </div>
    </div>
  )
}
