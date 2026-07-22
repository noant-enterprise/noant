import { useEffect, useState } from 'react'
import { useAPI } from '@/hooks/useAPI'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { useToast } from '@/components/ui/Toast'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { useModal } from '@/hooks/useModal'
import { cn } from '@/lib/utils'
import type { ProfileResponse, TeamResponse, APIKeysResponse } from '@/types/api'
import type { TeamMember, APIKey } from '@/types'
import BillingPage from '../billing/page'

export default function SetupPage() {
  const { open: showRemove, openModal: openRemove, closeModal: closeRemove } = useModal()
  const { open: showRevoke, openModal: openRevoke, closeModal: closeRevoke } = useModal()
  const [removeId, setRemoveId] = useState('')
  const [revokeId, setRevokeId] = useState('')
  const [actionLoading, setActionLoading] = useState(false)
  const [tab, setTab] = useState<'profile' | 'team' | 'billing' | 'api'>('profile')
  
  const { data: profile, get: getProfile, loading: profLoading } = useAPI<ProfileResponse>()
  
  const { data: teamData, get: getTeam, loading: teamLoading } = useAPI<TeamResponse>()
  
  const { data: keys, get: getKeys, loading: keyLoading } = useAPI<APIKeysResponse>()
  
  const { put } = useAPI()
  
  const { post } = useAPI()
  
  const { del } = useAPI()
  
  const { toast } = useToast()

  useEffect(() => {
    if (tab === 'profile') getProfile('/settings/profile')
    if (tab === 'team') getTeam('/settings/team')
    if (tab === 'api') getKeys('/settings/api-keys')
  }, [tab])

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const form = e.target as HTMLFormElement
      await put('/settings/profile', {
        first_name: (form.elements.namedItem('first') as HTMLInputElement).value,
        last_name: (form.elements.namedItem('last') as HTMLInputElement).value,
        company_name: (form.elements.namedItem('company') as HTMLInputElement).value,
        phone: (form.elements.namedItem('phone') as HTMLInputElement).value,
      })
      toast('Profile saved', 'success')
    } catch {
      toast('Failed to save profile', 'error')
    }
  }

  const handleInvite = async () => {
    const email = prompt('Email:')
    if (!email) return
    try {
      await post('/settings/team/invite', { email, role: 'agent' })
      toast('Invitation sent', 'success')
      getTeam('/settings/team')
    } catch {
      toast('Failed to send invitation', 'error')
    }
  }

  const handleRemoveClick = (id: string) => {
    setRemoveId(id)
    openRemove()
  }

  const handleRemoveConfirm = async () => {
    if (!removeId) return
    setActionLoading(true)
    try {
      await del(`/settings/team/${removeId}`)
      toast('Team member removed', 'success')
      getTeam('/settings/team')
    } catch {
      toast('Failed to remove member', 'error')
    } finally {
      setActionLoading(false)
      closeRemove()
      setRemoveId('')
    }
  }

  const handleGenerateKey = async () => {
    const name = prompt('Key name:') || 'New Key'
    try {
      await post('/settings/api-keys', { name })
      toast('API key generated', 'success')
      getKeys('/settings/api-keys')
    } catch {
      toast('Failed to generate key', 'error')
    }
  }

  const handleRevokeClick = (id: string) => {
    setRevokeId(id)
    openRevoke()
  }

  const handleRevokeConfirm = async () => {
    if (!revokeId) return
    setActionLoading(true)
    try {
      await del(`/settings/api-keys/${revokeId}`)
      toast('API key revoked', 'success')
      getKeys('/settings/api-keys')
    } catch {
      toast('Failed to revoke key', 'error')
    } finally {
      setActionLoading(false)
      closeRevoke()
      setRevokeId('')
    }
  }

  const members = teamData?.members || []
  const keyList = keys?.api_keys || []

  return (
    <div className="max-w-3xl mx-auto animate-fade-in pt-2">
      {/* Mobile tabs */}
      <div className="px-1 flex gap-0 border-b border-default mb-5 lg:mb-6 overflow-x-auto scrollbar-hide">
        {(['profile', 'team', 'billing', 'api'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={cn(
              'px-4 lg:px-5 py-3 text-sm font-semibold border-b-2 -mb-px transition-colors capitalize shrink-0 whitespace-nowrap',
              tab === t ? 'text-noant-sky-deep border-noant-sky' : 'text-tertiary border-transparent hover:text-secondary'
            )}
          >
            {t === 'api' ? 'API Keys' : t}
          </button>
        ))}
      </div>

      <div className="px-1 pb-4">
        {tab === 'profile' && (
          <Card>
            <CardHeader className="px-4 py-3 lg:px-6 lg:py-4"><CardTitle>Your profile</CardTitle></CardHeader>
            <CardBody className="p-3 lg:p-4">
              {profLoading ? <ProfileSkeleton /> : (
                <form onSubmit={handleSaveProfile} className="space-y-5">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
                    <div>
                      <label className="block text-sm font-semibold text-primary mb-2">First name</label>
                      <Input name="first" defaultValue={profile?.first_name} />
                    </div>
                    <div>
                      <label className="block text-sm font-semibold text-primary mb-2">Last name</label>
                      <Input name="last" defaultValue={profile?.last_name} />
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-semibold text-primary mb-2">Email</label>
                    <Input defaultValue={profile?.email} disabled />
                  </div>
                  <div>
                    <label className="block text-sm font-semibold text-primary mb-2">Company</label>
                    <Input name="company" defaultValue={profile?.company_name} />
                  </div>
                  <div>
                    <label className="block text-sm font-semibold text-primary mb-2">Phone</label>
                    <Input name="phone" defaultValue={profile?.phone} />
                  </div>
                  <Button type="submit">Save changes</Button>
                </form>
              )}
            </CardBody>
          </Card>
        )}

        {tab === 'team' && (
          <Card>
            <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
              <CardTitle>Team members</CardTitle>
              <Button size="sm" onClick={handleInvite}>+ Invite</Button>
            </CardHeader>
            <CardBody className="p-0">
              {teamLoading ? <TeamSkeleton /> : members.length === 0 ? (
                <EmptyTeam />
              ) : (
                members.map((m: TeamMember) => (
                  <div key={m.id} className="flex items-center justify-between p-4 border-b border-subtle last:border-b-0 gap-3">
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="w-10 h-10 rounded-full bg-noant-black text-white flex items-center justify-center font-semibold shrink-0">
                        {m.first_name[0]}{m.last_name[0]}
                      </div>
                      <div className="min-w-0">
                        <p className="font-semibold text-sm text-primary truncate">{m.first_name} {m.last_name}</p>
                        <p className="text-xs text-secondary truncate">{m.email}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <Badge variant="sky">{m.role}</Badge>
                      <Button variant="ghost" size="sm" className="text-red-600 px-2" onClick={() => handleRemoveClick(m.id)}>Remove</Button>
                    </div>
                  </div>
                ))
              )}
            </CardBody>
          </Card>
        )}

        {tab === 'billing' && (
          <BillingPage />
        )}

        {tab === 'api' && (
          <Card>
            <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
              <CardTitle>API Keys</CardTitle>
              <Button size="sm" onClick={handleGenerateKey}>+ Generate</Button>
            </CardHeader>
            <CardBody className="p-0">
              {keyLoading ? <ApiKeySkeleton /> : keyList.length === 0 ? (
                <EmptyApiKeys />
              ) : (
                keyList.map((k: APIKey) => (
                  <div key={k.id} className="flex items-center justify-between p-4 border-b border-subtle last:border-b-0 gap-3">
                    <div className="min-w-0 flex-1">
                      <p className="font-semibold text-sm text-primary">{k.name}</p>
                      <p className="font-mono text-xs text-secondary truncate">{k.key || '••••••••'}</p>
                      <p className="text-xs text-tertiary">Created {new Date(k.created_at).toLocaleDateString()}</p>
                    </div>
                    <Button variant="ghost" size="sm" className="text-red-600 shrink-0" onClick={() => handleRevokeClick(k.id)}>Revoke</Button>
                  </div>
                ))
              )}
            </CardBody>
          </Card>
        )}
      </div>

      <ConfirmModal
        open={showRemove}
        onClose={closeRemove}
        onConfirm={handleRemoveConfirm}
        title="Remove team member?"
        description="They will lose access to your Noant workspace immediately."
        variant="danger"
        confirmText="Remove"
        loading={actionLoading}
      />

      <ConfirmModal
        open={showRevoke}
        onClose={closeRevoke}
        onConfirm={handleRevokeConfirm}
        title="Revoke API key?"
        description="Any integrations using this key will stop working immediately."
        variant="warning"
        confirmText="Revoke"
        loading={actionLoading}
      />
    </div>
  )
}

