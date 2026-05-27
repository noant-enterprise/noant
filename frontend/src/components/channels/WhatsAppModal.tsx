import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { api } from '@/lib/api'
import { CheckCircle2, XCircle, Loader2, Zap } from 'lucide-react'

interface WhatsAppModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
}

type TestState = 'idle' | 'testing' | 'success' | 'error'

export function WhatsAppModal({ open, onClose, onConnect, loading }: WhatsAppModalProps) {
  const [phoneNumberId, setPhoneNumberId] = useState('')
  const [businessAccountId, setBusinessAccountId] = useState('')
  const [accessToken, setAccessToken] = useState('')
  const [testState, setTestState] = useState<TestState>('idle')
  const [testMessage, setTestMessage] = useState('')

  const handleTest = async () => {
    if (!phoneNumberId.trim() || !accessToken.trim()) return
    setTestState('testing')
    setTestMessage('')
    try {
      const res = await api.post<{ status: string; message: string }>(
        '/integrations/test/whatsapp',
        {
          config: {
            phone_number_id: phoneNumberId.trim(),
            access_token: accessToken.trim(),
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
    if (!phoneNumberId.trim() || !accessToken.trim()) return
    onConnect({
      phone_number_id: phoneNumberId.trim(),
      business_account_id: businessAccountId.trim(),
      access_token: accessToken.trim(),
    })
  }

  const handleClose = () => {
    setPhoneNumberId('')
    setBusinessAccountId('')
    setAccessToken('')
    setTestState('idle')
    setTestMessage('')
    onClose()
  }

  const canTest = phoneNumberId.trim() && accessToken.trim()

  return (
    <Modal open={open} onClose={handleClose} title="Connect WhatsApp Business API">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary">Setup Instructions:</p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Go to <b>developers.facebook.com</b> and create an App.</li>
            <li>Add the <b>WhatsApp</b> product and configure a test number.</li>
            <li>Copy your <b>Phone Number ID</b> and generate a <b>Permanent System Token</b>.</li>
            <li>Enter the details below and click <b>Test Connection</b> to verify.</li>
          </ol>
          <p className="font-mono text-[10px] text-tertiary bg-surface/60 px-2 py-1 rounded border border-default mt-1">
            API: https://graph.facebook.com/v21.0/&#123;PHONE_NUMBER_ID&#125;
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Phone Number ID <span className="text-red-500">*</span>
            </label>
            <Input
              value={phoneNumberId}
              onChange={(e) => { setPhoneNumberId(e.target.value); setTestState('idle'); setTestMessage('') }}
              placeholder="e.g. 1098239082398"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              WhatsApp Business Account ID
            </label>
            <Input
              value={businessAccountId}
              onChange={(e) => setBusinessAccountId(e.target.value)}
              placeholder="e.g. 2938492834928"
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Permanent Access Token <span className="text-red-500">*</span>
            </label>
            <Input
              type="password"
              value={accessToken}
              onChange={(e) => { setAccessToken(e.target.value); setTestState('idle'); setTestMessage('') }}
              placeholder="EAAG..."
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
              disabled={!canTest || loading || testState === 'testing'}
              className="gap-1.5"
            >
              <Zap className="w-3.5 h-3.5" />
              Test Connection
            </Button>
            <Button type="submit" loading={loading} disabled={!canTest || loading}>
              Connect Channel
            </Button>
          </div>
        </form>
      </div>
    </Modal>
  )
}
