import { describe, it, expect, vi, beforeAll, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ChatMessages } from '@/components/chat/ChatMessages'
import type { Message } from '@/types'

vi.mock('@/hooks/useInfiniteScroll', () => ({
  useInfiniteScroll: () => ({
    setSentinel: vi.fn(),
  }),
}))

vi.mock('@/components/chat/ConversationLoading', () => ({
  ConversationLoading: () => <div data-testid="conversation-loading" />,
}))

vi.mock('@/components/chat/TypingIndicator', () => ({
  TypingIndicator: ({ conversationId }: { conversationId: string }) => (
    <div data-testid="typing-indicator" data-conversation-id={conversationId} />
  ),
}))

const customerMessage: Message = {
  id: 'msg-1',
  conversation_id: 'conv-1',
  content: 'Hi, I need help',
  sender_type: 'customer',
  created_at: '2024-01-15T10:00:00Z',
}

const aiMessage: Message = {
  id: 'msg-2',
  conversation_id: 'conv-1',
  content: 'Hello! How can I assist you?',
  sender_type: 'ai',
  metadata: { confidence: 0.95 },
  created_at: '2024-01-15T10:01:00Z',
}

const systemMessage: Message = {
  id: 'msg-3',
  conversation_id: 'conv-1',
  content: 'Conversation transferred to agent',
  sender_type: 'system',
  created_at: '2024-01-15T10:02:00Z',
}

beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn()
})

describe('ChatMessages', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state when isLoading is true', () => {
    render(<ChatMessages messages={[]} isLoading />)
    expect(screen.getByTestId('conversation-loading')).toBeInTheDocument()
  })

  it('shows empty state when no messages', () => {
    render(<ChatMessages messages={[]} />)
    expect(screen.getByText('Start the conversation')).toBeInTheDocument()
  })

  it('shows empty state description', () => {
    render(<ChatMessages messages={[]} />)
    expect(screen.getByText(/Select a customer to view messages/)).toBeInTheDocument()
  })

  it('renders messages', () => {
    render(<ChatMessages messages={[customerMessage, aiMessage]} />)
    expect(screen.getByText('Hi, I need help')).toBeInTheDocument()
    expect(screen.getByText('Hello! How can I assist you?')).toBeInTheDocument()
  })

  it('renders customer messages aligned to the right', () => {
    const { container } = render(<ChatMessages messages={[customerMessage]} />)
    const msgDiv = container.querySelector('.self-end')
    expect(msgDiv).toBeInTheDocument()
  })

  it('renders AI messages aligned to the left', () => {
    const { container } = render(<ChatMessages messages={[aiMessage]} />)
    const msgDiv = container.querySelector('.self-start')
    expect(msgDiv).toBeInTheDocument()
  })

  it('renders system messages centered', () => {
    const { container } = render(<ChatMessages messages={[systemMessage]} />)
    const msgDiv = container.querySelector('.self-center')
    expect(msgDiv).toBeInTheDocument()
  })

  it('renders AI badge for AI messages', () => {
    render(<ChatMessages messages={[aiMessage]} />)
    expect(screen.getByText('AI')).toBeInTheDocument()
  })

  it('shows confidence percentage for AI messages with confidence', () => {
    render(<ChatMessages messages={[aiMessage]} />)
    expect(screen.getByText('95%')).toBeInTheDocument()
  })

  it('does not show source badge for customer messages', () => {
    render(<ChatMessages messages={[customerMessage]} />)
    expect(screen.queryByText('AI')).not.toBeInTheDocument()
    expect(screen.queryByText('Agent')).not.toBeInTheDocument()
  })

  it('renders system message with centered styling', () => {
    const { container } = render(<ChatMessages messages={[systemMessage]} />)
    const msgBubble = container.querySelector('.bg-amber-50')
    expect(msgBubble).toBeInTheDocument()
    expect(msgBubble).toHaveTextContent('Conversation transferred to agent')
  })

  it('renders TypingIndicator when conversationId is provided', () => {
    render(<ChatMessages messages={[aiMessage]} conversationId="conv-1" />)
    expect(screen.getByTestId('typing-indicator')).toBeInTheDocument()
  })

  it('does not render TypingIndicator when conversationId is null', () => {
    render(<ChatMessages messages={[aiMessage]} />)
    expect(screen.queryByTestId('typing-indicator')).not.toBeInTheDocument()
  })

  it('does not show loading and empty state simultaneously', () => {
    render(<ChatMessages messages={[]} isLoading />)
    expect(screen.queryByText('Start the conversation')).not.toBeInTheDocument()
    expect(screen.getByTestId('conversation-loading')).toBeInTheDocument()
  })
})
