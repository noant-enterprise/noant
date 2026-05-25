import { useState } from 'react'
import { Eye, EyeOff, Copy, Check } from 'lucide-react'

interface TokenDisplayProps {
  token?: string
  label?: string
}

export function TokenDisplay({ token, label = 'Token' }: TokenDisplayProps) {
  const [visible, setVisible] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (!token) return
    await navigator.clipboard.writeText(token)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const masked = token ? '••••••••••••••••••••' : 'Not configured'

  return (
    <div className="space-y-1.5">
      <label className="text-[10px] font-semibold uppercase tracking-wider text-tertiary">{label}</label>
      <div className="flex items-center gap-2">
        <div className="flex-1 min-w-0 bg-inset rounded-lg px-3 py-2 font-mono text-xs text-primary truncate">
          {visible ? token : masked}
        </div>
        {token && (
          <>
            <button
              onClick={() => setVisible(!visible)}
              className="w-8 h-8 rounded-lg border border-default flex items-center justify-center text-secondary hover:text-primary hover:border-noant-sky transition-all active:scale-95"
              title={visible ? 'Hide token' : 'Show token'}
            >
              {visible ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
            </button>
            <button
              onClick={handleCopy}
              className="w-8 h-8 rounded-lg border border-default flex items-center justify-center text-secondary hover:text-primary hover:border-noant-sky transition-all active:scale-95"
              title="Copy token"
            >
              {copied ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
            </button>
          </>
        )}
      </div>
    </div>
  )
}
