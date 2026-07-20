import { describe, it, expect, vi } from 'vitest'
import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MobileOverlay } from '@/components/layout/MobileOverlay'

vi.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('MobileOverlay', () => {
  it('renders the overlay element', () => {
    render(<MobileOverlay open={true} onClose={vi.fn()} />)
    const overlay = document.querySelector('.fixed.inset-0')
    expect(overlay).toBeInTheDocument()
  })

  it('is visible when open is true', () => {
    render(<MobileOverlay open={true} onClose={vi.fn()} />)
    const overlay = document.querySelector('.fixed.inset-0') as HTMLElement
    expect(overlay.className).toContain('opacity-100')
    expect(overlay.className).toContain('pointer-events-auto')
  })

  it('is hidden when open is false', () => {
    render(<MobileOverlay open={false} onClose={vi.fn()} />)
    const overlay = document.querySelector('.fixed.inset-0') as HTMLElement
    expect(overlay.className).toContain('opacity-0')
    expect(overlay.className).toContain('pointer-events-none')
  })

  it('calls onClose when clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()
    render(<MobileOverlay open={true} onClose={onClose} />)
    const overlay = document.querySelector('.fixed.inset-0') as HTMLElement
    await user.click(overlay)
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('has z-index 40', () => {
    render(<MobileOverlay open={true} onClose={vi.fn()} />)
    const overlay = document.querySelector('.fixed.inset-0') as HTMLElement
    expect(overlay.className).toContain('z-40')
  })

  it('is hidden on lg screens', () => {
    render(<MobileOverlay open={true} onClose={vi.fn()} />)
    const overlay = document.querySelector('.fixed.inset-0') as HTMLElement
    expect(overlay.className).toContain('lg:hidden')
  })
})
