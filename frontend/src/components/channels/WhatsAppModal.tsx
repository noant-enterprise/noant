import { useState, useEffect } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { api } from '@/lib/api'
import { CheckCircle2, XCircle, Loader2, Zap, Smartphone, RefreshCw, ShieldCheck, AlertTriangle } from 'lucide-react'

interface WhatsAppModalProps {
  open: boolean
  onClose: () => void
  loading?: boolean
  onConnect?: () => void
}

type Step = 'form' | 'qr' | 'verifying' | 'success' | 'error'

export function WhatsAppModal({ open, onClose, loading, onConnect }: WhatsAppModalProps) {
  const [phone, setPhone] = useState('')
  const [connectionName, setConnectionName] = useState('')
  const [step, setStep] = useState<Step>('form')
  const [errorMsg, setErrorMsg] = useState('')
  const [qrCode, setQrCode] = useState('')
  const [sessionId, setSessionId] = useState('')
  const [testState, setTestState] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')
  const [testMessage, setTestMessage] = useState('')
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [qrExpired, setQrExpired] = useState(false)
  const [sessionStatus, setSessionStatus] = useState<string>('')

  // Auto-poll while QR is displayed using long polling
  useEffect(() => {
    if (step !== 'qr' || !sessionId) return

    let active = true

    const poll = async () => {
      while (active) {
        try {
          const res = await api.get<{ status: string; connected: boolean; qr_code?: string }>(
            `/channels/whatsapp/status/${sessionId}?t=${Date.now()}`
          )

          if (!active) return
          setSessionStatus(res.status)

          if (res.connected) {
            setStep('success')
            onConnect?.()
            setTimeout(() => handleClose(), 2500)
            break
          } else if (res.status === 'expired') {
            setQrExpired(true)
            setQrCode('')
            setErrorMsg('QR code expired. Click "Refresh QR" to generate a new one.')
            break
          } else if (res.status === 'failed') {
            setQrCode('')
            setErrorMsg('Session initialization failed. Click "Refresh QR" or check your server status.')
            break
          } else if (res.status === 'disconnected') {
            setQrCode('')
            setErrorMsg('Session disconnected. Click "Refresh QR" to reconnect.')
            break
          } else if (res.qr_code && res.qr_code !== qrCode) {
            setQrCode(res.qr_code)
            setQrExpired(false)
          }
        } catch {
          // If request fails (e.g. network disconnect, server restart), wait 3s before retrying to prevent hot loop
          if (!active) return
          await new Promise((resolve) => setTimeout(resolve, 3000))
        }
      }
    }

    poll()

    return () => {
      active = false
    }
  }, [step, sessionId, qrCode])

  const handleConnect = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!phone.trim()) return
    setStep('form')
    setErrorMsg('')
    setQrCode('')
    try {
      const res = await api.post<{
        session_id: string
        qr_code?: string
        qrCode?: string
        status: string
      }>('/channels/whatsapp/connect', { phone: phone.trim() })

      setSessionId(res.session_id)
      setSessionStatus(res.status)
      const qr = res.qr_code || res.qrCode || ''

      if (res.status === 'connected') {
        setStep('success')
        onConnect?.()
        setTimeout(() => handleClose(), 2500)
        return
      }

      if (qr) {
        setQrCode(qr)
        setStep('qr')
      } else {
        setStep('qr')
      }
    } catch (err: any) {
      setStep('error')
      setErrorMsg((err as any)?.data?.message || err?.message || 'Failed to connect to WhatsApp')
    }
  }

  // Called when user clicks "I've scanned — Verify" 
  const handleVerify = async () => {
    if (!sessionId) return
    setStep('verifying')
    try {
      // First try a normal status check
      const res = await api.get<{ status: string; connected: boolean }>(
        `/channels/whatsapp/status/${sessionId}`
      )
      if (res.connected) {
        setStep('success')
        onConnect?.()
        setTimeout(() => handleClose(), 2500)
        return
      }

      // Phone says it's logged in but API still shows qr_ready — force confirm
      const forced = await api.get<{ status: string; connected: boolean }>(
        `/channels/whatsapp/status/${sessionId}?force=true`
      )
      if (forced.connected) {
        setStep('success')
        onConnect?.()
        setTimeout(() => handleClose(), 2500)
      } else {
        setStep('qr')
        setErrorMsg('Could not verify. Make sure you scanned the QR code with your phone.')
      }
    } catch (err: any) {
      setStep('qr')
      setErrorMsg((err as any)?.data?.message || err?.message || 'Verification failed')
    }
  }

  const handleRefreshQR = async () => {
    if (!sessionId || isRefreshing) return
    setIsRefreshing(true)
    setErrorMsg('')
    try {
      const res = await api.post<{ qr_code?: string; session_id?: string }>(`/channels/whatsapp/refresh/${sessionId}`)
      if (res.session_id) {
        setSessionId(res.session_id)
      }
      if (res.qr_code) {
        setQrCode(res.qr_code)
      }
      setQrExpired(false)
      setSessionStatus('initializing')
    } catch (err: any) {
      setErrorMsg((err as any)?.data?.message || err?.message || 'Failed to refresh QR code')
    } finally {
      setIsRefreshing(false)
    }
  }

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
      setTestMessage((err as any)?.data?.message || err?.message || 'Connection test failed')
    }
  }

  const handleClose = () => {
    setPhone('')
    setConnectionName('')
    setStep('form')
    setErrorMsg('')
    setQrCode('')
    setSessionId('')
    setTestState('idle')
    setTestMessage('')
    setQrExpired(false)
    setIsRefreshing(false)
    setSessionStatus('')
    onClose()
  }

  return (
    <Modal open={open} onClose={handleClose} title="Connect WhatsApp">
      <div className="space-y-4">

        {/* ── FORM STEP ── */}
        {(step === 'form' || step === 'error') && (
          <>
            <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
              <p className="font-semibold text-primary flex items-center gap-1.5">
                <Smartphone className="w-3.5 h-3.5" />
                OpenWA self-hosted WhatsApp
              </p>
              <ol className="list-decimal list-inside space-y-1">
                <li>Enter your business WhatsApp number</li>
                <li>Give this connection a name (optional)</li>
                <li>Click <b>Connect</b> to generate a QR code</li>
                <li>Scan with your phone to link the device</li>
              </ol>
            </div>

            <form onSubmit={handleConnect} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-secondary mb-1">
                  WhatsApp Number <span className="text-red-500">*</span>
                </label>
                <Input
                  value={phone}
                  onChange={(e) => { setPhone(e.target.value); setTestState('idle'); setTestMessage('') }}
                  placeholder="+44 7700 900123"
                  required
                  disabled={loading}
                />
                <p className="text-[10px] text-tertiary mt-1">Include country code (e.g. +44, +234)</p>
              </div>

              <div>
                <label className="block text-xs font-semibold text-secondary mb-1">
                  Connection Name <span className="text-tertiary font-normal">(optional)</span>
                </label>
                <Input
                  value={connectionName}
                  onChange={(e) => setConnectionName(e.target.value)}
                  placeholder="e.g. Business Line, Support Number"
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
                  <span>{testState === 'testing' ? 'Pinging server...' : testMessage}</span>
                </div>
              )}

              {step === 'error' && errorMsg && (
                <div className="flex items-start gap-2 px-3 py-2.5 rounded-lg text-xs font-medium border bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-700 dark:text-red-400">
                  <XCircle className="w-3.5 h-3.5 shrink-0 mt-0.5" />
                  <span>{errorMsg}</span>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-2">
                <Button type="button" variant="ghost" onClick={handleClose} disabled={loading}>Cancel</Button>
                <Button
                  type="button"
                  variant="ghost"
                  onClick={handleTest}
                  disabled={!phone.trim() || loading || testState === 'testing'}
                  className="gap-1.5"
                >
                  <Zap className="w-3.5 h-3.5" />
                  Test Server
                </Button>
                <Button type="submit" loading={loading} disabled={!phone.trim() || loading}>
                  Connect WhatsApp
                </Button>
              </div>
            </form>
          </>
        )}

        {/* ── QR STEP ── */}
        {step === 'qr' && (
          <div className="space-y-4">
            {/* QR Code */}
            <div className="flex flex-col items-center gap-4 py-4 bg-inset/50 rounded-2xl border border-dashed border-default p-4">
              {sessionStatus === 'expired' || qrExpired ? (
                <div className="w-52 h-52 bg-red-500/5 rounded-2xl flex flex-col items-center justify-center gap-2 border border-red-500/20 text-red-600 dark:text-red-400 p-4 text-center animate-fade-in">
                  <AlertTriangle className="w-8 h-8 text-red-500 animate-bounce" />
                  <span className="text-xs font-semibold">QR Code Expired</span>
                </div>
              ) : sessionStatus === 'initializing' ? (
                <div className="w-52 h-52 bg-inset rounded-2xl flex flex-col items-center justify-center gap-2 border border-default p-4 text-center">
                  <Loader2 className="w-8 h-8 animate-spin text-emerald-500" />
                  <span className="text-xs font-semibold text-secondary">Starting session...</span>
                  <span className="text-[10px] text-tertiary animate-pulse">Checking with self-hosted server</span>
                </div>
              ) : sessionStatus === 'failed' ? (
                <div className="w-52 h-52 bg-red-500/5 rounded-2xl flex flex-col items-center justify-center gap-2 border border-red-500/20 text-red-600 dark:text-red-400 p-4 text-center animate-fade-in">
                  <XCircle className="w-8 h-8 text-red-500 animate-pulse" />
                  <span className="text-xs font-semibold">Initialization Failed</span>
                  <span className="text-[10px] text-tertiary">Restart your self-hosted server</span>
                </div>
              ) : sessionStatus === 'disconnected' ? (
                <div className="w-52 h-52 bg-amber-500/5 rounded-2xl flex flex-col items-center justify-center gap-2 border border-amber-500/20 text-amber-600 dark:text-amber-400 p-4 text-center animate-fade-in">
                  <AlertTriangle className="w-8 h-8 text-amber-500 animate-pulse" />
                  <span className="text-xs font-semibold">Session Disconnected</span>
                  <span className="text-[10px] text-tertiary">Click "Refresh QR" to reconnect</span>
                </div>
              ) : qrCode ? (
                <div className="relative group">
                  <div className="absolute -inset-1.5 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-2xl blur-md opacity-25 group-hover:opacity-40 transition duration-1000" />
                  <div className="relative w-52 h-52 bg-white rounded-2xl p-3.5 flex items-center justify-center shadow-2xl border border-default">
                    <img src={qrCode} alt="WhatsApp QR Code" className="w-full h-full object-contain" />
                  </div>
                </div>
              ) : (
                <div className="w-52 h-52 bg-inset rounded-2xl flex items-center justify-center border border-default">
                  <Loader2 className="w-8 h-8 animate-spin text-tertiary" />
                </div>
              )}

              {/* Live scanning indicator */}
              {sessionStatus === 'initializing' ? (
                <div className="flex items-center gap-2 bg-blue-500/10 px-3.5 py-1.5 rounded-full border border-blue-500/20 text-xs font-semibold text-blue-600 dark:text-blue-400 animate-pulse">
                  <span className="relative flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500" />
                  </span>
                  Initializing...
                </div>
              ) : sessionStatus === 'failed' ? (
                <div className="flex items-center gap-2 bg-red-500/10 px-3.5 py-1.5 rounded-full border border-red-500/20 text-xs font-semibold text-red-600 dark:text-red-400">
                  <span className="relative flex h-2 w-2">
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500" />
                  </span>
                  Failed to start
                </div>
              ) : sessionStatus === 'disconnected' ? (
                <div className="flex items-center gap-2 bg-amber-500/10 px-3.5 py-1.5 rounded-full border border-amber-500/20 text-xs font-semibold text-amber-600 dark:text-amber-400">
                  <span className="relative flex h-2 w-2">
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-amber-500" />
                  </span>
                  Disconnected
                </div>
              ) : sessionStatus === 'expired' || qrExpired ? (
                <div className="flex items-center gap-2 bg-red-500/10 px-3.5 py-1.5 rounded-full border border-red-500/20 text-xs font-semibold text-red-600 dark:text-red-400">
                  <span className="relative flex h-2 w-2">
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500" />
                  </span>
                  Expired
                </div>
              ) : (
                <div className="flex items-center gap-2 bg-emerald-500/10 dark:bg-emerald-500/20 px-3.5 py-1.5 rounded-full border border-emerald-500/20 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                  <span className="relative flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
                  </span>
                  Waiting for scan...
                </div>
              )}

              <div className="space-y-1.5 text-center max-w-xs">
                <p className="text-xs font-medium text-secondary">
                  Open WhatsApp → <b>Linked Devices</b> → <b>Link a Device</b> → Scan this QR
                </p>
                <p className="text-[10px] font-mono text-tertiary">Session: {sessionId?.slice(0, 16)}...</p>
              </div>
            </div>

            {/* Error hint */}
            {errorMsg && (
              <div className="flex items-start gap-2 px-3 py-2.5 rounded-lg text-xs font-medium border bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800 text-amber-700 dark:text-amber-400">
                <XCircle className="w-3.5 h-3.5 shrink-0 mt-0.5" />
                <span>{errorMsg}</span>
              </div>
            )}

            <div className="flex flex-col gap-2">
              {/* Primary action — verify after scanning */}
              <Button
                type="button"
                onClick={handleVerify}
                className="w-full gap-2 justify-center"
              >
                <ShieldCheck className="w-4 h-4" />
                I've scanned — Verify Connection
              </Button>

              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  onClick={handleRefreshQR}
                  disabled={isRefreshing}
                  className="flex-1 gap-1.5 justify-center"
                >
                  <RefreshCw className={`w-3.5 h-3.5 ${isRefreshing ? 'animate-spin' : ''}`} />
                  {isRefreshing ? 'Refreshing...' : 'Refresh QR'}
                </Button>
                <Button type="button" variant="ghost" onClick={handleClose} className="flex-1 justify-center">
                  Cancel
                </Button>
              </div>
            </div>
          </div>
        )}

        {/* ── VERIFYING STEP ── */}
        {step === 'verifying' && (
          <div className="flex flex-col items-center gap-5 py-10">
            <div className="relative">
              <div className="w-16 h-16 rounded-full bg-emerald-500/10 flex items-center justify-center">
                <Loader2 className="w-8 h-8 animate-spin text-emerald-500" />
              </div>
            </div>
            <div className="text-center space-y-1">
              <p className="text-sm font-semibold text-primary">Verifying your connection...</p>
              <p className="text-xs text-secondary">Checking with WhatsApp — this takes a moment</p>
            </div>
          </div>
        )}

        {/* ── SUCCESS STEP ── */}
        {step === 'success' && (
          <div className="flex flex-col items-center gap-5 py-10">
            <div className="relative">
              <div className="absolute inset-0 bg-emerald-500/20 rounded-full animate-ping" />
              <div className="relative w-16 h-16 rounded-full bg-emerald-500/10 flex items-center justify-center">
                <CheckCircle2 className="w-8 h-8 text-emerald-500" />
              </div>
            </div>
            <div className="text-center space-y-1">
              <p className="text-sm font-semibold text-primary">WhatsApp Connected! 🎉</p>
              <p className="text-xs text-secondary">Your AI is now live on WhatsApp and ready to chat</p>
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}