import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Modal } from '@/components/ui/Modal'

function Wrapper({
  open = true,
  onClose = vi.fn(),
  title = 'Test Title',
  description,
  children,
  size,
  hideClose,
}: {
  open?: boolean
  onClose?: () => void
  title?: string
  description?: string
  children?: React.ReactNode
  size?: 'sm' | 'md' | 'lg'
  hideClose?: boolean
}) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      description={description}
      size={size}
      hideClose={hideClose}
    >
      {children}
    </Modal>
  )
}

describe('Modal', () => {
  it('renders when open', () => {
    render(<Wrapper open={true} />)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('Test Title')).toBeInTheDocument()
  })

  it('returns null when closed', () => {
    const { container } = render(<Wrapper open={false} />)
    expect(container.innerHTML).toBe('')
  })

  it('shows title and description', () => {
    render(<Wrapper title="My Title" description="My Description" />)
    expect(screen.getByText('My Title')).toBeInTheDocument()
    expect(screen.getByText('My Description')).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<Wrapper onClose={onClose} />)
    await user.click(screen.getByRole('button', { name: /close/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('closes on Escape key', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<Wrapper onClose={onClose} />)
    await user.keyboard('{Escape}')
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('closes on overlay click', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<Wrapper onClose={onClose} />)
    const overlay = document.querySelector('.fixed.inset-0') as HTMLElement
    await user.click(overlay)
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('does not close on dialog content click', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<Wrapper onClose={onClose}>Child content</Wrapper>)
    await user.click(screen.getByText('Child content'))
    expect(onClose).not.toHaveBeenCalled()
  })

  it('applies size classes for sm', () => {
    render(<Wrapper size="sm" />)
    expect(screen.getByRole('dialog').className).toContain('max-w-sm')
  })

  it('applies size classes for md', () => {
    render(<Wrapper size="md" />)
    expect(screen.getByRole('dialog').className).toContain('max-w-md')
  })

  it('applies size classes for lg', () => {
    render(<Wrapper size="lg" />)
    expect(screen.getByRole('dialog').className).toContain('max-w-lg')
  })

  it('hides close button when hideClose=true', () => {
    render(<Wrapper hideClose={true} />)
    expect(screen.queryByRole('button', { name: /close/i })).not.toBeInTheDocument()
  })

  it('shows close button by default', () => {
    render(<Wrapper />)
    expect(screen.getByRole('button', { name: /close/i })).toBeInTheDocument()
  })

  it('renders children', () => {
    render(<Wrapper><p>Modal body</p></Wrapper>)
    expect(screen.getByText('Modal body')).toBeInTheDocument()
  })
})
