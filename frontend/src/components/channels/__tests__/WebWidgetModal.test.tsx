import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WebWidgetModal } from '@/components/channels/WebWidgetModal'

vi.mock('@/components/ui/Modal', () => ({
  Modal: ({ open, onClose, title, children, size }: any) =>
    open ? (
      <div role="dialog" aria-label={title} data-size={size}>
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

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    user: { id: 'user-123', first_name: 'Test', last_name: 'User' },
  }),
}))

vi.mock('@/components/ui/Toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}))

const defaultProps = {
  open: true,
  onClose: vi.fn(),
  onConnect: vi.fn(),
  loading: false,
  isConnected: false,
  existingConfig: undefined,
}

describe('WebWidgetModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders when open', () => {
    render(<WebWidgetModal {...defaultProps} />)
    expect(screen.getByRole('dialog', { name: 'Configure Web Chat Widget' })).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    const { container } = render(<WebWidgetModal {...defaultProps} open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders bot name input with default value', () => {
    render(<WebWidgetModal {...defaultProps} />)
    const input = screen.getByPlaceholderText('Noant AI')
    expect(input).toHaveValue('Noant AI')
  })

  it('renders greeting input with default value', () => {
    render(<WebWidgetModal {...defaultProps} />)
    const input = screen.getByPlaceholderText(/How can I help you/)
    expect(input).toHaveValue('Hi! 👋 How can I help you?')
  })

  it('renders brand color input', () => {
    render(<WebWidgetModal {...defaultProps} />)
    expect(screen.getAllByDisplayValue('#0ea5e9').length).toBeGreaterThanOrEqual(1)
  })

  it('renders position select', () => {
    render(<WebWidgetModal {...defaultProps} />)
    expect(screen.getByText('Bottom Right')).toBeInTheDocument()
    expect(screen.getByText('Bottom Left')).toBeInTheDocument()
  })

  it('renders embed code section', () => {
    render(<WebWidgetModal {...defaultProps} />)
    expect(screen.getByText('Embed HTML Code')).toBeInTheDocument()
  })

  it('renders Copy button', () => {
    render(<WebWidgetModal {...defaultProps} />)
    expect(screen.getByText('Copy')).toBeInTheDocument()
  })

  it('renders Cancel button', () => {
    render(<WebWidgetModal {...defaultProps} />)
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('renders Connect Channel button when not connected', () => {
    render(<WebWidgetModal {...defaultProps} />)
    expect(screen.getByText('Connect Channel')).toBeInTheDocument()
  })

  it('renders Update Widget button when connected', () => {
    render(<WebWidgetModal {...defaultProps} isConnected />)
    expect(screen.getByText('Update Widget')).toBeInTheDocument()
  })

  it('calls onClose when Cancel is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<WebWidgetModal {...defaultProps} onClose={onClose} />)
    await user.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('submits form with widget config', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()
    render(<WebWidgetModal {...defaultProps} onConnect={onConnect} />)

    await user.click(screen.getByText('Connect Channel'))

    expect(onConnect).toHaveBeenCalledWith(
      expect.objectContaining({
        botName: 'Noant AI',
        greeting: 'Hi! 👋 How can I help you?',
        brandColor: '#0ea5e9',
        position: 'right',
      })
    )
  })

  it('allows editing bot name', async () => {
    const user = userEvent.setup()
    render(<WebWidgetModal {...defaultProps} />)
    const input = screen.getByPlaceholderText('Noant AI')
    await user.clear(input)
    await user.type(input, 'My Bot')
    expect(input).toHaveValue('My Bot')
  })

  it('allows editing greeting', async () => {
    const user = userEvent.setup()
    render(<WebWidgetModal {...defaultProps} />)
    const input = screen.getByPlaceholderText(/How can I help you/)
    await user.clear(input)
    await user.type(input, 'Hello!')
    expect(input).toHaveValue('Hello!')
  })

  it('loads existing config values', () => {
    render(
      <WebWidgetModal
        {...defaultProps}
        existingConfig={{
          botName: 'Custom Bot',
          greeting: 'Custom greeting',
          brandColor: '#ff0000',
          position: 'left',
        }}
      />
    )
    expect(screen.getByPlaceholderText('Noant AI')).toHaveValue('Custom Bot')
    expect(screen.getByPlaceholderText(/How can I help you/)).toHaveValue('Custom greeting')
    expect(screen.getAllByDisplayValue('#ff0000').length).toBeGreaterThanOrEqual(1)
  })

  it('opens modal with lg size', () => {
    render(<WebWidgetModal {...defaultProps} />)
    expect(screen.getByRole('dialog')).toHaveAttribute('data-size', 'lg')
  })
})
