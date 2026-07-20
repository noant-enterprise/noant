import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ConfirmModal } from '@/components/ui/ConfirmModal'

function Wrapper({
  open = true,
  onClose = vi.fn(),
  onConfirm = vi.fn(),
  title = 'Confirm Action',
  description,
  confirmText,
  cancelText,
  variant,
  loading,
  requireTypeConfirm,
  confirmPhrase,
  closeOnOverlayClick,
}: {
  open?: boolean
  onClose?: () => void
  onConfirm?: () => void
  title?: string
  description?: string
  confirmText?: string
  cancelText?: string
  variant?: 'danger' | 'warning' | 'success' | 'neutral'
  loading?: boolean
  requireTypeConfirm?: boolean
  confirmPhrase?: string
  closeOnOverlayClick?: boolean
} = {}) {
  return (
    <ConfirmModal
      open={open}
      onClose={onClose}
      onConfirm={onConfirm}
      title={title}
      description={description}
      confirmText={confirmText}
      cancelText={cancelText}
      variant={variant}
      loading={loading}
      requireTypeConfirm={requireTypeConfirm}
      confirmPhrase={confirmPhrase}
      closeOnOverlayClick={closeOnOverlayClick}
    />
  )
}

describe('ConfirmModal', () => {
  it('renders when open', () => {
    render(<Wrapper />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('Confirm Action')).toBeInTheDocument()
  })

  it('returns null when closed', () => {
    const { container } = render(<Wrapper open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('shows title and description', () => {
    render(<Wrapper title="Delete?" description="This cannot be undone" />)
    expect(screen.getByText('Delete?')).toBeInTheDocument()
    expect(screen.getByText('This cannot be undone')).toBeInTheDocument()
  })

  it('shows confirm and cancel buttons with default text', () => {
    render(<Wrapper />)
    expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('shows custom confirm and cancel text', () => {
    render(<Wrapper confirmText="Yes, do it" cancelText="Nope" />)
    expect(screen.getByRole('button', { name: /yes, do it/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /nope/i })).toBeInTheDocument()
  })

  it('calls onConfirm when confirm button is clicked', async () => {
    const onConfirm = vi.fn()
    const user = userEvent.setup()
    render(<Wrapper onConfirm={onConfirm} />)
    await user.click(screen.getByRole('button', { name: /confirm/i }))
    expect(onConfirm).toHaveBeenCalledOnce()
  })

  it('calls onClose when cancel button is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<Wrapper onClose={onClose} />)
    await user.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('calls onClose when close X button is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<Wrapper onClose={onClose} />)
    await user.click(screen.getByRole('button', { name: /close/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('renders danger variant icon', () => {
    render(<Wrapper variant="danger" />)
    const dialog = screen.getByRole('dialog')
    expect(dialog.querySelector('.text-red-500')).toBeInTheDocument()
  })

  it('renders with type-to-confirm', () => {
    render(<Wrapper requireTypeConfirm={true} confirmPhrase="DELETE" />)
    expect(screen.getByText('DELETE', { exact: false })).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: /type delete to confirm/i })).toBeInTheDocument()
  })

  it('confirm button is disabled until correct phrase is typed', async () => {
    const onConfirm = vi.fn()
    const user = userEvent.setup()
    render(
      <Wrapper
        requireTypeConfirm={true}
        confirmPhrase="RESET"
        onConfirm={onConfirm}
      />
    )
    const confirmBtn = screen.getByRole('button', { name: /confirm/i })
    expect(confirmBtn).toBeDisabled()

    await user.type(screen.getByRole('textbox'), 'WRONG')
    expect(confirmBtn).toBeDisabled()

    await user.clear(screen.getByRole('textbox'))
    await user.type(screen.getByRole('textbox'), 'RESET')
    expect(confirmBtn).not.toBeDisabled()

    await user.click(confirmBtn)
    expect(onConfirm).toHaveBeenCalledOnce()
  })

  it('disables confirm button while loading', () => {
    render(<Wrapper loading />)
    expect(screen.getByRole('button', { name: /please wait/i })).toBeDisabled()
  })

  it('disables cancel button while loading', () => {
    render(<Wrapper loading />)
    expect(screen.getByRole('button', { name: /cancel/i })).toBeDisabled()
  })
})
