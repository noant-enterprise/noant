import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { ConversationLoading } from '@/components/chat/ConversationLoading'

describe('ConversationLoading', () => {
  it('renders the loading spinner structure', () => {
    const { container } = render(<ConversationLoading />)
    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('renders with sm size', () => {
    const { container } = render(<ConversationLoading size="sm" />)
    const outerDiv = container.querySelector('.relative')
    expect(outerDiv).toHaveAttribute('style', expect.stringContaining('width: 20px'))
    expect(outerDiv).toHaveAttribute('style', expect.stringContaining('height: 20px'))
  })

  it('renders with md size (default)', () => {
    const { container } = render(<ConversationLoading />)
    const outerDiv = container.querySelector('.relative')
    expect(outerDiv).toHaveAttribute('style', expect.stringContaining('width: 60px'))
    expect(outerDiv).toHaveAttribute('style', expect.stringContaining('height: 60px'))
  })

  it('renders with lg size', () => {
    const { container } = render(<ConversationLoading size="lg" />)
    const outerDiv = container.querySelector('.relative')
    expect(outerDiv).toHaveAttribute('style', expect.stringContaining('width: 120px'))
    expect(outerDiv).toHaveAttribute('style', expect.stringContaining('height: 120px'))
  })

  it('applies custom className', () => {
    const { container } = render(<ConversationLoading className="my-custom-class" />)
    const outerDiv = container.querySelector('.my-custom-class')
    expect(outerDiv).toBeInTheDocument()
  })

  it('renders three pulsing dots', () => {
    const { container } = render(<ConversationLoading />)
    const dots = container.querySelectorAll('.rounded-full[style*="logoDotPulse"]')
    expect(dots).toHaveLength(3)
  })

  it('renders spinning dashed ring SVG', () => {
    const { container } = render(<ConversationLoading />)
    const svg = container.querySelector('svg')
    expect(svg).toBeInTheDocument()
    expect(svg?.querySelector('circle')).toBeInTheDocument()
  })

  it('renders style tag with keyframe animations', () => {
    const { container } = render(<ConversationLoading />)
    const style = container.querySelector('style')
    expect(style?.textContent).toContain('rotateLogoRing')
    expect(style?.textContent).toContain('logoDotPulse')
  })
})
