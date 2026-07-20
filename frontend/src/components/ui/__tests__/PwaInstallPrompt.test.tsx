import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PwaInstallPrompt } from '@/components/ui/PwaInstallPrompt'

describe('PwaInstallPrompt', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders nothing by default (no beforeinstallprompt event)', () => {
    const { container } = render(<PwaInstallPrompt />)
    expect(container.innerHTML).toBe('')
  })

  it('shows install prompt when beforeinstallprompt event fires', async () => {
    render(<PwaInstallPrompt />)

    act(() => {
      const event = new Event('beforeinstallprompt', { bubbles: true })
      ;(event as any).preventDefault = vi.fn()
      ;(event as any).prompt = vi.fn()
      ;(event as any).userChoice = Promise.resolve({ outcome: 'dismissed' })
      window.dispatchEvent(event)
    })

    expect(screen.getByText('Install Noant')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /install/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /not now/i })).toBeInTheDocument()
  })

  it('hides when Not Now is clicked', async () => {
    const user = userEvent.setup()
    render(<PwaInstallPrompt />)

    act(() => {
      const event = new Event('beforeinstallprompt', { bubbles: true })
      ;(event as any).preventDefault = vi.fn()
      ;(event as any).prompt = vi.fn()
      ;(event as any).userChoice = Promise.resolve({ outcome: 'dismissed' })
      window.dispatchEvent(event)
    })

    expect(screen.getByText('Install Noant')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: /not now/i }))
    expect(screen.queryByText('Install Noant')).not.toBeInTheDocument()
  })

  it('hides when dismiss X button is clicked', async () => {
    const user = userEvent.setup()
    render(<PwaInstallPrompt />)

    act(() => {
      const event = new Event('beforeinstallprompt', { bubbles: true })
      ;(event as any).preventDefault = vi.fn()
      ;(event as any).prompt = vi.fn()
      ;(event as any).userChoice = Promise.resolve({ outcome: 'dismissed' })
      window.dispatchEvent(event)
    })

    const dismissBtn = screen.getAllByRole('button').find(
      (btn) => !btn.textContent?.trim()
    )!
    await user.click(dismissBtn)
    expect(screen.queryByText('Install Noant')).not.toBeInTheDocument()
  })
})
