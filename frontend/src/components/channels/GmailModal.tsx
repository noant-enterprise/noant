import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { api } from '@/lib/api'
import { CheckCircle2, XCircle, Loader2, Zap, Mail } from 'lucide-react'

interface GmailModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
}

type TestState = 'idle' | 'testing' | 'success' | 'error'

export function GmailModal({ open, onClose, onConnect, loading }: GmailModalProps) {
  const [email, setEmail] = useState('')
  const [appPassword, setAppPassword] = useState('')
  const [testState, setTestState] = useState<TestState>('idle')
  const [testMessage, setTestMessage] = useState('')

  const handleTest = async () => {
    if (!email.trim() || !appPassword.trim()) return
    setTestState('testing')
    setTestMessage('')
    try {
      const res = await api.post<{ status: string; message: string }>(
        '/integrations/test/gmail',
        {
          config: {
            email: email.trim(),
            app_password: appPassword.trim(),
          },
        }
      )
      setTestState('success')
      setTestMessage(res.message || 'Connection successful!')
    } catch (err: any) {
      setTestState('error')
      const msg = (err as any)?.data?.message || err?.message || 'Connection test failed'
      setTestMessage(msg)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim() || !appPassword.trim()) return
    onConnect({
      email: email.trim(),
      app_password: appPassword.trim(),
    })
  }

  const handleClose = () => {
    setEmail('')
    setAppPassword('')
    setTestState('idle')
    setTestMessage('')
    onClose()
  }

  const canTest = email.trim() && appPassword.trim()

  return (
    <Modal open={open} onClose={handleClose} title="Connect Gmail">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary flex items-center gap-1.5">
            <Mail className="w-3.5 h-3.5" />
            Email Customer Support via IMAP/SMTP
          </p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Go to <b>myaccount.google.com</b> → Security</li>
            <li>Enable <b>2-Step Verification</b> (required)</li>
            <li>Go to <b>App passwords</b> → Generate new</li>
            <li>Select <b>Mail</b> → <b>Other device</b> → Name it "NOANT"</li>
            <li>Copy the 16-character password below</li>
          </ol>
          <p className="text-[10px] text-tertiary mt-1">
            Your Gmail password is NOT stored. Only the App Password is used for IMAP/SMTP access.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Gmail Address <span className="text-red-500">*</span>
            </label>
            <Input
              type="email"
              value={email}
              onChange={(e) => { setEmail(e.target.value); setTestState('idle'); setTestMessage('') }}
              placeholder="yourbusiness@gmail.com"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              App Password <span className="text-red-500">*</span>
            </label>
            <Input
              type="password"
              value={appPassword}
              onChange={(e) => { setAppPassword(e.target.value); setTestState('idle'); setTestMessage('') }}
              placeholder="xxxx xxxx xxxx xxxx"
              required
              disabled={loading}
            />
            <p className="text-[10px] text-tertiary mt-1">
              16-character App Password from Google Account settings
            </p>
          </div>

          {/* Test feedback */}
          {testState !== 'idle' && (
            <div className={`flex items-start gap-2 px-3 py-2.5 rounded-lg text-xs font-medium border ${
              testState === 'testing' ? 'bg-surface border-default text-secondary' :
              testState === 'success' ? 'bg-emerald-50 dark:bg-emerald-900/20 border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-400' :
              'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-700 dark:text-red-400'
            }`}>
              {testState === 'testing' && <Loader2 className="w-3.5 h-3.5 shrink-0 mt-0.5 animate-spin" />}
              {testState === 'success' && <CheckCircle2 className="w-3.5 h-3.5 shrink-0 mt-0.5" />}
              {testState === 'error' && <XCircle className="w-3.5 h-3.5 shrink-0 mt-0.5" />}
              <span>{testState === 'testing' ? 'Testing IMAP/SMTP connection...' : testMessage}</span>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="ghost" onClick={handleClose} disabled={loading}>
              Cancel
            </Button>
            <Button
              type="button"
              variant="ghost"
              onClick={handleTest}
              disabled={!canTest || loading || testState === 'testing'}
              className="gap-1.5"
            >
              <Zap className="w-3.5 h-3.5" />
              Test Connection
            </Button>
            <Button type="submit" loading={loading} disabled={!canTest || loading}>
              Connect Gmail
            </Button>
          </div>
        </form>
      </div>
    </Modal>
  )
}
