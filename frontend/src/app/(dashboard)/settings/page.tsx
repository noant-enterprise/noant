import { useState, useEffect } from 'react'
import { User, Shield, Bell, Palette, Lock, Save, Eye, EyeOff, Download, Trash2, Check, Globe, Smartphone, BellOff, ScrollText } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { api } from '../../../lib/api'
import { useToast } from '@/components/ui/Toast'
import { usePushNotifications } from '@/hooks/usePushNotifications'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { AuditLogTab } from '@/components/settings/AuditLogTab'

type Tab = 'profile' | 'security' | 'notifications' | 'appearance' | 'privacy' | 'audit'

const tabs: { id: Tab; label: string; icon: typeof User }[] = [
  { id: 'profile', label: 'Profile', icon: User },
  { id: 'security', label: 'Security', icon: Shield },
  { id: 'notifications', label: 'Notifications', icon: Bell },
  { id: 'appearance', label: 'Appearance', icon: Palette },
  { id: 'privacy', label: 'Privacy & Data', icon: Lock },
  { id: 'audit', label: 'Audit Log', icon: ScrollText },
]

function ProfileTab() {
  const { user, refreshUser } = useAuth()
  const { toast: showToast } = useToast()
  const [form, setForm] = useState({
    first_name: user?.first_name || '',
    last_name: user?.last_name || '',
    email: user?.email || '',
    company_name: user?.company_name || '',
    phone: user?.phone || '',
  })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (user) {
      setForm({
        first_name: user.first_name || '',
        last_name: user.last_name || '',
        email: user.email || '',
        company_name: user.company_name || '',
        phone: user?.phone || '',
      })
    }
  }, [user])

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.put('/settings/profile', form)
      await refreshUser?.()
      showToast('Profile updated successfully', 'success')
    } catch {
      showToast('Failed to update profile', 'error')
    } finally {
      setSaving(false)
    }
  }

  const initials = user ? `${user.first_name[0]}${user.last_name[0]}`.toUpperCase() : '--'

  return (
    <div className="space-y-6 animate-page-in">
      {/* Avatar */}
      <div className="flex items-center gap-5">
        <div className="w-20 h-20 rounded-2xl bg-noant-black text-white dark:bg-white dark:text-noant-black flex items-center justify-center text-2xl font-bold select-none shrink-0">
          {initials}
        </div>
        <div>
          <h3 className="font-semibold text-primary">{form.first_name} {form.last_name}</h3>
          <p className="text-sm text-secondary">{form.email}</p>
          <p className="text-xs text-tertiary mt-0.5 capitalize">{user?.plan || 'Free'} plan · {user?.role || 'owner'}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">First Name</label>
          <input
            type="text"
            value={form.first_name}
            onChange={e => setForm(f => ({ ...f, first_name: e.target.value }))}
            className="w-full px-3 py-2.5 rounded-xl border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all"
          />
        </div>
        <div>
          <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Last Name</label>
          <input
            type="text"
            value={form.last_name}
            onChange={e => setForm(f => ({ ...f, last_name: e.target.value }))}
            className="w-full px-3 py-2.5 rounded-xl border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all"
          />
        </div>
        <div>
          <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Email</label>
          <input
            type="email"
            value={form.email}
            disabled
            className="w-full px-3 py-2.5 rounded-xl border border-default bg-inset text-tertiary text-sm cursor-not-allowed"
          />
        </div>
        <div>
          <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Phone</label>
          <input
            type="tel"
            value={form.phone}
            onChange={e => setForm(f => ({ ...f, phone: e.target.value }))}
            placeholder="+234 800 000 0000"
            className="w-full px-3 py-2.5 rounded-xl border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all placeholder:text-tertiary"
          />
        </div>
        <div className="sm:col-span-2">
          <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Company / Brand Name</label>
          <input
            type="text"
            value={form.company_name}
            onChange={e => setForm(f => ({ ...f, company_name: e.target.value }))}
            placeholder="Your business name"
            className="w-full px-3 py-2.5 rounded-xl border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all placeholder:text-tertiary"
          />
        </div>
      </div>

      <button
        onClick={handleSave}
        disabled={saving}
        className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-noant-sky text-white text-sm font-semibold hover:bg-noant-sky-deep active:scale-[0.98] transition-all shadow-sky disabled:opacity-60"
      >
        <Save className="w-4 h-4" />
        {saving ? 'Saving…' : 'Save changes'}
      </button>
    </div>
  )
}

