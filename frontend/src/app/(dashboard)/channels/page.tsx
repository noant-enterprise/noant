import { useEffect, useState } from 'react'
import { useAPI } from '@/hooks/useAPI'
import {
  ChannelCard,
  WhatsAppModal,
  TelegramModal,
  FacebookModal,
  InstagramModal,
  WebWidgetModal
} from '@/components/channels'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import { useModal } from '@/hooks/useModal'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useWidgetConfig } from '@/contexts/WidgetConfigContext'

interface Integration {
  id?: string
  channel: string
  status: 'connected' | 'disconnected' | 'error'
  token?: string
  webhook_url?: string
  connected_at?: string
  config?: any
}

const channelConfig: Record<string, { name: string; desc: string }> = {
  telegram: { name: 'Telegram', desc: 'Bot integration with webhook' },
  whatsapp: { name: 'WhatsApp', desc: 'Business API messaging' },
  instagram: { name: 'Instagram', desc: 'Direct Messages automation' },
  facebook: { name: 'Facebook', desc: 'Messenger integration' },
  discord: { name: 'Discord', desc: 'Server bot integration' },
  web: { name: 'Web Chat', desc: 'Embeddable widget for your site' },
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

  useEffect(() => {
    getIntegrations('/integrations/list')
  }, [])

  // Listen for real-time channel status updates
  useEffect(() => {
    const unsub = subscribe((msg) => {
      if (msg.type === 'integration_update') {
        getIntegrations('/integrations/list')
      }
    })
    return unsub
  }, [subscribe, getIntegrations])

  const [activeModal, setActiveModal] = useState<'telegram' | 'whatsapp' | 'instagram' | 'facebook' | 'web' | null>(null)
  const [connectLoading, setConnectLoading] = useState(false)

  const handleConnectClick = (channel: string) => {
    if (channel === 'discord') {
      handleConnectSubmit('discord', {})
    } else {
      setActiveModal(channel as any)
    }
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
      // error handled by API hook
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
        setWidgetConfig(c => {
          if (!c) return null
          const updated = { ...c, is_active: false }
          post('/widget/config', updated)
          return updated
        })
      }
      getIntegrations('/integrations/list')
    } catch {
      // error handled by API hook
    } finally {
      setDisconnectLoading(false)
      closeDisconnect()
      setDisconnectChannel('')
    }
  }

  const integrations: Integration[] = data?.integrations || []
  // Only show truly connected channels in the summary table
  const connectedIntegrations = integrations.filter((i: Integration) => i.status === 'connected')
  const integrationMap = new Map<string, Integration>(integrations.map((i: Integration) => [i.channel, i]))

  return (
    <div className="animate-page-in space-y-5 lg:space-y-6 pt-2">
      {/* Channel cards */}
      <div className="px-1">
        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 lg:gap-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="rounded-2xl border border-default bg-surface p-4 lg:p-5 space-y-4">
                <div className="flex items-center gap-3">
                  <Skeleton className="w-10 h-10 rounded-xl" />
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
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 lg:gap-4">
            {Object.entries(channelConfig).map(([key, cfg]) => {
              const integration = integrationMap.get(key)
              return (
                <ChannelCard
                  key={key}
                  channel={key}
                  name={cfg.name}
                  desc={cfg.desc}
                  status={integration?.status || 'disconnected'}
                  token={integration?.token}
                  webhookUrl={integration?.webhook_url}
                  connectedAt={integration?.connected_at}
                  onConnect={() => handleConnectClick(key)}
                  onDisconnect={() => handleDisconnectClick(key)}
                />
              )
            })}
          </div>
        )}
      </div>

      {/* Connected channels list */}
      <div className="px-1 pb-4">
        <Card>
          <CardHeader className="px-4 py-3 lg:px-6 lg:py-4">
            <CardTitle>Connected channels</CardTitle>
          </CardHeader>
          <CardBody className="p-0">
            {loading ? (
              <div className="p-4 space-y-3">
                {Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="flex items-center justify-between py-3">
                    <Skeleton className="h-4 w-32 rounded" />
                    <Skeleton className="h-6 w-20 rounded" />
                  </div>
                ))}
              </div>
            ) : connectedIntegrations.length === 0 ? (
              <EmptyChannels />
            ) : (
              <>
                {/* Desktop table */}
                <div className="hidden lg:block">
                  <table className="w-full">
                    <thead>
                      <tr className="text-left text-[11px] font-semibold uppercase tracking-wider text-tertiary bg-inset">
                        <th className="px-4 py-3 rounded-l-lg">Channel</th>
                        <th className="px-4 py-3">Status</th>
                        <th className="px-4 py-3">Token</th>
                        <th className="px-4 py-3">Connected</th>
                        <th className="px-4 py-3 rounded-r-lg text-right">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {connectedIntegrations.map((i: Integration, idx: number) => (
                        <tr key={`desktop-${i.id || i.channel}-${idx}`} className="border-b border-subtle hover:bg-surface-hover transition-colors">
                          <td className="px-4 py-3">
                            <span className="font-semibold text-sm text-primary capitalize">{i.channel}</span>
                          </td>
                          <td className="px-4 py-3">
                            <Badge variant={i.status === 'connected' ? 'success' : 'error'}>{i.status}</Badge>
                          </td>
                          <td className="px-4 py-3">
                            <span className="font-mono text-xs text-secondary">
                              {i.token ? '••••••••' : '—'}
                            </span>
                          </td>
                          <td className="px-4 py-3 text-sm text-secondary">
                            {i.connected_at ? new Date(i.connected_at).toLocaleDateString() : '-'}
                          </td>
                          <td className="px-4 py-3 text-right">
                            <Button variant="ghost" size="sm" onClick={() => handleDisconnectClick(i.channel)}>
                              Disconnect
                            </Button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* Mobile list */}
                <div className="lg:hidden divide-y divide-subtle">
                  {connectedIntegrations.map((i: Integration, idx: number) => (
                    <div key={`mobile-${i.id || i.channel}-${idx}`} className="flex items-center justify-between p-3">
                      <div className="flex items-center gap-3">
                        <span className="font-semibold text-sm text-primary capitalize">{i.channel}</span>
                        <Badge variant={i.status === 'connected' ? 'success' : 'error'} className="text-[9px]">{i.status}</Badge>
                      </div>
                      <Button variant="ghost" size="sm" className="text-red-600 px-2" onClick={() => handleDisconnectClick(i.channel)}>
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                          <path d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </Button>
                    </div>
                  ))}
                </div>
              </>
            )}
          </CardBody>
        </Card>
      </div>

      {/* Disconnect Confirmation Modal */}
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

      {/* Connection Modals */}
      <WhatsAppModal
        open={activeModal === 'whatsapp'}
        onClose={() => setActiveModal(null)}
        onConnect={(config) => handleConnectSubmit('whatsapp', config)}
        loading={connectLoading}
      />

      <TelegramModal
        open={activeModal === 'telegram'}
        onClose={() => setActiveModal(null)}
        onConnect={(config) => handleConnectSubmit('telegram', config)}
        loading={connectLoading}
      />

      <FacebookModal
        open={activeModal === 'facebook'}
        onClose={() => setActiveModal(null)}
        onConnect={(config) => handleConnectSubmit('facebook', config)}
        loading={connectLoading}
      />

      <InstagramModal
        open={activeModal === 'instagram'}
        onClose={() => setActiveModal(null)}
        onConnect={(config) => handleConnectSubmit('instagram', config)}
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
            : integrationMap.get('web')?.config
        }
      />
    </div>
  )
}

function EmptyChannels() {
  return (
    <div className="flex flex-col items-center justify-center py-12 lg:py-16 text-center px-4">
      <div className="w-14 h-14 lg:w-16 lg:h-16 bg-inset rounded-2xl flex items-center justify-center mb-4 animate-float">
        <svg className="w-7 h-7 lg:w-8 lg:h-8 text-tertiary" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
          <path d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" />
        </svg>
      </div>
      <p className="text-base lg:text-lg font-semibold text-primary mb-1">No channels connected</p>
      <p className="text-sm text-secondary max-w-xs">
        Connect your first channel to start receiving customer messages.
      </p>
    </div>
  )
}