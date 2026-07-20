import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { Skeleton, StatSkeleton } from '@/components/ui/Skeleton'

describe('Skeleton', () => {
  it('renders a div with shimmer class', () => {
    const { container } = render(<Skeleton />)
    const el = container.firstChild as HTMLElement
    expect(el.tagName).toBe('DIV')
    expect(el.className).toContain('animate-shimmer-slow')
  })

  it('applies custom className', () => {
    const { container } = render(<Skeleton className="h-4 w-32" />)
    const el = container.firstChild as HTMLElement
    expect(el.className).toContain('h-4')
    expect(el.className).toContain('w-32')
  })
})

describe('StatSkeleton', () => {
  it('renders with stat-like structure', () => {
    const { container } = render(<StatSkeleton />)
    const wrapper = container.firstChild as HTMLElement
    expect(wrapper.tagName).toBe('DIV')
    expect(wrapper.className).toContain('rounded-xl')

    const shimmerDivs = wrapper.querySelectorAll('.animate-shimmer-slow')
    expect(shimmerDivs.length).toBe(2)
  })
})
