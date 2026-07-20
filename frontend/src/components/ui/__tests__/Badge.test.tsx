import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Badge } from '@/components/ui/Badge'

describe('Badge', () => {
  it('renders children text', () => {
    render(<Badge>Active</Badge>)
    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('applies neutral variant by default', () => {
    render(<Badge>Neutral</Badge>)
    const el = screen.getByText('Neutral')
    expect(el.className).toContain('bg-slate-100')
  })

  it('applies sky variant', () => {
    render(<Badge variant="sky">Sky</Badge>)
    expect(screen.getByText('Sky').className).toContain('bg-sky-50')
  })

  it('applies success variant', () => {
    render(<Badge variant="success">OK</Badge>)
    expect(screen.getByText('OK').className).toContain('bg-emerald-50')
  })

  it('applies warning variant', () => {
    render(<Badge variant="warning">Warn</Badge>)
    expect(screen.getByText('Warn').className).toContain('bg-amber-50')
  })

  it('applies error variant', () => {
    render(<Badge variant="error">Err</Badge>)
    expect(screen.getByText('Err').className).toContain('bg-red-50')
  })

  it('accepts custom className', () => {
    render(<Badge className="my-class">Custom</Badge>)
    expect(screen.getByText('Custom').className).toContain('my-class')
  })

  it('spreads additional HTML attributes', () => {
    render(<Badge data-testid="badge">Test</Badge>)
    expect(screen.getByTestId('badge')).toBeInTheDocument()
  })
})