function SecurityTab() {
  const { toast: showToast } = useToast()
  const [form, setForm] = useState({ current_password: '', new_password: '', confirm_password: '' })
  const [show, setShow] = useState({ current: false, new: false, confirm: false })
  const [saving, setSaving] = useState(false)

  const handleChange = async () => {
    if (form.new_password !== form.confirm_password) {
      showToast('New passwords do not match', 'error')
      return
    }
    if (form.new_password.length < 8) {
      showToast('Password must be at least 8 characters', 'error')
      return
    }
    setSaving(true)
    try {
      await api.post('/auth/change-password', {
        current_password: form.current_password,
        new_password: form.new_password,
      })
      showToast('Password changed successfully', 'success')
      setForm({ current_password: '', new_password: '', confirm_password: '' })
    } catch {
      showToast('Failed to change password. Check your current password.', 'error')
    } finally {
      setSaving(false)
    }
  }

  const PasswordInput = ({ field, label, showKey }: { field: keyof typeof form, label: string, showKey: keyof typeof show }) => (
    <div>
      <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">{label}</label>
      <div className="relative">
        <input
          type={show[showKey] ? 'text' : 'password'}
          value={form[field]}
          onChange={e => setForm(f => ({ ...f, [field]: e.target.value }))}
          className="w-full pr-10 px-3 py-2.5 rounded-xl border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all"
        />
        <button
          type="button"
          onClick={() => setShow(s => ({ ...s, [showKey]: !s[showKey] }))}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-tertiary hover:text-secondary transition-colors"
        >
          {show[showKey] ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
        </button>
      </div>
    </div>
  )

  return (
    <div className="space-y-6 animate-page-in">
      <div className="p-4 rounded-xl border border-default bg-inset/50">
        <h3 className="font-semibold text-primary mb-1">Change Password</h3>
        <p className="text-sm text-secondary mb-4">Use a strong, unique password that you don't use elsewhere.</p>
        <div className="space-y-4 max-w-md">
          <PasswordInput field="current_password" label="Current Password" showKey="current" />
          <PasswordInput field="new_password" label="New Password" showKey="new" />
          <PasswordInput field="confirm_password" label="Confirm New Password" showKey="confirm" />
          <button
            onClick={handleChange}
            disabled={saving || !form.current_password || !form.new_password}
            className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-noant-sky text-white text-sm font-semibold hover:bg-noant-sky-deep active:scale-[0.98] transition-all shadow-sky disabled:opacity-60"
          >
            <Shield className="w-4 h-4" />
            {saving ? 'Updating…' : 'Update Password'}
          </button>
        </div>
      </div>

      <div className="p-4 rounded-xl border border-default bg-inset/50">
        <h3 className="font-semibold text-primary mb-1">Two-Factor Authentication</h3>
        <p className="text-sm text-secondary mb-3">Add an extra layer of security to your account.</p>
        <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-amber-500/10 text-amber-600 dark:text-amber-400 text-xs font-medium border border-amber-500/20">
          Coming Soon
        </div>
      </div>
    </div>
  )
}

function NotificationsTab() {
  const { toast: showToast } = useToast()
  const push = usePushNotifications()
  const [prefs, setPrefs] = useState({
    notif_escalation: true,
    notif_unknown_questions: true,
    notif_payment: true,
    notif_security: true,
    notif_team_invite: true,
  })
  const [saving, setSaving] = useState(false)
  const [loadError, setLoadError] = useState(false)

  useEffect(() => {
    api.get<typeof prefs>('/settings/notifications').then(res => {
      if (res) setPrefs(prev => ({ ...prev, ...res }))
      setLoadError(false)
    }).catch(() => setLoadError(true))
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.put('/settings/notifications', prefs)
      showToast('Notification preferences saved', 'success')
    } catch {
      showToast('Failed to save preferences', 'error')
    } finally {
      setSaving(false)
    }
  }

  const items = [
    { key: 'notif_escalation' as const, label: 'Conversation Escalations', desc: 'When a customer asks for human support' },
    { key: 'notif_unknown_questions' as const, label: 'Unknown Questions', desc: 'When your AI can\'t answer a question' },
    { key: 'notif_payment' as const, label: 'Payment & Billing', desc: 'Subscription renewals, failed charges' },
    { key: 'notif_security' as const, label: 'Security Alerts', desc: 'New logins, password changes' },
    { key: 'notif_team_invite' as const, label: 'Team Invitations', desc: 'When someone accepts your invite' },
  ]

  return (
    <div className="space-y-4 animate-page-in">
      {/* Push / Desktop Notifications */}
      {push.supported && (
        <div className="rounded-xl border border-default overflow-hidden">
          <div className="flex items-center justify-between px-4 py-3.5 bg-inset/30 border-b border-default">
            <div className="flex items-center gap-2">
              <Smartphone className="w-4 h-4 text-noant-sky" />
              <div>
                <div className="text-sm font-medium text-primary">Desktop & Push Notifications</div>
                <div className="text-xs text-secondary mt-0.5">Receive notifications even when the app is closed</div>
              </div>
            </div>
            <button
              onClick={() => push.subscribed ? push.unsubscribe() : push.subscribe()}
              disabled={push.loading}
              className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent cursor-pointer transition-colors duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-noant-sky ${push.subscribed ? 'bg-noant-sky' : 'bg-border-strong'} disabled:opacity-60`}
              aria-checked={push.subscribed}
              aria-label="Desktop notifications"
              role="switch"
            >
              <span className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow-lg transform transition duration-200 ${push.subscribed ? 'translate-x-4' : 'translate-x-0'}`} />
            </button>
          </div>
          {push.permission === 'denied' && (
            <div className="px-4 py-2.5 text-xs text-amber-500 bg-amber-500/5 flex items-center gap-2">
              <BellOff className="w-3.5 h-3.5 shrink-0" />
              Notifications blocked. Enable them in your browser settings.
            </div>
          )}
        </div>
      )}

      {loadError && (
        <div className="flex items-center gap-2 text-red-500 text-sm px-1">
          <BellOff className="w-4 h-4 shrink-0" /> Failed to load preferences
        </div>
      )}

      <p className="text-sm text-secondary">Choose which events send you email notifications.</p>
      <div className="rounded-xl border border-default overflow-hidden">
        {items.map((item, i) => (
          <div key={item.key} className={`flex items-center justify-between px-4 py-3.5 ${i < items.length - 1 ? 'border-b border-default' : ''} hover:bg-inset/50 transition-colors`}>
            <div>
              <div className="text-sm font-medium text-primary">{item.label}</div>
              <div className="text-xs text-secondary mt-0.5">{item.desc}</div>
            </div>
            <button
              onClick={() => setPrefs(p => ({ ...p, [item.key]: !p[item.key] }))}
              className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent cursor-pointer transition-colors duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-noant-sky ${prefs[item.key] ? 'bg-noant-sky' : 'bg-border-strong'}`}
              aria-checked={prefs[item.key]}
              aria-label="Email notifications"
              role="switch"
            >
              <span className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow-lg transform transition duration-200 ${prefs[item.key] ? 'translate-x-4' : 'translate-x-0'}`} />
            </button>
          </div>
        ))}
      </div>
      <button
        onClick={handleSave}
        disabled={saving}
        className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-noant-sky text-white text-sm font-semibold hover:bg-noant-sky-deep active:scale-[0.98] transition-all shadow-sky disabled:opacity-60"
      >
        <Check className="w-4 h-4" />
        {saving ? 'Saving…' : 'Save preferences'}
      </button>
    </div>
  )
}

