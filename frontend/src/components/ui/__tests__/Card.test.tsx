import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card'

describe('Card', () => {
  it('renders children', () => {
    render(<Card><p>Content</p></Card>)
    expect(screen.getByText('Content')).toBeInTheDocument()
  })

  it('applies base classes', () => {
    render(<Card data-testid="card">X</Card>)
    const el = screen.getByTestId('card')
    expect(el.tagName).toBe('DIV')
    expect(el.className).toContain('rounded-lg')
    expect(el.className).toContain('border')
  })

  it('accepts custom className', () => {
    render(<Card className="custom">X</Card>)
    expect(screen.getByText('X').className).toContain('custom')
  })
})

describe('CardHeader', () => {
  it('renders children', () => {
    render(<CardHeader><span>Title Here</span></CardHeader>)
    expect(screen.getByText('Title Here')).toBeInTheDocument()
  })

  it('applies border-b', () => {
    render(<CardHeader data-testid="h">X</CardHeader>)
    expect(screen.getByTestId('h').className).toContain('border-b')
  })
})

describe('CardTitle', () => {
  it('renders as h3', () => {
    render(<CardTitle>My Title</CardTitle>)
    const el = screen.getByRole('heading', { level: 3 })
    expect(el).toHaveTextContent('My Title')
  })

  it('applies font-semibold', () => {
    render(<CardTitle data-testid="t">X</CardTitle>)
    expect(screen.getByTestId('t').className).toContain('font-semibold')
  })
})

describe('CardBody', () => {
  it('renders children', () => {
    render(<CardBody><p>Body content</p></CardBody>)
    expect(screen.getByText('Body content')).toBeInTheDocument()
  })

  it('applies padding', () => {
    render(<CardBody data-testid="b">X</CardBody>)
    expect(screen.getByTestId('b').className).toContain('p-4')
  })
})

describe('Card composition', () => {
  it('renders nested Card > CardHeader + CardBody', () => {
    render(
      <Card>
        <CardHeader><CardTitle>Header</CardTitle></CardHeader>
        <CardBody><p>Body text</p></CardBody>
      </Card>
    )
    expect(screen.getByRole('heading', { level: 3 })).toHaveTextContent('Header')
    expect(screen.getByText('Body text')).toBeInTheDocument()
  })
})
