import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { Spinner } from '@/components/ui/Spinner'

describe('Spinner', () => {
  it('renders an SVG element', () => {
    render(<Spinner />)
    const svg = document.querySelector('svg')
    expect(svg).toBeInTheDocument()
  })

  it('applies md size by default', () => {
    render(<Spinner />)
    const svg = document.querySelector('svg')
    expect(svg?.classList.contains('w-5')).toBe(true)
    expect(svg?.classList.contains('h-5')).toBe(true)
  })

  it('applies sm size', () => {
    render(<Spinner size="sm" />)
    const svg = document.querySelector('svg')
    expect(svg?.classList.contains('w-4')).toBe(true)
    expect(svg?.classList.contains('h-4')).toBe(true)
  })

  it('applies lg size', () => {
    render(<Spinner size="lg" />)
    const svg = document.querySelector('svg')
    expect(svg?.classList.contains('w-8')).toBe(true)
    expect(svg?.classList.contains('h-8')).toBe(true)
  })

  it('applies animate-spin', () => {
    render(<Spinner />)
    const svg = document.querySelector('svg')
    expect(svg?.classList.contains('animate-spin')).toBe(true)
  })

  it('accepts custom className', () => {
    render(<Spinner className="text-red-500" />)
    const svg = document.querySelector('svg')
    expect(svg?.classList.contains('text-red-500')).toBe(true)
  })
})
