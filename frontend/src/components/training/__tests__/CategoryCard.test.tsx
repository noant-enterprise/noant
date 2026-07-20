import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CategoryCard } from '@/components/training/CategoryCard'

const mockCategory = {
  id: 'cat-1',
  name: 'Billing',
  color: '#ff6600',
  qa_count: 12,
  created_at: '2026-01-15T10:00:00Z',
}

describe('CategoryCard', () => {
  it('renders the category name', () => {
    render(<CategoryCard category={mockCategory} />)
    expect(screen.getByText('Billing')).toBeInTheDocument()
  })

  it('renders the QA count', () => {
    render(<CategoryCard category={mockCategory} />)
    expect(screen.getByText('12 Q&A pairs')).toBeInTheDocument()
  })

  it('renders the color dot', () => {
    render(<CategoryCard category={mockCategory} />)
    const dot = screen.getByText('Billing').closest('div')?.querySelector('.rounded-full')
    expect(dot).toHaveStyle({ background: '#ff6600' })
  })

  it('calls onClick when clicked', async () => {
    const user = userEvent.setup()
    const handleClick = vi.fn()
    render(<CategoryCard category={mockCategory} onClick={handleClick} />)
    await user.click(screen.getByText('Billing'))
    expect(handleClick).toHaveBeenCalledOnce()
  })

  it('renders without crashing when onClick is not provided', () => {
    render(<CategoryCard category={mockCategory} />)
    expect(screen.getByText('Billing')).toBeInTheDocument()
  })

  it('renders the formatted date', () => {
    render(<CategoryCard category={mockCategory} />)
    const dateStr = new Date('2026-01-15T10:00:00Z').toLocaleDateString()
    expect(screen.getByText(dateStr)).toBeInTheDocument()
  })
})
