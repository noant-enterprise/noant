import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ChatInput } from '@/components/chat/ChatInput'

vi.mock('@/hooks/useOffline', () => ({
  useOffline: () => false,
}))

describe('ChatInput', () => {
  const defaultProps = {
    onSend: vi.fn(),
    onTakeover: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the text input', () => {
    render(<ChatInput {...defaultProps} />)
    expect(screen.getByPlaceholderText('Message...')).toBeInTheDocument()
  })

  it('renders the send button as submit', () => {
    const { container } = render(<ChatInput {...defaultProps} />)
    const sendBtn = container.querySelector('button[type="submit"]')
    expect(sendBtn).toBeInTheDocument()
  })

  it('renders the take over button', () => {
    render(<ChatInput {...defaultProps} />)
    expect(screen.getByText('Take over')).toBeInTheDocument()
  })

  it('calls onSend with trimmed text on submit', async () => {
    const onSend = vi.fn()
    const user = userEvent.setup()
    render(<ChatInput {...defaultProps} onSend={onSend} />)

    const input = screen.getByPlaceholderText('Message...')
    await user.type(input, 'Hello world')
    await user.keyboard('{Enter}')

    expect(onSend).toHaveBeenCalledWith('Hello world')
  })

  it('clears input after sending', async () => {
    const user = userEvent.setup()
    render(<ChatInput {...defaultProps} />)

    const input = screen.getByPlaceholderText('Message...')
    await user.type(input, 'Test message')
    await user.keyboard('{Enter}')

    expect(input).toHaveValue('')
  })

  it('does not call onSend for empty/whitespace text', async () => {
    const onSend = vi.fn()
    const user = userEvent.setup()
    render(<ChatInput {...defaultProps} onSend={onSend} />)

    await user.type(screen.getByPlaceholderText('Message...'), '   ')
    await user.keyboard('{Enter}')

    expect(onSend).not.toHaveBeenCalled()
  })

  it('disables input when disabled prop is true', () => {
    render(<ChatInput {...defaultProps} disabled />)
    expect(screen.getByPlaceholderText('Message...')).toBeDisabled()
  })

  it('disables send button when disabled prop is true', () => {
    const { container } = render(<ChatInput {...defaultProps} disabled />)
    const sendBtn = container.querySelector('button[type="submit"]')
    expect(sendBtn).toBeDisabled()
  })

  it('disables take over button when disabled prop is true', () => {
    render(<ChatInput {...defaultProps} disabled />)
    expect(screen.getByText('Take over')).toBeDisabled()
  })

  it('calls onTakeover when take over button is clicked', async () => {
    const onTakeover = vi.fn()
    const user = userEvent.setup()
    render(<ChatInput {...defaultProps} onTakeover={onTakeover} />)

    await user.click(screen.getByText('Take over'))
    expect(onTakeover).toHaveBeenCalledOnce()
  })
})
