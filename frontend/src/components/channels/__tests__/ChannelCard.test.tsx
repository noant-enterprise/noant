import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ChannelCard } from '@/components/channels/ChannelCard'

vi.mock('@/components/ui/Button', () => ({
  Button: ({ children, onClick, disabled, ...props }: any) => (
    <button onClick={onClick} disabled={disabled} data-variant={props.variant}>
      {children}
    </button>
  ),
}))

vi.mock('@/components/ui/Badge', () => ({
  Badge: ({ children, variant, ...props }: any) => (
    <span data-variant={variant} {...props}>{children}</span>
  ),
}))

vi.mock('@/components/channels/ChannelIcon', () => ({
  ChannelIcon: ({ channel, size }: any) => (
    <div data-testid="channel-icon" data-channel={channel} data-size={size}>
      Icon
    </div>
  ),
}))

const defaultProps = {
  channel: 'whatsapp',
  name: 'WhatsApp',
  desc: 'Connect your WhatsApp Business account',
  status: 'disconnected' as const,
  onConnect: vi.fn(),
  onDisconnect: vi.fn(),
}

describe('ChannelCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders channel name', () => {
    render(<ChannelCard {...defaultProps} />)
    expect(screen.getByText('WhatsApp')).toBeInTheDocument()
  })

  it('renders channel description', () => {
    render(<ChannelCard {...defaultProps} />)
    expect(screen.getByText('Connect your WhatsApp Business account')).toBeInTheDocument()
  })

  it('renders ChannelIcon component', () => {
    render(<ChannelCard {...defaultProps} />)
    const icon = screen.getByTestId('channel-icon')
    expect(icon).toHaveAttribute('data-channel', 'whatsapp')
  })

  it('shows Connect button when disconnected', () => {
    render(<ChannelCard {...defaultProps} />)
    expect(screen.getByText('Connect')).toBeInTheDocument()
  })

  it('shows Disconnect button when connected', () => {
    render(<ChannelCard {...defaultProps} status="connected" />)
    expect(screen.getByText('Disconnect')).toBeInTheDocument()
  })

  it('calls onConnect when Connect button is clicked', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()
    render(<ChannelCard {...defaultProps} onConnect={onConnect} />)

    await user.click(screen.getByText('Connect'))
    expect(onConnect).toHaveBeenCalledOnce()
  })

  it('calls onDisconnect when Disconnect button is clicked', async () => {
    const onDisconnect = vi.fn()
    const user = userEvent.setup()
    render(<ChannelCard {...defaultProps} status="connected" onDisconnect={onDisconnect} />)

    await user.click(screen.getByText('Disconnect'))
    expect(onDisconnect).toHaveBeenCalledOnce()
  })

  it('shows connected badge when status is connected', () => {
    render(<ChannelCard {...defaultProps} status="connected" />)
    expect(screen.getByText('connected')).toBeInTheDocument()
  })

  it('shows connected badge when status is active', () => {
    render(<ChannelCard {...defaultProps} status="active" />)
    expect(screen.getByText('connected')).toBeInTheDocument()
  })

  it('does not show connected badge when disconnected', () => {
    render(<ChannelCard {...defaultProps} />)
    expect(screen.queryByText('connected')).not.toBeInTheDocument()
  })

  it('renders details when connected and details provided', () => {
    render(
      <ChannelCard
        {...defaultProps}
        status="connected"
        details={[{ label: 'Phone', value: '+1234567890' }]}
      />
    )
    expect(screen.getByText('Phone:')).toBeInTheDocument()
    expect(screen.getByText('+1234567890')).toBeInTheDocument()
  })

  it('renders webhook URL when connected with details', () => {
    render(
      <ChannelCard
        {...defaultProps}
        status="connected"
        details={[{ label: 'Phone', value: '+123' }]}
        webhookUrl="https://example.com/webhook"
      />
    )
    expect(screen.getByText('https://example.com/webhook')).toBeInTheDocument()
  })

  it('renders connected date when provided with details', () => {
    render(
      <ChannelCard
        {...defaultProps}
        status="connected"
        details={[{ label: 'Phone', value: '+123' }]}
        connectedAt="2024-01-15T10:00:00Z"
      />
    )
    expect(screen.getByText(/Connected/)).toBeInTheDocument()
  })
})
