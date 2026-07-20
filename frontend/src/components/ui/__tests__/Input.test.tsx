import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Input } from '@/components/ui/Input'

describe('Input', () => {
  it('renders an input element', () => {
    render(<Input />)
    expect(screen.getByRole('textbox')).toBeInTheDocument()
  })

  it('applies placeholder', () => {
    render(<Input placeholder="Enter email" />)
    expect(screen.getByPlaceholderText('Enter email')).toBeInTheDocument()
  })

  it('applies type prop', () => {
    render(<Input type="password" />)
    const input = document.querySelector('input')
    expect(input).toHaveAttribute('type', 'password')
  })

  it('applies custom className', () => {
    render(<Input className="extra" />)
    expect(screen.getByRole('textbox').className).toContain('extra')
  })

  it('spreads additional HTML attributes', () => {
    render(<Input data-testid="email-input" aria-label="Email" />)
    expect(screen.getByTestId('email-input')).toBeInTheDocument()
  })

  it('is focusable via keyboard', () => {
    render(<Input />)
    const input = screen.getByRole('textbox')
    input.focus()
    expect(input).toHaveFocus()
  })

  it('forwards ref', () => {
    const ref = { current: null }
    render(<Input ref={ref} />)
    expect(ref.current).toBeInstanceOf(HTMLInputElement)
  })
})