function ProfileSkeleton() {
  return (
    <div className="space-y-5">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <Skeleton className="h-12 rounded-lg" />
        <Skeleton className="h-12 rounded-lg" />
      </div>
      <Skeleton className="h-12 rounded-lg" />
      <Skeleton className="h-12 rounded-lg" />
      <Skeleton className="h-12 rounded-lg" />
      <Skeleton className="h-10 w-32 rounded-lg" />
    </div>
  )
}

function TeamSkeleton() {
  return (
    <div className="divide-y divide-subtle">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="flex items-center justify-between p-4 gap-3">
          <div className="flex items-center gap-3 min-w-0">
            <Skeleton className="w-10 h-10 rounded-full shrink-0" />
            <div className="space-y-1 min-w-0">
              <Skeleton className="h-4 w-32 rounded" />
              <Skeleton className="h-3 w-24 rounded" />
            </div>
          </div>
          <div className="flex gap-2 shrink-0">
            <Skeleton className="h-6 w-16 rounded" />
            <Skeleton className="h-8 w-16 rounded" />
          </div>
        </div>
      ))}
    </div>
  )
}

function ApiKeySkeleton() {
  return (
    <div className="divide-y divide-subtle">
      {Array.from({ length: 2 }).map((_, i) => (
        <div key={i} className="flex items-center justify-between p-4 gap-3">
          <div className="space-y-2 min-w-0 flex-1">
            <Skeleton className="h-4 w-24 rounded" />
            <Skeleton className="h-3 w-48 rounded" />
            <Skeleton className="h-3 w-32 rounded" />
          </div>
          <Skeleton className="h-8 w-16 rounded shrink-0" />
        </div>
      ))}
    </div>
  )
}

function EmptyTeam() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center px-4">
      <p className="text-secondary text-sm">No team members yet. Invite someone to collaborate.</p>
    </div>
  )
}

function EmptyApiKeys() {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center px-4">
      <p className="text-secondary text-sm">No API keys yet. Generate one to get started.</p>
    </div>
  )
}