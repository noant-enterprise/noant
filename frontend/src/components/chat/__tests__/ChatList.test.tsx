import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ChatList } from '@/components/chat/ChatList'
import type { Conversation } from '@/types'

vi.mock('react-router-dom', () => ({
  useNavigate: () => vi.fn(),
}))

vi.mock('@/hooks/useInfiniteScroll', () => ({
  useInfiniteScroll: () => ({
    setSentinel: vi.fn(),
  }),
}))

const mockConversations: Conversation[] = [
  {
    id: '1',
    customer_name: 'Alice Johnson',
    channel: 'whatsapp',
    status: 'active',
    is_ai_transferred: false,
    last_message: 'Hello, I need help',
    unread: 2,
    priority: 'medium',
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-15T10:05:00Z',
  },
  {
    id: '2',
    customer_name: 'Bob Smith',
    channel: 'telegram',
    status: 'resolved',
    is_ai_transferred: false,
    last_message: 'Thanks for the help!',
    unread: 0,
    priority: 'low',
    created_at: '2024-01-14T10:00:00Z',
    updated_at: '2024-01-14T10:10:00Z',
  },
]

describe('ChatList', () => {
  const defaultProps = {
    conversations: mockConversations,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders conversation list', () => {
    render(<ChatList {...defaultProps} />)
    expect(screen.getByText('Alice Johnson')).toBeInTheDocument()
    expect(screen.getByText('Bob Smith')).toBeInTheDocument()
  })

  it('renders last message preview', () => {
    render(<ChatList {...defaultProps} />)
    expect(screen.getByText('Hello, I need help')).toBeInTheDocument()
    expect(screen.getByText('Thanks for the help!')).toBeInTheDocument()
  })

  it('shows empty state when no conversations', () => {
    render(<ChatList conversations={[]} />)
    expect(screen.getByText('No conversations found')).toBeInTheDocument()
  })

  it('shows unread badge for conversations with unread messages', () => {
    render(<ChatList {...defaultProps} />)
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('does not show unread badge when unread is 0', () => {
    render(<ChatList conversations={[mockConversations[1]]} />)
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  it('renders the search input', () => {
    render(<ChatList {...defaultProps} />)
    expect(screen.getByPlaceholderText('Search conversations...')).toBeInTheDocument()
  })

  it('filters conversations by search term', async () => {
    const user = userEvent.setup()
    render(<ChatList {...defaultProps} />)

    await user.type(screen.getByPlaceholderText('Search conversations...'), 'Alice')

    expect(screen.getByText('Alice Johnson')).toBeInTheDocument()
    expect(screen.queryByText('Bob Smith')).not.toBeInTheDocument()
  })

  it('renders filter buttons', () => {
    render(<ChatList {...defaultProps} />)
    expect(screen.getByRole('button', { name: /all/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /active/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /escalated/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /resolved/i })).toBeInTheDocument()
  })

  it('shows clear all button when onClearAll is provided and conversations exist', () => {
    render(<ChatList {...defaultProps} onClearAll={vi.fn()} />)
    expect(screen.getByTitle('Clear all chats')).toBeInTheDocument()
  })

  it('does not show clear all button without onClearAll', () => {
    render(<ChatList {...defaultProps} />)
    expect(screen.queryByTitle('Clear all chats')).not.toBeInTheDocument()
  })

  it('shows empty state message when filtered results are empty', async () => {
    const user = userEvent.setup()
    render(<ChatList {...defaultProps} />)

    await user.type(screen.getByPlaceholderText('Search conversations...'), 'zzzznotfound')
    expect(screen.getByText('No conversations found')).toBeInTheDocument()
  })

  it('shows "No messages yet" for conversations without last_message', () => {
    const convs = [{ ...mockConversations[0], last_message: undefined }]
    render(<ChatList conversations={convs} />)
    expect(screen.getByText('No messages yet')).toBeInTheDocument()
  })

  it('renders the Messages header', () => {
    render(<ChatList {...defaultProps} />)
    expect(screen.getByText('Messages')).toBeInTheDocument()
  })
})
