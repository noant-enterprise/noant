import { useAuth } from '@/lib/hooks/useAuth'
import { useSystemHealth } from '@/lib/hooks/useSystemHealth'
import { SkeletonCard } from '@/components/ui/Skeleton'
import { Shield, Key, Users, Edit2, X } from 'lucide-react'
import { useState } from 'react'
import { formatVersion } from '@/lib/utils'

export default function SettingsPage() {
  const { user } = useAuth()
  const { data: system, loading: systemLoading } = useSystemHealth()
  const [activeTab, setActiveTab] = useState<'profile' | 'team' | 'api-keys'>('profile')

  if (!user) {
    return (
      <div className="space-y-6">
        <div>
          <SkeletonCard className="h-8 w-32" />
          <SkeletonCard className="mt-2 h-4 w-64" />
        </div>
        <SkeletonCard className="h-10 w-72" />
        <SkeletonCard className="h-48" />
      </div>
    )
  }

  const tabs = [
    { id: 'profile' as const, label: 'Profile', icon: Shield },
    { id: 'team' as const, label: 'Team', icon: Users },
    { id: 'api-keys' as const, label: 'API Keys', icon: Key },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Settings</h1>
        <p className="text-sm text-text-tertiary">Manage your admin account and platform configuration</p>
      </div>

      <div className="flex gap-1 rounded-lg border border-border bg-bg-surface p-1">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'bg-bg-inset text-text-primary'
                : 'text-text-tertiary hover:text-text-secondary'
            }`}
          >
            <tab.icon className="h-4 w-4" />
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'profile' && (
        <div className="rounded-xl border border-border bg-bg-surface p-6">
          <h3 className="mb-4 text-sm font-medium text-text-secondary">Admin Profile</h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-xs text-text-tertiary">Email</label>
              <input
                type="email"
                value={user?.email ?? ''}
                disabled
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-text-tertiary">Role</label>
              <input
                type="text"
                value={user?.role ?? ''}
                disabled
                className="w-full rounded-lg border border-border bg-bg-inset px-3 py-2 text-sm text-text-primary"
              />
            </div>
          </div>
          <div className="mt-4 flex items-center gap-2 text-xs text-text-tertiary">
            <Shield className="h-3.5 w-3.5" />
            Admin access is managed through the JWT token role claim
          </div>
        </div>
      )}

      {activeTab === 'team' && (
        <div className="rounded-xl border border-border bg-bg-surface p-6">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="text-sm font-medium text-text-secondary">Platform Admins</h3>
          </div>
          <div className="rounded-lg border border-border bg-bg-inset p-8 text-center">
            <Users className="mx-auto mb-2 h-8 w-8 text-text-tertiary" />
            <p className="text-sm font-medium text-text-secondary">No team members</p>
            <p className="mt-1 text-xs text-text-tertiary">Team management is available through the main app settings.</p>
            <p className="mt-1 text-xs text-text-tertiary">Admin users are determined by the <code className="rounded bg-bg-surface px-1">role</code> field in the users table.</p>
          </div>
        </div>
      )}

      {activeTab === 'api-keys' && (
        <div className="rounded-xl border border-border bg-bg-surface p-6">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="text-sm font-medium text-text-secondary">Platform API Keys</h3>
          </div>
          <div className="rounded-lg border border-border bg-bg-inset p-8 text-center">
            <Key className="mx-auto mb-2 h-8 w-8 text-text-tertiary" />
            <p className="text-sm text-text-tertiary">API keys are managed per-user through the main app settings.</p>
            <p className="mt-1 text-xs text-text-tertiary">This admin panel uses cookie-based auth and doesn't require API keys.</p>
          </div>
        </div>
      )}

      <div className="rounded-xl border border-border bg-bg-surface p-6">
        <h3 className="mb-4 text-sm font-medium text-text-secondary">System Info</h3>
        <div className="flex items-center justify-between rounded-lg border border-border bg-bg-inset p-4">
          <div>
            <p className="text-sm font-medium text-text-primary">Platform Status</p>
            <p className="text-xs text-text-tertiary">
              {systemLoading ? 'Checking...' : system.api.status === 'healthy' ? 'All systems operational' : 'Some issues detected'}
              {systemLoading ? '' : ` • Version ${formatVersion()}`}
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
