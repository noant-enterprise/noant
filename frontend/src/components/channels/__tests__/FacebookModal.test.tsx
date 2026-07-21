import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FacebookModal } from '@/components/channels/FacebookModal'

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

describe('FacebookModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders when open', () => {
    render(<FacebookModal {...defaultProps} />)
    expect(screen.getByRole('dialog', { name: 'Connect Facebook Messenger' })).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    const { container } = render(<FacebookModal {...defaultProps} open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('renders Page ID input', () => {
    render(<FacebookModal {...defaultProps} />)
    expect(screen.getByPlaceholderText('e.g. 10928392839283')).toBeInTheDocument()
  })

  it('renders Page Access Token input', () => {
    render(<FacebookModal {...defaultProps} />)
    expect(screen.getByPlaceholderText('EAAG...')).toBeInTheDocument()
  })

  it('renders setup instructions', () => {
    render(<FacebookModal {...defaultProps} />)
    expect(screen.getByText('Setup Instructions:')).toBeInTheDocument()
    expect(screen.getByText(/Facebook Business Page/)).toBeInTheDocument()
  })

  it('renders Connect Channel button', () => {
    render(<FacebookModal {...defaultProps} />)
    expect(screen.getByText('Connect Channel')).toBeInTheDocument()
  })

  it('renders Test Connection button', () => {
    render(<FacebookModal {...defaultProps} />)
    expect(screen.getByRole('button', { name: 'Test Connection' })).toBeInTheDocument()
  })

  it('renders Cancel button', () => {
    render(<FacebookModal {...defaultProps} />)
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('calls onClose when Cancel is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<FacebookModal {...defaultProps} onClose={onClose} />)
    await user.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('submits form with Page ID and token', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()
    render(<FacebookModal {...defaultProps} onConnect={onConnect} />)

    await user.type(screen.getByPlaceholderText('e.g. 10928392839283'), '12345')
    await user.type(screen.getByPlaceholderText('EAAG...'), 'mytoken')
    await user.click(screen.getByText('Connect Channel'))

    expect(onConnect).toHaveBeenCalledWith({
      page_id: '12345',
      page_access_token: 'mytoken',
    })
  })

  it('does not submit when fields are empty', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()
    render(<FacebookModal {...defaultProps} onConnect={onConnect} />)
    await user.click(screen.getByText('Connect Channel'))
    expect(onConnect).not.toHaveBeenCalled()
  })

  it('disables inputs when loading', () => {
    render(<FacebookModal {...defaultProps} loading />)
    expect(screen.getByPlaceholderText('e.g. 10928392839283')).toBeDisabled()
    expect(screen.getByPlaceholderText('EAAG...')).toBeDisabled()
  })

  it('disables Test Connection when fields are empty', () => {
    render(<FacebookModal {...defaultProps} />)
    expect(screen.getByRole('button', { name: 'Test Connection' })).toBeDisabled()
  })
})
