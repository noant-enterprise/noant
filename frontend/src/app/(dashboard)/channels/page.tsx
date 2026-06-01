import { useEffect, useState } from 'react'
import type { Integration } from '@/types'
import { useAPI } from '@/hooks/useAPI'
import {
  ChannelCard,
  WhatsAppModal,
  TelegramModal,
  WebWidgetModal,
  GmailModal,
} from '@/components/channels'
import { ChannelIcon } from '@/components/channels/ChannelIcon'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { useModal } from '@/hooks/useModal'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useWidgetConfig } from '@/contexts/WidgetConfigContext'
import {
  AlertTriangle,
  Check,
  Copy,
  ShieldCheck,
  Unplug,
  Plus,
} from 'lucide-react'

const channelConfig: Record<string, { name: string; desc: string }> = {
  whatsapp: { name: 'WhatsApp', desc: 'OpenWA self-hosted WhatsApp API' },
  telegram: { name: 'Telegram', desc: 'Bot integration with webhook or local polling' },
  gmail: { name: 'Gmail', desc: 'Email customer support via IMAP/SMTP' },
  web: { name: 'Web Widget', desc: 'Embeddable chat widget for your site' },
}

type IntegrationConfig = Record<string, unknown>

function isConnected(status: string) {
  return status === 'connected' || status === 'active'
}

function getConfigValue(config: IntegrationConfig | undefined, keys: string[]) {
  if (!config) return ''
  for (const key of keys) {
    const raw = config[key]
    if (typeof raw === 'string' && raw.trim()) return raw.trim()
    if (typeof raw === 'number' && Number.isFinite(raw)) return String(raw)
    if (typeof raw === 'boolean') return raw ? 'true' : 'false'
  }
  return ''
}

function formatShortDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

function shorten(value: string, head = 10, tail = 8) {
  if (value.length <= head + tail + 3) return value
  return `${value.slice(0, head)}...${value.slice(-tail)}`
}

