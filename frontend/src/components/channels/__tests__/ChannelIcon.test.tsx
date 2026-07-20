import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { ChannelIcon } from '@/components/channels/ChannelIcon'

describe('ChannelIcon', () => {
  it('renders WhatsApp icon', () => {
    const { container } = render(<ChannelIcon channel="whatsapp" />)
    const div = container.querySelector('div')
    expect(div?.className).toContain('bg-[#25D366]')
    expect(div?.querySelector('svg')).toBeInTheDocument()
  })

  it('renders Instagram icon', () => {
    const { container } = render(<ChannelIcon channel="instagram" />)
    const div = container.querySelector('div')
    expect(div?.className).toContain('from-[#833AB4]')
  })

  it('renders Facebook icon', () => {
    const { container } = render(<ChannelIcon channel="facebook" />)
    const div = container.querySelector('div')
    expect(div?.className).toContain('bg-[#1877F2]')
  })

  it('renders Telegram icon', () => {
    const { container } = render(<ChannelIcon channel="telegram" />)
    const div = container.querySelector('div')
    expect(div?.className).toContain('bg-[#0088CC]')
  })

  it('renders Gmail icon', () => {
    const { container } = render(<ChannelIcon channel="gmail" />)
    const div = container.querySelector('div')
    expect(div?.className).toContain('bg-[#EA4335]')
  })

  it('renders Web widget icon', () => {
    const { container } = render(<ChannelIcon channel="web" />)
    const div = container.querySelector('div')
    expect(div?.className).toContain('bg-noant-black')
  })

  it('renders default fallback for unknown channel', () => {
    const { container } = render(<ChannelIcon channel="unknown" />)
    const div = container.querySelector('div')
    expect(div?.className).toContain('bg-tertiary/20')
    expect(div?.textContent).toBe('u')
  })

  it('applies sm size', () => {
    const { container } = render(<ChannelIcon channel="whatsapp" size="sm" />)
    expect(container.querySelector('div')?.className).toContain('w-8')
  })

  it('applies md size by default', () => {
    const { container } = render(<ChannelIcon channel="whatsapp" />)
    expect(container.querySelector('div')?.className).toContain('w-10')
  })

  it('applies lg size', () => {
    const { container } = render(<ChannelIcon channel="whatsapp" size="lg" />)
    expect(container.querySelector('div')?.className).toContain('w-12')
  })

  it('accepts custom className', () => {
    const { container } = render(<ChannelIcon channel="whatsapp" className="extra" />)
    expect(container.querySelector('div')?.className).toContain('extra')
  })

  it('handles case-insensitive channel names', () => {
    const { container } = render(<ChannelIcon channel="WhatsApp" />)
    expect(container.querySelector('div')?.className).toContain('bg-[#25D366]')
  })
})
