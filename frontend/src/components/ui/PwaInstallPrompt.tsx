import { useState, useEffect } from 'react'
import { X, Download } from 'lucide-react'

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

export function PwaInstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [show, setShow] = useState(false)
  const [dismissed, setDismissed] = useState(false)

  useEffect(() => {
    const handler = (e: Event) => {
      e.preventDefault()
      setDeferredPrompt(e as BeforeInstallPromptEvent)
      if (!dismissed) setShow(true)
    }
    window.addEventListener('beforeinstallprompt', handler)
    window.addEventListener('appinstalled', () => setShow(false))
    return () => window.removeEventListener('beforeinstallprompt', handler)
  }, [dismissed])

  if (!show || !deferredPrompt) return null

  const handleInstall = async () => {
    deferredPrompt.prompt()
    const { outcome } = await deferredPrompt.userChoice
    if (outcome === 'accepted') setShow(false)
    setDeferredPrompt(null)
  }

  return (
    <div className="fixed bottom-4 right-4 z-50 animate-in slide-in-from-bottom-2 fade-in duration-300">
      <div className="bg-surface border border-default rounded-xl shadow-xl p-4 max-w-xs backdrop-blur-sm">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-noant-sky/10 flex items-center justify-center shrink-0">
              <Download className="w-5 h-5 text-noant-sky" />
            </div>
            <div>
              <p className="text-sm font-semibold text-primary">Install Noant</p>
              <p className="text-xs text-secondary mt-0.5">Get the app for a better experience</p>
            </div>
          </div>
          <button
            onClick={() => { setShow(false); setDismissed(true) }}
            className="p-1 rounded-lg hover:bg-inset text-secondary hover:text-primary transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="flex items-center gap-2 mt-3">
          <button
            onClick={handleInstall}
            className="flex-1 px-4 py-2 rounded-lg bg-noant-sky text-white text-sm font-semibold hover:bg-noant-sky-deep active:scale-[0.98] transition-all"
          >
            Install
          </button>
          <button
            onClick={() => setShow(false)}
            className="px-4 py-2 rounded-lg text-sm font-medium text-secondary hover:text-primary hover:bg-inset active:scale-[0.98] transition-all"
          >
            Not now
          </button>
        </div>
      </div>
    </div>
  )
}
