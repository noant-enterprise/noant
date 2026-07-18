import { useState, useEffect, useCallback } from 'react'
import { UserPlus, Crown, Shield, User, MoreVertical, Trash2, Mail, X, ChevronDown } from 'lucide-react'
import { api } from '../../../lib/api'
import { useAuth } from '@/hooks/useAuth'
import { useToast } from '@/components/ui/Toast'
import { useConfirm } from '@/hooks/useConfirm'
import { UpgradeModal } from '@/components/ui'

interface Member {
  user_id: string
  email: string
  first_name: string
  last_name: string
  role: string
  joined_at: string
}

const roleConfig: Record<string, { icon: typeof Crown; label: string; color: string; bg: string }> = {
  owner:  { icon: Crown,  label: 'Owner',  color: 'text-amber-500', bg: 'bg-amber-500/10' },
  admin:  { icon: Shield, label: 'Admin',  color: 'text-blue-500',  bg: 'bg-blue-500/10'  },
  member: { icon: User,   label: 'Member', color: 'text-secondary', bg: 'bg-inset'         },
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function InviteModal({ onClose, onSuccess }: { onClose: () => void; onSuccess: () => void }) {
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('member')
  const [sending, setSending] = useState(false)
  const { toast: showToast } = useToast()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email) return
    setSending(true)
    try {
      await api.post('/settings/team/invite', { email, role })
      showToast(`Invite sent to ${email}`, 'success')
      onSuccess()
      onClose()
    } catch {
      showToast('Failed to send invite. Please try again.', 'error')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-overlay" onClick={onClose} />
      <div className="relative bg-surface rounded-2xl border border-default shadow-lg w-full max-w-md animate-slide-up">
        <div className="flex items-center justify-between px-5 py-4 border-b border-default">
          <h3 className="font-semibold text-primary">Invite team member</h3>
          <button onClick={onClose} className="w-7 h-7 rounded-lg hover:bg-inset flex items-center justify-center text-secondary hover:text-primary transition-all">
            <X className="w-4 h-4" />
          </button>
        </div>
        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          <div>
            <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Email address</label>
            <div className="relative">
              <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-tertiary" />
              <input
                type="email"
                value={email}
                onChange={e => setEmail(e.target.value)}
                placeholder="colleague@company.com"
                required
                autoFocus
                className="w-full pl-9 pr-3 py-2.5 rounded-lg border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all placeholder:text-tertiary"
              />
            </div>
          </div>
          <div>
            <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Role</label>
            <div className="relative">
              <select
                value={role}
                onChange={e => setRole(e.target.value)}
                className="w-full px-3 py-2.5 rounded-lg border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky transition-all appearance-none pr-8"
              >
                <option value="member">Member — can view and reply to chats</option>
                <option value="admin">Admin — full access except billing</option>
              </select>
              <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-tertiary pointer-events-none" />
            </div>
          </div>
          <div className="flex gap-2 pt-2">
            <button type="button" onClick={onClose} className="flex-1 py-2.5 rounded-lg border border-default text-sm font-medium text-secondary hover:bg-inset hover:text-primary transition-all">
              Cancel
            </button>
            <button
              type="submit"
              disabled={sending || !email}
              className="flex-1 py-2.5 rounded-lg bg-noant-sky text-white text-sm font-semibold hover:bg-noant-sky-deep active:scale-[0.98] transition-all shadow-sky disabled:opacity-60"
            >
              {sending ? 'Sending…' : 'Send invite'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function TeamPage() {
  const [members, setMembers] = useState<Member[]>([])
  const [loading, setLoading] = useState(true)
  const [showInvite, setShowInvite] = useState(false)
  const [showUpgradeModal, setShowUpgradeModal] = useState(false)
  const [menuOpen, setMenuOpen] = useState<string | null>(null)
  const { user } = useAuth()
  const { toast: showToast } = useToast()
  const confirm = useConfirm()

  const handleInviteClick = () => {
    const currentPlan = user?.plan_id || user?.plan || 'free'
    if (currentPlan === 'free' || currentPlan === 'pulse') {
      setShowUpgradeModal(true)
      return
    }
    setShowInvite(true)
  }

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await api.get<{ members: Member[] }>('/settings/team')
      setMembers(res.members || [])
    } catch {
      showToast('Failed to load team members', 'error')
    } finally {
      setLoading(false)
    }
  }, [showToast])

  useEffect(() => { load() }, [load])

  const handleRemove = (memberId: string, memberEmail: string) => {
    confirm({
      title: 'Remove team member?',
      body: `Remove ${memberEmail} from your team? They will lose access to all shared resources.`,
      variant: 'danger',
      confirmText: 'Remove',
      onConfirm: async () => {
        try {
          await api.delete(`/settings/team/${memberId}`)
          setMembers(prev => prev.filter(m => m.user_id !== memberId))
          showToast('Team member removed', 'success')
        } catch {
          showToast('Failed to remove member', 'error')
        }
        setMenuOpen(null)
      },
    })
  }

  const isOwner = user?.role === 'owner'

  if (loading) {
    return (
      <div className="min-h-screen p-4 lg:p-6">
        <div className="max-w-2xl mx-auto space-y-3">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-16 rounded-xl animate-shimmer-slow" />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen p-4 lg:p-6 animate-page-in" onClick={() => setMenuOpen(null)}>
      <div className="max-w-2xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-bold text-primary">Team</h1>
            <p className="text-sm text-secondary mt-0.5">{members.length} member{members.length !== 1 ? 's' : ''}</p>
          </div>
          {isOwner && (
            <button
              onClick={handleInviteClick}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-noant-sky text-white text-sm font-semibold hover:bg-noant-sky-deep active:scale-[0.98] transition-all shadow-sky"
            >
              <UserPlus className="w-4 h-4 mr-1.5" /> Invite member
            </button>
          )}
        </div>

        {/* Members list */}
        <div className="bg-surface rounded-2xl border border-default overflow-hidden">
          {members.length === 0 ? (
            <div className="p-10 text-center">
              <div className="w-14 h-14 rounded-2xl bg-inset flex items-center justify-center mx-auto mb-4">
                <UserPlus className="w-6 h-6 text-tertiary" />
              </div>
              <h3 className="font-semibold text-primary mb-1">No team members yet</h3>
              <p className="text-sm text-secondary mb-4">Invite colleagues to collaborate on NOANT.</p>
              {isOwner && (
                <button
                  onClick={handleInviteClick}
                  className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-noant-sky text-white text-sm font-semibold hover:bg-noant-sky-deep transition-all"
                >
                  <UserPlus className="w-4 h-4 mr-1.5" />
                  Invite your first member
                </button>
              )}
            </div>
          ) : (
            members.map((member, i) => {
              const rc = roleConfig[member.role] || roleConfig.member
              const RoleIcon = rc.icon
              const isMe = member.user_id === user?.id
              const initials = `${member.first_name[0]}${member.last_name[0]}`.toUpperCase()

              return (
                <div
                  key={member.user_id}
                  className={`flex items-center gap-3 px-4 py-3.5 ${i < members.length - 1 ? 'border-b border-default' : ''} hover:bg-inset/30 transition-colors`}
                >
                  {/* Avatar */}
                  <div className="w-9 h-9 rounded-xl bg-noant-black text-white dark:bg-white dark:text-noant-black flex items-center justify-center text-xs font-bold shrink-0">
                    {initials}
                  </div>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-semibold text-primary truncate">
                        {member.first_name} {member.last_name}
                      </span>
                      {isMe && (
                        <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-noant-sky/10 text-noant-sky">You</span>
                      )}
                    </div>
                    <div className="text-xs text-secondary truncate">{member.email}</div>
                  </div>

                  {/* Role badge */}
                  <div className={`hidden sm:flex items-center gap-1 px-2.5 py-1 rounded-lg ${rc.bg} shrink-0`}>
                    <RoleIcon className={`w-3 h-3 ${rc.color}`} />
                    <span className={`text-xs font-semibold ${rc.color}`}>{rc.label}</span>
                  </div>

                  {/* Joined */}
                  <div className="hidden lg:block text-xs text-tertiary shrink-0 ml-2">
                    Joined {formatDate(member.joined_at)}
                  </div>

                  {/* Actions */}
                  {isOwner && !isMe && (
                    <div className="relative" onClick={e => e.stopPropagation()}>
                      <button
                        onClick={() => setMenuOpen(menuOpen === member.user_id ? null : member.user_id)}
                        className="w-7 h-7 rounded-lg hover:bg-inset flex items-center justify-center text-tertiary hover:text-secondary transition-all"
                      >
                        <MoreVertical className="w-4 h-4" />
                      </button>
                      {menuOpen === member.user_id && (
                        <div className="absolute right-0 top-8 z-20 w-40 bg-surface border border-default rounded-xl shadow-lg overflow-hidden animate-fade-in">
                          <button
                            onClick={() => handleRemove(member.user_id, member.email)}
                            className="flex items-center gap-2 w-full px-3 py-2.5 text-sm text-red-500 hover:bg-red-500/10 transition-colors"
                          >
                            <Trash2 className="w-4 h-4" />
                            Remove member
                          </button>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )
            })
          )}
        </div>

        {/* Plan info */}
        <p className="mt-4 text-xs text-tertiary text-center">
          Team members are available on Pro and Enterprise plans.{' '}
          <a href="/billing" className="text-noant-sky hover:underline">View plans →</a>
        </p>
      </div>

      {showInvite && (
        <InviteModal onClose={() => setShowInvite(false)} onSuccess={load} />
      )}

      {showUpgradeModal && (
        <UpgradeModal
          open={showUpgradeModal}
          onClose={() => setShowUpgradeModal(false)}
          title="Unlock Team Collaboration"
          description="Team member management is exclusive to Pro and Enterprise plans. Upgrade now to collaborate with your team."
          featureList={[
            'Add unlimited team members',
            'Assign custom roles (Admin, Member)',
            'Collaborative inbox & lead scoring',
            'Priority developer support',
          ]}
        />
      )}
    </div>
  )
}
