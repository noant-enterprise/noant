import { useState, useEffect, useCallback } from 'react'
import { Code2, Copy, Check, MessageSquare, RefreshCw, Palette, AlignRight, AlignLeft, Save } from 'lucide-react'
import { api } from '../../../lib/api'
import { useToast } from '@/components/ui/Toast'

interface WidgetConfig {
  brand_color: string
  greeting: string
  bot_name: string
  position: string
  widget_api_key: string
  is_active: boolean
}

const defaultConfig: WidgetConfig = {
  brand_color: '#3b82f6',
  greeting: 'Hello! How can I help you today? 👋',
  bot_name: 'Noant AI',
  position: 'bottom-right',
  widget_api_key: '',
  is_active: true,
}

function WidgetPreview({ config }: { config: WidgetConfig }) {
  const [open, setOpen] = useState(true)
  const [input, setInput] = useState('')
  const [messages] = useState([
    { from: 'bot', text: config.greeting },
    { from: 'user', text: 'Do you ship to Lagos?' },
    { from: 'bot', text: 'Yes! We ship to all states in Nigeria. Delivery takes 2-5 business days to Lagos.' },
  ])

  return (
    <div className="relative w-full h-[400px] bg-inset rounded-2xl border border-default overflow-hidden flex items-end justify-end p-4">
      {/* Background pattern */}
      <div className="absolute inset-0 opacity-5">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_1px_1px,var(--border-strong)_1px,transparent_0)] bg-[size:20px_20px]" />
      </div>

      {open ? (
        <div
          className="w-72 bg-white rounded-2xl shadow-xl overflow-hidden flex flex-col border border-gray-100"
          style={{ maxHeight: '340px' }}
        >
          {/* Widget header */}
          <div className="flex items-center gap-2.5 px-4 py-3" style={{ background: config.brand_color }}>
            <div className="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center">
              <MessageSquare className="w-4 h-4 text-white" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="text-white text-sm font-semibold">{config.bot_name}</div>
              <div className="text-white/70 text-[10px]">Usually replies instantly</div>
            </div>
            <button
              onClick={() => setOpen(false)}
              className="text-white/70 hover:text-white text-lg leading-none"
            >×</button>
          </div>

          {/* Messages */}
          <div className="flex-1 overflow-y-auto p-3 space-y-2 bg-gray-50">
            {messages.map((msg, i) => (
              <div key={i} className={`flex ${msg.from === 'user' ? 'justify-end' : 'justify-start'}`}>
                <div
                  className={`max-w-[80%] px-3 py-1.5 rounded-xl text-xs leading-relaxed ${
                    msg.from === 'user'
                      ? 'text-white rounded-br-sm'
                      : 'bg-white text-gray-700 border border-gray-200 rounded-bl-sm'
                  }`}
                  style={msg.from === 'user' ? { background: config.brand_color } : {}}
                >
                  {msg.text}
                </div>
              </div>
            ))}
          </div>

          {/* Input */}
          <div className="px-3 py-2.5 bg-white border-t border-gray-100 flex gap-2">
            <input
              value={input}
              onChange={e => setInput(e.target.value)}
              placeholder="Type your message…"
              className="flex-1 text-xs text-gray-700 bg-transparent outline-none placeholder:text-gray-400"
            />
            <button
              className="w-6 h-6 rounded-full flex items-center justify-center text-white text-xs shrink-0"
              style={{ background: config.brand_color }}
            >
              ›
            </button>
          </div>
        </div>
      ) : (
        <button
          onClick={() => setOpen(true)}
          className="w-14 h-14 rounded-2xl shadow-lg flex items-center justify-center text-white hover:scale-105 active:scale-95 transition-transform"
          style={{ background: config.brand_color }}
        >
          <MessageSquare className="w-6 h-6" />
        </button>
      )}
    </div>
  )
}

function CodeBlock({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)
  const { toast: showToast } = useToast()

  const copy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    showToast('Embed code copied!', 'success')
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative">
      <pre className="text-xs bg-zinc-900 text-green-400 p-4 rounded-xl overflow-x-auto leading-relaxed font-mono border border-zinc-700">
        {code}
      </pre>
      <button
        onClick={copy}
        className="absolute top-2.5 right-2.5 flex items-center gap-1 px-2.5 py-1.5 rounded-lg bg-zinc-700 hover:bg-zinc-600 text-white text-xs font-medium transition-all"
      >
        {copied ? <Check className="w-3 h-3 text-green-400" /> : <Copy className="w-3 h-3" />}
        {copied ? 'Copied!' : 'Copy'}
      </button>
    </div>
  )
}