function AppearanceTab() {
  const [theme, setTheme] = useState<'light' | 'dark' | 'system'>('system')
  const [language, setLanguage] = useState('en')
  const { toast: showToast } = useToast()

  useEffect(() => {
    const saved = localStorage.getItem('noant_theme')
    if (saved === 'dark' || saved === 'light') setTheme(saved)
    else setTheme('system')
    const savedLang = localStorage.getItem('noant_language') || 'en'
    setLanguage(savedLang)
  }, [])

  const applyTheme = (t: 'light' | 'dark' | 'system') => {
    setTheme(t)
    const html = document.documentElement
    if (t === 'dark') {
      html.classList.add('dark'); html.classList.remove('light')
      localStorage.setItem('noant_theme', 'dark')
    } else if (t === 'light') {
      html.classList.add('light'); html.classList.remove('dark')
      localStorage.setItem('noant_theme', 'light')
    } else {
      html.classList.remove('dark', 'light')
      localStorage.removeItem('noant_theme')
    }
    showToast('Theme updated', 'success')
  }

  const themes = [
    { id: 'light' as const, label: 'Light', preview: 'bg-white border-gray-200' },
    { id: 'dark' as const, label: 'Dark', preview: 'bg-zinc-900 border-zinc-700' },
    { id: 'system' as const, label: 'System', preview: 'bg-gradient-to-br from-white to-zinc-900 border-zinc-400' },
  ]

  return (
    <div className="space-y-6 animate-page-in">
      <div>
        <h3 className="font-semibold text-primary mb-1">Theme</h3>
        <p className="text-sm text-secondary mb-4">Choose how NOANT looks to you.</p>
        <div className="flex gap-3 flex-wrap">
          {themes.map(t => (
            <button
              key={t.id}
              onClick={() => applyTheme(t.id)}
              className={`flex flex-col items-center gap-2 p-3 rounded-xl border-2 transition-all active:scale-95 ${theme === t.id ? 'border-noant-sky bg-noant-sky/5' : 'border-default hover:border-strong'}`}
            >
              <div className={`w-20 h-12 rounded-lg border ${t.preview} relative overflow-hidden`}>
                <div className={`absolute left-0 top-0 h-full w-8 ${t.id === 'light' ? 'bg-gray-100' : t.id === 'dark' ? 'bg-zinc-800' : 'bg-gradient-to-b from-gray-100 to-zinc-800'}`} />
                <div className={`absolute right-0 top-0 h-full w-12 p-1 space-y-1`}>
                  <div className={`h-1.5 rounded-full ${t.id === 'dark' ? 'bg-zinc-700' : 'bg-gray-200'}`} />
                  <div className={`h-1.5 rounded-full w-3/4 ${t.id === 'dark' ? 'bg-zinc-700' : 'bg-gray-200'}`} />
                </div>
              </div>
              <div className="flex items-center gap-1.5">
                {theme === t.id && <div className="w-2 h-2 rounded-full bg-noant-sky" />}
                <span className={`text-xs font-medium ${theme === t.id ? 'text-noant-sky' : 'text-secondary'}`}>{t.label}</span>
              </div>
            </button>
          ))}
        </div>
      </div>

      <div>
        <h3 className="font-semibold text-primary mb-1">Language</h3>
        <p className="text-sm text-secondary mb-3">Interface language preference.</p>
        <div className="flex items-center gap-2 max-w-xs">
          <Globe className="w-4 h-4 text-secondary shrink-0" />
          <select
            value={language}
            onChange={e => { setLanguage(e.target.value); localStorage.setItem('noant_language', e.target.value) }}
            className="flex-1 px-3 py-2.5 rounded-xl border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky transition-all"
          >
            <option value="en">English</option>
            <option value="yo">Yoruba (coming soon)</option>
            <option value="ha">Hausa (coming soon)</option>
            <option value="ig">Igbo (coming soon)</option>
            <option value="hi">Hindi (coming soon)</option>
            <option value="pt">Portuguese (coming soon)</option>
          </select>
        </div>
      </div>
    </div>
  )
}

