import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { EmptyState } from '@/components/ui/EmptyState'

describe('EmptyState', () => {
  it('renders the title', () => {
    render(<EmptyState title="No results" />)
    expect(screen.getByText('No results')).toBeInTheDocument()
  })

  it('renders optional description', () => {
    render(<EmptyState title="Empty" description="Try adding something" />)
    expect(screen.getByText('Try adding something')).toBeInTheDocument()
  })

  it('hides description when not provided', () => {
    render(<EmptyState title="Empty" />)
    expect(screen.queryByText(/try adding/i)).not.toBeInTheDocument()
  })

  it('renders optional action', () => {
    render(
      <EmptyState
        title="Nothing here"
        action={<button>Add item</button>}
      />
    )
    expect(screen.getByRole('button', { name: /add item/i })).toBeInTheDocument()
  })

  it('hides action when not provided', () => {
    render(<EmptyState title="Empty" />)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('accepts custom className', () => {
    render(<EmptyState title="X" className="my-custom" />)
    expect(screen.getByText('X').closest('[class*="my-custom"]')).toBeInTheDocument()
  })

  it('accepts custom icon', () => {
    render(
      <EmptyState
        title="Custom"
        icon={<span data-testid="custom-icon">★</span>}
      />
    )
    expect(screen.getByTestId('custom-icon')).toBeInTheDocument()
  })

  it('shows default Inbox icon when no icon prop', () => {
    const { container } = render(<EmptyState title="Default icon" />)
    expect(container.querySelector('svg')).toBeInTheDocument()
  })
})
