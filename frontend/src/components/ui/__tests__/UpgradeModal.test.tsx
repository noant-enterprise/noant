import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { UpgradeModal } from '@/components/ui/UpgradeModal'

describe('UpgradeModal', () => {
  it('renders when open', () => {
    render(<UpgradeModal open={true} onClose={vi.fn()} />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('returns null when closed', () => {
    const { container } = render(<UpgradeModal open={false} onClose={vi.fn()} />)
    expect(container.innerHTML).toBe('')
  })

  it('shows default title and description', () => {
    render(<UpgradeModal open={true} onClose={vi.fn()} />)
    expect(screen.getByText(/upgrade to unlock/i)).toBeInTheDocument()
    expect(screen.getByText(/free plan/i)).toBeInTheDocument()
  })

  it('shows custom title and description', () => {
    render(
      <UpgradeModal
        open={true}
        onClose={vi.fn()}
        title="Custom Title"
        description="Custom Description"
      />
    )
    expect(screen.getByText('Custom Title')).toBeInTheDocument()
    expect(screen.getByText('Custom Description')).toBeInTheDocument()
  })

  it('renders feature list items', () => {
    render(<UpgradeModal open={true} onClose={vi.fn()} />)
    expect(screen.getByText(/unlimited ai responses/i)).toBeInTheDocument()
    expect(screen.getByText(/priority support/i)).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<UpgradeModal open={true} onClose={onClose} />)
    await user.click(screen.getByRole('button', { name: /close/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('calls onClose when Escape is pressed', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<UpgradeModal open={true} onClose={onClose} />)
    await user.keyboard('{Escape}')
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('calls onClose when Maybe Later is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<UpgradeModal open={true} onClose={onClose} />)
    await user.click(screen.getByRole('button', { name: /maybe later/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })
})
