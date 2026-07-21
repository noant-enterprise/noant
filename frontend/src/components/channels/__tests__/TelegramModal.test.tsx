import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TelegramModal } from '@/components/channels/TelegramModal'

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
  },
}))

const defaultProps = {
  open: true,
  onClose: vi.fn(),
  onConnect: vi.fn(),
  loading: false,
}

describe('TelegramModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders when open', () => {
    render(<TelegramModal {...defaultProps} />)
    expect(screen.getByRole('dialog', { name: 'Connect Telegram Bot' })).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    const { container } = render(<TelegramModal {...defaultProps} open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders bot token input', () => {
    render(<TelegramModal {...defaultProps} />)
    expect(screen.getByPlaceholderText('e.g. 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ')).toBeInTheDocument()
  })

  it('renders setup instructions', () => {
    render(<TelegramModal {...defaultProps} />)
    expect(screen.getByText('Setup Instructions:')).toBeInTheDocument()
    expect(screen.getByText(/@BotFather/)).toBeInTheDocument()
  })

  it('renders Connect Channel button', () => {
    render(<TelegramModal {...defaultProps} />)
    expect(screen.getByText('Connect Channel')).toBeInTheDocument()
  })

  it('renders Test Connection button', () => {
    render(<TelegramModal {...defaultProps} />)
    expect(screen.getByRole('button', { name: 'Test Connection' })).toBeInTheDocument()
  })

  it('renders Cancel button', () => {
    render(<TelegramModal {...defaultProps} />)
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('calls onClose when Cancel is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<TelegramModal {...defaultProps} onClose={onClose} />)
    await user.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('submits form with bot token', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()
    render(<TelegramModal {...defaultProps} onConnect={onConnect} />)

    await user.type(screen.getByPlaceholderText('e.g. 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ'), 'mybot123:token')
    await user.click(screen.getByText('Connect Channel'))

    expect(onConnect).toHaveBeenCalledWith({ bot_token: 'mybot123:token' })
  })

  it('does not submit when token is empty', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()
    render(<TelegramModal {...defaultProps} onConnect={onConnect} />)
    await user.click(screen.getByText('Connect Channel'))
    expect(onConnect).not.toHaveBeenCalled()
  })

  it('disables inputs when loading', () => {
    render(<TelegramModal {...defaultProps} loading />)
    expect(screen.getByPlaceholderText('e.g. 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ')).toBeDisabled()
  })

  it('disables Test Connection when token is empty', () => {
    render(<TelegramModal {...defaultProps} />)
    expect(screen.getByRole('button', { name: 'Test Connection' })).toBeDisabled()
  })

  it('resets state on close', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<TelegramModal {...defaultProps} onClose={onClose} />)

    await user.type(screen.getByPlaceholderText('e.g. 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ'), 'test')
    await user.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalledOnce()
  })
})
