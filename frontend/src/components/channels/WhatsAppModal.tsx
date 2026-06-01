import { useState, useEffect } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { api } from '@/lib/api'
import { CheckCircle2, XCircle, Loader2, Zap, Smartphone, RefreshCw } from 'lucide-react'

interface WhatsAppModalProps {
  open: boolean
  onClose: () => void
  loading?: boolean
}

type TestState = 'idle' | 'testing' | 'success' | 'error'

export function WhatsAppModal({ open, onClose, loading }: WhatsAppModalProps) {
  const [phone, setPhone] = useState('')
  const [testState, setTestState] = useState<TestState>('idle')
  const [testMessage, setTestMessage] = useState('')
  const [qrCode, setQrCode] = useState('')
  const [sessionId, setSessionId] = useState('')
  const [qrReady, setQrReady] = useState(false)

  const handleTest = async () => {
    if (!phone.trim()) return
    setTestState('testing')
    setTestMessage('')
    try {
      const res = await api.post<{ success: boolean; message: string }>(
        '/channels/whatsapp/ping',
        { phone: phone.trim() }
      )
      setTestState('success')
      setTestMessage(res.message || 'Connection successful!')
    } catch (err: any) {
      setTestState('error')
      const msg = (err as any)?.data?.message || err?.message || 'Connection test failed'
      setTestMessage(msg)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!phone.trim()) return
    setTestState('idle')
    setTestMessage('')
    setQrCode('')
    setQrReady(false)
    try {
      const res = await api.post<{
        session_id: string
        qr_code?: string
        qrCode?: string
        status: string
      }>('/channels/whatsapp/connect', { phone: phone.trim() })
      setSessionId(res.session_id)
      const qr = res.qr_code || res.qrCode || ''
      if (qr) {
        setQrCode(qr)
        setQrReady(true)
      }
    } catch (err: any) {
      setTestState('error')
      setTestMessage((err as any)?.data?.message || err?.message || 'Failed to connect')
    }
  }

  const handleClose = () => {
    setPhone('')
    setTestState('idle')
    setTestMessage('')
    setQrCode('')
    setSessionId('')
    setQrReady(false)
    onClose()
  }

  return (
    <Modal open={open} onClose={handleClose} title="Connect WhatsApp">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary flex items-center gap-1.5">
            <Smartphone className="w-3.5 h-3.5" />
            OpenWA self-hosted WhatsApp API
          </p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Enter your business WhatsApp number below</li>
            <li>Click <b>Test Connection</b> to verify your session</li>
            <li>Click <b>Connect WhatsApp</b> to generate a QR code</li>
            <li>Scan the QR with your phone to link the device</li>
          </ol>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              WhatsApp Number <span className="text-red-500">*</span>
            </label>
            <Input
              value={phone}
              onChange={(e) => { setPhone(e.target.value); setTestState('idle'); setTestMessage(''); setQrReady(false) }}
              placeholder="080 1234 5678"
              required
              disabled={loading || qrReady}
            />
            <p className="text-[10px] text-tertiary mt-1">
              Include country code. This is your business WhatsApp number.
            </p>
          </div>

          {/* QR Code after connect */}
          {qrReady && qrCode && (
            <div className="flex flex-col items-center gap-3 py-3">
              <div className="w-44 h-44 bg-white rounded-xl p-2 flex items-center justify-center shadow-lg">
                <img src={qrCode} alt="QR Code" className="w-full h-full" />
              </div>
              <p className="text-xs text-secondary text-center">
                Open WhatsApp → Linked Devices → Link a Device → Scan this QR
              </p>
              <p className="text-xs text-tertiary">Session: {sessionId?.slice(0, 8)}...</p>
            </div>
          )}

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
              <span>{testState === 'testing' ? 'Testing WhatsApp connection...' : testMessage}</span>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="ghost" onClick={handleClose} disabled={loading}>
              Cancel
            </Button>
            {qrReady ? (
              <Button
                type="button"
                variant="ghost"
                onClick={handleClose}
                className="gap-1.5"
              >
                <CheckCircle2 className="w-3.5 h-3.5" />
                Done
              </Button>
            ) : (
              <>
                <Button
                  type="button"
                  variant="ghost"
                  onClick={handleTest}
                  disabled={!phone.trim() || loading || testState === 'testing'}
                  className="gap-1.5"
                >
                  <Zap className="w-3.5 h-3.5" />
                  Test Connection
                </Button>
                <Button type="submit" loading={loading} disabled={!phone.trim() || loading}>
                  Connect WhatsApp
                </Button>
              </>
            )}
          </div>
        </form>
      </div>
    </Modal>
  )
}