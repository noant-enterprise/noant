import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuth } from '@/hooks/useAuth'
import { useToast } from '@/components/ui/Toast'

interface WebWidgetModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
  isConnected?: boolean
  existingConfig?: any
}

export function WebWidgetModal({ open, onClose, onConnect, loading, isConnected, existingConfig }: WebWidgetModalProps) {
  const { user } = useAuth()
  const { toast } = useToast()
  const [botName, setBotName] = useState(existingConfig?.botName || 'Noant AI')
  const [greeting, setGreeting] = useState(existingConfig?.greeting || 'Hi! 👋 How can I help you?')
  const [brandColor, setBrandColor] = useState(existingConfig?.brandColor || '#0ea5e9')
  const [position, setPosition] = useState(existingConfig?.position || 'right')
  const [copied, setCopied] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await onConnect({
      botName: botName.trim(),
      greeting: greeting.trim(),
      brandColor: brandColor.trim(),
      position: position,
    })
  }

  // Auto-generate embed code snippet
  const embedCode = `<script\n  src="${window.location.protocol}//${window.location.host}/widget.js"\n  data-id="${user?.id || ''}"\n  data-bot-name="${botName}"\n  data-greeting="${greeting}"\n  data-brand-color="${brandColor}"\n  data-position="${position}"\n></script>`

  const handleCopy = () => {
    navigator.clipboard.writeText(embedCode)
    setCopied(true)
    toast('Embed code copied to clipboard!', 'success')
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Modal open={open} onClose={onClose} title="Configure Web Chat Widget" size="lg">
      <div className="space-y-5">
        <form onSubmit={handleSubmit} className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-4">
            <h4 className="text-xs font-bold uppercase tracking-wider text-tertiary">Customization</h4>

            <div>
              <label className="block text-xs font-semibold text-secondary mb-1">
                Widget Bot Name <span className="text-red-500">*</span>
              </label>
              <Input
                value={botName}
                onChange={(e) => setBotName(e.target.value)}
                placeholder="Noant AI"
                required
                disabled={loading}
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-secondary mb-1">
                Greeting Message <span className="text-red-500">*</span>
              </label>
              <Input
                value={greeting}
                onChange={(e) => setGreeting(e.target.value)}
                placeholder="Hi! 👋 How can I help you?"
                required
                disabled={loading}
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-semibold text-secondary mb-1">
                  Brand Color
                </label>
                <div className="flex gap-2 items-center">
                  <input
                    type="color"
                    value={brandColor}
                    onChange={(e) => setBrandColor(e.target.value)}
                    className="w-8 h-8 rounded border border-default cursor-pointer p-0 bg-transparent"
                    disabled={loading}
                  />
                  <input
                    type="text"
                    value={brandColor}
                    onChange={(e) => setBrandColor(e.target.value)}
                    className="w-full text-xs font-mono border border-default rounded px-2 py-1 outline-none text-primary"
                    disabled={loading}
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-secondary mb-1">
                  Widget Position
                </label>
                <select
                  value={position}
                  onChange={(e) => setPosition(e.target.value)}
                  className="w-full text-sm px-3 py-2 rounded-lg border border-default outline-none bg-inset focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all text-primary"
                  disabled={loading}
                >
                  <option value="right">Bottom Right</option>
                  <option value="left">Bottom Left</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end gap-3 pt-2">
              <Button type="button" variant="ghost" onClick={onClose} disabled={loading}>
                Cancel
              </Button>
              <Button type="submit" loading={loading} disabled={loading}>
                {isConnected ? 'Update Widget' : 'Connect Channel'}
              </Button>
            </div>
          </div>

          <div className="space-y-4 flex flex-col justify-between">
            <div className="space-y-4">
              <h4 className="text-xs font-bold uppercase tracking-wider text-tertiary">Embed HTML Code</h4>
              <p className="text-xs text-secondary leading-relaxed">
                Copy this code and paste it right before the closing <code>&lt;/body&gt;</code> tag on every page of your website.
              </p>

              <div className="relative group">
                <pre className="bg-inset border border-default rounded-xl p-3.5 font-mono text-[10px] text-secondary overflow-x-auto whitespace-pre select-all max-h-[160px]">
                  {embedCode}
                </pre>
                <button
                  type="button"
                  onClick={handleCopy}
                  className="absolute top-2 right-2 px-2 py-1 rounded bg-surface hover:bg-surface-hover border border-default text-[10px] font-semibold text-secondary transition-all"
                >
                  {copied ? 'Copied' : 'Copy'}
                </button>
              </div>
            </div>

            <div className="bg-inset rounded-xl p-3 text-[11px] text-secondary">
              <p className="font-semibold text-primary mb-1">💡 Real-Time Testing</p>
              Once embedded, your website widget will establish a WebSocket connection and generate responses from your AI instantly.
            </div>
          </div>
        </form>
      </div>
    </Modal>
  )
}
