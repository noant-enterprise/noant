import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

interface WhatsAppModalProps {
  open: boolean
  onClose: () => void
  onConnect: (config: any) => Promise<void>
  loading?: boolean
}

export function WhatsAppModal({ open, onClose, onConnect, loading }: WhatsAppModalProps) {
  const [phoneNumberId, setPhoneNumberId] = useState('')
  const [businessAccountId, setBusinessAccountId] = useState('')
  const [accessToken, setAccessToken] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!phoneNumberId.trim() || !accessToken.trim()) return
    onConnect({
      phone_number_id: phoneNumberId.trim(),
      business_account_id: businessAccountId.trim(),
      access_token: accessToken.trim(),
    })
  }

  return (
    <Modal open={open} onClose={onClose} title="Connect WhatsApp Business API">
      <div className="space-y-4">
        <div className="bg-inset rounded-xl p-3.5 text-xs text-secondary space-y-2">
          <p className="font-semibold text-primary">Setup Instructions:</p>
          <ol className="list-decimal list-inside space-y-1">
            <li>Go to the Facebook Developers Console.</li>
            <li>Create an App and select the "WhatsApp" product.</li>
            <li>Retrieve your Phone Number ID and System Access Token.</li>
            <li>Enter details below to generate your unique Webhook URL.</li>
          </ol>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary mb-1">
              Phone Number ID <span className="text-red-500">*</span>
            </label>
            <Input
              value={phoneNumberId}
              onChange={(e) => setPhoneNumberId(e.target.value)}
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
              onChange={(e) => setAccessToken(e.target.value)}
              placeholder="EAAG..."
              required
              disabled={loading}
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="ghost" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" loading={loading} disabled={!phoneNumberId.trim() || !accessToken.trim() || loading}>
              Connect Channel
            </Button>
          </div>
        </form>
      </div>
    </Modal>
  )
}
