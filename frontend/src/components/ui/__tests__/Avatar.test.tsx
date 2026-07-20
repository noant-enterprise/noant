import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Avatar } from '@/components/ui/Avatar'

describe('Avatar', () => {
  it('shows initials when no src', () => {
    render(<Avatar name="John Doe" />)
    expect(screen.getByText('JD')).toBeInTheDocument()
  })

  it('shows single initial for single-word name', () => {
    render(<Avatar name="Alice" />)
    expect(screen.getByText('A')).toBeInTheDocument()
  })

  it('limits initials to 2 characters', () => {
    render(<Avatar name="Robert Downey Junior" />)
    expect(screen.getByText('RD')).toBeInTheDocument()
  })

  it('shows image when src is provided', () => {
    render(<Avatar name="User" src="https://example.com/avatar.jpg" />)
    const img = screen.getByRole('img')
    expect(img).toHaveAttribute('src', 'https://example.com/avatar.jpg')
    expect(img).toHaveAttribute('alt', 'User')
  })

  it('applies sm size classes to initials', () => {
    render(<Avatar name="S M" size="sm" />)
    const initials = screen.getByText('SM')
    expect(initials.className).toContain('w-7')
    expect(initials.className).toContain('h-7')
  })

  it('applies md size classes by default', () => {
    render(<Avatar name="M D" />)
    const initials = screen.getByText('MD')
    expect(initials.className).toContain('w-8')
    expect(initials.className).toContain('h-8')
  })

  it('applies lg size classes', () => {
    render(<Avatar name="L G" size="lg" />)
    const initials = screen.getByText('LG')
    expect(initials.className).toContain('w-10')
    expect(initials.className).toContain('h-10')
  })

  it('applies xl size classes', () => {
    render(<Avatar name="X L" size="xl" />)
    const initials = screen.getByText('XL')
    expect(initials.className).toContain('w-12')
    expect(initials.className).toContain('h-12')
  })

  it('renders gradient background for initials', () => {
    render(<Avatar name="Gradient User" />)
    const initials = screen.getByText('GU')
    expect(initials.className).toContain('bg-gradient-to-br')
  })

  it('applies className to wrapper', () => {
    const { container } = render(<Avatar name="C" className="extra" />)
    const wrapper = container.firstElementChild!
    expect(wrapper.className).toContain('extra')
  })

  it('shows channel badge when showChannel is true', () => {
    render(
      <Avatar
        name="User"
        showChannel
        channelColor="#25D366"
        channelIcon={<span>W</span>}
      />
    )
    expect(screen.getByText('W')).toBeInTheDocument()
  })

  it('hides channel badge when showChannel is false', () => {
    render(
      <Avatar
        name="User"
        channelColor="#25D366"
        channelIcon={<span>W</span>}
      />
    )
    expect(screen.queryByText('W')).not.toBeInTheDocument()
  })

  it('uses consistent gradient for same name', () => {
    const { container: c1 } = render(<Avatar name="Same Name" />)
    const { container: c2 } = render(<Avatar name="Same Name" />)
    const g1 = c1.querySelector('[class*="bg-gradient"]')
    const g2 = c2.querySelector('[class*="bg-gradient"]')
    expect(g1?.className).toBe(g2?.className)
  })
})
