import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface TelegramModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
}

export function TelegramModal({ open, onClose, onConnect, loading }: TelegramModalProps) {
  const [botToken, setBotToken] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!botToken.trim()) return
    onConnect({
      bot_token: botToken.trim(),
    })
  }

  return (
    <Modal open={open} onClose={onClose} title="Connect Telegram Bot">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary">Setup Instructions:</p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Search for <b>@BotFather</b> on Telegram.</li>
            <li>Send the command <code>/newbot</code> and follow instructions to set a name and username.</li>
            <li>Copy the generated HTTP API Bot Token and paste it below.</li>
            <li>Webhook settings are automatically registered upon saving.</li>
          </ol>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Bot Token <span className="text-red-500">*</span>
            </label>
            <Input
              value={botToken}
              onChange={(e) => setBotToken(e.target.value)}
              placeholder="e.g. 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
              required
              disabled={loading}
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="ghost" onClick={onClose} disabled={loading}>
              Cancel
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
