import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { GmailModal } from '@/components/channels/GmailModal'

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

describe('GmailModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders when open', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('Connect Gmail')).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    const { container } = render(<GmailModal {...defaultProps} open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders email input', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByPlaceholderText('yourbusiness@gmail.com')).toBeInTheDocument()
  })

  it('renders app password input', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByPlaceholderText('xxxx xxxx xxxx xxxx')).toBeInTheDocument()
  })

  it('renders setup instructions', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByText(/Email Customer Support via IMAP\/SMTP/)).toBeInTheDocument()
  })

  it('renders Connect Gmail button', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByText('Connect Gmail')).toBeInTheDocument()
  })

  it('renders Test Connection button', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByText('Test Connection')).toBeInTheDocument()
  })

  it('renders Cancel button', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('calls onClose when Cancel is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<GmailModal {...defaultProps} onClose={onClose} />)
    await user.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('submits form with email and app password', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()
    render(<GmailModal {...defaultProps} onConnect={onConnect} />)

    await user.type(screen.getByPlaceholderText('yourbusiness@gmail.com'), 'test@gmail.com')
    await user.type(screen.getByPlaceholderText('xxxx xxxx xxxx xxxx'), 'abcd1234efgh5678')
    await user.click(screen.getByText('Connect Gmail'))

    expect(onConnect).toHaveBeenCalledWith({
      email: 'test@gmail.com',
      app_password: 'abcd1234efgh5678',
    })
  })

  it('does not submit when fields are empty', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()
    render(<GmailModal {...defaultProps} onConnect={onConnect} />)
    await user.click(screen.getByText('Connect Gmail'))
    expect(onConnect).not.toHaveBeenCalled()
  })

  it('disables inputs when loading', () => {
    render(<GmailModal {...defaultProps} loading />)
    expect(screen.getByPlaceholderText('yourbusiness@gmail.com')).toBeDisabled()
    expect(screen.getByPlaceholderText('xxxx xxxx xxxx xxxx')).toBeDisabled()
  })

  it('disables Test Connection when fields are empty', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByText('Test Connection')).toBeDisabled()
  })

  it('shows app password info text', () => {
    render(<GmailModal {...defaultProps} />)
    expect(screen.getByText(/16-character App Password/)).toBeInTheDocument()
  })
})
