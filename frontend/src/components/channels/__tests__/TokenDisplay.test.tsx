import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TokenDisplay } from '@/components/channels/TokenDisplay'

describe('TokenDisplay', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders the label', () => {
    render(<TokenDisplay token="abc123" />)
    expect(screen.getByText('Token')).toBeInTheDocument()
  })

  it('renders custom label', () => {
    render(<TokenDisplay token="abc123" label="API Key" />)
    expect(screen.getByText('API Key')).toBeInTheDocument()
  })

  it('renders masked token by default', () => {
    const { container } = render(<TokenDisplay token="abc123" />)
    const display = container.querySelector('.font-mono')
    expect(display?.textContent).toContain('•')
  })

  it('shows actual token when eye button is clicked', async () => {
    const user = userEvent.setup()
    render(<TokenDisplay token="abc123secret" />)

    await user.click(screen.getByTitle('Show token'))
    expect(screen.getByText('abc123secret')).toBeInTheDocument()
  })

  it('hides token when eye button clicked again', async () => {
    const user = userEvent.setup()
    render(<TokenDisplay token="abc123secret" />)

    await user.click(screen.getByTitle('Show token'))
    expect(screen.getByText('abc123secret')).toBeInTheDocument()

    await user.click(screen.getByTitle('Hide token'))
    const { container } = render(<TokenDisplay token="abc123secret" />)
    const display = container.querySelector('.font-mono')
    expect(display?.textContent).toContain('•')
  })

  it('has a copy button', () => {
    render(<TokenDisplay token="abc123" />)
    expect(screen.getByTitle('Copy token')).toBeInTheDocument()
  })

  it('shows "Not configured" when no token is provided', () => {
    render(<TokenDisplay />)
    expect(screen.getByText('Not configured')).toBeInTheDocument()
  })

  it('does not show toggle/copy buttons when no token', () => {
    render(<TokenDisplay />)
    expect(screen.queryByTitle('Show token')).not.toBeInTheDocument()
    expect(screen.queryByTitle('Copy token')).not.toBeInTheDocument()
  })
})
