import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import { TypingIndicator } from '@/components/chat/TypingIndicator'

const mockSubscribe = vi.fn()

vi.mock('@/hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    subscribe: mockSubscribe,
  }),
}))

vi.mock('@/components/chat/ConversationLoading', () => ({
  ConversationLoading: (props: any) => (
    <div data-testid="conversation-loading" data-size={props.size}>
      Loading
    </div>
  ),
}))

describe('TypingIndicator', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockSubscribe.mockReturnValue(() => {})
  })

  it('returns null when not typing', () => {
    const { container } = render(<TypingIndicator conversationId="conv-1" />)
    expect(container.firstChild).toBeNull()
  })

  it('renders typing UI when typing message is received', () => {
    let messageHandler: (msg: any) => void = () => {}
    mockSubscribe.mockImplementation((handler) => {
      messageHandler = handler
      return () => {}
    })

    render(<TypingIndicator conversationId="conv-1" />)

    act(() => {
      messageHandler({
        type: 'typing_indicator',
        data: { conversation_id: 'conv-1', is_typing: true },
      })
    })

    expect(screen.getByText('AI is thinking...')).toBeInTheDocument()
    expect(screen.getByTestId('conversation-loading')).toBeInTheDocument()
  })

  it('hides when is_typing becomes false', () => {
    let messageHandler: (msg: any) => void = () => {}
    mockSubscribe.mockImplementation((handler) => {
      messageHandler = handler
      return () => {}
    })

    render(<TypingIndicator conversationId="conv-1" />)

    act(() => {
      messageHandler({
        type: 'typing_indicator',
        data: { conversation_id: 'conv-1', is_typing: true },
      })
    })
    expect(screen.getByText('AI is thinking...')).toBeInTheDocument()

    act(() => {
      messageHandler({
        type: 'typing_indicator',
        data: { conversation_id: 'conv-1', is_typing: false },
      })
    })
    expect(screen.queryByText('AI is thinking...')).not.toBeInTheDocument()
  })

  it('ignores messages for other conversations', () => {
    let messageHandler: (msg: any) => void = () => {}
    mockSubscribe.mockImplementation((handler) => {
      messageHandler = handler
      return () => {}
    })

    render(<TypingIndicator conversationId="conv-1" />)

    act(() => {
      messageHandler({
        type: 'typing_indicator',
        data: { conversation_id: 'conv-2', is_typing: true },
      })
    })
    expect(screen.queryByText('AI is thinking...')).not.toBeInTheDocument()
  })

  it('subscribes to websocket when conversationId is provided', () => {
    render(<TypingIndicator conversationId="conv-1" />)
    expect(mockSubscribe).toHaveBeenCalled()
  })

  it('does not subscribe when conversationId is null', () => {
    render(<TypingIndicator conversationId={null} />)
    expect(mockSubscribe).not.toHaveBeenCalled()
  })

  it('passes sm size to ConversationLoading', () => {
    let messageHandler: (msg: any) => void = () => {}
    mockSubscribe.mockImplementation((handler) => {
      messageHandler = handler
      return () => {}
    })

    render(<TypingIndicator conversationId="conv-1" />)

    act(() => {
      messageHandler({
        type: 'typing_indicator',
        data: { conversation_id: 'conv-1', is_typing: true },
      })
    })

    const loading = screen.getByTestId('conversation-loading')
    expect(loading).toHaveAttribute('data-size', 'sm')
  })
})