function PrivacyTab() {
  const { signOut } = useAuth()
  const { toast: showToast } = useToast()
  const [exporting, setExporting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const handleExport = async () => {
    setExporting(true)
    try {
      const res = await api.get('/settings/account/export')
      const blob = new Blob([JSON.stringify(res, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `noant-data-export-${new Date().toISOString().split('T')[0]}.json`
      a.click()
      URL.revokeObjectURL(url)
      showToast('Data exported successfully', 'success')
    } catch {
      showToast('Export failed. Please try again.', 'error')
    } finally {
      setExporting(false)
    }
  }

  const handleDeleteAccount = async () => {
    setDeleting(true)
    try {
      await api.delete('/settings/account')
      showToast('Account deleted. Goodbye!', 'success')
      setTimeout(() => signOut(), 1500)
    } catch {
      showToast('Failed to delete account', 'error')
      setDeleting(false)
      setShowDeleteConfirm(false)
    }
  }

  return (
    <div className="space-y-5 animate-page-in">
      <div className="p-4 rounded-xl border border-default bg-inset/50">
        <h3 className="font-semibold text-primary mb-1">Export Your Data</h3>
        <p className="text-sm text-secondary mb-4">Download a copy of all your data including conversations, training data, and account info (GDPR Article 20).</p>
        <button
          onClick={handleExport}
          disabled={exporting}
          className="flex items-center gap-2 px-4 py-2 rounded-lg border border-default text-sm font-medium text-primary hover:bg-inset hover:border-noant-sky hover:text-noant-sky active:scale-[0.98] transition-all disabled:opacity-60"
        >
          <Download className="w-4 h-4" />
          {exporting ? 'Preparing export…' : 'Download my data'}
        </button>
      </div>

      <div className="p-4 rounded-xl border border-red-500/30 bg-red-500/5">
        <h3 className="font-semibold text-red-500 mb-1">Delete Account</h3>
        <p className="text-sm text-secondary mb-4">Permanently delete your NOANT account and all associated data. This action cannot be undone.</p>
        <button
          onClick={() => setShowDeleteConfirm(true)}
          disabled={deleting}
          className="flex items-center gap-2 px-4 py-2 rounded-lg bg-red-500 text-white text-sm font-semibold hover:bg-red-600 active:scale-[0.98] transition-all disabled:opacity-60"
        >
          <Trash2 className="w-4 h-4" />
          Delete my account
        </button>
      </div>

      <ConfirmModal
        open={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        onConfirm={handleDeleteAccount}
        title="Delete Account"
        description="This action cannot be undone. All your data, conversations, and settings will be permanently deleted."
        confirmText="Delete Account"
        requireTypeConfirm
        confirmPhrase="DELETE"
        loading={deleting}
      />
    </div>
  )
}

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<Tab>('profile')

  const TabContent = () => {
    switch (activeTab) {
      case 'profile': return <ProfileTab />
      case 'security': return <SecurityTab />
      case 'notifications': return <NotificationsTab />
      case 'appearance': return <AppearanceTab />
      case 'privacy': return <PrivacyTab />
      case 'audit': return <AuditLogTab />
    }
  }

  return (
    <div className="min-h-screen p-4 lg:p-6 animate-page-in">
      <div className="max-w-3xl mx-auto">
        <div className="mb-6">
          <h1 className="text-xl font-bold text-primary">Settings</h1>
          <p className="text-sm text-secondary mt-0.5">Manage your account, preferences, and privacy.</p>
        </div>

        {/* Tab bar — scrollable on mobile */}
        <div className="flex gap-1 mb-6 overflow-x-auto pb-1 scrollbar-thin">
          {tabs.map(tab => {
            const Icon = tab.icon
            const isActive = activeTab === tab.id
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-all active:scale-95 shrink-0 ${
                  isActive
                    ? 'bg-noant-sky/10 text-noant-sky-deep dark:text-noant-sky border border-noant-sky/20'
                    : 'text-secondary hover:bg-inset hover:text-primary'
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? 'text-noant-sky' : ''}`} />
                {tab.label}
              </button>
            )
          })}
        </div>

        {/* Content */}
        <div className="bg-surface rounded-2xl border border-default p-5 lg:p-6">
          <TabContent />
        </div>
      </div>
    </div>
  )
}
