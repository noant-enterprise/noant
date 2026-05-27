import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { api } from '@/lib/api'
import { CheckCircle2, XCircle, Loader2, Zap } from 'lucide-react'

interface InstagramModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
}

type TestState = 'idle' | 'testing' | 'success' | 'error'

export function InstagramModal({ open, onClose, onConnect, loading }: InstagramModalProps) {
  const [instagramId, setInstagramId] = useState('')
  const [pageAccessToken, setPageAccessToken] = useState('')
  const [testState, setTestState] = useState<TestState>('idle')
  const [testMessage, setTestMessage] = useState('')

  const handleTest = async () => {
    if (!instagramId.trim() || !pageAccessToken.trim()) return
    setTestState('testing')
    setTestMessage('')
    try {
      const res = await api.post<{ status: string; message: string }>(
        '/integrations/test/instagram',
        {
          config: {
            instagram_id: instagramId.trim(),
            page_access_token: pageAccessToken.trim(),
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
    if (!instagramId.trim() || !pageAccessToken.trim()) return
    onConnect({
      instagram_id: instagramId.trim(),
      page_access_token: pageAccessToken.trim(),
    })
  }

  const handleClose = () => {
    setInstagramId('')
    setPageAccessToken('')
    setTestState('idle')
    setTestMessage('')
    onClose()
  }

  const canTest = instagramId.trim() && pageAccessToken.trim()

  return (
    <Modal open={open} onClose={handleClose} title="Connect Instagram Direct Messages">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary">Setup Instructions:</p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Ensure your Instagram account is a <b>Professional/Business</b> account.</li>
            <li>Link your Instagram account to a <b>Facebook Business Page</b>.</li>
            <li>Enable <b>Allow Access to Messages</b> in Instagram app settings.</li>
            <li>Enter your Instagram Account ID and Meta Page Access Token below.</li>
          </ol>
          <p className="font-mono text-[10px] text-tertiary bg-surface/60 px-2 py-1 rounded border border-default mt-1">
            API: https://graph.facebook.com/v21.0/&#123;INSTAGRAM_ID&#125;?fields=id,username
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Instagram Account ID <span className="text-red-500">*</span>
            </label>
            <Input
              value={instagramId}
              onChange={(e) => { setInstagramId(e.target.value); setTestState('idle'); setTestMessage('') }}
              placeholder="e.g. 17841400000000000"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Page Access Token <span className="text-red-500">*</span>
            </label>
            <Input
              type="password"
              value={pageAccessToken}
              onChange={(e) => { setPageAccessToken(e.target.value); setTestState('idle'); setTestMessage('') }}
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
