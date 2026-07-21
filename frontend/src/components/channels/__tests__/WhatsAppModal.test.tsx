import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WhatsAppModal } from '@/components/channels/WhatsAppModal'

vi.mock('@/components/ui/Modal', () => ({
  Modal: ({ open, onClose, title, children }: any) =>
    open ? (
      <div role="dialog" aria-label={title}>
        <button onClick={onClose} aria-label="Close">×</button>
        {children}
      </div>
    ) : null,
}))

vi.mock('@/components/ui/Button', () => ({
  Button: ({ children, onClick, disabled, type, loading, variant, ...props }: any) => (
    <button
      onClick={onClick}
      disabled={disabled || loading}
      type={type}
      data-variant={variant}
      {...props}
    >
      {children}
    </button>
  ),
}))

vi.mock('@/components/ui/Input', () => ({
  Input: ({ value, onChange, placeholder, disabled, type, required }: any) => (
    <input
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      disabled={disabled}
      type={type}
      required={required}
    />
  ),
}))

vi.mock('@/lib/api', () => ({
  api: {
    post: vi.fn(),
    get: vi.fn(),
  },
}))

vi.mock('@/hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    subscribe: vi.fn(() => vi.fn()),
  }),
}))

vi.mock('@/hooks/useOffline', () => ({
  useOffline: () => false,
}))

vi.mock('@/components/chat', () => ({
  ConversationLoading: ({ size }: any) => (
    <div data-testid="conversation-loading" data-size={size}>
      Loading...
    </div>
  ),
}))

const defaultProps = {
  open: true,
  onClose: vi.fn(),
  loading: false,
  onConnect: vi.fn(),
}

describe('WhatsAppModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders when open', () => {
    render(<WhatsAppModal {...defaultProps} />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('Connect WhatsApp')).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    const { container } = render(<WhatsAppModal {...defaultProps} open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders phone number input', () => {
    render(<WhatsAppModal {...defaultProps} />)
    expect(screen.getByPlaceholderText('+44 7700 900123')).toBeInTheDocument()
  })

  it('renders connection name input', () => {
    render(<WhatsAppModal {...defaultProps} />)
    expect(screen.getByPlaceholderText('e.g. Business Line, Support Number')).toBeInTheDocument()
  })

  it('renders instructions', () => {
    render(<WhatsAppModal {...defaultProps} />)
    expect(screen.getByText('OpenWA self-hosted WhatsApp')).toBeInTheDocument()
    expect(screen.getByText('Enter your business WhatsApp number')).toBeInTheDocument()
  })

  it('renders Connect WhatsApp button', () => {
    render(<WhatsAppModal {...defaultProps} />)
    expect(screen.getByText('Connect WhatsApp')).toBeInTheDocument()
  })

  it('renders Cancel button', () => {
    render(<WhatsAppModal {...defaultProps} />)
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('renders Test Server button', () => {
    render(<WhatsAppModal {...defaultProps} />)
    expect(screen.getByText('Test Server')).toBeInTheDocument()
  })

  it('calls onClose when Cancel is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<WhatsAppModal {...defaultProps} onClose={onClose} />)
    await user.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('shows validation error for number starting with 0', async () => {
    const user = userEvent.setup()
    render(<WhatsAppModal {...defaultProps} />)
    await user.type(screen.getByPlaceholderText('+44 7700 900123'), '07700900123')
    expect(screen.getByText(/Please replace the leading 0/)).toBeInTheDocument()
  })

  it('shows validation error for number without + prefix', async () => {
    const user = userEvent.setup()
    render(<WhatsAppModal {...defaultProps} />)
    await user.type(screen.getByPlaceholderText('+44 7700 900123'), '447700900123')
    expect(screen.getByText(/Phone number must start with \+/)).toBeInTheDocument()
  })

  it('shows validation error for number with non-digit characters after +', async () => {
    const user = userEvent.setup()
    render(<WhatsAppModal {...defaultProps} />)
    await user.type(screen.getByPlaceholderText('+44 7700 900123'), '+44abc')
    expect(screen.getByText(/Phone number must contain only numbers/)).toBeInTheDocument()
  })

  it('shows validation error for too short number', async () => {
    const user = userEvent.setup()
    render(<WhatsAppModal {...defaultProps} />)
    await user.type(screen.getByPlaceholderText('+44 7700 900123'), '+1234')
    expect(screen.getByText(/valid phone number/)).toBeInTheDocument()
  })

  it('clears validation when valid number entered', async () => {
    const user = userEvent.setup()
    render(<WhatsAppModal {...defaultProps} />)
    await user.type(screen.getByPlaceholderText('+44 7700 900123'), '0')
    expect(screen.getByText(/Please replace/)).toBeInTheDocument()
    await user.clear(screen.getByPlaceholderText('+44 7700 900123'))
    await user.type(screen.getByPlaceholderText('+44 7700 900123'), '+447700900123')
    expect(screen.queryByText(/Please replace/)).not.toBeInTheDocument()
  })

  it('shows loading state when loading prop is true', () => {
    render(<WhatsAppModal {...defaultProps} loading />)
    const phoneInput = screen.getByPlaceholderText('+44 7700 900123')
    expect(phoneInput).toBeDisabled()
  })

  it('renders modal title', () => {
    render(<WhatsAppModal {...defaultProps} />)
    expect(screen.getByText('Connect WhatsApp')).toBeInTheDocument()
  })
})
