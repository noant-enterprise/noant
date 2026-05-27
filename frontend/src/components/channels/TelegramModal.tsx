import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { api } from '@/lib/api'
import { CheckCircle2, XCircle, Loader2, Zap } from 'lucide-react'

interface TelegramModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
}

type TestState = 'idle' | 'testing' | 'success' | 'error'

export function TelegramModal({ open, onClose, onConnect, loading }: TelegramModalProps) {
  const [botToken, setBotToken] = useState('')
  const [testState, setTestState] = useState<TestState>('idle')
  const [testMessage, setTestMessage] = useState('')

  const handleTest = async () => {
    if (!botToken.trim()) return
    setTestState('testing')
    setTestMessage('')
    try {
      const res = await api.post<{ status: string; message: string }>(
        '/integrations/test/telegram',
        { config: { bot_token: botToken.trim() } }
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
    if (!botToken.trim()) return
    onConnect({ bot_token: botToken.trim() })
  }

  const handleClose = () => {
    setBotToken('')
    setTestState('idle')
    setTestMessage('')
    onClose()
  }

  return (
    <Modal open={open} onClose={handleClose} title="Connect Telegram Bot">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary">Setup Instructions:</p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Search for <b>@BotFather</b> on Telegram.</li>
            <li>Send <code>/newbot</code> and follow the instructions.</li>
            <li>Copy the HTTP API Token provided.</li>
            <li>Paste it below, then click <b>Test Connection</b> to verify.</li>
          </ol>
          <p className="font-mono text-[10px] text-tertiary bg-surface/60 px-2 py-1 rounded border border-default mt-1">
            API: https://api.telegram.org/bot&#123;TOKEN&#125;/getMe
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Bot Token <span className="text-red-500">*</span>
            </label>
            <Input
              value={botToken}
              onChange={(e) => { setBotToken(e.target.value); setTestState('idle'); setTestMessage('') }}
              placeholder="e.g. 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
              required
              disabled={loading}
            />
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
              <span>{testState === 'testing' ? 'Testing connection...' : testMessage}</span>
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
              disabled={!botToken.trim() || loading || testState === 'testing'}
              className="gap-1.5"
            >
              <Zap className="w-3.5 h-3.5" />
              Test Connection
            </Button>
            <Button type="submit" loading={loading} disabled={!botToken.trim() || loading}>
              Connect Channel
            </Button>
          </div>
        </form>
      </div>
    </Modal>
  )
}
