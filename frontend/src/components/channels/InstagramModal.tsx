import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface InstagramModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
}

export function InstagramModal({ open, onClose, onConnect, loading }: InstagramModalProps) {
  const [instagramId, setInstagramId] = useState('')
  const [pageAccessToken, setPageAccessToken] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!instagramId.trim() || !pageAccessToken.trim()) return
    onConnect({
      instagram_id: instagramId.trim(),
      page_access_token: pageAccessToken.trim(),
    })
  }

  return (
    <Modal open={open} onClose={onClose} title="Connect Instagram Direct Messages">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary">Setup Instructions:</p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Ensure your Instagram account is a <b>Professional/Business</b> account.</li>
            <li>Link your Instagram Professional account to a Facebook Business Page.</li>
            <li>Enable "Allow Access to Messages" in Instagram app settings.</li>
            <li>Paste your Meta Page Access Token and Instagram Account ID below.</li>
          </ol>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Instagram Account ID <span className="text-red-500">*</span>
            </label>
            <Input
              value={instagramId}
              onChange={(e) => setInstagramId(e.target.value)}
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
              onChange={(e) => setPageAccessToken(e.target.value)}
              placeholder="EAAG..."
              required
              disabled={loading}
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="ghost" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" loading={loading} disabled={!instagramId.trim() || !pageAccessToken.trim() || loading}>
              Connect Channel
            </Button>
          </div>
        </form>
      </div>
    </Modal>
  )
}
