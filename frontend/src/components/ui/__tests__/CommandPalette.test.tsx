import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { CommandPalette } from '@/components/ui/CommandPalette'

vi.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

function renderWithRouter(ui: React.ReactNode) {
  return render(<MemoryRouter>{ui}</MemoryRouter>)
}

describe('CommandPalette', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not render initially (closed)', () => {
    renderWithRouter(<CommandPalette />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('opens on Ctrl+K', async () => {
    renderWithRouter(<CommandPalette />)
    await act(async () => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
      )
    })
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Search commands...')).toBeInTheDocument()
  })

  it('closes on Escape', async () => {
    renderWithRouter(<CommandPalette />)
    await act(async () => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
      )
    })
    expect(screen.getByRole('dialog')).toBeInTheDocument()

    await act(async () => {
      const input = screen.getByRole('combobox')
      input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('shows all navigation commands', async () => {
    renderWithRouter(<CommandPalette />)
    await act(async () => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
      )
    })
    expect(screen.getByText('Go to Overview')).toBeInTheDocument()
    expect(screen.getByText('Go to Conversations')).toBeInTheDocument()
    expect(screen.getByText('Go to Teach your Noant')).toBeInTheDocument()
    expect(screen.getByText('Go to Insights')).toBeInTheDocument()
    expect(screen.getByText('Go to Your channels')).toBeInTheDocument()
    expect(screen.getByText('Go to Your setup')).toBeInTheDocument()
    expect(screen.getByText('Sign out')).toBeInTheDocument()
  })

  it('filters commands by search', async () => {
    const user = userEvent.setup()
    renderWithRouter(<CommandPalette />)
    await act(async () => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
      )
    })

    const searchInput = screen.getByRole('combobox')
    await user.type(searchInput, 'conversation')

    expect(screen.getByText('Go to Conversations')).toBeInTheDocument()
    expect(screen.queryByText('Go to Insights')).not.toBeInTheDocument()
  })

  it('shows no commands found for unmatched search', async () => {
    const user = userEvent.setup()
    renderWithRouter(<CommandPalette />)
    await act(async () => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
      )
    })

    const searchInput = screen.getByRole('combobox')
    await user.type(searchInput, 'zzzzzzz')

    expect(screen.getByText('No commands found')).toBeInTheDocument()
  })

  it('has keyboard shortcuts displayed', async () => {
    renderWithRouter(<CommandPalette />)
    await act(async () => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
      )
    })
    expect(screen.getByText('ESC')).toBeInTheDocument()
  })

  it('shows keyboard navigation hints', async () => {
    renderWithRouter(<CommandPalette />)
    await act(async () => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
      )
    })
    expect(screen.getByText(/to navigate/)).toBeInTheDocument()
    expect(screen.getByText(/to select/)).toBeInTheDocument()
    expect(screen.getByText(/to close/)).toBeInTheDocument()
  })

  it('closes on overlay click', async () => {
    renderWithRouter(<CommandPalette />)
    await act(async () => {
      window.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'k', ctrlKey: true })
      )
    })
    const dialog = screen.getByRole('dialog')
    await userEvent.click(dialog)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})