export default function WidgetPage() {
  const [config, setConfig] = useState<WidgetConfig>(defaultConfig)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const { toast: showToast } = useToast()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await api.get<WidgetConfig>('/widget/config')
      setConfig({ ...defaultConfig, ...res })
    } catch {
      // Use defaults if no config yet
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.post('/widget/config', config)
      showToast('Widget configuration saved!', 'success')
    } catch {
      showToast('Failed to save configuration', 'error')
    } finally {
      setSaving(false)
    }
  }

  const embedCode = config.widget_api_key
    ? `<!-- NOANT Web Widget -->
<script>
  window.NOANT_CONFIG = {
    apiKey: "${config.widget_api_key}",
    position: "${config.position}"
  };
</script>
<script src="https://cdn.noant.ai/widget.js" async></script>`
    : `<!-- Save your config first to generate your embed code -->`

  if (loading) {
    return (
      <div className="min-h-screen p-4 lg:p-6">
        <div className="max-w-4xl mx-auto">
          <div className="h-10 w-48 rounded-lg animate-shimmer-slow mb-6" />
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="h-80 rounded-2xl animate-shimmer-slow" />
            <div className="h-80 rounded-2xl animate-shimmer-slow" />
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen p-4 lg:p-6 animate-page-in">
      <div className="max-w-4xl mx-auto">
        <div className="mb-6">
          <h1 className="text-xl font-bold text-primary">Web Chat Widget</h1>
          <p className="text-sm text-secondary mt-0.5">Embed a live AI chat widget on your website in minutes.</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Config panel */}
          <div className="space-y-5">
            <div className="bg-surface rounded-2xl border border-default p-5 space-y-4">
              <h2 className="font-semibold text-primary flex items-center gap-2">
                <Palette className="w-4 h-4 text-noant-sky" />
                Customise
              </h2>

              <div>
                <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Bot Name</label>
                <input
                  type="text"
                  value={config.bot_name}
                  onChange={e => setConfig(c => ({ ...c, bot_name: e.target.value }))}
                  className="w-full px-3 py-2.5 rounded-lg border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Welcome Message</label>
                <textarea
                  value={config.greeting}
                  onChange={e => setConfig(c => ({ ...c, greeting: e.target.value }))}
                  rows={2}
                  className="w-full px-3 py-2.5 rounded-lg border border-default bg-inset text-primary text-sm focus:outline-none focus:border-noant-sky focus:ring-1 focus:ring-noant-sky/20 transition-all resize-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Brand Color</label>
                <div className="flex items-center gap-3">
                  <input
                    type="color"
                    value={config.brand_color}
                    onChange={e => setConfig(c => ({ ...c, brand_color: e.target.value }))}
                    className="w-10 h-10 rounded-lg border border-default cursor-pointer p-0.5 bg-inset"
                  />
                  <input
                    type="text"
                    value={config.brand_color}
                    onChange={e => setConfig(c => ({ ...c, brand_color: e.target.value }))}
                    className="flex-1 px-3 py-2.5 rounded-lg border border-default bg-inset text-primary text-sm font-mono focus:outline-none focus:border-noant-sky transition-all"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-secondary uppercase tracking-wide mb-1.5">Position</label>
                <div className="flex gap-2">
                  {(['bottom-right', 'bottom-left'] as const).map(pos => (
                    <button
                      key={pos}
                      onClick={() => setConfig(c => ({ ...c, position: pos }))}
                      className={`flex-1 flex items-center justify-center gap-1.5 py-2.5 rounded-lg border text-xs font-medium transition-all ${
                        config.position === pos
                          ? 'border-noant-sky bg-noant-sky/10 text-noant-sky-deep dark:text-noant-sky'
                          : 'border-default text-secondary hover:border-strong hover:text-primary'
                      }`}
                    >
                      {pos === 'bottom-right' ? <AlignRight className="w-3.5 h-3.5" /> : <AlignLeft className="w-3.5 h-3.5" />}
                      {pos === 'bottom-right' ? 'Bottom right' : 'Bottom left'}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex items-center justify-between pt-1">
                <div>
                  <div className="text-sm font-medium text-primary">Widget Active</div>
                  <div className="text-xs text-secondary">Show widget on your website</div>
                </div>
                <button
                  onClick={() => setConfig(c => ({ ...c, is_active: !c.is_active }))}
                  className={`relative inline-flex h-5 w-9 rounded-full border-2 border-transparent cursor-pointer transition-colors duration-200 ${config.is_active ? 'bg-noant-sky' : 'bg-border-strong'}`}
                >
                  <span className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow-lg transform transition duration-200 ${config.is_active ? 'translate-x-4' : 'translate-x-0'}`} />
                </button>
              </div>

              <button
                onClick={handleSave}
                disabled={saving}
                className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl bg-noant-sky text-white text-sm font-semibold hover:bg-noant-sky-deep active:scale-[0.98] transition-all shadow-sky disabled:opacity-60"
              >
                <Save className="w-4 h-4" />
                {saving ? 'Saving…' : 'Save configuration'}
              </button>
            </div>

            {/* Embed code */}
            <div className="bg-surface rounded-2xl border border-default p-5">
              <div className="flex items-center justify-between mb-3">
                <h2 className="font-semibold text-primary flex items-center gap-2">
                  <Code2 className="w-4 h-4 text-noant-sky" />
                  Embed Code
                </h2>
                <button onClick={load} className="text-tertiary hover:text-secondary transition-colors">
                  <RefreshCw className="w-3.5 h-3.5" />
                </button>
              </div>
              <p className="text-xs text-secondary mb-3">
                Paste this snippet before the <code className="text-noant-sky bg-noant-sky/10 px-1 py-0.5 rounded">{'</body>'}</code> tag on your website.
              </p>
              <CodeBlock code={embedCode} />
              {config.widget_api_key && (
                <div className="mt-3 p-2.5 rounded-lg bg-inset border border-default">
                  <div className="text-[10px] text-tertiary uppercase font-semibold tracking-wide mb-0.5">Widget API Key</div>
                  <div className="text-xs font-mono text-secondary break-all">{config.widget_api_key}</div>
                </div>
              )}
            </div>
          </div>

          {/* Live preview */}
          <div>
            <div className="bg-surface rounded-2xl border border-default p-5">
              <h2 className="font-semibold text-primary mb-3 flex items-center gap-2">
                <MessageSquare className="w-4 h-4 text-noant-sky" />
                Live Preview
              </h2>
              <WidgetPreview config={config} />
              <p className="text-xs text-tertiary text-center mt-3">This is how the widget will look on your website</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