export default function ChannelsPage() {
  const intAPI = useAPI() as any
  const { data, get: getIntegrations, loading } = intAPI

  const postAPI = useAPI() as any
  const { post } = postAPI

  const { config: widgetConfig, setConfig: setWidgetConfig } = useWidgetConfig()
  const { subscribe } = useWebSocket()
  const { open: showDisconnect, openModal: openDisconnect, closeModal: closeDisconnect } = useModal()

  const [disconnectChannel, setDisconnectChannel] = useState('')
  const [disconnectLoading, setDisconnectLoading] = useState(false)
  const [copiedWebhookId, setCopiedWebhookId] = useState('')
  const [activeModal, setActiveModal] = useState<'telegram' | 'whatsapp' | 'gmail' | 'web' | null>(null)
  const [connectLoading, setConnectLoading] = useState(false)

  useEffect(() => {
    getIntegrations('/integrations/list')
  }, [getIntegrations])

  useEffect(() => {
    const unsub = subscribe((msg) => {
      if (msg.type === 'integration_update') {
        getIntegrations('/integrations/list')
      }
    })
    return unsub
  }, [subscribe, getIntegrations])

  const integrations: Integration[] = data?.integrations || []
  const connectedIntegrations = integrations.filter((i) => isConnected(i.status))
  const integrationMap = new Map<string, Integration>(
    integrations.map((i) => [i.channel, i])
  )

  const issueCount = integrations.filter((i) => i.status === 'error').length

  let lastUpdated = ''
  for (const integration of connectedIntegrations) {
    const candidate = integration.updated_at || integration.created_at || integration.connected_at || ''
    if (candidate && candidate > lastUpdated) {
      lastUpdated = candidate
    }
  }

  const handleConnectClick = (channel: string) => {
    setActiveModal(channel as any)
  }

  const handleConnectSubmit = async (channel: string, config: any) => {
    setConnectLoading(true)
    try {
      await post('/integrations/connect', { channel, config })

      if (channel === 'web') {
        const updatedWidget = {
          bot_name: config.botName,
          greeting: config.greeting,
          brand_color: config.brandColor,
          position: config.position === 'left' ? 'bottom-left' : 'bottom-right',
          is_active: true,
          widget_api_key: widgetConfig?.widget_api_key || '',
        }
        setWidgetConfig(updatedWidget)
        await post('/widget/config', updatedWidget)
      }

      getIntegrations('/integrations/list')
      setActiveModal(null)
    } catch {
      // handled by API hook
    } finally {
      setConnectLoading(false)
    }
  }

  const handleDisconnectClick = (channel: string) => {
    setDisconnectChannel(channel)
    openDisconnect()
  }

  const handleDisconnectConfirm = async () => {
    if (!disconnectChannel) return

    setDisconnectLoading(true)
    try {
      await post(`/integrations/disconnect/${disconnectChannel}`)
      if (disconnectChannel === 'web') {
        setWidgetConfig((current) => {
          if (!current) return null
          const updated = { ...current, is_active: false }
          post('/widget/config', updated)
          return updated
        })
      }
      getIntegrations('/integrations/list')
    } catch {
      // handled by API hook
    } finally {
      setDisconnectLoading(false)
      closeDisconnect()
      setDisconnectChannel('')
    }
  }

  const handleCopyWebhook = async (integration: Integration, webhookUrl: string) => {
    try {
      await navigator.clipboard.writeText(webhookUrl)
      const copyKey = integration.id || integration.channel
      setCopiedWebhookId(copyKey)
      window.setTimeout(() => {
        setCopiedWebhookId((current) => (current === copyKey ? '' : current))
      }, 1800)
    } catch {
      // non-blocking
    }
  }

  function getCardDetails(integration: Integration) {
    const config = (integration.config || {}) as IntegrationConfig
    const details: Array<{ label: string; value: string }> = []

    if (integration.channel === 'whatsapp') {
      const phone = getConfigValue(config, ['phone'])
      const sessionId = getConfigValue(config, ['session_id'])
      if (phone) details.push({ label: 'Number', value: phone })
      if (sessionId) details.push({ label: 'Session', value: shorten(sessionId) })
    }

    if (integration.channel === 'telegram') {
      const botUsername = getConfigValue(config, ['bot_username'])
      if (botUsername) {
        details.push({
          label: 'Bot',
          value: botUsername.startsWith('@') ? botUsername : `@${botUsername}`,
        })
      }
    }

    if (integration.channel === 'gmail') {
      const email = getConfigValue(config, ['email', 'username'])
      if (email) details.push({ label: 'Mailbox', value: email })
    }

    if (integration.channel === 'web') {
      const widgetKey = getConfigValue(config, ['widget_api_key'])
      if (widgetKey) details.push({ label: 'Widget key', value: shorten(widgetKey) })
    }

    return details
  }

  return (
    <div className="animate-page-in space-y-5 lg:space-y-6 pt-2">
      {/* Channel Cards Grid — shows all available channels */}
      <div className="px-1">
        {loading ? (
          <div className="grid grid-cols-1 gap-3 lg:grid-cols-3 lg:gap-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="rounded-2xl border border-default bg-surface p-4 lg:p-5 space-y-4">
                <div className="flex items-center gap-3">
                  <Skeleton className="h-10 w-10 rounded-xl" />
                  <div className="space-y-1.5 flex-1">
                    <Skeleton className="h-4 w-24 rounded" />
                    <Skeleton className="h-3 w-32 rounded" />
                  </div>
                </div>
                <Skeleton className="h-16 rounded-lg" />
                <Skeleton className="h-10 w-full rounded-xl" />
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-3 lg:grid-cols-3 lg:gap-4">
            {Object.entries(channelConfig).map(([key, cfg]) => {
              const integration = integrationMap.get(key)
              const config = (integration?.config || {}) as IntegrationConfig
              const connected = integration && isConnected(integration.status)

              return (
                <ChannelCard
                  key={key}
                  channel={key}
                  name={cfg.name}
                  desc={cfg.desc}
                  status={connected ? 'connected' : 'disconnected'}
                  details={connected ? getCardDetails(integration) : undefined}
                  webhookUrl={connected ? (integration.webhook_url || getConfigValue(config, ['webhook_url'])) : undefined}
                  connectedAt={connected ? (integration.updated_at || integration.created_at || integration.connected_at) : undefined}
                  onConnect={() => handleConnectClick(key)}
                  onDisconnect={() => handleDisconnectClick(key)}
                />
              )
            })}
          </div>
        )}
      </div>

      {/* Connected Channels List */}
      {connectedIntegrations.length > 0 && (
        <div className="px-1 pb-4">
          <Card className="overflow-hidden">
            <CardHeader className="flex flex-col gap-3 px-4 py-4 lg:flex-row lg:items-start lg:justify-between lg:px-6 lg:py-5">
              <div className="space-y-1.5">
                <div className="flex flex-wrap items-center gap-2">
                  <CardTitle>Integration details</CardTitle>
                  <Badge variant="success">{connectedIntegrations.length} live</Badge>
                </div>
                <p className="max-w-2xl text-sm text-secondary">
                  Real-time view of every channel handling customer messages.
                </p>
              </div>

              <div className="flex flex-wrap items-center gap-2 text-xs text-secondary">
                {issueCount > 0 && (
                  <span className="inline-flex items-center gap-1.5 rounded-full border border-red-100 bg-red-50 px-3 py-1.5 text-red-700">
                    <AlertTriangle className="h-3.5 w-3.5" />
                    {issueCount} need attention
                  </span>
                )}
                {lastUpdated && (
                  <span className="inline-flex items-center gap-1.5 rounded-full border border-default bg-inset px-3 py-1.5">
                    <ShieldCheck className="h-3.5 w-3.5 text-tertiary" />
                    Updated {formatShortDate(lastUpdated)}
                  </span>
                )}
              </div>
            </CardHeader>

            <CardBody className="p-0">
              <div className="space-y-3 p-4 lg:p-6">
                {connectedIntegrations.map((integration) => {
                  const meta = channelConfig[integration.channel] || {
                    name: integration.channel,
                    desc: 'Connected integration',
                  }
                  const config = (integration.config || {}) as IntegrationConfig
                  const webhookUrl = integration.webhook_url || getConfigValue(config, ['webhook_url'])
                  const copyKey = integration.id || integration.channel

                  return (
                    <div
                      key={copyKey}
                      className="group rounded-2xl border border-default bg-surface p-4 lg:p-5 transition-all duration-200 hover:border-noant-sky/30 hover:shadow-md"
                    >
                      <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
                        <div className="flex min-w-0 gap-4">
                          <div className="relative shrink-0">
                            <ChannelIcon channel={integration.channel} size="lg" className="shadow-sm ring-1 ring-black/5" />
                            <span className="absolute -right-1 -bottom-1 h-3.5 w-3.5 rounded-full border-2 border-surface bg-emerald-500" />
                          </div>

                          <div className="min-w-0 space-y-3">
                            <div className="flex flex-wrap items-center gap-2">
                              <h3 className="text-sm font-semibold text-primary">{meta.name}</h3>
                              <Badge variant="success" className="capitalize">connected</Badge>
                            </div>

                            <p className="max-w-2xl text-sm leading-relaxed text-secondary">{meta.desc}</p>

                            <div className="flex flex-wrap gap-2">
                              {getCardDetails(integration).map((detail) => (
                                <span
                                  key={`${copyKey}-${detail.label}`}
                                  className="inline-flex items-center gap-2 rounded-full border border-default bg-inset px-3 py-1.5 text-xs text-secondary"
                                >
                                  <span className="font-semibold text-primary">{detail.label}:</span>
                                  <span className="max-w-[18rem] truncate">{detail.value}</span>
                                </span>
                              ))}
                            </div>

                            {integration.last_error && (
                              <div className="flex items-start gap-2 rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-xs text-red-700">
                                <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                                <p className="leading-relaxed">{integration.last_error}</p>
                              </div>
                            )}
                          </div>
                        </div>

                        <div className="flex shrink-0 flex-col gap-2 sm:flex-row xl:flex-col">
                          {webhookUrl && (
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => handleCopyWebhook(integration, webhookUrl)}
                              className="min-w-[132px] justify-center"
                            >
                              {copiedWebhookId === copyKey ? (
                                <>
                                  <Check className="h-3.5 w-3.5" />
                                  Copied
                                </>
                              ) : (
                                <>
                                  <Copy className="h-3.5 w-3.5" />
                                  Copy webhook
                                </>
                              )}
                            </Button>
                          )}

                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDisconnectClick(integration.channel)}
                            className="min-w-[132px] justify-center text-red-600 hover:border-red-200 hover:text-red-700"
                          >
                            <Unplug className="h-3.5 w-3.5" />
                            Disconnect
                          </Button>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </CardBody>
          </Card>
        </div>
      )}

      <ConfirmModal
        open={showDisconnect}
        onClose={closeDisconnect}
        onConfirm={handleDisconnectConfirm}
        title="Disconnect channel?"
        description={`Your ${disconnectChannel} integration will be removed. You can reconnect anytime.`}
        variant="warning"
        confirmText="Disconnect"
        loading={disconnectLoading}
      />

      <WhatsAppModal
        open={activeModal === 'whatsapp'}
        onClose={() => setActiveModal(null)}
        loading={connectLoading}
      />

      <TelegramModal
        open={activeModal === 'telegram'}
        onClose={() => setActiveModal(null)}
        onConnect={(config) => handleConnectSubmit('telegram', config)}
        loading={connectLoading}
      />

      <WebWidgetModal
        open={activeModal === 'web'}
        onClose={() => setActiveModal(null)}
        onConnect={(config) => handleConnectSubmit('web', config)}
        loading={connectLoading}
        isConnected={widgetConfig?.is_active ?? integrationMap.has('web')}
        existingConfig={
          widgetConfig
            ? {
                botName: widgetConfig.bot_name,
                greeting: widgetConfig.greeting,
                brandColor: widgetConfig.brand_color,
                position: widgetConfig.position === 'bottom-left' ? 'left' : 'right',
              }
            : null
        }
      />

      <GmailModal
        open={activeModal === 'gmail'}
        onClose={() => setActiveModal(null)}
        onConnect={(config) => handleConnectSubmit('gmail', config)}
        loading={connectLoading}
      />
    </div>
  )
}