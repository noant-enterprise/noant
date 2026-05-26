import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface FacebookModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
}

export function FacebookModal({ open, onClose, onConnect, loading }: FacebookModalProps) {
  const [pageId, setPageId] = useState('')
  const [pageAccessToken, setPageAccessToken] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!pageId.trim() || !pageAccessToken.trim()) return
    onConnect({
      page_id: pageId.trim(),
      page_access_token: pageAccessToken.trim(),
    })
  }

  return (
    <Modal open={open} onClose={onClose} title="Connect Facebook Messenger">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary">Setup Instructions:</p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Create or log into a Facebook Business Page.</li>
            <li>Go to the Meta App Dashboard, and add the "Messenger" product.</li>
            <li>Link your Facebook Page to your developer app.</li>
            <li>Generate and copy your Page Access Token, and paste it below.</li>
          </ol>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Page ID <span className="text-red-500">*</span>
            </label>
            <Input
              value={pageId}
              onChange={(e) => setPageId(e.target.value)}
              placeholder="e.g. 10928392839283"
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
            <Button type="submit" loading={loading} disabled={!pageId.trim() || !pageAccessToken.trim() || loading}>
              Connect Channel
            </Button>
          </div>
        </form>
      </div>
    </Modal>
  )
}
