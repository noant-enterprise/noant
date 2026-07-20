import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CustomerInfo } from '@/components/chat/CustomerInfo'
import type { Conversation } from '@/types'

vi.mock('@/components/ui/Avatar', () => ({
  Avatar: ({ name, size }: { name: string; size?: string }) => (
    <div data-testid="avatar" data-name={name} data-size={size}>
      {name}
    </div>
  ),
}))

const mockConversation: Conversation = {
  id: 'conv-1',
  customer_name: 'Alice Johnson',
  channel: 'whatsapp',
  status: 'active',
  is_ai_transferred: false,
  last_message: 'Hello',
  unread: 0,
  intent: 'support',
  priority: 'high',
  created_at: '2024-01-15T10:00:00Z',
  updated_at: '2024-01-15T10:05:00Z',
}

describe('CustomerInfo', () => {
  const defaultProps = {
    conversation: mockConversation,
    open: true,
    onClose: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns null when open is false', () => {
    const { container } = render(
      <CustomerInfo conversation={mockConversation} open={false} onClose={vi.fn()} />
    )
    expect(container.innerHTML).toBe('')
  })

  it('renders customer name', () => {
    render(<CustomerInfo {...defaultProps} />)
    const names = screen.getAllByText('Alice Johnson')
    expect(names.length).toBeGreaterThan(0)
  })

  it('renders channel type', () => {
    render(<CustomerInfo {...defaultProps} />)
    const channels = screen.getAllByText('whatsapp')
    expect(channels.length).toBeGreaterThan(0)
  })

  it('renders status info', () => {
    render(<CustomerInfo {...defaultProps} />)
    expect(screen.getAllByText('active').length).toBeGreaterThan(0)
  })

  it('renders priority info', () => {
    render(<CustomerInfo {...defaultProps} />)
    expect(screen.getAllByText('high').length).toBeGreaterThan(0)
  })

  it('renders intent info', () => {
    render(<CustomerInfo {...defaultProps} />)
    expect(screen.getAllByText('support').length).toBeGreaterThan(0)
  })

  it('shows "Unknown" when intent is not provided', () => {
    const conv = { ...mockConversation, intent: undefined }
    render(<CustomerInfo conversation={conv} open onClose={vi.fn()} />)
    expect(screen.getAllByText('Unknown').length).toBeGreaterThan(0)
  })

  it('calls onClose when close button is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<CustomerInfo {...defaultProps} onClose={onClose} />)

    const closeBtn = screen.getByRole('button', { name: /close customer info/i })
    await user.click(closeBtn)
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('renders the Customer header label', () => {
    render(<CustomerInfo {...defaultProps} />)
    expect(screen.getByText('Customer')).toBeInTheDocument()
  })

  it('shows placeholder when conversation is null but open', () => {
    render(<CustomerInfo conversation={null} open onClose={vi.fn()} />)
    const placeholders = screen.getAllByText('Select a conversation to see details')
    expect(placeholders.length).toBeGreaterThan(0)
  })

  it('renders avatar with customer name', () => {
    render(<CustomerInfo {...defaultProps} />)
    const avatars = screen.getAllByTestId('avatar')
    const matching = avatars.filter((a) => a.getAttribute('data-name') === 'Alice Johnson')
    expect(matching.length).toBeGreaterThan(0)
  })
})
