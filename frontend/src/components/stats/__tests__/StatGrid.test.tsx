import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatGrid } from '@/components/stats/StatGrid'

describe('StatGrid', () => {
  it('renders children', () => {
    render(
      <StatGrid>
        <div>Child 1</div>
        <div>Child 2</div>
      </StatGrid>
    )
    expect(screen.getByText('Child 1')).toBeInTheDocument()
    expect(screen.getByText('Child 2')).toBeInTheDocument()
  })

  it('applies grid layout classes', () => {
    const { container } = render(
      <StatGrid>
        <div>Item</div>
      </StatGrid>
    )
    const gridEl = container.firstChild as HTMLElement
    expect(gridEl.className).toContain('grid')
    expect(gridEl.className).toContain('grid-cols-2')
    expect(gridEl.className).toContain('lg:grid-cols-5')
  })

  it('applies gap classes', () => {
    const { container } = render(
      <StatGrid>
        <div>Item</div>
      </StatGrid>
    )
    const gridEl = container.firstChild as HTMLElement
    expect(gridEl.className).toContain('gap-3')
    expect(gridEl.className).toContain('sm:gap-4')
    expect(gridEl.className).toContain('lg:gap-5')
  })

  it('applies bottom margin', () => {
    const { container } = render(
      <StatGrid>
        <div>Item</div>
      </StatGrid>
    )
    const gridEl = container.firstChild as HTMLElement
    expect(gridEl.className).toContain('mb-6')
  })

  it('renders a single child', () => {
    render(
      <StatGrid>
        <span>Only child</span>
      </StatGrid>
    )
    expect(screen.getByText('Only child')).toBeInTheDocument()
  })
})
